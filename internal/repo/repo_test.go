package repo_test

import (
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/danielmrdev/dtasks-cli/internal/db"
	"github.com/danielmrdev/dtasks-cli/internal/repo"
)

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	f, err := os.CreateTemp("", "dtasks-repo-test-*.db")
	if err != nil {
		t.Fatal(err)
	}
	f.Close()
	name := f.Name()
	t.Cleanup(func() { os.Remove(name) })

	d, err := db.Open(name)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { d.Close() })
	return d
}

// --- List tests ---

func TestListCreate(t *testing.T) {
	d := openTestDB(t)

	l, err := repo.ListCreate(d, "Work", nil)
	if err != nil {
		t.Fatalf("ListCreate() error = %v", err)
	}
	if l.Name != "Work" {
		t.Errorf("expected Name=Work, got %q", l.Name)
	}
	if l.ID <= 0 {
		t.Errorf("expected positive ID, got %d", l.ID)
	}
}

func TestListCreate_Duplicate(t *testing.T) {
	d := openTestDB(t)

	if _, err := repo.ListCreate(d, "Personal", nil); err != nil {
		t.Fatal(err)
	}
	_, err := repo.ListCreate(d, "Personal", nil)
	if err == nil {
		t.Error("expected error for duplicate list name, got nil")
	}
}

func TestListAll(t *testing.T) {
	d := openTestDB(t)

	for _, name := range []string{"Alpha", "Beta", "Gamma"} {
		if _, err := repo.ListCreate(d, name, nil); err != nil {
			t.Fatal(err)
		}
	}

	lists, err := repo.ListAll(d)
	if err != nil {
		t.Fatalf("ListAll() error = %v", err)
	}
	if len(lists) != 3 {
		t.Errorf("expected 3 lists, got %d", len(lists))
	}
}

func TestListGet(t *testing.T) {
	d := openTestDB(t)

	created, err := repo.ListCreate(d, "MyList", nil)
	if err != nil {
		t.Fatal(err)
	}

	got, err := repo.ListGet(d, created.ID)
	if err != nil {
		t.Fatalf("ListGet() error = %v", err)
	}
	if got.Name != "MyList" {
		t.Errorf("expected Name=MyList, got %q", got.Name)
	}
}

func TestListEdit(t *testing.T) {
	d := openTestDB(t)

	color := "#ff0000"
	l, err := repo.ListCreate(d, "OldName", &color)
	if err != nil {
		t.Fatal(err)
	}

	// rename
	newName := "NewName"
	got, err := repo.ListPatchFields(d, l.ID, repo.ListPatch{Name: &newName})
	if err != nil {
		t.Fatalf("ListPatchFields() error = %v", err)
	}
	if got.Name != "NewName" {
		t.Errorf("expected Name=NewName, got %q", got.Name)
	}
	if got.Color == nil || *got.Color != "#ff0000" {
		t.Errorf("expected Color unchanged, got %v", got.Color)
	}

	// change color
	newColor := "#00ff00"
	got, err = repo.ListPatchFields(d, l.ID, repo.ListPatch{Color: &newColor})
	if err != nil {
		t.Fatalf("ListPatchFields() error = %v", err)
	}
	if got.Color == nil || *got.Color != "#00ff00" {
		t.Errorf("expected Color=#00ff00, got %v", got.Color)
	}

	// clear color
	empty := ""
	got, err = repo.ListPatchFields(d, l.ID, repo.ListPatch{Color: &empty})
	if err != nil {
		t.Fatalf("ListPatchFields() error = %v", err)
	}
	if got.Color != nil {
		t.Errorf("expected Color=nil after clear, got %v", got.Color)
	}
}

func TestListDelete(t *testing.T) {
	d := openTestDB(t)

	l, err := repo.ListCreate(d, "ToDelete", nil)
	if err != nil {
		t.Fatal(err)
	}

	if err := repo.ListDelete(d, l.ID); err != nil {
		t.Fatalf("ListDelete() error = %v", err)
	}

	lists, err := repo.ListAll(d)
	if err != nil {
		t.Fatal(err)
	}
	if len(lists) != 0 {
		t.Errorf("expected 0 lists after deletion, got %d", len(lists))
	}
}

func TestListEdit_NotFound(t *testing.T) {
	d := openTestDB(t)

	name := "ghost"
	if _, err := repo.ListPatchFields(d, 9999, repo.ListPatch{Name: &name}); err == nil {
		t.Error("expected error for non-existent list, got nil")
	}
}

func TestListDelete_NotFound(t *testing.T) {
	d := openTestDB(t)

	if err := repo.ListDelete(d, 9999); err == nil {
		t.Error("expected error for non-existent list, got nil")
	}
}

// --- Task tests ---

func TestTaskCreate(t *testing.T) {
	d := openTestDB(t)

	l, err := repo.ListCreate(d, "Work", nil)
	if err != nil {
		t.Fatal(err)
	}

	task, err := repo.TaskCreate(d, repo.TaskInput{
		ListID: l.ID,
		Title:  "Write report",
	})
	if err != nil {
		t.Fatalf("TaskCreate() error = %v", err)
	}
	if task.Title != "Write report" {
		t.Errorf("expected Title=%q, got %q", "Write report", task.Title)
	}
	if task.ListID != l.ID {
		t.Errorf("expected ListID=%d, got %d", l.ID, task.ListID)
	}
	if task.Completed {
		t.Error("expected task to be pending, got completed")
	}
}

func TestTaskCreate_WithDue(t *testing.T) {
	d := openTestDB(t)

	l, _ := repo.ListCreate(d, "Test", nil)
	due := "2026-03-01"
	dueTime := "10:00"

	task, err := repo.TaskCreate(d, repo.TaskInput{
		ListID:  l.ID,
		Title:   "With due",
		DueDate: &due,
		DueTime: &dueTime,
	})
	if err != nil {
		t.Fatalf("TaskCreate() error = %v", err)
	}
	if task.DueDate == nil || *task.DueDate != due {
		t.Errorf("expected DueDate=%q, got %v", due, task.DueDate)
	}
	if task.DueTime == nil || *task.DueTime != dueTime {
		t.Errorf("expected DueTime=%q, got %v", dueTime, task.DueTime)
	}
}

func TestTaskGet(t *testing.T) {
	d := openTestDB(t)

	l, _ := repo.ListCreate(d, "Test", nil)
	created, _ := repo.TaskCreate(d, repo.TaskInput{ListID: l.ID, Title: "Fetch me"})

	got, err := repo.TaskGet(d, created.ID)
	if err != nil {
		t.Fatalf("TaskGet() error = %v", err)
	}
	if got.Title != "Fetch me" {
		t.Errorf("expected Title=%q, got %q", "Fetch me", got.Title)
	}
}

