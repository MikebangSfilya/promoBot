package report

import (
	"context"
	"testing"
	"time"

	"github.com/MikebangSfilya/promoBot/internal/audit"
	cfg "github.com/MikebangSfilya/promoBot/internal/config"
	"github.com/MikebangSfilya/promoBot/internal/model"

	tgbotapi "github.com/OvyFlash/telegram-bot-api"
	"github.com/kozalosev/goSadTgBot/base"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReporter_ScheduleInvalidSpec(t *testing.T) {
	for _, spec := range []string{"", "not a cron expression", "99 * * * *"} {
		t.Run(spec, func(t *testing.T) {
			scheduler, err := newTestReporter(&reporterStub{}, &base.FakeBotAPI{}).Schedule(spec)

			require.Error(t, err)
			assert.Nil(t, scheduler)
		})
	}
}

func TestReporter_ScheduleValidSpec(t *testing.T) {
	scheduler, err := newTestReporter(&reporterStub{}, &base.FakeBotAPI{}).Schedule("0 10 * * 1")

	require.NoError(t, err)
	require.NotNil(t, scheduler)
	assert.Len(t, scheduler.cron.Entries(), 1)
}

// The default schedule is a valid expression that fires when it claims to, and
// is interpreted in UTC regardless of the host's local zone. A default that
// parses but fires at the wrong time would not fail the startup, so it is
// pinned here instead.
func TestReporter_ScheduleDefaultIsSaturdayInUTC(t *testing.T) {
	scheduler, err := newTestReporter(&reporterStub{}, &base.FakeBotAPI{}).
		Schedule(cfg.DefaultReportCron)
	require.NoError(t, err)

	// Monday 2026-06-15; the next run is the Saturday that follows.
	next := scheduler.cron.Entries()[0].Schedule.
		Next(time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC)).UTC()

	assert.Equal(t, time.Date(2026, 6, 20, 10, 0, 0, 0, time.UTC), next)
	assert.Equal(t, time.Saturday, next.Weekday())
}

func TestScheduler_RunStopsOnContextCancellation(t *testing.T) {
	scheduler, err := newTestReporter(&reporterStub{}, &base.FakeBotAPI{}).Schedule("0 10 * * 1")
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		defer close(done)
		scheduler.Run(ctx)
	}()

	cancel()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Run did not return after the context was cancelled")
	}
}

func TestScheduler_RunSendsOnSchedule(t *testing.T) {
	bot := &base.FakeBotAPI{}
	repo := &reporterStub{stats: []model.PromoActivationStat{stat("AAA", 1, 1)}}
	created := &creationsStub{logs: []audit.Log{creation("AAA", "boss")}}

	// The job is invoked directly instead of waiting for a real minute to pass.
	scheduler, err := newTestReporterWith(repo, created, bot).Schedule("* * * * *")
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	scheduler.ctx = ctx
	scheduler.send()

	sent, ok := bot.GetOutput().([]tgbotapi.Chattable)
	require.True(t, ok)
	require.Len(t, sent, 1)
	assert.Equal(t, 1, repo.calls)
}
