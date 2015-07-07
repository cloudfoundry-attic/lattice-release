package config_helpers

import "path/filepath"

func ConfigFileLocation(homeDir string) string {
	configDir := filepath.Join(homeDir, ".lattice")
	return filepath.Join(configDir, "config.json")
}