func TestTaskList(t *testing.T) {
	d := openTestDB(t)

	l, _ := repo.ListCreate(d, "Test", nil)
	for _, title := range []string{"Task A", "Task B", "Task C"} {
		repo.TaskCreate(d, repo.TaskInput{ListID: l.ID, Title: title})
	}

	false_ := false
	tasks, err := repo.TaskList(d, repo.TaskListOptions{Completed: &false_})
	if err != nil {
		t.Fatalf("TaskList() error = %v", err)
	}
	if len(tasks) != 3 {
		t.Errorf("expected 3 tasks, got %d", len(tasks))
	}
}

func TestTaskList_FilterByList(t *testing.T) {
	d := openTestDB(t)

	l1, _ := repo.ListCreate(d, "List1", nil)
	l2, _ := repo.ListCreate(d, "List2", nil)
	repo.TaskCreate(d, repo.TaskInput{ListID: l1.ID, Title: "In L1"})
	repo.TaskCreate(d, repo.TaskInput{ListID: l2.ID, Title: "In L2"})

	tasks, err := repo.TaskList(d, repo.TaskListOptions{ListID: &l1.ID})
	if err != nil {
		t.Fatal(err)
	}
	if len(tasks) != 1 {
		t.Errorf("expected 1 task for list1, got %d", len(tasks))
	}
	if tasks[0].Title != "In L1" {
		t.Errorf("expected Title=%q, got %q", "In L1", tasks[0].Title)
	}
}

func TestTaskDone(t *testing.T) {
	d := openTestDB(t)

	l, _ := repo.ListCreate(d, "Test", nil)
	task, _ := repo.TaskCreate(d, repo.TaskInput{ListID: l.ID, Title: "Do it"})

	if err := repo.TaskDone(d, task.ID, true); err != nil {
		t.Fatalf("TaskDone() error = %v", err)
	}

	got, _ := repo.TaskGet(d, task.ID)
	if !got.Completed {
		t.Error("expected task to be completed")
	}

	if err := repo.TaskDone(d, task.ID, false); err != nil {
		t.Fatalf("TaskDone(false) error = %v", err)
	}
	got, _ = repo.TaskGet(d, task.ID)
	if got.Completed {
		t.Error("expected task to be pending after undone")
	}
}

func TestTaskDelete(t *testing.T) {
	d := openTestDB(t)

	l, _ := repo.ListCreate(d, "Test", nil)
	task, _ := repo.TaskCreate(d, repo.TaskInput{ListID: l.ID, Title: "Delete me"})

	if err := repo.TaskDelete(d, task.ID); err != nil {
		t.Fatalf("TaskDelete() error = %v", err)
	}

	_, err := repo.TaskGet(d, task.ID)
	if err == nil {
		t.Error("expected error after deletion, got nil")
	}
}

func TestTaskDelete_NotFound(t *testing.T) {
	d := openTestDB(t)

	if err := repo.TaskDelete(d, 9999); err == nil {
		t.Error("expected error for non-existent task, got nil")
	}
}

func TestTaskPatchFields(t *testing.T) {
	d := openTestDB(t)

	l, _ := repo.ListCreate(d, "Test", nil)
	task, _ := repo.TaskCreate(d, repo.TaskInput{ListID: l.ID, Title: "Original"})

	newTitle := "Updated"
	got, err := repo.TaskPatchFields(d, task.ID, repo.TaskPatch{Title: &newTitle})
	if err != nil {
		t.Fatalf("TaskPatchFields() error = %v", err)
	}
	if got.Title != "Updated" {
		t.Errorf("expected Title=Updated, got %q", got.Title)
	}
}

func TestTaskSetRecur(t *testing.T) {
	d := openTestDB(t)

	l, _ := repo.ListCreate(d, "Test", nil)
	task, _ := repo.TaskCreate(d, repo.TaskInput{ListID: l.ID, Title: "Recurring task"})

	err := repo.TaskSetRecur(d, task.ID, repo.RecurInput{
		Type:     "daily",
		Interval: 1,
		EndsType: "never",
	})
	if err != nil {
		t.Fatalf("TaskSetRecur() error = %v", err)
	}

	got, _ := repo.TaskGet(d, task.ID)
	if !got.Recurring {
		t.Error("expected task to be recurring")
	}
	if got.RecurType == nil || *got.RecurType != "daily" {
		t.Errorf("expected RecurType=daily, got %v", got.RecurType)
	}
}

func TestTaskRemoveRecur(t *testing.T) {
	d := openTestDB(t)

	l, _ := repo.ListCreate(d, "Test", nil)
	task, _ := repo.TaskCreate(d, repo.TaskInput{ListID: l.ID, Title: "Was recurring"})
	repo.TaskSetRecur(d, task.ID, repo.RecurInput{Type: "daily", Interval: 1, EndsType: "never"})

	if err := repo.TaskRemoveRecur(d, task.ID); err != nil {
		t.Fatalf("TaskRemoveRecur() error = %v", err)
	}

	got, _ := repo.TaskGet(d, task.ID)
	if got.Recurring {
		t.Error("expected task to not be recurring")
	}
}

func TestTaskCreate_Subtask(t *testing.T) {
	d := openTestDB(t)

	l, _ := repo.ListCreate(d, "Test", nil)
	parent, _ := repo.TaskCreate(d, repo.TaskInput{ListID: l.ID, Title: "Parent"})

	child, err := repo.TaskCreate(d, repo.TaskInput{
		ListID:       l.ID,
		ParentTaskID: &parent.ID,
		Title:        "Child task",
	})
	if err != nil {
		t.Fatalf("TaskCreate subtask error = %v", err)
	}
	if child.ParentTaskID == nil || *child.ParentTaskID != parent.ID {
		t.Errorf("expected ParentTaskID=%d, got %v", parent.ID, child.ParentTaskID)
	}
}

func TestListDelete_CascadesTasks(t *testing.T) {
	d := openTestDB(t)

	l, _ := repo.ListCreate(d, "ToDelete", nil)
	repo.TaskCreate(d, repo.TaskInput{ListID: l.ID, Title: "Task in list"})

	if err := repo.ListDelete(d, l.ID); err != nil {
		t.Fatal(err)
	}

	false_ := false
	tasks, err := repo.TaskList(d, repo.TaskListOptions{Completed: &false_})
	if err != nil {
		t.Fatal(err)
	}
	if len(tasks) != 0 {
		t.Errorf("expected 0 tasks after list deletion, got %d", len(tasks))
	}
}

// --- Scheduler tests ---

