package output_test

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/danielmrdev/dtasks-cli/internal/models"
	"github.com/danielmrdev/dtasks-cli/internal/output"
	"github.com/danielmrdev/dtasks-cli/internal/repo"
)

// captureStdout replaces os.Stdout temporarily and returns the captured output.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	old := os.Stdout
	os.Stdout = w
	fn()
	w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

func TestPrintLists_Empty(t *testing.T) {
	output.JSONMode = false
	got := captureStdout(t, func() {
		output.PrintLists(nil)
	})
	if !strings.Contains(got, "No lists found") {
		t.Errorf("expected 'No lists found', got %q", got)
	}
}

func TestPrintLists_Table(t *testing.T) {
	output.JSONMode = false
	lists := []models.List{
		{ID: 1, Name: "Work", CreatedAt: time.Now()},
		{ID: 2, Name: "Personal", CreatedAt: time.Now()},
	}
	got := captureStdout(t, func() {
		output.PrintLists(lists)
	})
	if !strings.Contains(got, "Work") {
		t.Errorf("expected 'Work' in output, got %q", got)
	}
	if !strings.Contains(got, "Personal") {
		t.Errorf("expected 'Personal' in output, got %q", got)
	}
}

func TestPrintLists_JSON(t *testing.T) {
	output.JSONMode = true
	defer func() { output.JSONMode = false }()

	lists := []models.List{
		{ID: 1, Name: "Work", CreatedAt: time.Now()},
	}
	got := captureStdout(t, func() {
		output.PrintLists(lists)
	})

	var result map[string]any
	if err := json.Unmarshal([]byte(got), &result); err != nil {
		t.Fatalf("output is not valid JSON: %v\nOutput: %q", err, got)
	}
	if _, ok := result["lists"]; !ok {
		t.Error("expected 'lists' key in JSON output")
	}
}

func TestPrintList_Single(t *testing.T) {
	output.JSONMode = false
	l := &models.List{ID: 3, Name: "Home", CreatedAt: time.Now()}
	got := captureStdout(t, func() {
		output.PrintList(l)
	})
	if !strings.Contains(got, "Home") {
		t.Errorf("expected 'Home' in output, got %q", got)
	}
}

func TestPrintTasks_Empty(t *testing.T) {
	output.JSONMode = false
	got := captureStdout(t, func() {
		output.PrintTasks(nil)
	})
	if !strings.Contains(got, "No tasks found") {
		t.Errorf("expected 'No tasks found', got %q", got)
	}
}

func TestPrintTasks_Table(t *testing.T) {
	output.JSONMode = false
	tasks := []models.Task{
		{ID: 1, ListID: 1, ListName: "Work", Title: "Write report", Completed: false},
		{ID: 2, ListID: 1, ListName: "Work", Title: "Review PR", Completed: true},
	}
	got := captureStdout(t, func() {
		output.PrintTasks(tasks)
	})
	if !strings.Contains(got, "Write report") {
		t.Errorf("expected 'Write report' in output, got %q", got)
	}
	if !strings.Contains(got, "Review PR") {
		t.Errorf("expected 'Review PR' in output, got %q", got)
	}
}

func TestPrintTasks_JSON(t *testing.T) {
	output.JSONMode = true
	defer func() { output.JSONMode = false }()

	tasks := []models.Task{
		{ID: 1, ListID: 1, ListName: "Work", Title: "Task 1"},
	}
	got := captureStdout(t, func() {
		output.PrintTasks(tasks)
	})
	var result map[string]any
	if err := json.Unmarshal([]byte(got), &result); err != nil {
		t.Fatalf("output is not valid JSON: %v\nOutput: %q", err, got)
	}
	if _, ok := result["tasks"]; !ok {
		t.Error("expected 'tasks' key in JSON output")
	}
}

func TestPrintTask_Detail(t *testing.T) {
	output.JSONMode = false
	notes := "important notes"
	due := "2026-03-01"
	task := &models.Task{
		ID:       42,
		ListID:   1,
		ListName: "Work",
		Title:    "Detailed task",
		Notes:    &notes,
		DueDate:  &due,
	}
	got := captureStdout(t, func() {
		output.PrintTask(task)
	})
	if !strings.Contains(got, "Detailed task") {
		t.Errorf("expected title in output, got %q", got)
	}
	if !strings.Contains(got, "important notes") {
		t.Errorf("expected notes in output, got %q", got)
	}
	if !strings.Contains(got, "2026-03-01") {
		t.Errorf("expected due date in output, got %q", got)
	}
}

func TestPrintSuccess(t *testing.T) {
	output.JSONMode = false
	got := captureStdout(t, func() {
		output.PrintSuccess("done!")
	})
	if !strings.Contains(got, "done!") {
		t.Errorf("expected 'done!' in output, got %q", got)
	}
}

