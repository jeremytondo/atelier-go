package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Create a temporary config directory
	tmpDir, err := os.MkdirTemp("", "atelier-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	t.Cleanup(func() {
		_ = os.RemoveAll(tmpDir)
	})

	// Create a dummy config.yaml
	configContent := `
editor: nano
theme:
  primary: "#ff0000"
`
	err = os.WriteFile(filepath.Join(tmpDir, "config.yaml"), []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	// Mock GetConfigDir to return our temp dir
	// We need to refactor GetConfigDir or use an environment variable to override it.
	t.Setenv("XDG_CONFIG_HOME", filepath.Dir(tmpDir))
	// The current GetConfigDir appends "atelier-go" to XDG_CONFIG_HOME
	// So if tmpDir is /tmp/atelier-test-123, we want GetConfigDir to return it.
	// Let's create the 'atelier-go' subdir inside tmpDir.
	atelierDir := filepath.Join(tmpDir, "atelier-go")
	if err := os.MkdirAll(atelierDir, 0755); err != nil {
		t.Fatalf("failed to create atelier dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(atelierDir, "config.yaml"), []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write atelier config: %v", err)
	}
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if cfg.Editor != "nano" {
		t.Errorf("expected editor nano, got %s", cfg.Editor)
	}

	if cfg.Theme.Primary != "#ff0000" {
		t.Errorf("expected primary #ff0000, got %s", cfg.Theme.Primary)
	}
}

func TestLoadConfig_NoFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "atelier-test-none-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	t.Cleanup(func() {
		_ = os.RemoveAll(tmpDir)
	})
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// Should use defaults from defaults.go
	if cfg.Editor != "vim" {
		t.Errorf("expected default editor vim, got %s", cfg.Editor)
	}
	// Verify corrected theme defaults
	if cfg.Theme.Primary != "#89b4fa" {
		t.Errorf("expected default theme.primary #89b4fa, got %s", cfg.Theme.Primary)
	}
}

func TestLoadConfig_LocalOverride(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "atelier-test-merge-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	t.Cleanup(func() {
		_ = os.RemoveAll(tmpDir)
	})
	t.Setenv("XDG_CONFIG_HOME", tmpDir)
	atelierDir := filepath.Join(tmpDir, "atelier-go")
	if err := os.MkdirAll(atelierDir, 0755); err != nil {
		t.Fatalf("failed to create atelier dir: %v", err)
	}

	// Global config
	globalContent := `
editor: vim
projects:
  - name: global-prj
    path: /global
actions:
  - name: a1
    command: global
`
	if err := os.WriteFile(filepath.Join(atelierDir, "config.yaml"), []byte(globalContent), 0644); err != nil {
		t.Fatalf("failed to write global config: %v", err)
	}

	// Local override config
	localContent := `
editor: nano
projects:
  - name: global-prj
    path: /local
  - name: local-prj
    path: /local
actions:
  - name: a1
    command: local
  - name: a2
    command: extra
`
	if err := os.WriteFile(filepath.Join(atelierDir, "config.local.yaml"), []byte(localContent), 0644); err != nil {
		t.Fatalf("failed to write local config: %v", err)
	}

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if cfg.Editor != "nano" {
		t.Errorf("expected editor nano, got %s", cfg.Editor)
	}

	if len(cfg.Projects) != 2 {
		t.Errorf("expected 2 projects, got %d", len(cfg.Projects))
	}
	if cfg.Projects[0].Path != "/local" {
		t.Errorf("expected global-prj overridden to /local, got %s", cfg.Projects[0].Path)
	}
	if cfg.Projects[1].Name != "local-prj" {
		t.Errorf("expected local-prj appended, got %s", cfg.Projects[1].Name)
	}
	if len(cfg.Actions) != 2 {
		t.Errorf("expected 2 actions, got %d", len(cfg.Actions))
	}
	if cfg.Actions[0].Command != "local" {
		t.Errorf("expected a1 overridden to local, got %s", cfg.Actions[0].Command)
	}
	if cfg.Actions[1].Name != "a2" {
		t.Errorf("expected a2 appended, got %s", cfg.Actions[1].Name)
	}
}

