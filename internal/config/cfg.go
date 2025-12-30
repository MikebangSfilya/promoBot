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
)
