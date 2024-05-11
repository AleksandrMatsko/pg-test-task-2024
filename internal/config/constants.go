package config

const (
	envPrefix = "EXECUTOR"
)

const (
	hostEnv             = envPrefix + "_HOST"
	portEnv             = envPrefix + "_PORT"
	cmdDirEnv           = envPrefix + "_CMD_DIR"
	dbConnStrEnv        = envPrefix + "_DB_CONN_STR"
	migrationsSourceEnv = envPrefix + "_MIGRATIONS_SOURCE"
)

const (
	defaultHost             = "0.0.0.0"
	defaultPort             = "8081"
	defaultCmdDir           = "/tmp/commands/"
	defaultMigrationsSource = "file://scripts/migrations"
)