func TestScheduler_Daily(t *testing.T) {
	d := openTestDB(t)
	l, _ := repo.ListCreate(d, "Test", nil)
	due := "2026-02-26"
	task, _ := repo.TaskCreate(d, repo.TaskInput{ListID: l.ID, Title: "Daily task", DueDate: &due})
	repo.TaskSetRecur(d, task.ID, repo.RecurInput{Type: "daily", Interval: 1, EndsType: "never"})
	repo.TaskDone(d, task.ID, true)

	next, err := repo.TaskScheduleNext(d, task.ID)
	if err != nil {
		t.Fatalf("TaskScheduleNext() error = %v", err)
	}
	if next == nil {
		t.Fatal("expected next occurrence, got nil")
	}
	if next.DueDate == nil || *next.DueDate != "2026-02-27" {
		t.Errorf("expected DueDate=2026-02-27, got %v", next.DueDate)
	}
}

func TestScheduler_Weekly(t *testing.T) {
	d := openTestDB(t)
	l, _ := repo.ListCreate(d, "Test", nil)
	due := "2026-02-26"
	task, _ := repo.TaskCreate(d, repo.TaskInput{ListID: l.ID, Title: "Weekly task", DueDate: &due})
	repo.TaskSetRecur(d, task.ID, repo.RecurInput{Type: "weekly", Interval: 2, EndsType: "never"})

	next, err := repo.TaskScheduleNext(d, task.ID)
	if err != nil {
		t.Fatalf("TaskScheduleNext() error = %v", err)
	}
	if next == nil {
		t.Fatal("expected next occurrence, got nil")
	}
	if next.DueDate == nil || *next.DueDate != "2026-03-12" {
		t.Errorf("expected DueDate=2026-03-12, got %v", next.DueDate)
	}
}

func TestScheduler_Monthly(t *testing.T) {
	d := openTestDB(t)
	l, _ := repo.ListCreate(d, "Test", nil)
	due := "2026-01-15"
	task, _ := repo.TaskCreate(d, repo.TaskInput{ListID: l.ID, Title: "Monthly task", DueDate: &due})
	repo.TaskSetRecur(d, task.ID, repo.RecurInput{Type: "monthly", Interval: 1, EndsType: "never"})

	next, err := repo.TaskScheduleNext(d, task.ID)
	if err != nil {
		t.Fatalf("TaskScheduleNext() error = %v", err)
	}
	if next == nil {
		t.Fatal("expected next occurrence, got nil")
	}
	if next.DueDate == nil || *next.DueDate != "2026-02-15" {
		t.Errorf("expected DueDate=2026-02-15, got %v", next.DueDate)
	}
}

func TestScheduler_Monthly_DayClamp(t *testing.T) {
	d := openTestDB(t)
	l, _ := repo.ListCreate(d, "Test", nil)
	due := "2026-01-31"
	day := 31
	task, _ := repo.TaskCreate(d, repo.TaskInput{ListID: l.ID, Title: "End of month", DueDate: &due})
	repo.TaskSetRecur(d, task.ID, repo.RecurInput{
		Type: "monthly", Interval: 1, EndsType: "never", DayOfMonth: &day,
	})

	next, err := repo.TaskScheduleNext(d, task.ID)
	if err != nil {
		t.Fatalf("TaskScheduleNext() error = %v", err)
	}
	if next == nil {
		t.Fatal("expected next occurrence, got nil")
	}
	if next.DueDate == nil || *next.DueDate != "2026-02-28" {
		t.Errorf("expected DueDate=2026-02-28, got %v", next.DueDate)
	}
}

func TestScheduler_EndsAfterN_Creates(t *testing.T) {
	d := openTestDB(t)
	l, _ := repo.ListCreate(d, "Test", nil)
	due := "2026-02-26"
	endsAfter := 3
	count := 2
	task, _ := repo.TaskCreate(d, repo.TaskInput{ListID: l.ID, Title: "Limited", DueDate: &due})
	repo.TaskSetRecur(d, task.ID, repo.RecurInput{
		Type: "daily", Interval: 1, EndsType: "after_n", EndsAfter: &endsAfter, Count: &count,
	})

	next, err := repo.TaskScheduleNext(d, task.ID)
	if err != nil {
		t.Fatalf("TaskScheduleNext() error = %v", err)
	}
	if next == nil {
		t.Fatal("expected next occurrence when count < endsAfter")
	}
	if next.RecurCount != 3 {
		t.Errorf("expected RecurCount=3, got %d", next.RecurCount)
	}
}

func TestScheduler_EndsAfterN_Stops(t *testing.T) {
	d := openTestDB(t)
	l, _ := repo.ListCreate(d, "Test", nil)
	due := "2026-02-26"
	endsAfter := 3
	count := 3
	task, _ := repo.TaskCreate(d, repo.TaskInput{ListID: l.ID, Title: "Done", DueDate: &due})
	repo.TaskSetRecur(d, task.ID, repo.RecurInput{
		Type: "daily", Interval: 1, EndsType: "after_n", EndsAfter: &endsAfter, Count: &count,
	})

	next, err := repo.TaskScheduleNext(d, task.ID)
	if err != nil {
		t.Fatalf("TaskScheduleNext() error = %v", err)
	}
	if next != nil {
		t.Errorf("expected nil when count >= endsAfter, got task %d", next.ID)
	}
}

func TestScheduler_EndsOnDate_Creates(t *testing.T) {
	d := openTestDB(t)
	l, _ := repo.ListCreate(d, "Test", nil)
	due := "2026-03-05"
	endsDate := "2026-03-15"
	task, _ := repo.TaskCreate(d, repo.TaskInput{ListID: l.ID, Title: "On date", DueDate: &due})
	repo.TaskSetRecur(d, task.ID, repo.RecurInput{
		Type: "daily", Interval: 5, EndsType: "on_date", EndsDate: &endsDate,
	})

	next, err := repo.TaskScheduleNext(d, task.ID)
	if err != nil {
		t.Fatalf("TaskScheduleNext() error = %v", err)
	}
	if next == nil {
		t.Fatal("expected next occurrence when nextDate <= endsDate")
	}
}

func TestScheduler_EndsOnDate_Stops(t *testing.T) {
	d := openTestDB(t)
	l, _ := repo.ListCreate(d, "Test", nil)
	due := "2026-03-01"
	endsDate := "2026-03-01"
	task, _ := repo.TaskCreate(d, repo.TaskInput{ListID: l.ID, Title: "Expired", DueDate: &due})
	repo.TaskSetRecur(d, task.ID, repo.RecurInput{
		Type: "daily", Interval: 5, EndsType: "on_date", EndsDate: &endsDate,
	})

	next, err := repo.TaskScheduleNext(d, task.ID)
	if err != nil {
		t.Fatalf("TaskScheduleNext() error = %v", err)
	}
	if next != nil {
		t.Errorf("expected nil when nextDate > endsDate, got task %d", next.ID)
	}
}

