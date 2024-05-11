package config

import (
	"fmt"
	"os"
)

func GetHost() string {
	s := os.Getenv(hostEnv)
	if s == "" {
		return defaultHost
	}
	return s
}

func GetPort() string {
	s := os.Getenv(portEnv)
	if s == "" {
		return defaultPort
	}
	return s
}

func GetCmdDir() string {
	s := os.Getenv(cmdDirEnv)
	if s == "" {
		return defaultCmdDir
	}
	return s
}

func GetDbConnStr() string {
	s := os.Getenv(dbConnStrEnv)
	if s == "" {
		panic(fmt.Errorf("db connection string is empty"))
	}
	return s
}

func GetMigrationsSource() string {
	s := os.Getenv(migrationsSourceEnv)
	if s == "" {
		return defaultMigrationsSource
	}
	return s
}
