# Testing Patterns

**Analysis Date:** 2026-03-06

## Test Framework

**Runner:**
- Go `testing` package (standard library)
- Config: none needed (Go built-in)

**Assertion Library:**
- Manual assertions with `if` checks and `t.Errorf()` / `t.Fatalf()`
- No external assertion library (testify, assert, etc.)

**Run Commands:**
```bash
go test ./...              # Run all tests
go test ./internal/... -v  # Verbose (individual test output)
go test -race ./...        # Race condition detection
go test -cover ./...       # Coverage summary
```

## Test File Organization

**Location:**
- Co-located with implementation (same package)
- `*_test.go` suffix in same directory as code being tested

**Naming:**
- Test files: `{module}_test.go` (e.g. `repo_test.go`, `db_test.go`, `config_test.go`, `output_test.go`)
- Functions: `Test{FunctionName}` or `Test{FunctionName}_{Scenario}` (e.g. `TestListCreate`, `TestListCreate_Duplicate`, `TestListDelete_NotFound`)

**Structure:**
```
internal/
├── db/
│   ├── db.go
│   └── db_test.go
├── config/
│   ├── config.go
│   └── config_test.go
├── repo/
│   ├── list.go
│   ├── task.go
│   ├── recur_scheduler.go
│   └── repo_test.go
├── output/
│   ├── output.go
│   └── output_test.go
└── models/
    └── models.go          # (no tests; simple data structures)
```

## Test Structure

**Setup helper pattern:**

From `internal/repo/repo_test.go:13-29`:
```go
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
```

**Pattern:**
- Helper functions marked with `t.Helper()` to hide them from test output
- Setup in function body (not setUp methods)
- Cleanup with `t.Cleanup()` deferred callbacks
- Temporary files created with `os.CreateTemp()`

**Standard test structure:**
```go
func TestOperation(t *testing.T) {
	d := openTestDB(t)  // Setup

	// Execute
	result, err := repo.ListCreate(d, "Test", nil)

	// Assert
	if err != nil {
		t.Fatalf("ListCreate() error = %v", err)
	}
	if result.Name != "Test" {
		t.Errorf("expected Name=Test, got %q", result.Name)
	}
}
```

**Assertion patterns:**
- Use `t.Fatal()` for setup errors (stops test immediately)
- Use `t.Errorf()` for test assertions (continues, logs failure)
- Use `t.Fatalf()` for critical validation failures
- Error format: descriptive message with actual value (e.g. `expected Name=Test, got %q`)

## Table-Driven Tests

Not used in current codebase. Individual test functions per scenario instead.

## Mocking

**Framework:** None (use temporary files, real SQLite instances)

**Pattern:**
- Real database in temporary files for integration tests
- `os.CreateTemp()` for test DB creation (auto-cleanup via `t.Cleanup()`)
- No mock libraries (gomock, testify/mock) used

**What to mock:**
- External network services (not relevant to current codebase)

**What NOT to mock:**
- Database: use real temporary SQLite (fast, more realistic)
- File I/O: use real temp files for config tests
- Internal repo functions: test via direct calls, not mocks

**Example real DB test from `internal/db/db_test.go:10-27`:**
```go
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
```

## Fixtures and Factories

**Test data:**
- Hard-coded in test functions
- Lists created with `repo.ListCreate(d, "Name", nil)`
- Tasks created with `repo.TaskCreate(d, repo.TaskInput{...})`

**Example from `internal/repo/repo_test.go:243-258`:**
```go
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
```

**Location:**
- Inline in test functions (no separate fixture files)
- Via repo CRUD functions (leverages actual implementation)

## Coverage

**Requirements:** Not enforced (no coverage tool configured)

**View Coverage:**
```bash
go test -cover ./...                    # Summary per package
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out        # HTML report
```

**Gaps:**
- Command handlers (`cmd/`) not tested (integration would require CLI parsing)
- Output formatting has basic tests but not comprehensive

## Test Types

**Unit Tests:**
- Scope: Individual repo functions (Create, Get, List, Delete, Done, etc.)
- Approach: Real SQLite DB in temp file, test one function per test
- Examples: `TestTaskCreate`, `TestTaskDone`, `TestListEdit`

**Integration Tests:**
- Scope: Full workflows (create task → mark done → schedule next occurrence)
- Approach: Real DB with multiple repo calls in sequence
- Example: `TestTaskRecur_*` tests in `repo_test.go` (task scheduling with recurrence)

**End-to-End Tests:**
- Not implemented (would require CLI process execution)
- Could test via subprocess (`os/exec`) if needed in future

**Command Tests:**
- Minimal (no tests for `cmd/` package handlers)
- Would require mocking stdout or full CLI parsing
- Recommended future improvement: add CLI integration tests

## Async Testing

Not applicable (CLI tool, no goroutines in main code path).

## Error Testing

**Pattern:**
- Test for `err == nil` or `err != nil` as appropriate
- For expected errors: `if err == nil { t.Error("expected error, got nil") }`

**Example from `internal/repo/repo_test.go:159-165`:**
```go
func TestListEdit_NotFound(t *testing.T) {
	d := openTestDB(t)

	name := "ghost"
	if _, err := repo.ListPatchFields(d, 9999, repo.ListPatch{Name: &name}); err == nil {
		t.Error("expected error for non-existent list, got nil")
	}
}
```

**Database error patterns:**
- Schema/constraint violations bubble up wrapped with `fmt.Errorf()`
- Not-found scenarios return explicit error messages (e.g. "list 9999 not found")

## Output Testing

**Pattern:** Capture stdout using `os.Pipe()`

From `internal/output/output_test.go:16-31`:
```go
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
```

**Usage:**
```go
output.JSONMode = false
got := captureStdout(t, func() {
	output.PrintLists(lists)
})
if !strings.Contains(got, "Work") {
	t.Errorf("expected 'Work' in output, got %q", got)
}
```

**Testing both formats:**
- Table output: check string contains expected values
- JSON output: unmarshal and validate structure

**Example from `internal/output/output_test.go:60-78`:**
```go
func TestPrintLists_JSON(t *testing.T) {
	output.JSONMode = true
	defer func() { output.JSONMode = false }()

	lists := []models.List{...}
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
```

## Test Coverage by Package

| Package | Coverage | Key Tests |
|---------|----------|-----------|
| `internal/db` | Good | Open, directory creation, schema migration |
| `internal/config` | Good | DefaultDBPath, EnvFilePath, Load, wizard |
| `internal/repo` | Excellent | Full CRUD (lists/tasks), filters, done/undone, subtasks, recurrence, scheduling, autocomplete |
| `internal/output` | Basic | Table and JSON output for lists/tasks/success/error |
| `cmd` | None | (no tests; would need CLI integration tests) |
| `internal/models` | Not applicable | (data structures only) |

## Test Database Cleanup

**Pattern:**
- Temporary file created with `os.CreateTemp()`
- Cleanup registered via `t.Cleanup(func() { os.Remove(name) })`
- Database connection closed in cleanup

**Ensures:**
- No test pollution (each test gets fresh DB)
- Files cleaned up even if test fails
- Multiple cleanup handlers run in LIFO order

## Common Test Patterns

**Simple CRUD test:**
```go
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
}
```

**Filter/options test:**
```go
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
}
```

## Test Execution Notes

- All tests use `testing.T` (no custom test runners)
- Parallel execution: tests don't conflict (each has temp DB)
- Safe to run with `-race` flag (no global state in repo)
- No setup/teardown hooks (all in helper functions)
