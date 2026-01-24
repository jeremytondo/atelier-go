package config

import (
	"testing"

	"github.com/go-viper/mapstructure/v2"
)

func TestConfigStructs(t *testing.T) {
	t.Run("Action Struct", func(t *testing.T) {
		a := Action{Name: "test", Command: "echo test"}
		if a.Name != "test" {
			t.Errorf("expected Name to be test, got %s", a.Name)
		}
	})

	t.Run("Theme Struct", func(t *testing.T) {
		th := Theme{Primary: "#ffffff"}
		if th.Primary != "#ffffff" {
			t.Errorf("expected Primary to be #ffffff, got %s", th.Primary)
		}
	})

	t.Run("Project Struct", func(t *testing.T) {
		p := Project{Name: "test-project", Path: "/tmp"}
		if p.Name != "test-project" {
			t.Errorf("expected Name to be test-project, got %s", p.Name)
		}
	})

	t.Run("Config Struct", func(t *testing.T) {
		c := Config{Editor: "nvim"}
		if c.Editor != "nvim" {
			t.Errorf("expected Editor to be nvim, got %s", c.Editor)
		}
	})

	t.Run("Mapstructure Unmarshal", func(t *testing.T) {
		data := map[string]interface{}{
			"editor": "vscode",
			"theme": map[string]interface{}{
				"primary": "#000000",
			},
			"projects": []map[string]interface{}{
				{
					"name": "prj1",
					"path": "/home/user/prj1",
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
		if cfg.Theme.Primary != "#000000" {
			t.Errorf("expected Theme.Primary to be #000000, got %s", cfg.Theme.Primary)
		}
		if len(cfg.Projects) != 1 || cfg.Projects[0].Name != "prj1" {
			t.Errorf("expected 1 project named prj1, got %v", cfg.Projects)
		}
	})
}