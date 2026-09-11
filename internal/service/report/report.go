// Package report builds and delivers the periodic digest of promo code activity:
// which codes were created within the reported window, which of them were
// actually used, and which were never activated at all.
package report

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/MikebangSfilya/promoBot/internal/audit"
	"github.com/MikebangSfilya/promoBot/internal/formatter"
	"github.com/MikebangSfilya/promoBot/internal/model"

	tgbotapi "github.com/OvyFlash/telegram-bot-api"
	"github.com/kozalosev/goSadTgBot/base"
	"github.com/loctools/go-l10n/loc"
	"github.com/samber/lo"
)

// Translation keys of the report.
const (
	reportTitle               = "reportTitle"
	reportSectionCreated      = "reportSectionCreated"
	reportSectionActivated    = "reportSectionActivated"
	reportSectionNotActivated = "reportSectionNotActivated"
	reportSectionCount        = "reportSectionCount"
	reportEntryCreated        = "reportEntryCreatedFormat"
	reportEntryActivated      = "reportEntryActivatedFormat"
	reportEntryNotActivated   = "reportEntryNotActivatedFormat"
	reportEmpty               = "reportEmpty"
	reportUnknownAuthor       = "reportUnknownAuthor"
	dateEndless               = "dateEndless"
)

// sendTimeout bounds a single delivery attempt, including the database query.
const sendTimeout = 30 * time.Second

// PromoReporter is the slice of the repository the report needs.
type PromoReporter interface {
	GetActivationStats(ctx context.Context, since time.Time, codes []string) ([]model.PromoActivationStat, error)
}

// LogFinder tells which promo codes were created within the window, and by
// whom. The database keeps no such record, so it comes from the audit log.
type LogFinder interface {
	FindLogs(action string, since time.Time) ([]audit.Log, error)
}

// Reporter renders the digest and pushes it into the admins chat.
type Reporter struct {
	repo   PromoReporter
	logs   LogFinder
	bot    base.ExtendedBotAPI
	chatID int64
	lang   *loc.Context
	period time.Duration
}

func New(
	repo PromoReporter,
	logs LogFinder,
	bot base.ExtendedBotAPI,
	chatID int64,
	lang *loc.Context,
	period time.Duration,
) *Reporter {
	return &Reporter{
		repo:   repo,
		logs:   logs,
		bot:    bot,
		chatID: chatID,
		lang:   lang,
		period: period,
	}
}

// Build renders the report covering the period ending at now.
//
// Creation facts come from the audit log and activation counters from the
// database; the two are joined on the promo code.
func (r *Reporter) Build(ctx context.Context, now time.Time) (string, error) {
	since := now.Add(-r.period)

	creations, err := r.logs.FindLogs(model.ActionCreate, since)
	if err != nil {
		return "", fmt.Errorf("failed to read the audit log for the report: %w", err)
	}

	entries, err := r.collect(ctx, since, creations)
	if err != nil {
		return "", err
	}

	// An empty window is still worth reporting: a missing weekly message is
	// indistinguishable from a broken scheduler.
	if len(entries) == 0 {
		return fmt.Sprintf(r.lang.Tr(reportEmpty), formatDate(&since)), nil
	}

	activated, notActivated := lo.FilterReject(entries, func(entry model.ReportEntry, _ int) bool {
		return entry.ActivationsInWindow > 0
	})

	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintf(r.lang.Tr(reportTitle), formatDate(&since)))
	sb.WriteString("\n\n")
	sb.WriteString(r.section(reportSectionCreated, entries, r.formatCreated))
	sb.WriteString("\n\n")
	sb.WriteString(r.section(reportSectionActivated, activated, r.formatActivated))
	sb.WriteString("\n\n")
	sb.WriteString(r.section(reportSectionNotActivated, notActivated, r.formatNotActivated))

	return sb.String(), nil
}

