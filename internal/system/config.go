package system

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// GetConfigDir returns the configuration directory for atelier-go.
// It checks XDG_CONFIG_HOME, falling back to ~/.config/atelier-go.
func GetConfigDir() (string, error) {
	configDir := os.Getenv("XDG_CONFIG_HOME")
	if configDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		configDir = filepath.Join(home, ".config")
	}

	return filepath.Join(configDir, "atelier-go"), nil
}

// LoadConfig initializes viper with the given config name (e.g. "server" or "client").
// It looks for the config file in the directory returned by GetConfigDir.
func LoadConfig(configName string) {
	// 1. Determine the config directory
	atelierConfigDir, err := GetConfigDir()
	if err != nil {
		// If we can't find home/config dir, we can't load config.
		// Just return and rely on defaults/flags.
		return
	}

	// 2. Setup Viper
	viper.AddConfigPath(atelierConfigDir)
	viper.SetConfigName(configName)
	viper.SetConfigType("toml")

	// 3. Read Config
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			// Config file was found but had an error (e.g. invalid syntax)
			fmt.Fprintf(os.Stderr, "Warning: failed to read config file: %v\n", err)
		}
	}
}
