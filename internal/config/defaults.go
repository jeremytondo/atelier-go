package config

import (
	"github.com/spf13/viper"
)

// SetDefaults sets the default values for the configuration using Viper.
func SetDefaults(v *viper.Viper) {
	v.SetDefault("editor", "vim")
	v.SetDefault("shell-default", false)

	// Theme defaults
	v.SetDefault("theme.primary", "#58a6ff")
	v.SetDefault("theme.accent", "#1f6feb")
	v.SetDefault("theme.highlight", "#238636")
	v.SetDefault("theme.text", "#c9d1d9")
	v.SetDefault("theme.subtext", "#8b949e")
}
