package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	_ "modernc.org/sqlite"
)

func Open(dbPath string) (*sql.DB, error) {
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return nil, fmt.Errorf("cannot create db directory: %w", err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("cannot open database: %w", err)
	}

	// Single writer to avoid WAL conflicts; reads are concurrent
	db.SetMaxOpenConns(1)

	if err := configure(db); err != nil {
		return nil, err
	}
	if err := migrate(db); err != nil {
		return nil, err
	}

	return db, nil
}

func configure(db *sql.DB) error {
	pragmas := []string{
		"PRAGMA journal_mode=WAL",
		"PRAGMA busy_timeout=5000",
		"PRAGMA foreign_keys=ON",
	}
	for _, p := range pragmas {
		if _, err := db.Exec(p); err != nil {
			return fmt.Errorf("pragma %q failed: %w", p, err)
		}
	}
	return nil
}

func migrate(db *sql.DB) error {
	_, err := db.Exec(`
	CREATE TABLE IF NOT EXISTS lists (
		id         INTEGER PRIMARY KEY AUTOINCREMENT,
		name       TEXT NOT NULL UNIQUE,
		color      TEXT,
		created_at DATETIME NOT NULL DEFAULT (datetime('now', 'localtime'))
	);

	CREATE TABLE IF NOT EXISTS tasks (
		id                 INTEGER PRIMARY KEY AUTOINCREMENT,
		list_id            INTEGER NOT NULL REFERENCES lists(id) ON DELETE CASCADE,
		parent_task_id     INTEGER REFERENCES tasks(id) ON DELETE CASCADE,
		title              TEXT NOT NULL,
		notes              TEXT,
		due_date           TEXT,          -- YYYY-MM-DD
		due_time           TEXT,          -- HH:MM (null = all-day)
		completed          INTEGER NOT NULL DEFAULT 0,
		completed_at       DATETIME,
		recurring          INTEGER NOT NULL DEFAULT 0,
		recur_type         TEXT,          -- daily | weekly | monthly
		recur_interval     INTEGER NOT NULL DEFAULT 1,
		recur_day_of_week  INTEGER,       -- 0-6
		recur_day_of_month INTEGER,       -- 1-31
		recur_starts       TEXT,          -- YYYY-MM-DD
		recur_ends_type    TEXT,          -- never | on_date | after_n
		recur_ends_date    TEXT,          -- YYYY-MM-DD
		recur_ends_after   INTEGER,
		recur_count        INTEGER NOT NULL DEFAULT 0,
		autocomplete       INTEGER NOT NULL DEFAULT 0,
		created_at         DATETIME NOT NULL DEFAULT (datetime('now', 'localtime'))
	);

	CREATE INDEX IF NOT EXISTS idx_tasks_list    ON tasks(list_id);
	CREATE INDEX IF NOT EXISTS idx_tasks_parent  ON tasks(parent_task_id);
	CREATE INDEX IF NOT EXISTS idx_tasks_due     ON tasks(due_date);
	`)
	if err != nil {
		return err
	}

	// Idempotent: add color column to existing DBs that predate this migration.
	if _, alterErr := db.Exec(`ALTER TABLE lists ADD COLUMN color TEXT`); alterErr != nil {
		if !strings.Contains(alterErr.Error(), "duplicate column name") {
			return fmt.Errorf("add lists.color: %w", alterErr)
		}
	}

	// Add autocomplete column if missing (existing DBs)
	var count int
	db.QueryRow(`SELECT COUNT(*) FROM pragma_table_info('tasks') WHERE name='autocomplete'`).Scan(&count)
	if count == 0 {
		if _, err := db.Exec(`ALTER TABLE tasks ADD COLUMN autocomplete INTEGER NOT NULL DEFAULT 0`); err != nil {
			return fmt.Errorf("migrate autocomplete column: %w", err)
		}
	}

	// Add priority column if missing (existing DBs)
	var pCount int
	db.QueryRow(`SELECT COUNT(*) FROM pragma_table_info('tasks') WHERE name='priority'`).Scan(&pCount)
	if pCount == 0 {
		if _, err := db.Exec(`ALTER TABLE tasks ADD COLUMN priority TEXT`); err != nil {
			return fmt.Errorf("migrate priority column: %w", err)
		}
	}

	// Drop recur_time column if still present (removed in favour of inheriting due_time)
	var rcCount int
	db.QueryRow(`SELECT COUNT(*) FROM pragma_table_info('tasks') WHERE name='recur_time'`).Scan(&rcCount)
	if rcCount > 0 {
		if _, err := db.Exec(`ALTER TABLE tasks DROP COLUMN recur_time`); err != nil {
			return fmt.Errorf("migrate drop column recur_time: %w", err)
		}
	}

	// Drop legacy date/time columns if still present
	for _, col := range []string{"date", "time"} {
		var n int
		db.QueryRow(`SELECT COUNT(*) FROM pragma_table_info('tasks') WHERE name=?`, col).Scan(&n)
		if n > 0 {
			if _, err := db.Exec(`ALTER TABLE tasks DROP COLUMN ` + col); err != nil {
				return fmt.Errorf("migrate drop column %s: %w", col, err)
			}
		}
	}

	return nil
}
