package pf

import (
	"os"
)

const defaultPfConfPath = "/etc/pf.conf"

func (c *Config) getPath() string {
	if c.FilePath != "" {
		return c.FilePath
	}
	return defaultPfConfPath
}

func backupFile(path string, content []byte) error {
	return os.WriteFile(path+".bak", content, 0644)
}