func TestScheduler_NonRecurring(t *testing.T) {
	d := openTestDB(t)
	l, _ := repo.ListCreate(d, "Test", nil)
	task, _ := repo.TaskCreate(d, repo.TaskInput{ListID: l.ID, Title: "One shot"})

	next, err := repo.TaskScheduleNext(d, task.ID)
	if err != nil {
		t.Fatalf("unexpected error = %v", err)
	}
	if next != nil {
		t.Errorf("expected nil for non-recurring task, got %v", next)
	}
}

func TestScheduler_InheritsFields(t *testing.T) {
	d := openTestDB(t)
	l, _ := repo.ListCreate(d, "Test", nil)
	due := "2026-02-26"
	dueTime := "09:00"
	notes := "important notes"
	task, _ := repo.TaskCreate(d, repo.TaskInput{ListID: l.ID, Title: "With notes", DueDate: &due, DueTime: &dueTime, Notes: &notes})
	repo.TaskSetRecur(d, task.ID, repo.RecurInput{Type: "daily", Interval: 1, EndsType: "never"})

	next, err := repo.TaskScheduleNext(d, task.ID)
	if err != nil {
		t.Fatalf("TaskScheduleNext() error = %v", err)
	}
	if next == nil {
		t.Fatal("expected next occurrence")
	}
	if next.Notes == nil || *next.Notes != notes {
		t.Errorf("expected Notes=%q, got %v", notes, next.Notes)
	}
	if next.DueTime == nil || *next.DueTime != dueTime {
		t.Errorf("expected DueTime=%q, got %v", dueTime, next.DueTime)
	}
}

func TestScheduler_NilDueDate(t *testing.T) {
	d := openTestDB(t)
	l, _ := repo.ListCreate(d, "Test", nil)
	task, _ := repo.TaskCreate(d, repo.TaskInput{ListID: l.ID, Title: "No due"})
	repo.TaskSetRecur(d, task.ID, repo.RecurInput{Type: "daily", Interval: 1, EndsType: "never"})

	next, err := repo.TaskScheduleNext(d, task.ID)
	if err != nil {
		t.Fatalf("unexpected error with nil due date: %v", err)
	}
	if next == nil {
		t.Fatal("expected next occurrence")
	}
	// Should have a due date (today + 1)
	if next.DueDate == nil {
		t.Error("expected DueDate to be set")
	}
}

// --- Autocomplete tests ---

func TestAutocomplete_MarksAsDone(t *testing.T) {
	d := openTestDB(t)
	l, _ := repo.ListCreate(d, "Test", nil)
	yesterday := "2026-02-25"
	task, _ := repo.TaskCreate(d, repo.TaskInput{
		ListID:       l.ID,
		Title:        "Overdue autocomplete",
		DueDate:      &yesterday,
		Autocomplete: true,
	})

	if err := repo.ProcessAutocompleteTasks(d); err != nil {
		t.Fatalf("ProcessAutocompleteTasks() error = %v", err)
	}

	got, _ := repo.TaskGet(d, task.ID)
	if !got.Completed {
		t.Error("expected task to be completed by autocomplete")
	}
}

func TestAutocomplete_NotYetDue(t *testing.T) {
	d := openTestDB(t)
	l, _ := repo.ListCreate(d, "Test", nil)
	tomorrow := time.Now().AddDate(0, 0, 1).Format("2006-01-02")
	task, _ := repo.TaskCreate(d, repo.TaskInput{
		ListID:       l.ID,
		Title:        "Future task",
		DueDate:      &tomorrow,
		Autocomplete: true,
	})

	if err := repo.ProcessAutocompleteTasks(d); err != nil {
		t.Fatalf("ProcessAutocompleteTasks() error = %v", err)
	}

	got, _ := repo.TaskGet(d, task.ID)
	if got.Completed {
		t.Error("expected task NOT to be completed (not yet due)")
	}
}

func TestAutocomplete_RecurringChain(t *testing.T) {
	d := openTestDB(t)
	l, _ := repo.ListCreate(d, "Test", nil)
	yesterday := "2026-02-25"
	task, _ := repo.TaskCreate(d, repo.TaskInput{
		ListID:       l.ID,
		Title:        "Weekly reminder",
		DueDate:      &yesterday,
		Autocomplete: true,
	})
	repo.TaskSetRecur(d, task.ID, repo.RecurInput{Type: "weekly", Interval: 1, EndsType: "never"})

	if err := repo.ProcessAutocompleteTasks(d); err != nil {
		t.Fatalf("ProcessAutocompleteTasks() error = %v", err)
	}

	got, _ := repo.TaskGet(d, task.ID)
	if !got.Completed {
		t.Error("expected original task to be completed")
	}

	false_ := false
	all, err := repo.TaskList(d, repo.TaskListOptions{Completed: &false_})
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 1 {
		t.Errorf("expected 1 pending (next occurrence), got %d", len(all))
	}
	if all[0].Autocomplete != true {
		t.Error("expected next occurrence to inherit autocomplete=true")
	}
}

func TestRecurCount_UpdatedOnOriginalAfterSpawn(t *testing.T) {
	d := openTestDB(t)
	l, _ := repo.ListCreate(d, "Test", nil)
	yesterday := "2026-02-25"
	endsAfter := 2
	task, _ := repo.TaskCreate(d, repo.TaskInput{
		ListID:  l.ID,
		Title:   "Counter test",
		DueDate: &yesterday,
	})
	repo.TaskSetRecur(d, task.ID, repo.RecurInput{Type: "daily", Interval: 1, EndsType: "after_n", EndsAfter: &endsAfter})

	next, err := repo.TaskScheduleNext(d, task.ID)
	if err != nil {
		t.Fatalf("TaskScheduleNext() error = %v", err)
	}
	if next == nil {
		t.Fatal("expected a next occurrence to be created")
	}

	original, _ := repo.TaskGet(d, task.ID)
	if original.RecurCount != 1 {
		t.Errorf("expected original recur_count = 1, got %d", original.RecurCount)
	}
	if next.RecurCount != 1 {
		t.Errorf("expected next recur_count = 1, got %d", next.RecurCount)
	}
}

