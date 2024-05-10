package config

const (
	envPrefix = "EXECUTOR"
)

const (
	hostEnv   = envPrefix + "_HOST"
	portEnv   = envPrefix + "_PORT"
	cmdDirEnv = envPrefix + "_CMD_DIR"
)

const (
	defaultHost   = "0.0.0.0"
	defaultPort   = "8081"
	defaultCmdDir = "/tmp/commands/"
)
