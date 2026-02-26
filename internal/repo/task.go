package repo

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/danielmrdev/dtasks-cli/internal/models"
)

// TaskCreateInput holds fields for creating/updating a task.
type TaskInput struct {
	ListID       int64
	ParentTaskID *int64
	Title        string
	Notes        *string
	DueDate      *string
	DueTime      *string
	Autocomplete bool
}

func TaskCreate(db *sql.DB, in TaskInput) (*models.Task, error) {
	res, err := db.Exec(`
		INSERT INTO tasks (list_id, parent_task_id, title, notes, due_date, due_time, autocomplete)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		in.ListID, in.ParentTaskID, in.Title, in.Notes,
		in.DueDate, in.DueTime, boolToInt(in.Autocomplete),
	)
	if err != nil {
		return nil, fmt.Errorf("create task: %w", err)
	}
	id, _ := res.LastInsertId()
	return TaskGet(db, id)
}

func TaskGet(db *sql.DB, id int64) (*models.Task, error) {
	row := db.QueryRow(taskSelectSQL+` WHERE t.id = ?`, id)
	return scanTask(row)
}

// TaskList returns tasks filtered by options.
type TaskListOptions struct {
	ListID    *int64
	ParentID  *int64
	OnlyRoot  bool // no subtasks
	Completed *bool
	DueToday  bool
}

func TaskList(db *sql.DB, opts TaskListOptions) ([]models.Task, error) {
	query := taskSelectSQL + ` WHERE 1=1`
	args := []any{}

	if opts.ListID != nil {
		query += ` AND t.list_id = ?`
		args = append(args, *opts.ListID)
	}
	if opts.ParentID != nil {
		query += ` AND t.parent_task_id = ?`
		args = append(args, *opts.ParentID)
	}
	if opts.OnlyRoot {
		query += ` AND t.parent_task_id IS NULL`
	}
	if opts.Completed != nil {
		query += ` AND t.completed = ?`
		if *opts.Completed {
			args = append(args, 1)
		} else {
			args = append(args, 0)
		}
	}
	if opts.DueToday {
		today := time.Now().Format("2006-01-02")
		query += ` AND t.due_date <= ?`
		args = append(args, today)
	}

	query += ` ORDER BY t.due_date ASC, t.due_time ASC, t.created_at ASC`

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []models.Task
	for rows.Next() {
		t, err := scanTaskRow(rows)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, *t)
	}
	return tasks, rows.Err()
}

func TaskUpdate(db *sql.DB, id int64, in TaskInput) (*models.Task, error) {
	res, err := db.Exec(`
		UPDATE tasks SET
			list_id = ?, parent_task_id = ?, title = ?, notes = ?,
			due_date = ?, due_time = ?
		WHERE id = ?`,
		in.ListID, in.ParentTaskID, in.Title, in.Notes,
		in.DueDate, in.DueTime,
		id,
	)
	if err != nil {
		return nil, fmt.Errorf("update task: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return nil, fmt.Errorf("task %d not found", id)
	}
	return TaskGet(db, id)
}

// TaskPatch updates only non-nil fields.
type TaskPatch struct {
	Title        *string
	Notes        *string
	DueDate      *string
	DueTime      *string
	ListID       *int64
	Autocomplete *bool
}

func TaskPatchFields(db *sql.DB, id int64, p TaskPatch) (*models.Task, error) {
	// Build dynamic update
	set := ""
	args := []any{}
	add := func(col string, val any) {
		if set != "" {
			set += ", "
		}
		set += col + " = ?"
		args = append(args, val)
	}

	if p.Title != nil {
		add("title", *p.Title)
	}
	if p.Notes != nil {
		add("notes", *p.Notes)
	}
	if p.DueDate != nil {
		add("due_date", *p.DueDate)
	}
	if p.DueTime != nil {
		add("due_time", *p.DueTime)
	}
	if p.ListID != nil {
		add("list_id", *p.ListID)
	}
	if p.Autocomplete != nil {
		add("autocomplete", boolToInt(*p.Autocomplete))
	}

	if set == "" {
		return TaskGet(db, id)
	}

	args = append(args, id)
	res, err := db.Exec(`UPDATE tasks SET `+set+` WHERE id = ?`, args...)
	if err != nil {
		return nil, fmt.Errorf("patch task: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return nil, fmt.Errorf("task %d not found", id)
	}
	return TaskGet(db, id)
}

func TaskDone(db *sql.DB, id int64, done bool) error {
	var err error
	if done {
		_, err = db.Exec(`UPDATE tasks SET completed = 1, completed_at = datetime('now', 'localtime') WHERE id = ?`, id)
	} else {
		_, err = db.Exec(`UPDATE tasks SET completed = 0, completed_at = NULL WHERE id = ?`, id)
	}
	return err
}

func TaskDelete(db *sql.DB, id int64) error {
	res, err := db.Exec(`DELETE FROM tasks WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete task: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("task %d not found", id)
	}
	return nil
}

