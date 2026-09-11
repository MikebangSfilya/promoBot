package config

import (
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testLanguages = []string{"en", "ru"}

func TestNewReportConfig(t *testing.T) {
	t.Run("is disabled when the admins chat is not set", func(t *testing.T) {
		// Set explicitly rather than relying on the ambient environment.
		t.Setenv(EnvAdminsChatID, "")

		cfg, err := NewReportConfig(testLanguages, "ru")

		require.NoError(t, err)
		assert.Nil(t, cfg)
	})

	t.Run("is disabled when the admins chat is only whitespace", func(t *testing.T) {
		t.Setenv(EnvAdminsChatID, "   ")

		cfg, err := NewReportConfig(testLanguages, "ru")

		require.NoError(t, err)
		assert.Nil(t, cfg)
	})

	t.Run("applies the defaults", func(t *testing.T) {
		t.Setenv(EnvAdminsChatID, "-1001234567890")

		cfg, err := NewReportConfig(testLanguages, "ru")

		require.NoError(t, err)
		require.NotNil(t, cfg)
		assert.Equal(t, int64(-1001234567890), cfg.ChatID)
		assert.Equal(t, DefaultReportCron, cfg.Cron)
		assert.Equal(t, "ru", cfg.Lang)
		assert.Equal(t, time.Duration(DefaultReportPeriodWeeks)*week, cfg.Period)
	})

	t.Run("reads every setting", func(t *testing.T) {
		t.Setenv(EnvAdminsChatID, "42")
		t.Setenv(EnvReportCron, "0 9 * * 5")
		t.Setenv(EnvReportPeriodWeek, "3")
		t.Setenv(EnvReportLang, "en")

		cfg, err := NewReportConfig(testLanguages, "ru")

		require.NoError(t, err)
		require.NotNil(t, cfg)
		assert.Equal(t, int64(42), cfg.ChatID)
		assert.Equal(t, "0 9 * * 5", cfg.Cron)
		assert.Equal(t, "en", cfg.Lang)
		assert.Equal(t, 21*24*time.Hour, cfg.Period)
	})

	// os.Getenv does not trim, and REPORT_CRON legitimately contains spaces, so
	// a stray one in .env must not break the startup.
	t.Run("ignores whitespace around the values", func(t *testing.T) {
		t.Setenv(EnvAdminsChatID, "  42\t")
		t.Setenv(EnvReportCron, "  0 9 * * 5  ")
		t.Setenv(EnvReportPeriodWeek, " 3 ")
		t.Setenv(EnvReportLang, " en ")

		cfg, err := NewReportConfig(testLanguages, "ru")

		require.NoError(t, err)
		require.NotNil(t, cfg)
		assert.Equal(t, int64(42), cfg.ChatID)
		assert.Equal(t, "0 9 * * 5", cfg.Cron)
		assert.Equal(t, "en", cfg.Lang)
		assert.Equal(t, 21*24*time.Hour, cfg.Period)
	})

	t.Run("falls back to the default language", func(t *testing.T) {
		t.Setenv(EnvAdminsChatID, "42")

		for _, lang := range []string{"", "de", "EN"} {
			t.Setenv(EnvReportLang, lang)

			cfg, err := NewReportConfig(testLanguages, "ru")

			require.NoError(t, err)
			require.NotNil(t, cfg)
			assert.Equal(t, "ru", cfg.Lang, "language %q", lang)
		}
	})

	t.Run("rejects an unusable chat id", func(t *testing.T) {
		for _, value := range []string{"not-a-number", "1.5", "12abc"} {
			t.Setenv(EnvAdminsChatID, value)

			_, err := NewReportConfig(testLanguages, "ru")

			require.Error(t, err, "value %q", value)
			assert.Contains(t, err.Error(), EnvAdminsChatID)
		}
	})

	// A period outside the range would overflow into a negative duration, which
	// would put the window's start in the future and empty every report.
	t.Run("rejects a period outside the supported range", func(t *testing.T) {
		t.Setenv(EnvAdminsChatID, "42")

		for _, value := range []string{"0", "-1", "not-a-number", "15251"} {
			t.Setenv(EnvReportPeriodWeek, value)

			_, err := NewReportConfig(testLanguages, "ru")

			require.Error(t, err, "value %q", value)
			assert.Contains(t, err.Error(), EnvReportPeriodWeek)
		}
	})

	t.Run("keeps the period positive at the upper bound", func(t *testing.T) {
		t.Setenv(EnvAdminsChatID, "42")
		t.Setenv(EnvReportPeriodWeek, strconv.Itoa(MaxReportPeriodWeeks))

		cfg, err := NewReportConfig(testLanguages, "ru")

		require.NoError(t, err)
		require.NotNil(t, cfg)
		assert.Positive(t, cfg.Period)
		assert.Equal(t, time.Duration(MaxReportPeriodWeeks)*week, cfg.Period)
	})
}
