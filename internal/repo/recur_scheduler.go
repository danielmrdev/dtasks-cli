package repo

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/danielmrdev/dtasks-cli/internal/models"
)

// calcNextDate returns the next due date based on task recurrence, using base as start.
// If base is empty, today is used.
func calcNextDate(base string, t *models.Task) (string, error) {
	if base == "" {
		base = time.Now().Format("2006-01-02")
	}
	start, err := time.Parse("2006-01-02", base)
	if err != nil {
		return "", fmt.Errorf("parse base date %q: %w", base, err)
	}

	interval := t.RecurInterval
	if interval <= 0 {
		interval = 1
	}

	switch *t.RecurType {
	case "daily":
		return start.AddDate(0, 0, interval).Format("2006-01-02"), nil
	case "weekly":
		return start.AddDate(0, 0, 7*interval).Format("2006-01-02"), nil
	case "monthly":
		// Use day 1 of target month to avoid overflow (e.g. Jan 31 + 1 month)
		targetMonth := time.Date(start.Year(), start.Month()+time.Month(interval), 1, 0, 0, 0, 0, time.UTC)
		targetDay := start.Day()
		if t.RecurDayOfMonth != nil {
			targetDay = *t.RecurDayOfMonth
		}
		// Clamp to last day of target month
		lastDay := time.Date(targetMonth.Year(), targetMonth.Month()+1, 0, 0, 0, 0, 0, time.UTC).Day()
		if targetDay > lastDay {
			targetDay = lastDay
		}
		return time.Date(targetMonth.Year(), targetMonth.Month(), targetDay, 0, 0, 0, 0, time.UTC).Format("2006-01-02"), nil
	default:
		return "", fmt.Errorf("unknown recur_type: %q", *t.RecurType)
	}
}

// shouldSpawn reports whether a new occurrence should be created given the next date.
func shouldSpawn(t *models.Task, nextDate string) bool {
	if t.RecurEndsType == nil || *t.RecurEndsType == "never" {
		return true
	}
	switch *t.RecurEndsType {
	case "on_date":
		return t.RecurEndsDate != nil && nextDate <= *t.RecurEndsDate
	case "after_n":
		return t.RecurEndsAfter != nil && t.RecurCount < *t.RecurEndsAfter
	default:
		return true
	}
}

// TaskScheduleNext creates the next occurrence of a recurring task after it is completed.
// Returns nil, nil if the task is not recurring or no more occurrences should be created.
func TaskScheduleNext(db *sql.DB, id int64) (*models.Task, error) {
	t, err := TaskGet(db, id)
	if err != nil {
		return nil, err
	}
	if !t.Recurring || t.RecurType == nil {
		return nil, nil
	}

	base := ""
	if t.DueDate != nil {
		base = *t.DueDate
	}
	nextDate, err := calcNextDate(base, t)
	if err != nil {
		return nil, err
	}

	if !shouldSpawn(t, nextDate) {
		return nil, nil
	}

	tx, err := db.Begin()
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	autocompleteVal := boolToInt(t.Autocomplete)
	res, err := tx.Exec(`
		INSERT INTO tasks (list_id, title, notes, due_date, due_time, autocomplete)
		VALUES (?, ?, ?, ?, ?, ?)`,
		t.ListID, t.Title, t.Notes, nextDate, t.DueTime, autocompleteVal,
	)
	if err != nil {
		return nil, fmt.Errorf("insert next occurrence: %w", err)
	}
	newID, _ := res.LastInsertId()

	newCount := t.RecurCount + 1
	_, err = tx.Exec(`
		UPDATE tasks SET
			recurring = 1,
			recur_type = ?, recur_interval = ?,
			recur_day_of_week = ?, recur_day_of_month = ?,
			recur_starts = ?, recur_ends_type = ?,
			recur_ends_date = ?, recur_ends_after = ?,
			recur_count = ?
		WHERE id = ?`,
		t.RecurType, t.RecurInterval,
		t.RecurDayOfWeek, t.RecurDayOfMonth,
		t.RecurStarts, t.RecurEndsType,
		t.RecurEndsDate, t.RecurEndsAfter,
		newCount, newID,
	)
	if err != nil {
		return nil, fmt.Errorf("set recur on next occurrence: %w", err)
	}

	// Reflect the spawned count back on the original task so its display is accurate.
	if _, err = tx.Exec(`UPDATE tasks SET recur_count = ? WHERE id = ?`, newCount, id); err != nil {
		return nil, fmt.Errorf("update recur_count on original task: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	return TaskGet(db, newID)
}

// ProcessAutocompleteTasks marks overdue autocomplete tasks as done and spawns next occurrences.
// Errors on individual tasks are logged to stderr but do not abort processing.
func ProcessAutocompleteTasks(db *sql.DB) error {
	now := time.Now()
	today := now.Format("2006-01-02")
	currentTime := now.Format("15:04")
	rows, err := db.Query(
		`SELECT id FROM tasks WHERE autocomplete = 1 AND completed = 0
		 AND (
		   due_date < ?
		   OR (due_date = ? AND (due_time IS NULL OR due_time <= ?))
		 )`,
		today, today, currentTime,
	)
	if err != nil {
		return fmt.Errorf("query autocomplete tasks: %w", err)
	}
	defer rows.Close()

	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return err
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	for _, id := range ids {
		if err := TaskDone(db, id, true); err != nil {
			fmt.Fprintf(os.Stderr, "warning: autocomplete task %d done: %v\n", id, err)
			continue
		}
		if _, err := TaskScheduleNext(db, id); err != nil {
			fmt.Fprintf(os.Stderr, "warning: schedule next for task %d: %v\n", id, err)
		}
	}
	return nil
}