func TestAutocomplete_NonAutocomplete(t *testing.T) {
	d := openTestDB(t)
	l, _ := repo.ListCreate(d, "Test", nil)
	yesterday := "2026-02-25"
	task, _ := repo.TaskCreate(d, repo.TaskInput{
		ListID:  l.ID,
		Title:   "Manual task",
		DueDate: &yesterday,
		// Autocomplete: false (default)
	})

	if err := repo.ProcessAutocompleteTasks(d); err != nil {
		t.Fatalf("ProcessAutocompleteTasks() error = %v", err)
	}

	got, _ := repo.TaskGet(d, task.ID)
	if got.Completed {
		t.Error("expected non-autocomplete task NOT to be completed")
	}
}

func TestAutocomplete_DueTimeNotYetPassed(t *testing.T) {
	d := openTestDB(t)
	l, _ := repo.ListCreate(d, "Test", nil)
	today := time.Now().Format("2006-01-02")
	futureTime := time.Now().Add(2 * time.Hour).Format("15:04")
	task, _ := repo.TaskCreate(d, repo.TaskInput{
		ListID:       l.ID,
		Title:        "Today but future time",
		DueDate:      &today,
		DueTime:      &futureTime,
		Autocomplete: true,
	})

	if err := repo.ProcessAutocompleteTasks(d); err != nil {
		t.Fatalf("ProcessAutocompleteTasks() error = %v", err)
	}

	got, _ := repo.TaskGet(d, task.ID)
	if got.Completed {
		t.Error("expected task NOT to be completed (due_time not yet passed)")
	}
}

func TestAutocomplete_DueTimePassed(t *testing.T) {
	d := openTestDB(t)
	l, _ := repo.ListCreate(d, "Test", nil)
	today := time.Now().Format("2006-01-02")
	pastTime := time.Now().Add(-2 * time.Hour).Format("15:04")
	task, _ := repo.TaskCreate(d, repo.TaskInput{
		ListID:       l.ID,
		Title:        "Today but past time",
		DueDate:      &today,
		DueTime:      &pastTime,
		Autocomplete: true,
	})

	if err := repo.ProcessAutocompleteTasks(d); err != nil {
		t.Fatalf("ProcessAutocompleteTasks() error = %v", err)
	}

	got, _ := repo.TaskGet(d, task.ID)
	if !got.Completed {
		t.Error("expected task to be completed (due_time already passed)")
	}
}

// --- Phase 1 Querying tests ---

func TestTaskList_FilterToday(t *testing.T) {
	d := openTestDB(t)
	l, _ := repo.ListCreate(d, "Test", nil)

	today := time.Now().Format("2006-01-02")
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	tomorrow := time.Now().AddDate(0, 0, 1).Format("2006-01-02")

	repo.TaskCreate(d, repo.TaskInput{ListID: l.ID, Title: "due today", DueDate: &today})
	repo.TaskCreate(d, repo.TaskInput{ListID: l.ID, Title: "due yesterday (overdue)", DueDate: &yesterday})
	repo.TaskCreate(d, repo.TaskInput{ListID: l.ID, Title: "due tomorrow"})
	repo.TaskCreate(d, repo.TaskInput{ListID: l.ID, Title: "due tomorrow explicit", DueDate: &tomorrow})
	repo.TaskCreate(d, repo.TaskInput{ListID: l.ID, Title: "no due date"})

	tasks, err := repo.TaskList(d, repo.TaskListOptions{DueToday: true})
	if err != nil {
		t.Fatalf("TaskList() error = %v", err)
	}
	if len(tasks) != 2 {
		t.Errorf("expected 2 tasks (today + overdue), got %d", len(tasks))
	}
	for _, task := range tasks {
		if task.DueDate == nil || *task.DueDate > today {
			t.Errorf("unexpected task in FilterToday results: %q (due=%v)", task.Title, task.DueDate)
		}
	}
}

func TestTaskList_FilterOverdue(t *testing.T) {
	d := openTestDB(t)
	l, _ := repo.ListCreate(d, "Test", nil)

	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	today := time.Now().Format("2006-01-02")

	repo.TaskCreate(d, repo.TaskInput{ListID: l.ID, Title: "overdue", DueDate: &yesterday})
	repo.TaskCreate(d, repo.TaskInput{ListID: l.ID, Title: "due today — not overdue", DueDate: &today})
	repo.TaskCreate(d, repo.TaskInput{ListID: l.ID, Title: "no due date"})

	tasks, err := repo.TaskList(d, repo.TaskListOptions{Overdue: true})
	if err != nil {
		t.Fatalf("TaskList() error = %v", err)
	}
	if len(tasks) != 1 {
		t.Errorf("expected 1 overdue task, got %d", len(tasks))
	}
	if tasks[0].Title != "overdue" {
		t.Errorf("expected Title=%q, got %q", "overdue", tasks[0].Title)
	}
}

func TestTaskList_FilterTomorrow(t *testing.T) {
	d := openTestDB(t)
	l, _ := repo.ListCreate(d, "Test", nil)

	today := time.Now().Format("2006-01-02")
	tomorrow := time.Now().AddDate(0, 0, 1).Format("2006-01-02")
	dayAfter := time.Now().AddDate(0, 0, 2).Format("2006-01-02")

	repo.TaskCreate(d, repo.TaskInput{ListID: l.ID, Title: "tomorrow", DueDate: &tomorrow})
	repo.TaskCreate(d, repo.TaskInput{ListID: l.ID, Title: "today", DueDate: &today})
	repo.TaskCreate(d, repo.TaskInput{ListID: l.ID, Title: "day after tomorrow", DueDate: &dayAfter})

	tasks, err := repo.TaskList(d, repo.TaskListOptions{DueTomorrow: true})
	if err != nil {
		t.Fatalf("TaskList() error = %v", err)
	}
	if len(tasks) != 1 {
		t.Errorf("expected 1 task due tomorrow, got %d", len(tasks))
	}
	if tasks[0].Title != "tomorrow" {
		t.Errorf("expected Title=%q, got %q", "tomorrow", tasks[0].Title)
	}
}

func TestTaskList_FilterWeek(t *testing.T) {
	d := openTestDB(t)
	l, _ := repo.ListCreate(d, "Test", nil)

	today := time.Now().Format("2006-01-02")
	in6days := time.Now().AddDate(0, 0, 6).Format("2006-01-02")
	in7days := time.Now().AddDate(0, 0, 7).Format("2006-01-02")
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")

	repo.TaskCreate(d, repo.TaskInput{ListID: l.ID, Title: "today", DueDate: &today})
	repo.TaskCreate(d, repo.TaskInput{ListID: l.ID, Title: "in 6 days", DueDate: &in6days})
	repo.TaskCreate(d, repo.TaskInput{ListID: l.ID, Title: "in 7 days — out of range", DueDate: &in7days})
	repo.TaskCreate(d, repo.TaskInput{ListID: l.ID, Title: "yesterday — out of range", DueDate: &yesterday})

	tasks, err := repo.TaskList(d, repo.TaskListOptions{DueWeek: true})
	if err != nil {
		t.Fatalf("TaskList() error = %v", err)
	}
	if len(tasks) != 2 {
		t.Errorf("expected 2 tasks in week range [today, today+6], got %d", len(tasks))
	}
}

