package repo_test

import (
	"database/sql"
	"os"
	"testing"

	"github.com/dtasks/dtasks/internal/db"
	"github.com/dtasks/dtasks/internal/repo"
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

	l, err := repo.ListCreate(d, "Work")
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

	if _, err := repo.ListCreate(d, "Personal"); err != nil {
		t.Fatal(err)
	}
	_, err := repo.ListCreate(d, "Personal")
	if err == nil {
		t.Error("expected error for duplicate list name, got nil")
	}
}

func TestListAll(t *testing.T) {
	d := openTestDB(t)

	for _, name := range []string{"Alpha", "Beta", "Gamma"} {
		if _, err := repo.ListCreate(d, name); err != nil {
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

	created, err := repo.ListCreate(d, "MyList")
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

func TestListRename(t *testing.T) {
	d := openTestDB(t)

	l, err := repo.ListCreate(d, "OldName")
	if err != nil {
		t.Fatal(err)
	}

	if err := repo.ListRename(d, l.ID, "NewName"); err != nil {
		t.Fatalf("ListRename() error = %v", err)
	}

	got, err := repo.ListGet(d, l.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != "NewName" {
		t.Errorf("expected Name=NewName, got %q", got.Name)
	}
}

func TestListDelete(t *testing.T) {
	d := openTestDB(t)

	l, err := repo.ListCreate(d, "ToDelete")
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

func TestListRename_NotFound(t *testing.T) {
	d := openTestDB(t)

	if err := repo.ListRename(d, 9999, "ghost"); err == nil {
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

	l, err := repo.ListCreate(d, "Work")
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

	l, _ := repo.ListCreate(d, "Test")
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

	l, _ := repo.ListCreate(d, "Test")
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

	l, _ := repo.ListCreate(d, "Test")
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

	l1, _ := repo.ListCreate(d, "List1")
	l2, _ := repo.ListCreate(d, "List2")
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

	l, _ := repo.ListCreate(d, "Test")
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

	l, _ := repo.ListCreate(d, "Test")
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

	l, _ := repo.ListCreate(d, "Test")
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

	l, _ := repo.ListCreate(d, "Test")
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

	l, _ := repo.ListCreate(d, "Test")
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

	l, _ := repo.ListCreate(d, "Test")
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

	l, _ := repo.ListCreate(d, "ToDelete")
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
