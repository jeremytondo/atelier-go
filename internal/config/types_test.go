package config

import (
	"testing"

	"github.com/go-viper/mapstructure/v2"
)

func TestConfig_Mapstructure(t *testing.T) {
	// This test verifies that the struct tags (mapstructure:"...") match 
	// the keys used in configuration files.
	data := map[string]any{
		"editor":         "vscode",
		"shell-default":  true,
		"theme": map[string]any{
			"primary": "#000000",
		},
		"projects": []map[string]any{
			{
				"name":            "prj1",
				"path":            "/home/user/prj1",
				"default-actions": false,
				"shell-default":   true,
			},
		},
		"actions": []map[string]any{
			{
				"name":    "build",
				"command": "go build",
			},
		},
	}

	var cfg Config
	err := mapstructure.Decode(data, &cfg)
	if err != nil {
		t.Fatalf("failed to decode: %v", err)
	}

	if cfg.Editor != "vscode" {
		t.Errorf("expected Editor to be vscode, got %s", cfg.Editor)
	}
	if cfg.ShellDefault == nil || *cfg.ShellDefault != true {
		t.Errorf("expected root ShellDefault to be true")
	}
	if cfg.Theme.Primary != "#000000" {
		t.Errorf("expected Theme.Primary to be #000000, got %s", cfg.Theme.Primary)
	}
	
	if len(cfg.Projects) != 1 {
		t.Fatalf("expected 1 project, got %d", len(cfg.Projects))
	}
	p := cfg.Projects[0]
	if p.Name != "prj1" || p.Path != "/home/user/prj1" {
		t.Errorf("project field mismatch: %+v", p)
	}
	if p.DefaultActions == nil || *p.DefaultActions != false {
		t.Errorf("expected project DefaultActions to be false")
	}
	if p.ShellDefault == nil || *p.ShellDefault != true {
		t.Errorf("expected project ShellDefault to be true")
	}

	if len(cfg.Actions) != 1 || cfg.Actions[0].Name != "build" {
		t.Errorf("expected 1 action 'build', got %v", cfg.Actions)
	}
}