func TestTaskList_Sort(t *testing.T) {
	d := openTestDB(t)
	l, _ := repo.ListCreate(d, "Test", nil)

	d1 := "2026-03-10"
	d2 := "2026-03-05"
	d3 := "2026-03-20"

	t1, _ := repo.TaskCreate(d, repo.TaskInput{ListID: l.ID, Title: "Due March 10", DueDate: &d1})
	t2, _ := repo.TaskCreate(d, repo.TaskInput{ListID: l.ID, Title: "Due March 5", DueDate: &d2})
	t3, _ := repo.TaskCreate(d, repo.TaskInput{ListID: l.ID, Title: "Due March 20", DueDate: &d3})

	tasks, err := repo.TaskList(d, repo.TaskListOptions{SortBy: "due"})
	if err != nil {
		t.Fatalf("TaskList() error = %v", err)
	}
	if len(tasks) != 3 {
		t.Fatalf("expected 3 tasks, got %d", len(tasks))
	}
	if tasks[0].ID != t2.ID || tasks[1].ID != t1.ID || tasks[2].ID != t3.ID {
		t.Errorf("expected sort order [March5, March10, March20], got IDs [%d, %d, %d]",
			tasks[0].ID, tasks[1].ID, tasks[2].ID)
	}

	tasksByCreated, err := repo.TaskList(d, repo.TaskListOptions{SortBy: "created"})
	if err != nil {
		t.Fatalf("TaskList() error = %v", err)
	}
	if tasksByCreated[0].ID != t1.ID || tasksByCreated[1].ID != t2.ID || tasksByCreated[2].ID != t3.ID {
		t.Errorf("expected creation order [t1, t2, t3], got IDs [%d, %d, %d]",
			tasksByCreated[0].ID, tasksByCreated[1].ID, tasksByCreated[2].ID)
	}
}

func TestTaskList_SortReverse(t *testing.T) {
	d := openTestDB(t)
	l, _ := repo.ListCreate(d, "Test", nil)

	d1 := "2026-03-10"
	d2 := "2026-03-05"
	d3 := "2026-03-20"

	t1, _ := repo.TaskCreate(d, repo.TaskInput{ListID: l.ID, Title: "Due March 10", DueDate: &d1})
	t2, _ := repo.TaskCreate(d, repo.TaskInput{ListID: l.ID, Title: "Due March 5", DueDate: &d2})
	t3, _ := repo.TaskCreate(d, repo.TaskInput{ListID: l.ID, Title: "Due March 20", DueDate: &d3})

	tasks, err := repo.TaskList(d, repo.TaskListOptions{SortBy: "due", Reverse: true})
	if err != nil {
		t.Fatalf("TaskList() error = %v", err)
	}
	if len(tasks) != 3 {
		t.Fatalf("expected 3 tasks, got %d", len(tasks))
	}
	if tasks[0].ID != t3.ID || tasks[1].ID != t1.ID || tasks[2].ID != t2.ID {
		t.Errorf("expected reverse due order [March20, March10, March5], got IDs [%d, %d, %d]",
			tasks[0].ID, tasks[1].ID, tasks[2].ID)
	}
}

func TestTaskSearch_Keyword(t *testing.T) {
	d := openTestDB(t)
	l, _ := repo.ListCreate(d, "Test", nil)

	notes := "call the doctor"
	repo.TaskCreate(d, repo.TaskInput{ListID: l.ID, Title: "Buy groceries"})
	repo.TaskCreate(d, repo.TaskInput{ListID: l.ID, Title: "Gym session", Notes: &notes})
	repo.TaskCreate(d, repo.TaskInput{ListID: l.ID, Title: "gym"})

	tasks, err := repo.TaskSearch(d, repo.TaskSearchOptions{Keyword: "grocer"})
	if err != nil {
		t.Fatalf("TaskSearch() error = %v", err)
	}
	if len(tasks) != 1 {
		t.Errorf("expected 1 match for 'grocer', got %d", len(tasks))
	}
	if tasks[0].Title != "Buy groceries" {
		t.Errorf("expected Title=%q, got %q", "Buy groceries", tasks[0].Title)
	}

	tasks, err = repo.TaskSearch(d, repo.TaskSearchOptions{Keyword: "DOCTOR"})
	if err != nil {
		t.Fatalf("TaskSearch() error = %v", err)
	}
	if len(tasks) != 1 {
		t.Errorf("expected 1 match for 'DOCTOR' in notes, got %d", len(tasks))
	}

	tasks, err = repo.TaskSearch(d, repo.TaskSearchOptions{Keyword: "shopping"})
	if err != nil {
		t.Fatalf("TaskSearch() error = %v", err)
	}
	if len(tasks) != 0 {
		t.Errorf("expected 0 matches for 'shopping', got %d", len(tasks))
	}
}

func TestTaskSearch_List(t *testing.T) {
	d := openTestDB(t)
	lA, _ := repo.ListCreate(d, "ListA", nil)
	lB, _ := repo.ListCreate(d, "ListB", nil)

	repo.TaskCreate(d, repo.TaskInput{ListID: lA.ID, Title: "Task in A"})
	repo.TaskCreate(d, repo.TaskInput{ListID: lA.ID, Title: "Another in A"})
	repo.TaskCreate(d, repo.TaskInput{ListID: lB.ID, Title: "Task in B"})

	tasks, err := repo.TaskSearch(d, repo.TaskSearchOptions{Keyword: "Task", ListID: &lA.ID})
	if err != nil {
		t.Fatalf("TaskSearch() error = %v", err)
	}
	if len(tasks) != 1 {
		t.Errorf("expected 1 task in ListA matching 'Task', got %d", len(tasks))
	}
	if tasks[0].ListID != lA.ID {
		t.Errorf("expected ListID=%d, got %d", lA.ID, tasks[0].ListID)
	}

	allTasks, err := repo.TaskSearch(d, repo.TaskSearchOptions{Keyword: "Task"})
	if err != nil {
		t.Fatalf("TaskSearch() error = %v", err)
	}
	if len(allTasks) != 2 {
		t.Errorf("expected 2 tasks matching 'Task' across all lists, got %d", len(allTasks))
	}
}

