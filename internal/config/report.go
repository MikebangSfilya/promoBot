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
	// Period is how far back a report looks.
	Period time.Duration
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

	period, err := reportPeriod()
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

	return &ReportConfig{ChatID: chatID, Cron: cron, Period: period, Lang: lang}, nil
}

func reportPeriod() (time.Duration, error) {
	weeks := DefaultReportPeriodWeeks

	if raw := trimmedEnv(EnvReportPeriodWeek); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed <= 0 || parsed > MaxReportPeriodWeeks {
			return 0, fmt.Errorf("%s must be a whole number of weeks between 1 and %d, got %q",
				EnvReportPeriodWeek, MaxReportPeriodWeeks, raw)
		}
		weeks = parsed
	}

	return time.Duration(weeks) * week, nil
}

func trimmedEnv(key string) string {
	return strings.TrimSpace(os.Getenv(key))
}