func TestConfig_Merge(t *testing.T) {
	global := Config{
		Editor: "vim",
		Projects: []Project{
			{Name: "p1", Path: "/p1"},
			{Name: "p2", Path: "/p2"},
		},
		Actions: []Action{
			{Name: "a0", Command: "c0"},
			{Name: "a1", Command: "c1"},
		},
		Theme: Theme{
			Primary: "red",
			Accent:  "blue",
		},
	}

	host := Config{
		Editor: "nano", // Override
		Projects: []Project{
			{Name: "p1", Path: "/p1-host"}, // Override
			{Name: "p3", Path: "/p3"},      // Append
		},
		Actions: []Action{
			{Name: "a1", Command: "c1-host"}, // Override
			{Name: "a2", Command: "c2"},      // Append
		},
		Theme: Theme{
			Primary: "green", // Override
		},
	}

	global.Merge(host)

	// Check Editor
	if global.Editor != "nano" {
		t.Errorf("expected editor nano, got %s", global.Editor)
	}

	// Check Projects
	if len(global.Projects) != 3 {
		t.Errorf("expected 3 projects, got %d", len(global.Projects))
	}
	// Verify p1 was overridden (order might vary depending on impl, but usually p1 stays at index 0)
	// Our mergeProjects implementation preserves order for existing and appends new.
	if global.Projects[0].Path != "/p1-host" {
		t.Errorf("expected p1 path /p1-host, got %s", global.Projects[0].Path)
	}

	// Check Actions
	if len(global.Actions) != 3 {
		t.Errorf("expected 3 actions, got %d", len(global.Actions))
	}
	// Global actions are preserved in order, but overridden by host values.
	if global.Actions[0].Name != "a0" || global.Actions[0].Command != "c0" {
		t.Errorf("expected a0 at index 0, got %v", global.Actions[0])
	}
	if global.Actions[1].Name != "a1" || global.Actions[1].Command != "c1-host" {
		t.Errorf("expected a1 overridden at index 1, got %v", global.Actions[1])
	}
	if global.Actions[2].Name != "a2" || global.Actions[2].Command != "c2" {
		t.Errorf("expected a2 appended at index 2, got %v", global.Actions[2])
	}

	// Check Theme
	if global.Theme.Primary != "green" {
		t.Errorf("expected theme primary green, got %s", global.Theme.Primary)
	}
	if global.Theme.Accent != "blue" {
		t.Errorf("expected theme accent blue, got %s", global.Theme.Accent)
	}
}

func TestLoadConfig_Invalid(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "atelier-test-invalid-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	t.Cleanup(func() {
		_ = os.RemoveAll(tmpDir)
	})
	t.Setenv("XDG_CONFIG_HOME", tmpDir)
	atelierDir := filepath.Join(tmpDir, "atelier-go")
	if err := os.MkdirAll(atelierDir, 0755); err != nil {
		t.Fatalf("failed to create atelier dir: %v", err)
	}

	// Invalid config (missing path)
	content := `
projects:
  - name: invalid-prj
    path: ""
`
	if err := os.WriteFile(filepath.Join(atelierDir, "config.yaml"), []byte(content), 0644); err != nil {
		t.Fatalf("failed to write invalid config: %v", err)
	}

	_, err = LoadConfig()
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
}

func TestLoadConfig_LocalOnly(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "atelier-test-local-only-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	t.Cleanup(func() {
		_ = os.RemoveAll(tmpDir)
	})
	t.Setenv("XDG_CONFIG_HOME", tmpDir)
	atelierDir := filepath.Join(tmpDir, "atelier-go")
	if err := os.MkdirAll(atelierDir, 0755); err != nil {
		t.Fatalf("failed to create atelier dir: %v", err)
	}

	localContent := `
editor: micro
projects:
  - name: local-prj
    path: /local
`
	if err := os.WriteFile(filepath.Join(atelierDir, "config.local.yaml"), []byte(localContent), 0644); err != nil {
		t.Fatalf("failed to write local config: %v", err)
	}

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if cfg.Editor != "micro" {
		t.Errorf("expected editor micro, got %s", cfg.Editor)
	}
	if len(cfg.Projects) != 1 {
		t.Errorf("expected 1 project, got %d", len(cfg.Projects))
	}
}

func TestLoadConfig_InvalidLocal(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "atelier-test-local-invalid-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	t.Cleanup(func() {
		_ = os.RemoveAll(tmpDir)
	})
	t.Setenv("XDG_CONFIG_HOME", tmpDir)
	atelierDir := filepath.Join(tmpDir, "atelier-go")
	if err := os.MkdirAll(atelierDir, 0755); err != nil {
		t.Fatalf("failed to create atelier dir: %v", err)
	}

	globalContent := `
editor: vim
`
	if err := os.WriteFile(filepath.Join(atelierDir, "config.yaml"), []byte(globalContent), 0644); err != nil {
		t.Fatalf("failed to write global config: %v", err)
	}

	invalidContent := "projects: ["
	if err := os.WriteFile(filepath.Join(atelierDir, "config.local.yaml"), []byte(invalidContent), 0644); err != nil {
		t.Fatalf("failed to write local config: %v", err)
	}

	_, err = LoadConfig()
	if err == nil {
		t.Fatal("expected local config error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to read local config") {
		t.Fatalf("expected local config read error, got %v", err)
	}
}
