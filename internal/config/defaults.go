package config

import (
	"github.com/spf13/viper"
)

// SetDefaults sets the default values for the configuration using Viper.
func SetDefaults(v *viper.Viper) {
	v.SetDefault("editor", "vim")
	v.SetDefault("shell-default", false)

	// Theme defaults
	v.SetDefault("theme.primary", "#89b4fa")
	v.SetDefault("theme.accent", "#74c7ec")
	v.SetDefault("theme.highlight", "#cba6f7")
	v.SetDefault("theme.text", "#ffffff")
	v.SetDefault("theme.subtext", "240")
}