func TestPrintSuccess_JSON(t *testing.T) {
	output.JSONMode = true
	defer func() { output.JSONMode = false }()

	got := captureStdout(t, func() {
		output.PrintSuccess("ok")
	})
	var result map[string]string
	if err := json.Unmarshal([]byte(got), &result); err != nil {
		t.Fatalf("not valid JSON: %v", err)
	}
	if result["status"] != "ok" {
		t.Errorf("expected status=ok, got %q", result["status"])
	}
}

// --- Phase 2 Richness tests (TDD red phase) ---

func ptrStr(s string) *string { return &s }

// TestPrintTasks_Priority — PRIO-03
func TestPrintTasks_Priority(t *testing.T) {
	output.JSONMode = false
	tasks := []models.Task{
		{ID: 1, ListID: 1, ListName: "Work", Title: "High task", Priority: ptrStr("high")},
		{ID: 2, ListID: 1, ListName: "Work", Title: "Medium task", Priority: ptrStr("medium")},
		{ID: 3, ListID: 1, ListName: "Work", Title: "No priority task"},
	}
	got := captureStdout(t, func() {
		output.PrintTasks(tasks)
	})
	if !strings.Contains(got, "PRIO") {
		t.Errorf("expected PRIO header in output, got %q", got)
	}
	if !strings.Contains(got, "!") {
		t.Errorf("expected '!' for high priority in output, got %q", got)
	}
	if !strings.Contains(got, "~") {
		t.Errorf("expected '~' for medium priority in output, got %q", got)
	}
}

// TestPrintDeletedCount — MAINT-04
func TestPrintDeletedCount(t *testing.T) {
	// JSON mode
	output.JSONMode = true
	gotJSON := captureStdout(t, func() {
		output.PrintDeletedCount(3)
	})
	output.JSONMode = false

	if !strings.Contains(gotJSON, `"deleted"`) {
		t.Errorf("expected JSON with 'deleted' key, got %q", gotJSON)
	}
	if !strings.Contains(gotJSON, `3`) {
		t.Errorf("expected JSON with value 3, got %q", gotJSON)
	}
	var jsonResult map[string]any
	if err := json.Unmarshal([]byte(gotJSON), &jsonResult); err != nil {
		t.Fatalf("output is not valid JSON: %v\nOutput: %q", err, gotJSON)
	}
	if jsonResult["deleted"] != float64(3) {
		t.Errorf("expected deleted=3, got %v", jsonResult["deleted"])
	}

	// text mode
	gotText := captureStdout(t, func() {
		output.PrintDeletedCount(3)
	})
	if !strings.Contains(gotText, "Deleted 3 task(s).") {
		t.Errorf("expected 'Deleted 3 task(s).', got %q", gotText)
	}
}

// TestPrintStats_Table — STAT-02 (table mode)
func TestPrintStats_Table(t *testing.T) {
	output.JSONMode = false
	summary := &repo.StatsSummary{
		Total:   5,
		Pending: 3,
		Done:    2,
		ByList: []repo.ListStats{
			{ListID: 1, ListName: "Work", Total: 3, Pending: 2, Done: 1},
			{ListID: 2, ListName: "Personal", Total: 2, Pending: 1, Done: 1},
		},
	}
	got := captureStdout(t, func() {
		output.PrintStats(summary)
	})
	if !strings.Contains(got, "Work") {
		t.Errorf("expected 'Work' in output, got %q", got)
	}
	if !strings.Contains(got, "Personal") {
		t.Errorf("expected 'Personal' in output, got %q", got)
	}
	// check for column headers
	if !strings.Contains(got, "Total") || !strings.Contains(got, "Pending") || !strings.Contains(got, "Done") {
		t.Errorf("expected Total/Pending/Done columns in output, got %q", got)
	}
}

// TestPrintStats_JSON — STAT-02 (JSON mode)
func TestPrintStats_JSON(t *testing.T) {
	output.JSONMode = true
	defer func() { output.JSONMode = false }()

	summary := &repo.StatsSummary{
		Total:   5,
		Pending: 3,
		Done:    2,
		ByList: []repo.ListStats{
			{ListID: 1, ListName: "Work", Total: 3, Pending: 2, Done: 1},
		},
	}
	got := captureStdout(t, func() {
		output.PrintStats(summary)
	})
	var result map[string]any
	if err := json.Unmarshal([]byte(got), &result); err != nil {
		t.Fatalf("output is not valid JSON: %v\nOutput: %q", err, got)
	}
	for _, key := range []string{"total", "pending", "done", "by_list"} {
		if _, ok := result[key]; !ok {
			t.Errorf("expected key %q in JSON output", key)
		}
	}
}
