// Package report builds the periodic digests of promo code activity: which
// codes were created within the reported window, which of them were actually
// used, which were never activated at all, and who changed them.
package report

import (
	"context"
	"fmt"
	"time"

	"github.com/MikebangSfilya/promoBot/internal/audit"
	"github.com/MikebangSfilya/promoBot/internal/model"

	"github.com/iafan/Plurr/go/plurr"
	"github.com/loctools/go-l10n/loc"
	"github.com/samber/lo"
)

// Translation keys of the report.
const (
	reportTitle               = "reportTitle"
	reportSectionCreated      = "reportSectionCreated"
	reportSectionActivated    = "reportSectionActivated"
	reportSectionNotActivated = "reportSectionNotActivated"
	reportEntryCreated        = "reportEntryCreatedFormat"
	reportEntryActivated      = "reportEntryActivatedFormat"
	reportEntryNotActivated   = "reportEntryNotActivatedFormat"
	reportEmpty               = "reportEmpty"
	reportUnknownAuthor       = "reportUnknownAuthor"
	reportPageSuffix          = "reportPageSuffix"
	reportDateLayout          = "reportDateLayout"
	reportTruncated           = "reportTruncated"
	dateEndless               = "dateEndless"
)

// PromoReporter is the slice of the repository the report needs.
type PromoReporter interface {
	GetActivationStats(ctx context.Context, since time.Time, codes []string) ([]model.PromoActivationStat, error)
}

// LogFinder tells what happened to the promo codes within the window, and who
// did it. The database keeps no such record, so it comes from the audit log.
type LogFinder interface {
	FindLogs(since time.Time, actions ...string) ([]audit.Log, error)
}

// ActivationReport renders the digest of newly created codes and their use.
type ActivationReport struct {
	repo     PromoReporter
	logs     LogFinder
	lang     *loc.Context
	period   time.Duration
	maxPages int
}

func NewActivationReport(
	repo PromoReporter,
	logs LogFinder,
	lang *loc.Context,
	period time.Duration,
	maxPages int,
) *ActivationReport {
	return &ActivationReport{
		repo:     repo,
		logs:     logs,
		lang:     lang,
		period:   period,
		maxPages: maxPages,
	}
}

func (*ActivationReport) Name() string { return "activation" }

// Build renders the report covering the period ending at now.
//
// Creation facts come from the audit log and activation counters from the
// database; the two are joined on the promo code.
func (r *ActivationReport) Build(ctx context.Context, now time.Time) ([]string, error) {
	since := now.Add(-r.period)

	creations, err := r.logs.FindLogs(since, model.ActionCreate)
	if err != nil {
		return nil, fmt.Errorf("failed to read the audit log for the report: %w", err)
	}

	entries, err := r.collect(ctx, since, creations)
	if err != nil {
		return nil, err
	}

	// An empty window is still worth reporting: a missing weekly message is
	// indistinguishable from a broken scheduler.
	if len(entries) == 0 {
		return []string{fmt.Sprintf(r.lang.Tr(reportEmpty), formatDate(r.lang, &since))}, nil
	}

	activated, notActivated := lo.FilterReject(entries, func(entry model.ReportEntry, _ int) bool {
		return entry.ActivationsInWindow > 0
	})

	blocks := lo.Filter([]block{
		r.section(reportSectionCreated, entries, r.formatCreated),
		r.section(reportSectionActivated, activated, r.formatActivated),
		r.section(reportSectionNotActivated, notActivated, r.formatNotActivated),
	}, func(b block, _ int) bool { return len(b.lines) > 0 })

	title := fmt.Sprintf(r.lang.Tr(reportTitle), formatDate(r.lang, &since))

	return paginate(r.pageOptions(title), blocks), nil
}

// collect joins the audited creations with the activation counters, keeping the
// creation order of the audit log.
//
// The audit log remembers a creation forever, but the promo code itself may
// have been deleted since. Such a code has no row left to report on, so it is
// dropped. If the same code was created more than once, the latest creation
// wins.
func (r *ActivationReport) collect(ctx context.Context, since time.Time, creations []audit.Log) ([]model.ReportEntry, error) {
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

// truncatedNotice says how many pages were left out, in the report's language.
func (r *ActivationReport) truncatedNotice(dropped int) string {
	return r.lang.Format(reportTruncated, plurr.Params{"n": dropped})
}

func (r *ActivationReport) pageOptions(title string) pageOptions {
	return pageOptions{
		title:     title,
		suffix:    r.lang.Tr(reportPageSuffix),
		truncated: r.truncatedNotice,
		limit:     maxMessageLen,
		maxPages:  r.maxPages,
	}
}

// section renders one section as a block, so that a long one can be continued
// on the next page under its own title.
//
// A section with nothing in it comes back empty and is dropped by the caller:
// an empty heading tells the reader nothing.
func (r *ActivationReport) section(titleKey string, entries []model.ReportEntry, format func(model.ReportEntry) string) block {
	if len(entries) == 0 {
		return block{}
	}

	return block{
		header: r.lang.Tr(titleKey),
		lines: lo.Map(entries, func(entry model.ReportEntry, i int) string {
			return fmt.Sprintf("%d. %s", i+1, format(entry))
		}),
	}
}

func (r *ActivationReport) formatCreated(entry model.ReportEntry) string {
	author := entry.CreatedBy
	if author == "" {
		author = r.lang.Tr(reportUnknownAuthor)
	}

	return r.lang.Format(reportEntryCreated, plurr.Params{
		"code":     escapeHTML(entry.Code),
		"author":   escapeHTML(author),
		"length":   entry.BonusLength,
		"capacity": entry.InitialCapacity(),
		"created":  formatDate(r.lang, &entry.CreatedAt),
		"since":    formatDate(r.lang, entry.Since),
		"until":    r.formatUntil(entry.Until),
	})
}

func (r *ActivationReport) formatActivated(entry model.ReportEntry) string {
	return r.lang.Format(reportEntryActivated, plurr.Params{
		"code": escapeHTML(entry.Code),
		"n":    entry.ActivationsInWindow,
	})
}

func (r *ActivationReport) formatNotActivated(entry model.ReportEntry) string {
	return r.lang.Format(reportEntryNotActivated, plurr.Params{
		"code": escapeHTML(entry.Code),
	})
}

// formatUntil renders an open-ended promo code as "endless" rather than a blank.
func (r *ActivationReport) formatUntil(until *time.Time) string {
	if until == nil {
		return r.lang.Tr(dateEndless)
	}
	return formatDate(r.lang, until)
}

// formatDate renders a date the way the report's language writes them.
func formatDate(lang *loc.Context, t *time.Time) string {
	if t == nil {
		return "?"
	}
	return t.Format(lang.Tr(reportDateLayout))
}
