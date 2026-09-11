package report

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/MikebangSfilya/promoBot/internal/audit"
	"github.com/MikebangSfilya/promoBot/internal/model"

	tgbotapi "github.com/OvyFlash/telegram-bot-api"
	"github.com/kozalosev/goSadTgBot/base"
	"github.com/loctools/go-l10n/loc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testChatID int64 = -100500

var (
	testNow  = time.Date(2026, 6, 15, 12, 0, 0, 0, time.UTC)
	testCrtd = time.Date(2026, 6, 10, 9, 0, 0, 0, time.UTC)
	testSnc  = time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC)
	testUntl = time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
)

// reporterStub stands in for the repository, the way tableGetterStub does in
// the handlers package.
type reporterStub struct {
	stats []model.PromoActivationStat
	err   error
	since time.Time
	codes []string
	calls int
}

func (s *reporterStub) GetActivationStats(_ context.Context, since time.Time, codes []string) ([]model.PromoActivationStat, error) {
	s.calls++
	s.since = since
	s.codes = codes
	return s.stats, s.err
}

// creationsStub stands in for the audit log.
type creationsStub struct {
	logs   []audit.Log
	err    error
	since  time.Time
	action string
	calls  int
}

func (s *creationsStub) FindLogs(action string, since time.Time) ([]audit.Log, error) {
	s.calls++
	s.since = since
	s.action = action
	return s.logs, s.err
}

func creation(code, by string) audit.Log {
	return audit.Log{Code: code, Action: model.ActionCreate, By: by, At: testCrtd}
}

func testLang() *loc.Context {
	pool := loc.NewPool("en")
	pool.Resources["en"] = map[string]string{
		reportTitle:               "Report since %s",
		reportSectionCreated:      "Created",
		reportSectionActivated:    "Activated",
		reportSectionNotActivated: "Not activated",
		reportSectionCount:        "Total: %d",
		reportEntryCreated:        "%s by %s, %d cm, %d activations, created %s, valid %s - %s",
		reportEntryActivated:      "%s: %d",
		reportEntryNotActivated:   "%s",
		reportEmpty:               "Nothing since %s",
		reportUnknownAuthor:       "unknown",
		dateEndless:               "endless",
	}
	return pool.GetContext("en")
}

func newTestReporter(repo PromoReporter, bot base.ExtendedBotAPI) *Reporter {
	return newTestReporterWith(repo, &creationsStub{}, bot)
}

func newTestReporterWith(repo PromoReporter, created LogFinder, bot base.ExtendedBotAPI) *Reporter {
	return New(repo, created, bot, testChatID, testLang(), 14*24*time.Hour)
}

func stat(code string, inWindow, total int) model.PromoActivationStat {
	since, until := testSnc, testUntl
	return model.PromoActivationStat{
		Code:                code,
		BonusLength:         10,
		Capacity:            5,
		Since:               &since,
		Until:               &until,
		ActivationsInWindow: inWindow,
		ActivationsTotal:    total,
	}
}

func TestReporter_Build(t *testing.T) {
	tests := []struct {
		name        string
		creations   []audit.Log
		stats       []model.PromoActivationStat
		wantContain []string
		wantAbsent  []string
	}{
		{
			name:        "empty window",
			wantContain: []string{"Nothing since 2026-06-01"},
			wantAbsent:  []string{"Created", "Activated"},
		},
		{
			name:      "all activated",
			creations: []audit.Log{creation("AAA", "boss"), creation("BBB", "boss")},
			stats:     []model.PromoActivationStat{stat("AAA", 3, 3), stat("BBB", 1, 1)},
			wantContain: []string{
				"Report since 2026-06-01",
				"AAA by boss, 10 cm, 8 activations, created 2026-06-10, valid 2026-06-10 - 2026-07-01",
				"AAA: 3",
				"BBB: 1",
				"Not activated: \n\n\nTotal: 0",
			},
		},
		{
			name:      "none activated",
			creations: []audit.Log{creation("AAA", "boss")},
			stats:     []model.PromoActivationStat{stat("AAA", 0, 0)},
			wantContain: []string{
				"Activated: \n\n\nTotal: 0",
				"Not activated: \n\n1. AAA\n\nTotal: 1",
			},
		},
		{
			name:      "mixed",
			creations: []audit.Log{creation("AAA", "boss"), creation("BBB", "boss")},
			stats:     []model.PromoActivationStat{stat("AAA", 2, 4), stat("BBB", 0, 0)},
			wantContain: []string{
				"1. AAA: 2\n\nTotal: 1",
				"1. BBB\n\nTotal: 1",
				// Initial capacity is the remaining capacity plus every
				// activation ever, not just the ones inside the window.
				"AAA by boss, 10 cm, 9 activations",
			},
			wantAbsent: []string{"BBB: 0"},
		},
		{
			name:      "a code deleted since its creation is dropped",
			creations: []audit.Log{creation("AAA", "boss"), creation("GONE", "boss")},
			stats:     []model.PromoActivationStat{stat("AAA", 1, 1)},
			// Nothing was created that still exists, apart from AAA.
			wantContain: []string{"Created: \n\n1. AAA", "Total: 1"},
			wantAbsent:  []string{"GONE"},
		},
		{
			name:        "every created code was deleted",
			creations:   []audit.Log{creation("GONE", "boss")},
			stats:       nil,
			wantContain: []string{"Nothing since 2026-06-01"},
			wantAbsent:  []string{"GONE"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &reporterStub{stats: tt.stats}
			created := &creationsStub{logs: tt.creations}
			text, err := newTestReporterWith(repo, created, &base.FakeBotAPI{}).
				Build(context.Background(), testNow)

			require.NoError(t, err)

			// 14 days before testNow.
			window := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
			assert.Equal(t, 1, created.calls)
			assert.Equal(t, window, created.since)
			// The report is about created promo codes, nothing else.
			assert.Equal(t, model.ActionCreate, created.action)
			if len(tt.creations) > 0 {
				assert.Equal(t, 1, repo.calls)
				assert.Equal(t, window, repo.since)
			} else {
				// Nothing was created, so the database is never queried.
				assert.Zero(t, repo.calls)
			}

			for _, want := range tt.wantContain {
				assert.Contains(t, text, want)
			}
			for _, absent := range tt.wantAbsent {
				assert.NotContains(t, text, absent)
			}
		})
	}
}

