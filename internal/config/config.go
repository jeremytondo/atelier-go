package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Project represents a defined project with a name and a filesystem path.
type Project struct {
	Name string `mapstructure:"name"`
	Path string `mapstructure:"path"`
}

// Config represents the application configuration structure.
type Config struct {
	Projects []Project `mapstructure:"projects"`
}

// Load reads the configuration from ~/.config/atelier/config.toml.
// If the configuration file is missing, it returns an empty Config without error.
func Load() (*Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home directory: %w", err)
	}

	configDir := filepath.Join(home, ".config", "atelier")

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
