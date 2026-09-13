package report

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/MikebangSfilya/promoBot/internal/audit"
	"github.com/MikebangSfilya/promoBot/internal/model"

	"github.com/loctools/go-l10n/loc"
	"github.com/samber/lo"
)

// Translation keys of the events report.
const (
	eventsReportTitle   = "eventsReportTitle"
	eventsReportEmpty   = "eventsReportEmpty"
	eventsReportAuthor  = "eventsReportAuthor"
	eventsReportEntry   = "eventsReportEntryFormat"
	eventsReportChanges = "eventsReportChangesFormat"
	eventsReportChange  = "eventsReportChangeFormat"
	eventsTimeLayout    = "eventsTimeLayout"
	eventsActionCreate  = "eventsActionCreate"
	eventsActionUpdate  = "eventsActionUpdate"
	eventsActionDelete  = "eventsActionDelete"
)

// eventsFieldKeys translates the field names promoChanges can produce. Any other
// key is shown as it appears in the log, since the audit format is open-ended.
var eventsFieldKeys = map[string]string{
	"bonus_length": "eventsFieldBonusLength",
	"since":        "eventsFieldSince",
	"until":        "eventsFieldUntil",
	"capacity":     "eventsFieldCapacity",
}

// EventsReport tells who touched the promo codes during the window, grouped by
// person. It is built entirely from the audit log, the only place that records
// who did what.
type EventsReport struct {
	logs     LogFinder
	lang     *loc.Context
	period   time.Duration
	maxPages int
}

func NewEventsReport(logs LogFinder, lang *loc.Context, period time.Duration, maxPages int) *EventsReport {
	return &EventsReport{logs: logs, lang: lang, period: period, maxPages: maxPages}
}

func (*EventsReport) Name() string { return "events" }

// Build renders the report covering the period ending at now.
func (r *EventsReport) Build(_ context.Context, now time.Time) ([]string, error) {
	since := now.Add(-r.period)

	// No action filter: every kind of change belongs in this report.
	events, err := r.logs.FindLogs(since)
	if err != nil {
		return nil, fmt.Errorf("failed to read the audit log for the events report: %w", err)
	}

	// An empty window is still worth reporting: a missing weekly message is
	// indistinguishable from a broken scheduler.
	if len(events) == 0 {
		return []string{fmt.Sprintf(r.lang.Tr(eventsReportEmpty), formatDate(r.lang, &since))}, nil
	}

	// Uniq over the authors in log order, so the groups follow the order people
	// first appear rather than the random order of a map.
	authors := lo.Uniq(lo.Map(events, func(e audit.Log, _ int) string { return e.By }))
	eventsByAuthor := lo.GroupBy(events, func(e audit.Log) string { return e.By })

	blocks := lo.Map(authors, func(author string, _ int) block {
		return block{
			header: fmt.Sprintf(r.lang.Tr(eventsReportAuthor), r.authorName(author)),
			lines: lo.Map(eventsByAuthor[author], func(e audit.Log, _ int) string {
				return r.formatEvent(e)
			}),
		}
	})

	title := fmt.Sprintf(r.lang.Tr(eventsReportTitle), formatDate(r.lang, &since))

	return paginate(pageOptions{
		title:     title,
		suffix:    r.lang.Tr(reportPageSuffix),
		truncated: r.lang.Tr(reportTruncated),
		limit:     maxMessageLen,
		maxPages:  r.maxPages,
	}, blocks), nil
}

func (r *EventsReport) authorName(author string) string {
	if author == "" {
		return r.lang.Tr(reportUnknownAuthor)
	}
	return author
}

func (r *EventsReport) formatEvent(e audit.Log) string {
	line := fmt.Sprintf(r.lang.Tr(eventsReportEntry),
		e.At.UTC().Format(r.lang.Tr(eventsTimeLayout)),
		r.actionName(e.Action),
		e.Code,
	)

	if changes := r.formatChanges(e.Changes); changes != "" {
		line += fmt.Sprintf(r.lang.Tr(eventsReportChanges), changes)
	}

	return line
}

func (r *EventsReport) actionName(action string) string {
	switch action {
	case model.ActionCreate:
		return r.lang.Tr(eventsActionCreate)
	case model.ActionUpdate:
		return r.lang.Tr(eventsActionUpdate)
	case model.ActionDelete:
		return r.lang.Tr(eventsActionDelete)
	default:
		// The log may hold actions this code knows nothing about.
		return action
	}
}

// formatChanges renders the recorded field changes, sorted by field name.
// Changes is a map, so without sorting the order would differ between runs.
func (r *EventsReport) formatChanges(changes map[string]audit.Change) string {
	if len(changes) == 0 {
		return ""
	}

	fields := lo.Keys(changes)
	slices.Sort(fields)

	return strings.Join(lo.Map(fields, func(field string, _ int) string {
		change := changes[field]
		return fmt.Sprintf(r.lang.Tr(eventsReportChange),
			r.fieldName(field),
			r.changeValue(change.Old),
			r.changeValue(change.New),
		)
	}), ", ")
}

func (r *EventsReport) fieldName(field string) string {
	if key, known := eventsFieldKeys[field]; known {
		return r.lang.Tr(key)
	}
	return field
}

// changeValue renders the empty value an open-ended date is recorded with.
func (r *EventsReport) changeValue(value string) string {
	if value == "" {
		return r.lang.Tr(dateEndless)
	}
	return value
}
