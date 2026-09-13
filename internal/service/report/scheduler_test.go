package report

import (
	"context"
	"errors"
	"testing"
	"time"

	cfg "github.com/MikebangSfilya/promoBot/internal/config"

	tgbotapi "github.com/OvyFlash/telegram-bot-api"
	"github.com/kozalosev/goSadTgBot/base"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// reportStub stands in for a report, recording what it was asked for.
type reportStub struct {
	name  string
	pages []string
	err   error
	now   time.Time
	calls int
}

func (s *reportStub) Name() string { return s.name }

func (s *reportStub) Build(_ context.Context, now time.Time) ([]string, error) {
	s.calls++
	s.now = now
	return s.pages, s.err
}

func sentMessages(t *testing.T, bot *base.FakeBotAPI) []tgbotapi.MessageConfig {
	t.Helper()

	if bot.GetOutput() == nil {
		return nil
	}

	sent, ok := bot.GetOutput().([]tgbotapi.Chattable)
	require.True(t, ok)

	messages := make([]tgbotapi.MessageConfig, 0, len(sent))
	for _, c := range sent {
		msg, ok := c.(tgbotapi.MessageConfig)
		require.True(t, ok)
		messages = append(messages, msg)
	}

	return messages
}

func TestScheduleInvalidSpec(t *testing.T) {
	for _, spec := range []string{"", "not a cron expression", "99 * * * *"} {
		t.Run(spec, func(t *testing.T) {
			scheduler, err := Schedule(spec, &base.FakeBotAPI{}, testChatID)

			require.Error(t, err)
			assert.Nil(t, scheduler)
		})
	}
}

func TestScheduleValidSpec(t *testing.T) {
	scheduler, err := Schedule("0 10 * * 1", &base.FakeBotAPI{}, testChatID)

	require.NoError(t, err)
	require.NotNil(t, scheduler)
	assert.Len(t, scheduler.cron.Entries(), 1)
}

// The default schedule is a valid expression that fires when it claims to, and
// is interpreted in UTC regardless of the host's local zone. A default that
// parses but fires at the wrong time would not fail the startup, so it is
// pinned here instead.
func TestScheduleDefaultIsSaturdayInUTC(t *testing.T) {
	scheduler, err := Schedule(cfg.DefaultReportCron, &base.FakeBotAPI{}, testChatID)
	require.NoError(t, err)

	// Monday 2026-06-15; the next run is the Saturday that follows.
	next := scheduler.cron.Entries()[0].Schedule.
		Next(time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC)).UTC()

	assert.Equal(t, time.Date(2026, 6, 20, 10, 0, 0, 0, time.UTC), next)
	assert.Equal(t, time.Saturday, next.Weekday())
}

func TestSchedulerRunStopsOnContextCancellation(t *testing.T) {
	scheduler, err := Schedule("0 10 * * 1", &base.FakeBotAPI{}, testChatID)
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

func TestSchedulerSendsReportsInOrder(t *testing.T) {
	bot := &base.FakeBotAPI{}
	first := &reportStub{name: "first", pages: []string{"activation"}}
	second := &reportStub{name: "second", pages: []string{"events"}}

	scheduler, err := Schedule("* * * * *", bot, testChatID, first, second)
	require.NoError(t, err)

	// The job is invoked directly instead of waiting for a real minute to pass.
	scheduler.ctx = context.Background()
	scheduler.send()

	messages := sentMessages(t, bot)
	require.Len(t, messages, 2)
	assert.Equal(t, "activation", messages[0].Text)
	assert.Equal(t, "events", messages[1].Text)
	assert.Equal(t, testChatID, messages[0].ChatID)
	assert.Equal(t, testChatID, messages[1].ChatID)
}

// Every report ends its window at the same moment.
func TestSchedulerGivesEveryReportTheSameNow(t *testing.T) {
	first := &reportStub{name: "first", pages: []string{"a"}}
	second := &reportStub{name: "second", pages: []string{"b"}}

	scheduler, err := Schedule("* * * * *", &base.FakeBotAPI{}, testChatID, first, second)
	require.NoError(t, err)

	scheduler.ctx = context.Background()
	scheduler.send()

	assert.Equal(t, 1, first.calls)
	assert.Equal(t, 1, second.calls)
	assert.Equal(t, first.now, second.now)
	assert.False(t, first.now.IsZero())
}

func TestSchedulerSendsEveryPageOfAReport(t *testing.T) {
	bot := &base.FakeBotAPI{}
	paged := &reportStub{name: "paged", pages: []string{"page 1", "page 2", "page 3"}}
	next := &reportStub{name: "next", pages: []string{"after"}}

	scheduler, err := Schedule("* * * * *", bot, testChatID, paged, next)
	require.NoError(t, err)

	scheduler.ctx = context.Background()
	scheduler.send()

	messages := sentMessages(t, bot)
	require.Len(t, messages, 4)
	assert.Equal(t, "page 1", messages[0].Text)
	assert.Equal(t, "page 2", messages[1].Text)
	assert.Equal(t, "page 3", messages[2].Text)
	// The next report still follows the pages of the previous one.
	assert.Equal(t, "after", messages[3].Text)
}

// One broken report must not silence the others.
func TestSchedulerKeepsGoingAfterAFailingReport(t *testing.T) {
	bot := &base.FakeBotAPI{}
	broken := &reportStub{name: "broken", err: errors.New("boom")}
	working := &reportStub{name: "working", pages: []string{"events"}}

	scheduler, err := Schedule("* * * * *", bot, testChatID, broken, working)
	require.NoError(t, err)

	scheduler.ctx = context.Background()
	scheduler.send()

	messages := sentMessages(t, bot)
	require.Len(t, messages, 1)
	assert.Equal(t, "events", messages[0].Text)
	assert.Equal(t, 1, working.calls)
}
