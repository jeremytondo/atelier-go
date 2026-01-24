package config

import (
	"atelier-go/internal/utils"
	"os"
	"path/filepath"
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

func TestLoadConfig_MergeProjects(t *testing.T) {
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
projects:
  - name: global-prj
    path: /global
`
	if err := os.WriteFile(filepath.Join(atelierDir, "config.yaml"), []byte(globalContent), 0644); err != nil {
		t.Fatalf("failed to write global config: %v", err)
	}

	// Host config
	hostname, err := utils.GetHostname()
	if err != nil {
		t.Fatalf("failed to get hostname: %v", err)
	}
	hostContent := `
projects:
  - name: host-prj
    path: /host
`
	if err := os.WriteFile(filepath.Join(atelierDir, hostname+".yaml"), []byte(hostContent), 0644); err != nil {
		t.Fatalf("failed to write host config: %v", err)
	}

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// We need to decide if we WANT merging or overriding for slices.
	// Usually Viper overrides slices.
	if len(cfg.Projects) != 2 {
		t.Errorf("expected 2 projects, got %d", len(cfg.Projects))
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
	if len(global.Actions) != 2 {
		t.Errorf("expected 2 actions, got %d", len(global.Actions))
	}
	// Our MergeActions implementation puts specific (host) actions first
	if global.Actions[0].Command != "c1-host" {
		t.Errorf("expected a1 command c1-host, got %s", global.Actions[0].Command)
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
