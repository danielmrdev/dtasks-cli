package models

import "time"

type List struct {
	ID        int64
	Name      string
	CreatedAt time.Time
}

type Task struct {
	ID              int64
	ListID          int64
	ListName        string
	ParentTaskID    *int64
	Title           string
	Notes           *string
	DueDate         *string // YYYY-MM-DD
	DueTime         *string // HH:MM
	Completed       bool
	CompletedAt     *time.Time
	Recurring       bool
	RecurType       *string // daily | weekly | monthly
	RecurInterval   int     // every N
	RecurTime       *string // HH:MM
	RecurDayOfWeek  *int    // 0-6
	RecurDayOfMonth *int    // 1-31
	RecurStarts     *string // YYYY-MM-DD
	RecurEndsType   *string // never | on_date | after_n
	RecurEndsDate   *string // YYYY-MM-DD
	RecurEndsAfter  *int
	RecurCount      int
	Autocomplete    bool
	CreatedAt       time.Time
}
