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
	Name           string   `mapstructure:"name"`
	Path           string   `mapstructure:"path"`
	Actions        []Action `mapstructure:"actions"`
	DefaultActions *bool    `mapstructure:"default-actions"`
	ShellDefault   *bool    `mapstructure:"shell-default"`
}

// Action represents a runnable command associated with a project.
type Action struct {
	Name    string `mapstructure:"name"`
	Command string `mapstructure:"command"`
}

// Config represents the application configuration.
type Config struct {
	Projects     []Project `mapstructure:"projects"`
	Actions      []Action  `mapstructure:"actions"`
	ShellDefault *bool     `mapstructure:"shell-default"`
	Editor       string    `mapstructure:"editor"`
	Theme        Theme     `mapstructure:"theme"`
}

// Theme holds color settings for the UI.
type Theme struct {
	Primary   string `mapstructure:"primary"`
	Accent    string `mapstructure:"accent"`
	Highlight string `mapstructure:"highlight"`
	Text      string `mapstructure:"text"`
	Subtext   string `mapstructure:"subtext"`
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
				// Manually merge
				cfg.Projects = mergeProjects(cfg.Projects, hostCfg.Projects)
				cfg.Actions = MergeActions(cfg.Actions, hostCfg.Actions)
				cfg.Theme = mergeTheme(cfg.Theme, hostCfg.Theme)
			} else {
				return nil, fmt.Errorf("failed to unmarshal host config: %w", err)
			}
		} else if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read host config: %w", err)
		}
	}

	return &cfg, nil
}

// mergeTheme merges two themes. Host values override global.
func mergeTheme(global, host Theme) Theme {
	if host.Primary != "" {
		global.Primary = host.Primary
	}
	if host.Accent != "" {
		global.Accent = host.Accent
	}
	if host.Highlight != "" {
		global.Highlight = host.Highlight
	}
	if host.Text != "" {
		global.Text = host.Text
	}
	if host.Subtext != "" {
		global.Subtext = host.Subtext
	}
	return global
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

// MergeActions merges two action slices. specific actions (e.g. project actions)
// come first and override global actions by name. Matching is case-insensitive.
func MergeActions(global, specific []Action) []Action {
	actionMap := make(map[string]bool)
	merged := make([]Action, 0, len(global)+len(specific))

	// Add all specific actions first
	for _, a := range specific {
		merged = append(merged, a)
		actionMap[utils.Sanitize(a.Name)] = true
	}

	// Add global actions that don't overlap
	for _, a := range global {
		if !actionMap[utils.Sanitize(a.Name)] {
			merged = append(merged, a)
		}
	}

	return merged
}

// UseDefaultActions returns true if the project should use default actions.
func (p Project) UseDefaultActions() bool {
	if p.DefaultActions == nil {
		return true
	}
	return *p.DefaultActions
}

// GetShellDefault returns the shell-default setting for the project.
// If not set, it inherits from the root setting.
func (p Project) GetShellDefault(rootDefault bool) bool {
	if p.ShellDefault == nil {
		return rootDefault
	}
	return *p.ShellDefault
}

// GetShellDefault returns the root shell-default setting.
func (c *Config) GetShellDefault() bool {
	if c.ShellDefault == nil {
		return false
	}
	return *c.ShellDefault
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
