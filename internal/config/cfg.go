package config

// Environment variable names.
// Some constants may be unused in the current codebase but are kept
// for documentation and future use.
const (
	// Basic parameters

	EnvAPIToken        = "API_TOKEN"
	EnvDebug           = "DEBUG"
	EnvDevLevel        = "DEV_LVL"
	EnvAppPort         = "APP_PORT"
	EnvCommandStateTTL = "COMMAND_STATE_TTL"

	// Application specific

	EnvUsersConfigFile = "USERS_CONFIG_FILE"
	EnvAuditLogsDir    = "AUDIT_LOGS_DIR"

	// Weekly activation report

	EnvAdminsChatID     = "ADMINS_CHAT_ID"
	EnvReportCron       = "REPORT_CRON"
	EnvReportPeriodWeek = "REPORT_PERIOD_WEEKS"
	EnvReportLang       = "REPORT_LANG"

	// WebHook related

	EnvAppPath     = "APP_PATH"
	EnvWebhookHost = "WEBHOOK_HOST"
	EnvWebhookPort = "WEBHOOK_PORT"
	EnvWebhookPath = "WEBHOOK_PATH"

	// Redis connection options
	EnvRedisHost     = "REDIS_HOST"
	EnvRedisPort     = "REDIS_PORT"
	EnvRedisPassword = "REDIS_PASSWORD"

	// Database connection options
	EnvPostgresHost     = "POSTGRES_HOST"
	EnvPostgresPort     = "POSTGRES_PORT"
	EnvPostgresDB       = "POSTGRES_DB"
	EnvPostgresUser     = "POSTGRES_USER"
	EnvPostgresPassword = "POSTGRES_PASSWORD"
	EnvMigrationsRepo   = "MIGRATIONS_REPO"
)

const (
	LocalLogDir = "audit-logs"

	// DefaultReportCron sends the report every Saturday at 10:00 UTC.
	DefaultReportCron = "0 10 * * 6"
	// DefaultReportPeriodWeeks is how far back the report looks by default.
	DefaultReportPeriodWeeks = 2
	// MaxReportPeriodWeeks bounds REPORT_PERIOD_WEEKS. The window is turned
	// into a time.Duration, which wraps to a negative value past ~15250 weeks,
	// and a negative window would put the report's start date in the future and
	// make every report come back empty instead of failing.
	MaxReportPeriodWeeks = 520
)
