# Architecture

**Analysis Date:** 2026-03-06

## Pattern Overview

**Overall:** Layered CLI architecture with separation of concerns: CLI command layer, data repository layer, SQLite persistence layer.

**Key Characteristics:**
- Dependency injection: global `*sql.DB` passed to repository functions
- Functional repository API: pure functions (no methods on models)
- Single-responsibility commands: each Cobra command handles one operation
- Transactional safety: long operations wrapped in explicit transactions
- Idempotent migrations: schema changes survive multiple runs

## Layers

**Command Layer (CLI):**
- Purpose: Parse flags, validate user input, call repository functions, format output
- Location: `cmd/` — `root.go`, `list.go`, `task.go`, `recur.go`, `completion.go`
- Contains: Cobra command definitions, flag parsing, input validation
- Depends on: `internal/repo`, `internal/output`
- Used by: `main.go` → `cmd.Execute()`

**Repository Layer (Business Logic):**
- Purpose: Data access and domain logic — CRUD operations, recurrence scheduling, autocomplete processing
- Location: `internal/repo/` — `list.go`, `task.go`, `recur_scheduler.go`
- Contains: Functions like `TaskCreate`, `TaskList`, `TaskDone`, `TaskScheduleNext`, `ProcessAutocompleteTasks`
- Depends on: `internal/models`, `internal/db` (receives `*sql.DB`)
- Used by: Command layer only

**Database Layer:**
- Purpose: SQLite connection, schema management, pragmas configuration
- Location: `internal/db/db.go`
- Contains: `Open()` (initializes DB), `configure()` (PRAGMA settings), `migrate()` (CREATE/ALTER migrations)
- Depends on: `modernc.org/sqlite`
- Used by: `cmd/root.go` on startup

**Models Layer:**
- Purpose: Data structures for tasks and lists
- Location: `internal/models/models.go`
- Contains: `List` and `Task` structs with JSON tags for output
- Depends on: stdlib `time`
- Used by: All layers

**Output Layer:**
- Purpose: Format data for console (table + ANSI colors) or JSON
- Location: `internal/output/output.go`
- Contains: `PrintLists`, `PrintTasks`, `PrintTask`, `PrintSuccess`, `PrintError`, table rendering, color formatting
- Depends on: `internal/models`, `mattn/go-runewidth`
- Used by: Command layer only

**Config Layer:**
- Purpose: Load environment configuration and manage database path
- Location: `internal/config/config.go`
- Contains: `Load()` (reads `.env`), interactive wizard on first run, platform-specific paths
- Depends on: `github.com/joho/godotenv`
- Used by: `cmd/root.go` on startup

## Data Flow

**Command Execution:**

1. User invokes `dtasks <command> [flags] [args]`
2. `main.go` → `cmd.Execute(version)`
3. `cmd/root.go` PersistentPreRunE:
   - Skip DB init for help/completion commands
   - Load config via `config.Load()` (runs wizard if `.env` missing)
   - Open DB via `db.Open(dbPath)` → runs migrations
   - Call `repo.ProcessAutocompleteTasks(DB)` to handle overdue auto-complete tasks
4. Cobra dispatches to specific command (e.g., `addCmd`, `lsCmd`, `editCmd`)
5. Command parses flags, builds input struct (e.g., `TaskInput`, `ListPatch`)
6. Command calls repo function (e.g., `TaskCreate(DB, in)`)
7. Repo function executes SQL, scans rows, returns models
8. Command calls `output.Print*()` to format and display result
9. Exit with code 0 (success) or 1 (error)

**Task Completion & Scheduling:**

1. User runs `dtasks done <id>`
2. `doneCmd` calls `repo.TaskDone(DB, id, true)` → updates `completed=1, completed_at=NOW`
3. `doneCmd` calls `repo.TaskScheduleNext(DB, id)` if recurring
4. `TaskScheduleNext`:
   - Fetches task with `TaskGet`
   - Validates recurrence fields
   - Calculates next due date via `calcNextDate()` (handles daily/weekly/monthly + intervals)
   - Checks if more occurrences should spawn via `shouldSpawn()` (checks ends_type/ends_date/ends_after)
   - Begins transaction
   - Inserts new task with next due date
   - Copies recurrence config to new task, increments `recur_count`
   - Commits transaction
   - Returns new task

**Auto-Complete Processing:**

1. Every command invocation: `root.go` PersistentPreRunE calls `repo.ProcessAutocompleteTasks(DB)`
2. Query finds all tasks where:
   - `autocomplete=1` AND `completed=0`
   - `due_date < today` OR (`due_date = today` AND `due_time <= now`)
3. For each overdue task:
   - Mark done via `TaskDone(DB, id, true)`
   - Spawn next occurrence via `TaskScheduleNext(DB, id)` (errors logged to stderr, non-fatal)
4. Command proceeds normally

## Key Abstractions

