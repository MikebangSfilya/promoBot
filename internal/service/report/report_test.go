package report

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/MikebangSfilya/promoBot/internal/audit"
	"github.com/MikebangSfilya/promoBot/internal/model"

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
	logs    []audit.Log
	err     error
	since   time.Time
	actions []string
	calls   int
}

func (s *creationsStub) FindLogs(since time.Time, actions ...string) ([]audit.Log, error) {
	s.calls++
	s.since = since
	s.actions = actions
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
		reportPageSuffix:          "%s (page %d)",
		reportDateLayout:          "2006-01-02",
		reportTruncated:           "... %d more page(s) left out",
		dateEndless:               "endless",

		eventsReportTitle:        "Changes since %s",
		eventsReportEmpty:        "No changes since %s",
		eventsReportAuthor:       "@%s",
		eventsReportEntry:        "%s %s %s",
		eventsReportChanges:      ": %s",
		eventsReportChange:       "%s %s -> %s",
		eventsTimeLayout:         "2006-01-02 15:04",
		eventsActionCreate:       "created",
		eventsActionUpdate:       "updated",
		eventsActionDelete:       "deleted",
		"eventsFieldBonusLength": "length",
		"eventsFieldSince":       "start",
		"eventsFieldUntil":       "end",
		"eventsFieldCapacity":    "activations",
	}
	return pool.GetContext("en")
}

// testMaxPages is deliberately larger than any report these tests build, so
// that only the pagination tests deal with the cap.
const testMaxPages = 20

func newTestActivationReport(repo PromoReporter, logs LogFinder) *ActivationReport {
	return NewActivationReport(repo, logs, testLang(), 14*24*time.Hour, testMaxPages)
}

// buildPage renders a report and requires it to fit in a single message, which
// is what every test here expects. Pagination is covered on its own.
func buildPage(t *testing.T, r Report, now time.Time) string {
	t.Helper()

	pages, err := r.Build(context.Background(), now)
	require.NoError(t, err)
	require.Len(t, pages, 1)

	return pages[0]
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

func TestActivationReport_Build(t *testing.T) {
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
			text := buildPage(t, newTestActivationReport(repo, created), testNow)

			// 14 days before testNow.
			window := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
			assert.Equal(t, 1, created.calls)
			assert.Equal(t, window, created.since)
			// The report is about created promo codes, nothing else.
			assert.Equal(t, []string{model.ActionCreate}, created.actions)
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
func TestActivationReport_BuildLatestCreationWins(t *testing.T) {
	first := creation("AAA", "first")
	second := creation("AAA", "second")
	second.At = testCrtd.Add(24 * time.Hour)

	repo := &reporterStub{stats: []model.PromoActivationStat{stat("AAA", 0, 0)}}
	text := buildPage(t,
		newTestActivationReport(repo, &creationsStub{logs: []audit.Log{first, second}}), testNow)

	assert.Equal(t, []string{"AAA"}, repo.codes)
	assert.Contains(t, text, "AAA by second, 10 cm, 5 activations, created 2026-06-11")
	assert.NotContains(t, text, "by first")
}

func TestActivationReport_BuildAuditErrorIsReported(t *testing.T) {
	_, err := newTestActivationReport(&reporterStub{}, &creationsStub{err: errors.New("audit boom")}).
		Build(context.Background(), testNow)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "audit boom")
}

func TestActivationReport_BuildEndlessPromoAndUnknownAuthor(t *testing.T) {
	endless := stat("AAA", 0, 0)
	endless.Until = nil

	text := buildPage(t, newTestActivationReport(
		&reporterStub{stats: []model.PromoActivationStat{endless}},
		&creationsStub{logs: []audit.Log{creation("AAA", "")}},
	), testNow)

	assert.Contains(t, text, "AAA by unknown, 10 cm, 5 activations, created 2026-06-10, valid 2026-06-10 - endless")
}

func TestActivationReport_BuildRepositoryError(t *testing.T) {
	repo := &reporterStub{err: errors.New("boom")}

	_, err := newTestActivationReport(repo, &creationsStub{logs: []audit.Log{creation("AAA", "boss")}}).
		Build(context.Background(), testNow)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "boom")
}

