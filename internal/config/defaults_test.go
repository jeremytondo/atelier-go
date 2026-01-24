package config

import (
	"testing"

	"github.com/spf13/viper"
)

func TestSetDefaults(t *testing.T) {
	v := viper.New()
	SetDefaults(v)

	if v.GetString("editor") != "vim" {
		t.Errorf("expected default editor to be vim, got %s", v.GetString("editor"))
	}

	if v.GetBool("shell-default") != false {
		t.Errorf("expected default shell-default to be false, got %v", v.GetBool("shell-default"))
	}

	themeTests := []struct {
		key      string
		expected string
	}{
		{"theme.primary", "#58a6ff"},
		{"theme.accent", "#1f6feb"},
		{"theme.highlight", "#238636"},
		{"theme.text", "#c9d1d9"},
		{"theme.subtext", "#8b949e"},
	}

	for _, tt := range themeTests {
		if v.GetString(tt.key) != tt.expected {
			t.Errorf("expected default %s to be %s, got %s", tt.key, tt.expected, v.GetString(tt.key))
		}
	}
}