**TaskInput / TaskPatch:**
- Purpose: Separate structures for create vs. partial update operations
- `TaskInput`: full set of fields for creation
- `TaskPatch`: all fields optional (pointers), only non-nil fields are updated
- Example files: `internal/repo/task.go` lines 12–20, 119–126

**ListPatch:**
- Purpose: Partial list updates (name, color, or clear color)
- Pattern: `*string` allows three states: not provided (nil), provided as value, provided as empty string (clear)
- Example: `internal/repo/list.go` lines 49–83

**RecurInput:**
- Purpose: Encapsulate all recurrence configuration for `TaskSetRecur`
- Fields: type, interval, day constraints, start/end configuration
- Example: `internal/repo/task.go` lines 199–209

**Scanner Interface:**
- Purpose: Abstract over `*sql.Row` and `*sql.Rows` for unified task scanning
- Allows `scanTask()` to work with both single queries and row iteration
- Example: `internal/repo/task.go` lines 268–297

## Entry Points

**main.go:**
- Location: `/Users/danielmunoz/Projects/dtasks-cli/main.go`
- Triggers: Binary execution
- Responsibilities: Set version (build-time ldflags) and call `cmd.Execute(version)`

**root.go PersistentPreRunE:**
- Location: `cmd/root.go` lines 21–69
- Triggers: Every command (before the actual command runs)
- Responsibilities:
  - Initialize output mode from `--json` flag
  - Load configuration (run wizard if needed)
  - Open database (create directories, apply pragmas, run migrations)
  - Process auto-complete tasks
  - Set global `cmd.DB` so all commands can access it

**Cobra Command Definitions:**
- `addCmd` — Create task: `cmd/task.go` lines 23–70
- `lsCmd` — List tasks: `cmd/task.go` lines 80–112
- `showCmd` — Show task detail + subtasks: `cmd/task.go` lines 116–143
- `editCmd` — Edit task: `cmd/task.go` lines 147+
- `doneCmd` — Mark task complete + schedule next: `cmd/task.go`
- `undoneCmd` — Unmark task: `cmd/task.go`
- `rmCmd` — Delete task: `cmd/task.go`
- `listCmd` (group) with subcommands: `cmd/list.go`
  - `list create` — Create list: `cmd/list.go` lines 24–40
  - `list ls` — List all: `cmd/list.go` lines 42–53
  - `list edit` — Edit list: `cmd/list.go` lines 55–85
  - `list rm` — Delete list (cascades): `cmd/list.go` lines 87–103

## Error Handling

**Strategy:** Error propagation with context wrapping; fatal errors exit with code 1.

**Patterns:**

- **Validation at CLI boundary:** Flags checked with `cmd.Flags().Changed()` to distinguish "not provided" from "provided as empty"
  - Example: `cmd/task.go` line 28 checks `--due-time` requires `--due`

- **DB errors wrapped with context:** Format errors with `fmt.Errorf(...: %w, err)` for chaining
  - Example: `internal/repo/task.go` line 30 `"create task: %w"`

- **Non-fatal warnings:** Errors in `ProcessAutocompleteTasks` logged to stderr but don't abort
  - Example: `internal/repo/recur_scheduler.go` lines 172–177

- **Not-found handling:** Return descriptive error when resource doesn't exist
  - Example: `internal/repo/task.go` line 113 `"task %d not found"`

## Cross-Cutting Concerns

**Logging:**
- Only warning/error messages via `fmt.Fprintf(os.Stderr, ...)`
- No structured logging framework; plain text to stderr
- Auto-complete errors logged non-fatally: `internal/repo/recur_scheduler.go` lines 172–177

**Validation:**
- Flag validation in command handlers (e.g., `--due-time` requires `--due`)
- Date format validation in repo functions (`YYYY-MM-DD` for dates, `HH:MM` for times)
- SQL NOT NULL constraints enforced at DB level (e.g., `title TEXT NOT NULL`)

**Authentication:**
- Not applicable — single-user local CLI

**Date/Time Handling:**
- Dates: `YYYY-MM-DD` as `*string` (nullable)
- Times: `HH:MM` as `*string` (nullable, requires date)
- Stored and compared as strings for timezone safety
- `time.Now().Format("2006-01-02")` used for "today" comparisons

**Recurrence Logic:**
- Encapsulated in `internal/repo/recur_scheduler.go`
- `calcNextDate()` — computes next occurrence for daily/weekly/monthly with intervals
- `shouldSpawn()` — checks if more occurrences should be created (never/on_date/after_n)
- Spawning wrapped in transaction to ensure consistency
- New occurrences inherit recurrence config and increment `recur_count`

**Output Formatting:**
- Controlled by global `output.JSONMode` bool (set from `--json` flag)
- Table output: right-aligned columns, Unicode borders, ANSI color codes for list dots
- Column width calculation via `runewidth.StringWidth()` to handle wide characters
- Two rendering paths: `plain` (for width calc) and `styled` (with ANSI codes)

---

*Architecture analysis: 2026-03-06*
