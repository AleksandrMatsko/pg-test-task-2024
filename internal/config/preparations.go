package config

import "os"

// PrepareCmdDir uses MkdirAll to create dir in which
// files with scripts will be stored
func PrepareCmdDir(path string) error {
	return os.MkdirAll(path, 0700)
}
