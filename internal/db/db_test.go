package db_test

import (
	"os"
	"testing"

	"github.com/danielmrdev/dtasks-cli/internal/db"
)

func TestOpen(t *testing.T) {
	f, err := os.CreateTemp("", "dtasks-test-*.db")
	if err != nil {
		t.Fatal(err)
	}
	f.Close()
	defer os.Remove(f.Name())

	d, err := db.Open(f.Name())
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer d.Close()

	if err := d.Ping(); err != nil {
		t.Errorf("Ping() error = %v", err)
	}
}

func TestOpen_CreatesDir(t *testing.T) {
	dir, err := os.MkdirTemp("", "dtasks-test-dir-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	dbPath := dir + "/subdir/tasks.db"
	d, err := db.Open(dbPath)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer d.Close()

	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Errorf("expected db file to be created at %s", dbPath)
	}
}

func TestOpen_SchemaMigrated(t *testing.T) {
	f, err := os.CreateTemp("", "dtasks-test-*.db")
	if err != nil {
		t.Fatal(err)
	}
	f.Close()
	defer os.Remove(f.Name())

	d, err := db.Open(f.Name())
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer d.Close()

	// Verify tables exist
	for _, table := range []string{"lists", "tasks"} {
		var name string
		row := d.QueryRow(`SELECT name FROM sqlite_master WHERE type='table' AND name=?`, table)
		if err := row.Scan(&name); err != nil {
			t.Errorf("table %q not found: %v", table, err)
		}
	}
}
