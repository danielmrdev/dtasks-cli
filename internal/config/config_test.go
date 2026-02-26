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
	// Create a temp dir to act as config dir
	dir, err := os.MkdirTemp("", "dtasks-config-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	envFile := filepath.Join(dir, ".env")
	dbPath := filepath.Join(dir, "tasks.db")
	content := "DB_PATH=" + dbPath + "\n"
	if err := os.WriteFile(envFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	// Override XDG_CONFIG_HOME so Load() reads our temp file
	oldXDG := os.Getenv("XDG_CONFIG_HOME")
	// EnvFilePath() on Linux: $XDG_CONFIG_HOME/dtasks/.env
	// We need to place our .env at the right relative location.
	// Create dtasks subdir inside our temp config dir.
	dtasksDir := filepath.Join(dir, "dtasks")
	if err := os.MkdirAll(dtasksDir, 0755); err != nil {
		t.Fatal(err)
	}
	envFile2 := filepath.Join(dtasksDir, ".env")
	if err := os.WriteFile(envFile2, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	os.Setenv("XDG_CONFIG_HOME", dir)
	defer os.Setenv("XDG_CONFIG_HOME", oldXDG)

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.DBPath != dbPath {
		t.Errorf("expected DBPath=%q, got %q", dbPath, cfg.DBPath)
	}
}
