package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/danielmrdev/dtasks-cli/internal/config"
)

func TestDefaultDBPath_Linux(t *testing.T) {
	// On Linux, default path should be under ~/.local/share or $XDG_DATA_HOME
	path := config.DefaultDBPath()
	if path == "" {
		t.Error("DefaultDBPath() returned empty string")
	}
	if !filepath.IsAbs(path) {
		t.Errorf("expected absolute path, got %q", path)
	}
}

func TestEnvFilePath_Linux(t *testing.T) {
	path := config.EnvFilePath()
	if path == "" {
		t.Error("EnvFilePath() returned empty string")
	}
	if !filepath.IsAbs(path) {
		t.Errorf("expected absolute path, got %q", path)
	}
}

func TestLoad_FromEnvFile(t *testing.T) {
	// Create a temp dir to act as home/config dir
	dir, err := os.MkdirTemp("", "dtasks-config-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	dbPath := filepath.Join(dir, "tasks.db")
	content := "DB_PATH=" + dbPath + "\n"

	// Override HOME so EnvFilePath() resolves to our temp dir on macOS
	// (darwin path: ~/.dtasks/.env → <dir>/.dtasks/.env)
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", dir)
	defer os.Setenv("HOME", oldHome)

	dotDtasksDir := filepath.Join(dir, ".dtasks")
	if err := os.MkdirAll(dotDtasksDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dotDtasksDir, ".env"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	// Override XDG_CONFIG_HOME for Linux
	// (linux path: $XDG_CONFIG_HOME/dtasks/.env → <dir>/dtasks/.env)
	oldXDG := os.Getenv("XDG_CONFIG_HOME")
	os.Setenv("XDG_CONFIG_HOME", dir)
	defer os.Setenv("XDG_CONFIG_HOME", oldXDG)

	xdgDtasksDir := filepath.Join(dir, "dtasks")
	if err := os.MkdirAll(xdgDtasksDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(xdgDtasksDir, ".env"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.DBPath != dbPath {
		t.Errorf("expected DBPath=%q, got %q", dbPath, cfg.DBPath)
	}
}
