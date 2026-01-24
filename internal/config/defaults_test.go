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
		{"theme.primary", "#89b4fa"},
		{"theme.accent", "#74c7ec"},
		{"theme.highlight", "#cba6f7"},
		{"theme.text", "#ffffff"},
		{"theme.subtext", "240"},
	}

	for _, tt := range themeTests {
		if v.GetString(tt.key) != tt.expected {
			t.Errorf("expected default %s to be %s, got %s", tt.key, tt.expected, v.GetString(tt.key))
		}
	}
}
