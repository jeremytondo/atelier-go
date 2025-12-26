package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
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

// Config represents the application configuration structure.
type Config struct {
	Projects []Project `mapstructure:"projects"`
}

// Load reads the configuration from the XDG config directory.
// It looks in $XDG_CONFIG_HOME/atelier-go/config.toml, defaulting to ~/.config/atelier-go/config.toml.
// If the configuration file is missing, it returns an empty Config without error.
func Load() (*Config, error) {
	var configDir string
	xdgConfigHome := os.Getenv("XDG_CONFIG_HOME")

	if xdgConfigHome != "" {
		configDir = filepath.Join(xdgConfigHome, "atelier-go")
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get user home directory: %w", err)
		}
		configDir = filepath.Join(home, ".config", "atelier-go")
	}

	v := viper.New()
	v.AddConfigPath(configDir)
	v.SetConfigName("config")
	v.SetConfigType("toml")

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error and return empty config
			return &Config{Projects: []Project{}}, nil
		}
		// Config file was found but another error occurred
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}
