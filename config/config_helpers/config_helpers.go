package config_helpers

import (
	"path/filepath"
)

func ConfigFileLocation(homeDir string) string {
	configDir := filepath.Join(homeDir, ".diego")
	return filepath.Join(configDir, "config.json")
}