func TestTaskSearch_Regex(t *testing.T) {
	d := openTestDB(t)
	l, _ := repo.ListCreate(d, "Test", nil)

	repo.TaskCreate(d, repo.TaskInput{ListID: l.ID, Title: "Buy groceries"})
	repo.TaskCreate(d, repo.TaskInput{ListID: l.ID, Title: "Buy milk"})
	repo.TaskCreate(d, repo.TaskInput{ListID: l.ID, Title: "Sell house"})

	tasks, err := repo.TaskSearch(d, repo.TaskSearchOptions{Keyword: "^Buy", Regex: true})
	if err != nil {
		t.Fatalf("TaskSearch() error = %v", err)
	}
	if len(tasks) != 2 {
		t.Errorf("expected 2 tasks matching '^Buy', got %d", len(tasks))
	}

	tasks, err = repo.TaskSearch(d, repo.TaskSearchOptions{Keyword: "[invalid", Regex: true})
	if err == nil {
		t.Error("expected error for invalid regex, got nil")
	}
	if tasks != nil {
		t.Errorf("expected nil tasks for invalid regex, got %v", tasks)
	}

	tasks, err = repo.TaskSearch(d, repo.TaskSearchOptions{Keyword: "(?i)groceri", Regex: true})
	if err != nil {
		t.Fatalf("TaskSearch() error = %v", err)
	}
	if len(tasks) != 1 {
		t.Errorf("expected 1 match for '(?i)groceri', got %d", len(tasks))
	}
}

// --- Phase 2 Richness tests (TDD red phase) ---

func strPtr(s string) *string { return &s }

// TestTaskCreate_WithPriority — PRIO-01
func TestTaskCreate_WithPriority(t *testing.T) {
	d := openTestDB(t)
	l, _ := repo.ListCreate(d, "Test", nil)

	task, err := repo.TaskCreate(d, repo.TaskInput{
		ListID:   l.ID,
		Title:    "High priority task",
		Priority: strPtr("high"),
	})
	if err != nil {
		t.Fatalf("TaskCreate() error = %v", err)
	}

	got, err := repo.TaskGet(d, task.ID)
	if err != nil {
		t.Fatalf("TaskGet() error = %v", err)
	}
	if got.Priority == nil || *got.Priority != "high" {
		t.Errorf("expected Priority=high, got %v", got.Priority)
	}

	// nil priority
	task2, err := repo.TaskCreate(d, repo.TaskInput{
		ListID: l.ID,
		Title:  "No priority task",
	})
	if err != nil {
		t.Fatalf("TaskCreate() error = %v", err)
	}
	got2, _ := repo.TaskGet(d, task2.ID)
	if got2.Priority != nil {
		t.Errorf("expected Priority=nil, got %v", got2.Priority)
	}
}

// TestTaskPatchFields_Priority — PRIO-02
func TestTaskPatchFields_Priority(t *testing.T) {
	d := openTestDB(t)
	l, _ := repo.ListCreate(d, "Test", nil)

	task, _ := repo.TaskCreate(d, repo.TaskInput{ListID: l.ID, Title: "Task"})

	got, err := repo.TaskPatchFields(d, task.ID, repo.TaskPatch{Priority: strPtr("medium")})
	if err != nil {
		t.Fatalf("TaskPatchFields() error = %v", err)
	}
	if got.Priority == nil || *got.Priority != "medium" {
		t.Errorf("expected Priority=medium, got %v", got.Priority)
	}

	// clear priority by passing empty string
	got2, err := repo.TaskPatchFields(d, task.ID, repo.TaskPatch{Priority: strPtr("")})
	if err != nil {
		t.Fatalf("TaskPatchFields() error = %v", err)
	}
	if got2.Priority != nil {
		t.Errorf("expected Priority=nil after clear, got %v", got2.Priority)
	}
}

// TestTaskList_SortPriority — PRIO-04
func TestTaskList_SortPriority(t *testing.T) {
	d := openTestDB(t)
	l, _ := repo.ListCreate(d, "Test", nil)

	tHigh, _ := repo.TaskCreate(d, repo.TaskInput{ListID: l.ID, Title: "High", Priority: strPtr("high")})
	tLow, _ := repo.TaskCreate(d, repo.TaskInput{ListID: l.ID, Title: "Low", Priority: strPtr("low")})
	tMed, _ := repo.TaskCreate(d, repo.TaskInput{ListID: l.ID, Title: "Medium", Priority: strPtr("medium")})

	tasks, err := repo.TaskList(d, repo.TaskListOptions{SortBy: "priority"})
	if err != nil {
		t.Fatalf("TaskList() error = %v", err)
	}
	if len(tasks) != 3 {
		t.Fatalf("expected 3 tasks, got %d", len(tasks))
	}
	// expected order: high → medium → low
	if tasks[0].ID != tHigh.ID {
		t.Errorf("expected first task to be high priority (ID=%d), got ID=%d", tHigh.ID, tasks[0].ID)
	}
	if tasks[1].ID != tMed.ID {
		t.Errorf("expected second task to be medium priority (ID=%d), got ID=%d", tMed.ID, tasks[1].ID)
	}
	if tasks[2].ID != tLow.ID {
		t.Errorf("expected third task to be low priority (ID=%d), got ID=%d", tLow.ID, tasks[2].ID)
	}

	// nil priority sorts last
	tNil, _ := repo.TaskCreate(d, repo.TaskInput{ListID: l.ID, Title: "No priority"})
	tasks2, err := repo.TaskList(d, repo.TaskListOptions{SortBy: "priority"})
	if err != nil {
		t.Fatalf("TaskList() error = %v", err)
	}
	if len(tasks2) != 4 {
		t.Fatalf("expected 4 tasks, got %d", len(tasks2))
	}
	if tasks2[3].ID != tNil.ID {
		t.Errorf("expected nil-priority task last (ID=%d), got ID=%d", tNil.ID, tasks2[3].ID)
	}
}

// TestTaskDeleteCompleted — MAINT-01
func TestTaskDeleteCompleted(t *testing.T) {
	d := openTestDB(t)
	l, _ := repo.ListCreate(d, "Test", nil)

	t1, _ := repo.TaskCreate(d, repo.TaskInput{ListID: l.ID, Title: "Done task 1"})
	t2, _ := repo.TaskCreate(d, repo.TaskInput{ListID: l.ID, Title: "Done task 2"})

	d.Exec(`UPDATE tasks SET completed = 1, completed_at = ? WHERE id = ?`, "2026-02-15T10:00:00Z", t1.ID)
	d.Exec(`UPDATE tasks SET completed = 1, completed_at = ? WHERE id = ?`, "2026-02-15T10:00:00Z", t2.ID)

	result, err := repo.TaskDeleteCompleted(d, repo.DeleteCompletedOptions{Before: "2026-03-01"})
	if err != nil {
		t.Fatalf("TaskDeleteCompleted() error = %v", err)
	}
	if result.Deleted != 2 {
		t.Errorf("expected Deleted=2, got %d", result.Deleted)
	}

	_, err1 := repo.TaskGet(d, t1.ID)
	_, err2 := repo.TaskGet(d, t2.ID)
	if err1 == nil || err2 == nil {
		t.Error("expected tasks to be deleted from DB")
	}
}