// --- Recurrence ---

type RecurInput struct {
	Type       string // daily | weekly | monthly
	Interval   int
	DayOfWeek  *int
	DayOfMonth *int
	Starts     *string
	EndsType   string // never | on_date | after_n
	EndsDate   *string
	EndsAfter  *int
	Count      *int // nil → reset to 0; non-nil → explicit value
}

func TaskSetRecur(db *sql.DB, id int64, r RecurInput) error {
	count := 0
	if r.Count != nil {
		count = *r.Count
	}
	_, err := db.Exec(`
		UPDATE tasks SET
			recurring = 1,
			recur_type = ?, recur_interval = ?,
			recur_day_of_week = ?, recur_day_of_month = ?,
			recur_starts = ?, recur_ends_type = ?,
			recur_ends_date = ?, recur_ends_after = ?,
			recur_count = ?
		WHERE id = ?`,
		r.Type, r.Interval,
		r.DayOfWeek, r.DayOfMonth,
		r.Starts, r.EndsType,
		r.EndsDate, r.EndsAfter,
		count, id,
	)
	return err
}

func TaskRemoveRecur(db *sql.DB, id int64) error {
	_, err := db.Exec(`
		UPDATE tasks SET
			recurring = 0, recur_type = NULL, recur_interval = 1,
			recur_day_of_week = NULL, recur_day_of_month = NULL,
			recur_starts = NULL, recur_ends_type = NULL,
			recur_ends_date = NULL, recur_ends_after = NULL, recur_count = 0
		WHERE id = ?`, id)
	return err
}

// --- Helpers ---

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

const taskSelectSQL = `
SELECT
	t.id, t.list_id, l.name,
	t.parent_task_id, t.title, t.notes,
	t.due_date, t.due_time,
	t.completed, t.completed_at,
	t.recurring, t.recur_type, t.recur_interval,
	t.recur_day_of_week, t.recur_day_of_month,
	t.recur_starts, t.recur_ends_type, t.recur_ends_date, t.recur_ends_after,
	t.recur_count, t.autocomplete, t.created_at
FROM tasks t
JOIN lists l ON t.list_id = l.id
`

type scanner interface {
	Scan(dest ...any) error
}

func scanTask(s scanner) (*models.Task, error) {
	return scanTaskRow(s)
}

func scanTaskRow(s scanner) (*models.Task, error) {
	t := &models.Task{}
	var completedAt sql.NullString
	err := s.Scan(
		&t.ID, &t.ListID, &t.ListName,
		&t.ParentTaskID, &t.Title, &t.Notes,
		&t.DueDate, &t.DueTime,
		&t.Completed, &completedAt,
		&t.Recurring, &t.RecurType, &t.RecurInterval,
		&t.RecurDayOfWeek, &t.RecurDayOfMonth,
		&t.RecurStarts, &t.RecurEndsType, &t.RecurEndsDate, &t.RecurEndsAfter,
		&t.RecurCount, &t.Autocomplete, &t.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("scan task: %w", err)
	}
	if completedAt.Valid {
		parsed, _ := time.Parse("2006-01-02T15:04:05Z", completedAt.String)
		t.CompletedAt = &parsed
	}
	return t, nil
}
