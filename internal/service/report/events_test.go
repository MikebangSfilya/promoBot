package report

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/MikebangSfilya/promoBot/internal/audit"
	"github.com/MikebangSfilya/promoBot/internal/model"

	"github.com/loctools/go-l10n/loc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestEventsReport(logs LogFinder) *EventsReport {
	return NewEventsReport(logs, testLang(), 7*24*time.Hour, testMaxPages)
}

func event(code, action, by string, at time.Time) audit.Log {
	return audit.Log{Code: code, Action: action, By: by, At: at}
}

func TestEventsReport_Build(t *testing.T) {
	first := testNow.Add(-72 * time.Hour)
	second := testNow.Add(-48 * time.Hour)
	third := testNow.Add(-24 * time.Hour)

	logs := &creationsStub{logs: []audit.Log{
		event("SUMMER", model.ActionCreate, "boss", first),
		event("X1Y2Z3", model.ActionCreate, "auto", second),
		event("SUMMER", model.ActionDelete, "boss", third),
	}}

	text := buildPage(t, newTestEventsReport(logs), testNow)

	assert.Contains(t, text, "Changes since 2026-06-08")

	// Grouped by person, in the order people first appear, and every event of a
	// person is listed under their name.
	assert.Equal(t,
		"Changes since 2026-06-08\n\n"+
			"@boss\n"+
			"2026-06-12 12:00 created SUMMER\n"+
			"2026-06-14 12:00 deleted SUMMER\n\n"+
			"@auto\n"+
			"2026-06-13 12:00 created X1Y2Z3",
		text)

	// The window is the period before now, and every action is asked for.
	assert.Equal(t, 1, logs.calls)
	assert.Equal(t, testNow.Add(-7*24*time.Hour), logs.since)
	assert.Empty(t, logs.actions)
}

func TestEventsReport_BuildShowsChanges(t *testing.T) {
	update := event("SUMMER", model.ActionUpdate, "boss", testNow.Add(-time.Hour))
	update.Changes = map[string]audit.Change{
		"capacity":     {Old: "5", New: "10"},
		"bonus_length": {Old: "10", New: "15"},
		"until":        {Old: "2026-10-01", New: ""},
	}

	text := buildPage(t, newTestEventsReport(&creationsStub{logs: []audit.Log{update}}), testNow)

	// Fields are sorted by name, and the empty value of an open-ended date is
	// rendered rather than left blank.
	assert.Contains(t, text,
		"updated SUMMER: length 10 -> 15, activations 5 -> 10, end 2026-10-01 -> endless")
}

// Changes is a map, so the order has to be imposed rather than inherited.
func TestEventsReport_BuildChangeOrderIsStable(t *testing.T) {
	update := event("SUMMER", model.ActionUpdate, "boss", testNow.Add(-time.Hour))
	update.Changes = map[string]audit.Change{
		"capacity":     {Old: "1", New: "2"},
		"bonus_length": {Old: "3", New: "4"},
		"since":        {Old: "2026-01-01", New: "2026-02-01"},
		"until":        {Old: "2026-03-01", New: "2026-04-01"},
	}

	first := buildPage(t, newTestEventsReport(&creationsStub{logs: []audit.Log{update}}), testNow)

	for range 20 {
		assert.Equal(t, first,
			buildPage(t, newTestEventsReport(&creationsStub{logs: []audit.Log{update}}), testNow))
	}
}

func TestEventsReport_BuildUnknownAuthorAndAction(t *testing.T) {
	logs := &creationsStub{logs: []audit.Log{
		event("AAA", "archived", "", testNow.Add(-time.Hour)),
	}}

	text := buildPage(t, newTestEventsReport(logs), testNow)

	assert.Contains(t, text, "@unknown")
	// An action this code knows nothing about is shown as it was recorded.
	assert.Contains(t, text, "archived AAA")
}

func TestEventsReport_BuildEmptyWindow(t *testing.T) {
	text := buildPage(t, newTestEventsReport(&creationsStub{}), testNow)

	assert.Equal(t, "No changes since 2026-06-08", text)
	assert.NotContains(t, text, "@")
}

func TestEventsReport_BuildAuditErrorIsReported(t *testing.T) {
	_, err := newTestEventsReport(&creationsStub{err: errors.New("audit boom")}).
		Build(context.Background(), testNow)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "audit boom")
}

// A person with more events than fit in one message keeps their name on every
// page their events continue on.
func TestEventsReport_BuildPaginatesLongAuthor(t *testing.T) {
	var logs []audit.Log
	for i := range 400 {
		logs = append(logs, event("PROMOCODE", model.ActionUpdate, "boss",
			testNow.Add(-time.Duration(i+1)*time.Minute)))
	}

	pages, err := newTestEventsReport(&creationsStub{logs: logs}).Build(context.Background(), testNow)
	require.NoError(t, err)
	require.Greater(t, len(pages), 1)

	for i, page := range pages {
		assert.LessOrEqual(t, msgLen(page), maxMessageLen)
		assert.Contains(t, page, "@boss", "page %d lost the author", i+1)
		assert.Contains(t, page, "(page ")
	}

	// Every event survives the split.
	assert.Equal(t, len(logs), strings.Count(strings.Join(pages, "\n"), "updated PROMOCODE"))
}

// The timestamp is localized: English writes it the ISO way, Russian as
// dd.mm.yyyy. Both are UTC.
func TestEventsReport_BuildLocalizedTimeLayout(t *testing.T) {
	pool := loc.NewPool("ru")
	pool.Resources["ru"] = map[string]string{
		eventsReportTitle:  "Изменения с %s",
		eventsReportAuthor: "@%s",
		eventsReportEntry:  "%s %s %s",
		eventsTimeLayout:   "02.01.2006 15:04",
		reportDateLayout:   "02.01.2006",
		eventsActionCreate: "создал",
		reportPageSuffix:   "%s (%d)",
	}

	events := NewEventsReport(&creationsStub{logs: []audit.Log{
		event("AAA", model.ActionCreate, "boss", testNow.Add(-24*time.Hour)),
	}}, pool.GetContext("ru"), 7*24*time.Hour, testMaxPages)

	pages, err := events.Build(context.Background(), testNow)
	require.NoError(t, err)
	require.Len(t, pages, 1)

	assert.Contains(t, pages[0], "14.06.2026 12:00 создал AAA")
	assert.NotContains(t, pages[0], "2026-06-14")

	// The title date follows the same locale, so one message never mixes formats.
	assert.Contains(t, pages[0], "Изменения с 08.06.2026")
	assert.NotContains(t, pages[0], "2026-06-08")
}

func TestEventsReport_Name(t *testing.T) {
	assert.Equal(t, "events", newTestEventsReport(&creationsStub{}).Name())
}
