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
	defer os.RemoveAll(tmpDir)

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
	os.Setenv("XDG_CONFIG_HOME", filepath.Dir(tmpDir))
    // The current GetConfigDir appends "atelier-go" to XDG_CONFIG_HOME
    // So if tmpDir is /tmp/atelier-test-123, we want GetConfigDir to return it.
    // Let's create the 'atelier-go' subdir inside tmpDir.
    atelierDir := filepath.Join(tmpDir, "atelier-go")
    os.MkdirAll(atelierDir, 0755)
    os.WriteFile(filepath.Join(atelierDir, "config.yaml"), []byte(configContent), 0644)
    os.Setenv("XDG_CONFIG_HOME", tmpDir)

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
	tmpDir, _ := os.MkdirTemp("", "atelier-test-none-*")
	defer os.RemoveAll(tmpDir)
	os.Setenv("XDG_CONFIG_HOME", tmpDir)

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
	tmpDir, _ := os.MkdirTemp("", "atelier-test-merge-*")
	defer os.RemoveAll(tmpDir)
	os.Setenv("XDG_CONFIG_HOME", tmpDir)
	atelierDir := filepath.Join(tmpDir, "atelier-go")
	os.MkdirAll(atelierDir, 0755)

	// Global config
	globalContent := `
projects:
  - name: global-prj
    path: /global
`
	os.WriteFile(filepath.Join(atelierDir, "config.yaml"), []byte(globalContent), 0644)

	// Host config
	hostname, _ := utils.GetHostname()
	hostContent := `
projects:
  - name: host-prj
    path: /host
`
	os.WriteFile(filepath.Join(atelierDir, hostname+".yaml"), []byte(hostContent), 0644)

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

func TestLoadConfig_Invalid(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "atelier-test-invalid-*")
	defer os.RemoveAll(tmpDir)
	os.Setenv("XDG_CONFIG_HOME", tmpDir)
	atelierDir := filepath.Join(tmpDir, "atelier-go")
	os.MkdirAll(atelierDir, 0755)

	// Invalid config (missing path)
	content := `
projects:
  - name: invalid-prj
    path: ""
`
	os.WriteFile(filepath.Join(atelierDir, "config.yaml"), []byte(content), 0644)

	_, err := LoadConfig()
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
}
