package models

import "time"

type List struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type Task struct {
	ID              int64      `json:"id"`
	ListID          int64      `json:"list_id"`
	ListName        string     `json:"list_name"`
	ParentTaskID    *int64     `json:"parent_task_id,omitempty"`
	Title           string     `json:"title"`
	Notes           *string    `json:"notes,omitempty"`
	DueDate         *string    `json:"due_date,omitempty"`
	DueTime         *string    `json:"due_time,omitempty"`
	Completed       bool       `json:"completed"`
	CompletedAt     *time.Time `json:"completed_at,omitempty"`
	Recurring       bool       `json:"recurring"`
	RecurType       *string    `json:"recur_type,omitempty"`
	RecurInterval   int        `json:"recur_interval,omitempty"`
	RecurTime       *string    `json:"recur_time,omitempty"`
	RecurDayOfWeek  *int       `json:"recur_day_of_week,omitempty"`
	RecurDayOfMonth *int       `json:"recur_day_of_month,omitempty"`
	RecurStarts     *string    `json:"recur_starts,omitempty"`
	RecurEndsType   *string    `json:"recur_ends_type,omitempty"`
	RecurEndsDate   *string    `json:"recur_ends_date,omitempty"`
	RecurEndsAfter  *int       `json:"recur_ends_after,omitempty"`
	RecurCount      int        `json:"recur_count,omitempty"`
	Autocomplete    bool       `json:"autocomplete"`
	CreatedAt       time.Time  `json:"created_at"`
}
