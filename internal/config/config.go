package config

import "os"

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
