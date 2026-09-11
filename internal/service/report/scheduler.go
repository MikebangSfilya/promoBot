package report

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/robfig/cron/v3"
)

// Scheduler runs a [Reporter] on a cron schedule.
type Scheduler struct {
	reporter *Reporter
	cron     *cron.Cron
	spec     string
	// ctx is set by Run and read by the job it starts; the cron library gives
	// jobs no argument of their own.
	ctx context.Context
}

// Schedule prepares the report to be sent according to the given cron
// expression, interpreted in UTC. Parsing happens here rather than inside the
// running loop, so the caller can fail the startup on a malformed expression
// instead of leaving the bot silently never sending a report.
func (r *Reporter) Schedule(spec string) (*Scheduler, error) {
	const op = "Reporter.Schedule"

	schedule, err := cron.ParseStandard(spec)
	if err != nil {
		return nil, fmt.Errorf("%s, invalid cron expression %q: %w", op, spec, err)
	}

	// Pinned to UTC so the schedule does not shift with the host's local zone.
	s := &Scheduler{reporter: r, cron: cron.New(cron.WithLocation(time.UTC)), spec: spec}
	s.cron.Schedule(schedule, cron.FuncJob(s.send))

	return s, nil
}

// Run starts the schedule and blocks until the context is canceled, so the
// caller can run it as a goroutine registered on the application's WaitGroup.
func (s *Scheduler) Run(ctx context.Context) {
	const op = "Scheduler.Run"
	log := slog.With("op", op, "spec", s.spec, "chat_id", s.reporter.chatID)

	s.ctx = ctx
	s.cron.Start()

	if entries := s.cron.Entries(); len(entries) > 0 {
		log.Info("weekly activation report scheduled", "next_run", entries[0].Next)
	}

	<-ctx.Done()

	// The context returned by Stop() is done once the running job, if any, has
	// finished, so an in-flight report is not cut short by the shutdown that
	// follows.
	<-s.cron.Stop().Done()
	log.Info("weekly activation report scheduler stopped")
}

func (s *Scheduler) send() {
	const op = "Scheduler.send"

	if err := s.reporter.Send(s.ctx); err != nil {
		slog.With("op", op).Error("scheduled report failed",
			slog.Group("error",
				slog.String("message", err.Error()),
				slog.String("component", "Reporter.Send")))
	}
}
