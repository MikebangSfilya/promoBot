package config

import (
	"fmt"
	"log/slog"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"
)

const week = 7 * 24 * time.Hour

// ReportConfig holds the settings of the weekly activation report, already
// parsed and validated.
type ReportConfig struct {
	// ChatID is the chat the report is sent to.
	ChatID int64
	// Cron is the schedule expression, interpreted in UTC.
	Cron string
	// Period is how far back the activation report looks.
	Period time.Duration
	// EventsPeriod is how far back the events report looks.
	EventsPeriod time.Duration
	// MaxPages caps how many messages a single report may be split into.
	MaxPages int
	// Lang is the language the report is written in.
	Lang string
}

// NewReportConfig reads the weekly report settings from the environment.
//
// The report is opt-in: when ADMINS_CHAT_ID is not set there is nowhere to send
// it, so the call returns (nil, nil) and the caller is expected to skip the
// feature. Anything else that is set but unusable is an error, so that a typo
// stops the bot instead of quietly disabling the report.
//
// An unsupported language is the one exception: it falls back to defaultLang
// with a warning, since the report is still worth sending in another language.
func NewReportConfig(supportedLanguages []string, defaultLang string) (*ReportConfig, error) {
	rawChatID := trimmedEnv(EnvAdminsChatID)
	if rawChatID == "" {
		return nil, nil
	}

	chatID, err := strconv.ParseInt(rawChatID, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("%s is not a valid chat id: %q: %w", EnvAdminsChatID, rawChatID, err)
	}

	period, err := periodWeeks(EnvReportPeriodWeek, DefaultReportPeriodWeeks)
	if err != nil {
		return nil, err
	}

	eventsPeriod, err := periodWeeks(EnvReportEventsPeriodWeeks, DefaultReportEventsPeriodWeeks)
	if err != nil {
		return nil, err
	}

	maxPages, err := boundedInt(EnvReportMaxPages, DefaultReportMaxPages, MaxReportPagesAllowed)
	if err != nil {
		return nil, err
	}

	cron := trimmedEnv(EnvReportCron)
	if cron == "" {
		cron = DefaultReportCron
	}

	lang := trimmedEnv(EnvReportLang)
	if !slices.Contains(supportedLanguages, lang) {
		if lang != "" {
			slog.Warn("unsupported report language, falling back to the default one",
				slog.String("variable", EnvReportLang),
				slog.String("value", lang),
				slog.String("default", defaultLang))
		}
		lang = defaultLang
	}

	return &ReportConfig{
		ChatID:       chatID,
		Cron:         cron,
		Period:       period,
		EventsPeriod: eventsPeriod,
		MaxPages:     maxPages,
		Lang:         lang,
	}, nil
}

// periodWeeks reads a window given in whole weeks, falling back to def.
//
// The upper bound is what keeps the multiplication below from overflowing into a
// negative duration, which would put the window's start in the future.
func periodWeeks(key string, def int) (time.Duration, error) {
	weeks := def

	if raw := trimmedEnv(key); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed <= 0 || parsed > MaxReportPeriodWeeks {
			return 0, fmt.Errorf("%s must be a whole number of weeks between 1 and %d, got %q",
				key, MaxReportPeriodWeeks, raw)
		}
		weeks = parsed
	}

	return time.Duration(weeks) * week, nil
}

// boundedInt reads a whole number within 1..max, falling back to def.
func boundedInt(key string, def, max int) (int, error) {
	raw := trimmedEnv(key)
	if raw == "" {
		return def, nil
	}

	parsed, err := strconv.Atoi(raw)
	if err != nil || parsed <= 0 || parsed > max {
		return 0, fmt.Errorf("%s must be a whole number between 1 and %d, got %q", key, max, raw)
	}

	return parsed, nil
}

func trimmedEnv(key string) string {
	return strings.TrimSpace(os.Getenv(key))
}