// TestTaskDeleteCompleted_DryRun — MAINT-02
func TestTaskDeleteCompleted_DryRun(t *testing.T) {
	d := openTestDB(t)
	l, _ := repo.ListCreate(d, "Test", nil)

	t1, _ := repo.TaskCreate(d, repo.TaskInput{ListID: l.ID, Title: "Done task 1"})
	t2, _ := repo.TaskCreate(d, repo.TaskInput{ListID: l.ID, Title: "Done task 2"})

	d.Exec(`UPDATE tasks SET completed = 1, completed_at = ? WHERE id = ?`, "2026-02-15T10:00:00Z", t1.ID)
	d.Exec(`UPDATE tasks SET completed = 1, completed_at = ? WHERE id = ?`, "2026-02-15T10:00:00Z", t2.ID)

	result, err := repo.TaskDeleteCompleted(d, repo.DeleteCompletedOptions{
		Before: "2026-03-01",
		DryRun: true,
	})
	if err != nil {
		t.Fatalf("TaskDeleteCompleted() dry run error = %v", err)
	}
	if result.Deleted != 0 {
		t.Errorf("expected Deleted=0 in dry run, got %d", result.Deleted)
	}
	if len(result.Tasks) != 2 {
		t.Errorf("expected Tasks slice len=2 in dry run, got %d", len(result.Tasks))
	}

	// tasks must still exist
	got1, err1 := repo.TaskGet(d, t1.ID)
	got2, err2 := repo.TaskGet(d, t2.ID)
	if err1 != nil || err2 != nil {
		t.Errorf("tasks should still exist after dry run: err1=%v, err2=%v", err1, err2)
	}
	_ = got1
	_ = got2
}

// TestTaskDeleteCompleted_Scoped — MAINT-05
func TestTaskDeleteCompleted_Scoped(t *testing.T) {
	d := openTestDB(t)
	l1, _ := repo.ListCreate(d, "List1", nil)
	l2, _ := repo.ListCreate(d, "List2", nil)

	tA, _ := repo.TaskCreate(d, repo.TaskInput{ListID: l1.ID, Title: "Done in L1"})
	tB, _ := repo.TaskCreate(d, repo.TaskInput{ListID: l2.ID, Title: "Done in L2"})

	today := "2026-03-06"
	d.Exec(`UPDATE tasks SET completed = 1, completed_at = ? WHERE id = ?`, "2026-02-15T10:00:00Z", tA.ID)
	d.Exec(`UPDATE tasks SET completed = 1, completed_at = ? WHERE id = ?`, "2026-02-15T10:00:00Z", tB.ID)

	result, err := repo.TaskDeleteCompleted(d, repo.DeleteCompletedOptions{
		ListID: &l1.ID,
		Before: today,
	})
	if err != nil {
		t.Fatalf("TaskDeleteCompleted() scoped error = %v", err)
	}
	if result.Deleted != 1 {
		t.Errorf("expected Deleted=1 (only list1 task), got %d", result.Deleted)
	}

	// list1 task deleted
	_, err1 := repo.TaskGet(d, tA.ID)
	if err1 == nil {
		t.Error("expected list1 task to be deleted")
	}
	// list2 task still exists
	_, err2 := repo.TaskGet(d, tB.ID)
	if err2 != nil {
		t.Errorf("expected list2 task to still exist, got error: %v", err2)
	}
}

// TestTaskStats — STAT-01
func TestTaskStats(t *testing.T) {
	d := openTestDB(t)
	l1, _ := repo.ListCreate(d, "List1", nil)
	l2, _ := repo.ListCreate(d, "List2", nil)
	l3, _ := repo.ListCreate(d, "List3", nil) // empty list

	// list1: 2 pending + 1 done
	repo.TaskCreate(d, repo.TaskInput{ListID: l1.ID, Title: "Pending 1"})
	repo.TaskCreate(d, repo.TaskInput{ListID: l1.ID, Title: "Pending 2"})
	tDone, _ := repo.TaskCreate(d, repo.TaskInput{ListID: l1.ID, Title: "Done"})
	d.Exec(`UPDATE tasks SET completed = 1, completed_at = datetime('now') WHERE id = ?`, tDone.ID)

	// list2: 1 pending
	repo.TaskCreate(d, repo.TaskInput{ListID: l2.ID, Title: "Pending in L2"})

	_ = l3 // no tasks in l3

	summary, err := repo.TaskStats(d)
	if err != nil {
		t.Fatalf("TaskStats() error = %v", err)
	}

	if summary.Total != 4 {
		t.Errorf("expected Total=4, got %d", summary.Total)
	}
	if summary.Pending != 3 {
		t.Errorf("expected Pending=3, got %d", summary.Pending)
	}
	if summary.Done != 1 {
		t.Errorf("expected Done=1, got %d", summary.Done)
	}
	if len(summary.ByList) != 3 {
		t.Errorf("expected ByList len=3 (including empty list), got %d", len(summary.ByList))
	}

	// find list1 entry
	var list1Entry *repo.ListStats
	for i := range summary.ByList {
		if summary.ByList[i].ListID == l1.ID {
			list1Entry = &summary.ByList[i]
			break
		}
	}
	if list1Entry == nil {
		t.Fatal("list1 entry not found in ByList")
	}
	if list1Entry.Total != 3 {
		t.Errorf("list1: expected Total=3, got %d", list1Entry.Total)
	}
	if list1Entry.Pending != 2 {
		t.Errorf("list1: expected Pending=2, got %d", list1Entry.Pending)
	}
	if list1Entry.Done != 1 {
		t.Errorf("list1: expected Done=1, got %d", list1Entry.Done)
	}

	// find list2 entry
	var list2Entry *repo.ListStats
	for i := range summary.ByList {
		if summary.ByList[i].ListID == l2.ID {
			list2Entry = &summary.ByList[i]
			break
		}
	}
	if list2Entry == nil {
		t.Fatal("list2 entry not found in ByList")
	}
	if list2Entry.Total != 1 {
		t.Errorf("list2: expected Total=1, got %d", list2Entry.Total)
	}
	if list2Entry.Pending != 1 {
		t.Errorf("list2: expected Pending=1, got %d", list2Entry.Pending)
	}
	if list2Entry.Done != 0 {
		t.Errorf("list2: expected Done=0, got %d", list2Entry.Done)
	}
}