// The same code created twice keeps the most recent author and date.
func TestReporter_BuildLatestCreationWins(t *testing.T) {
	first := creation("AAA", "first")
	second := creation("AAA", "second")
	second.At = testCrtd.Add(24 * time.Hour)

	repo := &reporterStub{stats: []model.PromoActivationStat{stat("AAA", 0, 0)}}
	text, err := newTestReporterWith(repo, &creationsStub{logs: []audit.Log{first, second}}, &base.FakeBotAPI{}).
		Build(context.Background(), testNow)

	require.NoError(t, err)
	assert.Equal(t, []string{"AAA"}, repo.codes)
	assert.Contains(t, text, "AAA by second, 10 cm, 5 activations, created 2026-06-11")
	assert.NotContains(t, text, "by first")
}

func TestReporter_BuildAuditErrorIsReported(t *testing.T) {
	_, err := newTestReporterWith(
		&reporterStub{}, &creationsStub{err: errors.New("audit boom")}, &base.FakeBotAPI{},
	).Build(context.Background(), testNow)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "audit boom")
}

func TestReporter_BuildEndlessPromoAndUnknownAuthor(t *testing.T) {
	endless := stat("AAA", 0, 0)
	endless.Until = nil

	text, err := newTestReporterWith(
		&reporterStub{stats: []model.PromoActivationStat{endless}},
		&creationsStub{logs: []audit.Log{creation("AAA", "")}},
		&base.FakeBotAPI{},
	).Build(context.Background(), testNow)

	require.NoError(t, err)
	assert.Contains(t, text, "AAA by unknown, 10 cm, 5 activations, created 2026-06-10, valid 2026-06-10 - endless")
}

func TestReporter_BuildRepositoryError(t *testing.T) {
	repo := &reporterStub{err: errors.New("boom")}

	_, err := newTestReporterWith(repo, &creationsStub{logs: []audit.Log{creation("AAA", "boss")}}, &base.FakeBotAPI{}).
		Build(context.Background(), testNow)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "boom")
}

func TestReporter_Send(t *testing.T) {
	bot := &base.FakeBotAPI{}
	repo := &reporterStub{stats: []model.PromoActivationStat{stat("AAA", 1, 1)}}
	created := &creationsStub{logs: []audit.Log{creation("AAA", "boss")}}

	require.NoError(t, newTestReporterWith(repo, created, bot).Send(context.Background()))

	sent, ok := bot.GetOutput().([]tgbotapi.Chattable)
	require.True(t, ok)
	require.Len(t, sent, 1)

	msg, ok := sent[0].(tgbotapi.MessageConfig)
	require.True(t, ok)
	assert.Equal(t, testChatID, msg.ChatID)
	assert.Contains(t, msg.Text, "AAA: 1")
}

func TestReporter_SendRepositoryErrorIsNotDelivered(t *testing.T) {
	bot := &base.FakeBotAPI{}

	err := newTestReporterWith(
		&reporterStub{err: errors.New("boom")},
		&creationsStub{logs: []audit.Log{creation("AAA", "boss")}},
		bot,
	).Send(context.Background())

	require.Error(t, err)
	assert.Empty(t, bot.GetOutput())
}