// collect joins the audited creations with the activation counters, keeping the
// creation order of the audit log.
//
// The audit log remembers a creation forever, but the promo code itself may
// have been deleted since. Such a code has no row left to report on, so it is
// dropped. If the same code was created more than once, the latest creation
// wins.
func (r *Reporter) collect(ctx context.Context, since time.Time, creations []audit.Log) ([]model.ReportEntry, error) {
	if len(creations) == 0 {
		return nil, nil
	}

	// Uniq keeps the first occurrence, so the report follows the order the
	// codes were created in; KeyBy keeps the last, so a code created twice is
	// attributed to its most recent creation.
	codes := lo.Uniq(lo.Map(creations, func(c audit.Log, _ int) string { return c.Code }))
	creationByCode := lo.KeyBy(creations, func(c audit.Log) string { return c.Code })

	stats, err := r.repo.GetActivationStats(ctx, since, codes)
	if err != nil {
		return nil, fmt.Errorf("failed to collect activation stats for the report: %w", err)
	}

	statByCode := lo.KeyBy(stats, func(stat model.PromoActivationStat) string { return stat.Code })

	return lo.FilterMap(codes, func(code string, _ int) (model.ReportEntry, bool) {
		stat, stillExists := statByCode[code]
		if !stillExists {
			return model.ReportEntry{}, false
		}

		creation := creationByCode[code]

		return model.ReportEntry{
			PromoActivationStat: stat,
			CreatedBy:           creation.By,
			CreatedAt:           creation.At,
		}, true
	}), nil
}

// Send builds the report and delivers it to the configured chat.
func (r *Reporter) Send(ctx context.Context) error {
	const op = "Reporter.Send"
	log := slog.With("op", op, "chat_id", r.chatID)

	ctx, cancel := context.WithTimeout(ctx, sendTimeout)
	defer cancel()

	text, err := r.Build(ctx, time.Now())
	if err != nil {
		log.Error("failed to build the activation report",
			slog.Group("error",
				slog.String("message", err.Error()),
				slog.String("component", "Reporter.Build")))
		return err
	}

	// Reply* methods all require an incoming message, so an unsolicited push
	// has to go through Send with an explicitly built message.
	if _, err := r.bot.Send(tgbotapi.NewMessage(r.chatID, text)); err != nil {
		log.Error("failed to send the activation report",
			slog.Group("error",
				slog.String("message", err.Error()),
				slog.String("component", "Bot.Send")))
		return err
	}

	log.Info("activation report sent")
	return nil
}

func (r *Reporter) section(titleKey string, entries []model.ReportEntry, format func(model.ReportEntry) string) string {
	return formatter.FormatList(r.lang.Tr(titleKey), r.lang.Tr(reportSectionCount), entries, format)
}

func (r *Reporter) formatCreated(entry model.ReportEntry) string {
	author := entry.CreatedBy
	if author == "" {
		author = r.lang.Tr(reportUnknownAuthor)
	}
	return fmt.Sprintf(r.lang.Tr(reportEntryCreated),
		entry.Code,
		author,
		entry.BonusLength,
		entry.InitialCapacity(),
		formatDate(&entry.CreatedAt),
		formatDate(entry.Since),
		r.formatUntil(entry.Until),
	)
}

func (r *Reporter) formatActivated(entry model.ReportEntry) string {
	return fmt.Sprintf(r.lang.Tr(reportEntryActivated), entry.Code, entry.ActivationsInWindow)
}

func (r *Reporter) formatNotActivated(entry model.ReportEntry) string {
	return fmt.Sprintf(r.lang.Tr(reportEntryNotActivated), entry.Code)
}

// formatUntil renders an open-ended promo code as "endless" rather than a blank.
func (r *Reporter) formatUntil(until *time.Time) string {
	if until == nil {
		return r.lang.Tr(dateEndless)
	}
	return formatDate(until)
}

func formatDate(t *time.Time) string {
	if t == nil {
		return "?"
	}
	return t.Format(time.DateOnly)
}
