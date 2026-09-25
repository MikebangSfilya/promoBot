package report

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	tgbotapi "github.com/OvyFlash/telegram-bot-api"
	"github.com/kozalosev/goSadTgBot/base"
	"github.com/robfig/cron/v3"
)

// sendTimeout bounds a single report, including the queries it makes.
const sendTimeout = 30 * time.Second

// Report is everything the scheduler needs to know about a report.
type Report interface {
	// Name identifies the report in the logs.
	Name() string
	// Build renders the report as pages, one message each.
	Build(ctx context.Context, now time.Time) ([]string, error)
}

// Scheduler delivers a set of reports on a cron schedule.
type Scheduler struct {
	bot     base.ExtendedBotAPI
	chatID  int64
	reports []Report
	cron    *cron.Cron
	spec    string
	// ctx is set by Run and read by the job it starts; the cron library gives
	// jobs no argument of their own.
	ctx context.Context
}

// Schedule prepares the reports to be sent according to the given cron
// expression, interpreted in UTC. Parsing happens here rather than inside the
// running loop, so the caller can fail the startup on a malformed expression
// instead of leaving the bot silently never sending a report.
func Schedule(spec string, bot base.ExtendedBotAPI, chatID int64, reports ...Report) (*Scheduler, error) {
	const op = "report.Schedule"

	schedule, err := cron.ParseStandard(spec)
	if err != nil {
		return nil, fmt.Errorf("%s, invalid cron expression %q: %w", op, spec, err)
	}

	// Pinned to UTC so the schedule does not shift with the host's local zone.
	s := &Scheduler{
		bot:     bot,
		chatID:  chatID,
		reports: reports,
		cron:    cron.New(cron.WithLocation(time.UTC)),
		spec:    spec,
	}
	s.cron.Schedule(schedule, cron.FuncJob(s.send))

	return s, nil
}

// Run starts the schedule and blocks until the context is canceled, so the
// caller can run it as a goroutine registered on the application's WaitGroup.
func (s *Scheduler) Run(ctx context.Context) {
	const op = "Scheduler.Run"
	log := slog.With("op", op, "spec", s.spec, "chat_id", s.chatID)

	s.ctx = ctx
	s.cron.Start()

	if entries := s.cron.Entries(); len(entries) > 0 {
		log.Info("reports scheduled", "next_run", entries[0].Next)
	}

	<-ctx.Done()

	// The context returned by Stop() is done once the running job, if any, has
	// finished, so an in-flight report is not cut short by the shutdown that
	// follows.
	<-s.cron.Stop().Done()
	log.Info("report scheduler stopped")
}

// send delivers every report, in order, as its own message or messages.
//
// One moment is taken for all of them, so their windows end together. A report
// that fails is logged and skipped: the ones after it are still sent.
func (s *Scheduler) send() {
	const op = "Scheduler.send"
	log := slog.With("op", op, "chat_id", s.chatID)

	now := time.Now()

	for _, report := range s.reports {
		log := log.With(slog.String("report", report.Name()))

		if err := s.sendReport(s.ctx, now, report); err != nil {
			log.Error("scheduled report failed",
				slog.Group("error",
					slog.String("message", err.Error())))
			continue
		}

		log.Info("report sent")
	}
}

func (s *Scheduler) sendReport(ctx context.Context, now time.Time, report Report) error {
	ctx, cancel := context.WithTimeout(ctx, sendTimeout)
	defer cancel()

	pages, err := report.Build(ctx, now)
	if err != nil {
		return fmt.Errorf("failed to build the report: %w", err)
	}

	for i, page := range pages {
		// Reply* methods all require an incoming message, so an unsolicited push
		// has to go through Send with an explicitly built message.
		msg := tgbotapi.NewMessage(s.chatID, page)
		msg.ParseMode = tgbotapi.ModeHTML
		if _, err := s.bot.Send(msg); err != nil {
			return fmt.Errorf("failed to send page %d of %d: %w", i+1, len(pages), err)
		}
	}

	return nil
}
