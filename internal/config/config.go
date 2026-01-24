// Package config provides the application configuration structures and loading logic.
package config

import (
	"atelier-go/internal/utils"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// LoadConfig loads the configuration from the config directory.
// It loads config.yaml first, then merges <hostname>.yaml if it exists.
func LoadConfig() (*Config, error) {
	v := viper.New()
	SetDefaults(v)

	configDir, err := GetConfigDir()
	if err != nil {
		return nil, err
	}

	v.SetConfigType("yaml")
	v.AddConfigPath(configDir)
	v.SetConfigName("config")

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read global config: %w", err)
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
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
				cfg.Merge(hostCfg)
			} else {
				return nil, fmt.Errorf("failed to unmarshal host config: %w", err)
			}
		} else if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read host config: %w", err)
		}
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &cfg, nil
}

// Merge merges another configuration into the current one.
// The other configuration takes precedence for scalar values and specific slice items.
func (c *Config) Merge(other Config) {
	c.Projects = mergeProjects(c.Projects, other.Projects)
	c.Actions = MergeActions(c.Actions, other.Actions)
	c.Theme = mergeTheme(c.Theme, other.Theme)

	if other.Editor != "" {
		c.Editor = other.Editor
	}
	if other.ShellDefault != nil {
		c.ShellDefault = other.ShellDefault
	}
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

// MergeActions merges two action slices. Global actions are preserved in order,
// but overridden by specific actions if names match (case-insensitive).
// Strictly new specific actions are appended to the end.
func MergeActions(global, specific []Action) []Action {
	specificMap := make(map[string]Action)
	for _, a := range specific {
		specificMap[utils.Sanitize(a.Name)] = a
	}

	merged := make([]Action, 0, len(global)+len(specific))
	processed := make(map[string]bool)

	// Process global actions, overriding if exists in specific
	for _, a := range global {
		key := utils.Sanitize(a.Name)
		if sa, ok := specificMap[key]; ok {
			merged = append(merged, sa)
			processed[key] = true
		} else {
			merged = append(merged, a)
		}
	}

	// Append strictly new actions from specific
	for _, a := range specific {
		key := utils.Sanitize(a.Name)
		if !processed[key] {
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