// formatDate is shared by both reports, so the activation one is pinned in a
// non-English locale too.
func TestActivationReport_BuildLocalizedDates(t *testing.T) {
	pool := loc.NewPool("ru")
	pool.Resources["ru"] = map[string]string{
		reportTitle:               "Отчёт с %s",
		reportSectionCreated:      "Созданы",
		reportSectionActivated:    "Активированы",
		reportSectionNotActivated: "Не активированы",
		reportSectionCount:        "Всего: %d",
		reportEntryCreated:        "%s от %s, %d см, %d активаций, создан %s, действует %s - %s",
		reportEntryActivated:      "%s: %d",
		reportEntryNotActivated:   "%s",
		reportPageSuffix:          "%s (%d)",
		reportDateLayout:          "02.01.2006",
		dateEndless:               "бессрочно",
	}

	activation := NewActivationReport(
		&reporterStub{stats: []model.PromoActivationStat{stat("AAA", 0, 0)}},
		&creationsStub{logs: []audit.Log{creation("AAA", "boss")}},
		pool.GetContext("ru"),
		14*24*time.Hour,
		testMaxPages,
	)

	pages, err := activation.Build(context.Background(), testNow)
	require.NoError(t, err)
	require.Len(t, pages, 1)

	assert.Contains(t, pages[0], "Отчёт с 01.06.2026")
	assert.Contains(t, pages[0],
		"AAA от boss, 10 см, 5 активаций, создан 10.06.2026, действует 10.06.2026 - 01.07.2026")
	assert.NotContains(t, pages[0], "2026-06")
}

// manyCreatedCodes returns more created codes than fit in a single message.
func manyCreatedCodes(count int) ([]audit.Log, []model.PromoActivationStat, []string) {
	var (
		creations []audit.Log
		stats     []model.PromoActivationStat
		codes     []string
	)

	for i := range count {
		code := fmt.Sprintf("CODE%04d", i)
		codes = append(codes, code)
		creations = append(creations, creation(code, "boss"))
		stats = append(stats, stat(code, 0, 0))
	}

	return creations, stats, codes
}

// The activation report paginates the same way the events one does, except that
// its blocks are the sections, so a long section continues under its own title.
func TestActivationReport_BuildPaginatesLongSections(t *testing.T) {
	creations, stats, codes := manyCreatedCodes(200)

	pages, err := newTestActivationReport(&reporterStub{stats: stats}, &creationsStub{logs: creations}).
		Build(context.Background(), testNow)

	require.NoError(t, err)
	require.Greater(t, len(pages), 1)

	for i, page := range pages {
		assert.LessOrEqual(t, msgLen(page), maxMessageLen, "page %d is over the limit", i+1)
		assert.Contains(t, page, fmt.Sprintf("(page %d)", i+1))
	}

	joined := strings.Join(pages, "\n")

	// The section title comes back wherever the section continues.
	assert.Greater(t, strings.Count(joined, "Created"), 1)

	// Every code survives the split. Each is listed twice on purpose: once under
	// "Created", and once under the activation section it belongs to — here
	// "Not activated", since none of them was used.
	for _, code := range codes {
		assert.Equal(t, 2, strings.Count(joined, code), "code %s", code)
	}

	// The cap is generous here, so nothing was dropped.
	assert.NotContains(t, joined, "left out")
}

// The page cap applies to this report too, not only to the events one.
func TestActivationReport_BuildRespectsThePageCap(t *testing.T) {
	creations, stats, _ := manyCreatedCodes(200)

	capped := NewActivationReport(
		&reporterStub{stats: stats},
		&creationsStub{logs: creations},
		testLang(),
		14*24*time.Hour,
		2,
	)

	pages, err := capped.Build(context.Background(), testNow)

	require.NoError(t, err)
	require.Len(t, pages, 2)
	assert.Contains(t, pages[1], "more page(s) left out")

	for i, page := range pages {
		assert.LessOrEqual(t, msgLen(page), maxMessageLen, "page %d is over the limit", i+1)
	}
}

func TestActivationReport_Name(t *testing.T) {
	assert.Equal(t, "activation", newTestActivationReport(&reporterStub{}, &creationsStub{}).Name())
}
