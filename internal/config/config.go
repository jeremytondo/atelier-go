// Package config provides the application configuration structures and loading logic.
package config

import (
	"atelier-go/internal/utils"
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

// Config represents the application configuration.
type Config struct {
	Projects []Project `mapstructure:"projects"`
	Editor   string    `mapstructure:"editor"`
}

// LoadConfig loads the configuration from the config directory.
// It loads config.yaml first, then merges <hostname>.yaml if it exists.
func LoadConfig() (*Config, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return nil, err
	}

	// Load global config
	vGlobal := viper.New()
	vGlobal.SetConfigType("yaml")
	vGlobal.AddConfigPath(configDir)
	vGlobal.SetConfigName("config")

	if err := vGlobal.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read global config: %w", err)
		}
	}

	var cfg Config
	if err := vGlobal.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal global config: %w", err)
	}

	// Load host-specific config
	hostname, _ := utils.GetHostname()
	if hostname != "" {
		vHost := viper.New()
		vHost.SetConfigType("yaml")
		vHost.AddConfigPath(configDir)
		vHost.SetConfigName(hostname)

		if err := vHost.ReadInConfig(); err == nil {
			var hostCfg Config
			if err := vHost.Unmarshal(&hostCfg); err == nil {
				// Manually merge projects
				cfg.Projects = mergeProjects(cfg.Projects, hostCfg.Projects)
			} else {
				return nil, fmt.Errorf("failed to unmarshal host config: %w", err)
			}
		} else if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read host config: %w", err)
		}
	}

	return &cfg, nil
}

// mergeProjects merges two project slices. Projects in host override global by name.
func mergeProjects(global, host []Project) []Project {
	projectMap := make(map[string]int)
	merged := make([]Project, len(global))
	copy(merged, global)

	for i, p := range merged {
		projectMap[p.Name] = i
	}

	for _, hp := range host {
		if idx, exists := projectMap[hp.Name]; exists {
			// Override
			merged[idx] = hp
		} else {
			// Append
			merged = append(merged, hp)
		}
	}

	return merged
}

// GetEditor returns the configured editor or fallbacks.
func (c *Config) GetEditor() string {
	if c.Editor != "" {
		return c.Editor
	}
	if ed := os.Getenv("EDITOR"); ed != "" {
		return ed
	}
	return "vim"
}

// GetConfigDir returns the configuration directory respecting XDG_CONFIG_HOME.
func GetConfigDir() (string, error) {
	xdgConfigHome := os.Getenv("XDG_CONFIG_HOME")
	if xdgConfigHome != "" {
		expanded, err := utils.ExpandPath(xdgConfigHome)
		if err != nil {
			return "", fmt.Errorf("failed to expand XDG_CONFIG_HOME: %w", err)
		}
		return filepath.Join(expanded, "atelier-go"), nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}
	return filepath.Join(home, ".config", "atelier-go"), nil
}
