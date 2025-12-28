// Package config provides the application configuration structures and loading logic.
package config

import (
	"fmt"
	"os"
	"path/filepath"
)

// Project represents a defined project with a name and a filesystem path.
type Project struct {
	Name    string   `mapstructure:"name"`
	Path    string   `mapstructure:"path"`
	Actions []Action `mapstructure:"actions"`
}

// Action represents a runnable command associated with a project.
type Action struct {
	Name    string `mapstructure:"name"`
	Command string `mapstructure:"command"`
}

// GetConfigDir returns the configuration directory respecting XDG_CONFIG_HOME.
func GetConfigDir() (string, error) {
	xdgConfigHome := os.Getenv("XDG_CONFIG_HOME")
	if xdgConfigHome != "" {
		return filepath.Join(xdgConfigHome, "atelier-go"), nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}
	return filepath.Join(home, ".config", "atelier-go"), nil
}
