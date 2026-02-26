package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	DBPath string
}

// EnvFilePath returns the platform-specific config file path.
func EnvFilePath() string {
	home, _ := os.UserHomeDir()
	switch runtime.GOOS {
	case "darwin":
		return filepath.Join(home, ".dtasks", ".env")
	default: // linux
		xdgConfig := os.Getenv("XDG_CONFIG_HOME")
		if xdgConfig == "" {
			xdgConfig = filepath.Join(home, ".config")
		}
		return filepath.Join(xdgConfig, "dtasks", ".env")
	}
}

// DefaultDBPath returns the platform-specific default database path.
func DefaultDBPath() string {
	home, _ := os.UserHomeDir()
	switch runtime.GOOS {
	case "darwin":
		return filepath.Join(home, "Library", "Application Support", "dtasks", "tasks.db")
	default: // linux
		xdgData := os.Getenv("XDG_DATA_HOME")
		if xdgData == "" {
			xdgData = filepath.Join(home, ".local", "share")
		}
		return filepath.Join(xdgData, "dtasks", "tasks.db")
	}
}

// Load reads the env file and returns config. Runs wizard if not found.
func Load() (*Config, error) {
	envFile := EnvFilePath()

	if _, err := os.Stat(envFile); os.IsNotExist(err) {
		return runWizard(envFile)
	}

	if err := godotenv.Load(envFile); err != nil {
		return nil, fmt.Errorf("error loading %s: %w", envFile, err)
	}

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		return nil, fmt.Errorf("DB_PATH not set in %s", envFile)
	}

	return &Config{DBPath: dbPath}, nil
}

func runWizard(envFile string) (*Config, error) {
	fmt.Println("Welcome to dtasks! No configuration found.")
	fmt.Println()

	defaultDB := DefaultDBPath()
	fmt.Printf("Database path [%s]: ", defaultDB)

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	dbPath := defaultDB
	if input != "" {
		dbPath = input
	}

	// Expand ~ if present
	if strings.HasPrefix(dbPath, "~/") {
		home, _ := os.UserHomeDir()
		dbPath = filepath.Join(home, dbPath[2:])
	}

	// Create dirs
	if err := os.MkdirAll(filepath.Dir(envFile), 0755); err != nil {
		return nil, fmt.Errorf("cannot create config dir: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return nil, fmt.Errorf("cannot create db dir: %w", err)
	}

	// Write env file
	content := fmt.Sprintf("DB_PATH=%s\n", dbPath)
	if err := os.WriteFile(envFile, []byte(content), 0644); err != nil {
		return nil, fmt.Errorf("cannot write config: %w", err)
	}

	fmt.Printf("\nConfiguration saved to %s\n", envFile)
	fmt.Printf("Database will be created at %s\n\n", dbPath)

	return &Config{DBPath: dbPath}, nil
}
