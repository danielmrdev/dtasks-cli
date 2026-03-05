# Codebase Structure

**Analysis Date:** 2026-03-06

## Directory Layout

```
dtasks-cli/
├── main.go                   # Entry point (calls cmd.Execute)
├── cmd/                      # Cobra CLI commands
│   ├── root.go               # Root command, DB/config initialization, global flags
│   ├── list.go               # List subcommand: create, ls, edit, rm
│   ├── task.go               # Task commands: add, ls, show, edit, done, undone, rm
│   ├── recur.go              # Recurrence commands: recur (daily/weekly/monthly), recur rm
│   └── completion.go         # Shell completion helpers
├── internal/
│   ├── config/               # Configuration & environment
│   │   ├── config.go         # Load .env, platform-specific paths, first-run wizard
│   │   └── config_test.go
│   ├── db/                   # Database layer
│   │   ├── db.go             # SQLite: Open, configure (PRAGMAs), migrate
│   │   └── db_test.go
│   ├── models/               # Data structures
│   │   └── models.go         # List, Task structs (JSON-tagged)
│   ├── repo/                 # Repository/business logic
│   │   ├── list.go           # List CRUD: Create, Get, All, Patch, Delete
│   │   ├── task.go           # Task CRUD + recurrence: Create, Get, List, Update, Patch, Done/Undone, Delete, SetRecur, RemoveRecur
│   │   ├── recur_scheduler.go # Recurrence engine: calcNextDate, shouldSpawn, TaskScheduleNext, ProcessAutocompleteTasks
│   │   └── repo_test.go      # Integration tests for all repo functions
│   └── output/               # Formatted output
│       ├── output.go         # Print functions: tables (text/JSON), colors, formatting
│       └── output_test.go
├── assets/                   # Images/resources for docs
├── .github/
│   └── workflows/            # CI/CD (GitHub Actions)
├── Makefile                  # Build targets (build, install, build-all, release)
├── go.mod / go.sum           # Go module definition
├── CLAUDE.md                 # Architecture and conventions (this repo)
├── README.md                 # User-facing documentation
├── CHANGELOG.md              # Version history
├── CONTRIBUTING.md           # Contributor guidelines
├── LICENSE                   # Software license
└── .planning/codebase/       # GSD planning documents (generated)
    ├── ARCHITECTURE.md       # Architecture patterns and layers (this file)
    └── STRUCTURE.md          # Directory structure and naming (this file)
```

## Directory Purposes

**`cmd/`:**
- Purpose: Cobra CLI command definitions
- Contains: Root command setup, subcommands for lists/tasks/recurrence, flag definitions, shell completion
- Key files:
  - `root.go`: initializes DB and config on every invocation
  - `list.go`: subcommands create/ls/edit/rm for lists
  - `task.go`: subcommands add/ls/show/edit/done/undone/rm for tasks
  - `recur.go`: subcommands for task recurrence setup
  - `completion.go`: functions to generate shell completions and dynamic suggestions

**`internal/config/`:**
- Purpose: Configuration loading and first-run setup
- Contains: `.env` file reading, platform-specific path resolution, interactive wizard
- Key files: `config.go` exports `Load()`, `EnvFilePath()`, `DefaultDBPath()`

**`internal/db/`:**
- Purpose: SQLite connection and schema management
- Contains: Database opening, pragma configuration (WAL, foreign keys), idempotent migrations
- Key files: `db.go` exports `Open()` which handles all DB initialization

**`internal/models/`:**
- Purpose: Domain data structures
- Contains: `List` struct (id, name, color, created_at) and `Task` struct (25+ fields covering core + recurrence)
- JSON tags on all fields for `--json` output

**`internal/repo/`:**
- Purpose: Data access and domain logic
- Contains: CRUD operations, input structs (TaskInput, TaskPatch, ListPatch, RecurInput), recurrence scheduling
- Key files:
  - `list.go`: `ListCreate`, `ListGet`, `ListAll`, `ListPatchFields`, `ListDelete`
  - `task.go`: Task CRUD, partial updates, done/undone toggle, recurrence config
  - `recur_scheduler.go`: Next occurrence calculation and auto-complete processing

**`internal/output/`:**
- Purpose: Formatted console/JSON output
- Contains: Table rendering with Unicode borders, ANSI colors, JSON marshaling
- Key files: `output.go` exports `Print*()` functions (PrintLists, PrintTasks, PrintTask, PrintSuccess, PrintError)

**`assets/`:**
- Purpose: Images and documentation resources
- Contains: Screenshots or diagrams referenced in README/docs

## Key File Locations

**Entry Points:**
- `main.go`: Binary entry point — parses version, calls `cmd.Execute()`
- `cmd/root.go`: Root command — initializes DB, config, runs auto-complete on every invocation

**Configuration:**
- `internal/config/config.go`: Loads/creates `.env`, prompts for DB path on first run
- Platform config paths: `~/.dtasks/.env` (macOS), `~/.config/dtasks/.env` (Linux), `%AppData%\dtasks\.env` (Windows)

**Core Logic:**
- `internal/repo/task.go`: Task CRUD and state management
- `internal/repo/list.go`: List CRUD
- `internal/repo/recur_scheduler.go`: Recurrence calculation and auto-complete
- `internal/db/db.go`: Schema and DB setup

**Database Schema:**
- Defined inline in `internal/db/db.go` function `migrate()` (lines 51–86)
- Tables: `lists`, `tasks`
- Indices: `idx_tasks_list`, `idx_tasks_parent`, `idx_tasks_due`

**Output:**
- `internal/output/output.go`: All print functions and table rendering
- Uses `mattn/go-runewidth` for proper Unicode column width

**Testing:**
- `internal/config/config_test.go`: Config loading and paths
- `internal/db/db_test.go`: DB opening, migrations, schema evolution
- `internal/output/output_test.go`: Table and JSON formatting
- `internal/repo/repo_test.go`: CRUD, filters, recurrence, cascades, auto-complete

## Naming Conventions

**Files:**
- Package files are lowercase with underscores: `config.go`, `db.go`, `output.go`, `recur_scheduler.go`
- Test files follow Go convention: `*_test.go`
- No grouping prefixes (e.g., no `task_create.go`, `task_list.go` — all in `task.go`)

**Directories:**
- Package directories are lowercase, no hyphens: `internal/config/`, `internal/db/`, `internal/repo/`
- Purpose-based grouping: `config` for config, `db` for persistence, `repo` for logic, `models` for data, `output` for formatting

**Functions:**
- Public (exported): PascalCase — `TaskCreate`, `ListAll`, `PrintTasks`, `ProcessAutocompleteTasks`
- Private (unexported): camelCase — `scanTask`, `calcNextDate`, `shouldSpawn`, `boolToInt`, `formatDate`
- Input structs: PascalCase descriptive names — `TaskInput`, `TaskPatch`, `ListPatch`, `RecurInput`

**Variables:**
- Global vars: camelCase — `DB`, `jsonFlag`, `dbPathFlag` (in cmd/root.go)
- Local: camelCase — `in`, `opts`, `t`, `l`, `rows`, `id`
- SQL NULL types: lowercase with Null suffix — `completedAt` of type `sql.NullString`

**Types:**
- Structs: PascalCase — `Config`, `Task`, `List`, `TaskInput`
- Interfaces: lowercase (Go convention) — `scanner` for row scanning
- Options structs: suffixed with `Options` or `Input` or `Patch` for semantic clarity

**Constants:**
- Uppercase: `taskSelectSQL` for SQL query template

## Where to Add New Code

**New Feature (e.g., tags, priorities):**
- Primary code:
  - Schema changes: `internal/db/db.go` `migrate()` function (idempotent ALTER TABLE)
  - Models: Add fields to `Task` struct in `internal/models/models.go`
  - CRUD: Add repo functions to `internal/repo/task.go` following `TaskCreate`/`TaskPatch` pattern
  - CLI: Add flag and handler to appropriate `cmd/*.go` file (e.g., `cmd/task.go` for task-related flags)
  - Output: Update `PrintTask` in `internal/output/output.go` to display new field
- Tests: `internal/repo/repo_test.go` (integration tests)

**New Command (e.g., `dtasks search`):**
- New file: Create `cmd/search.go` with Cobra command definition
- Register: Add to `rootCmd.AddCommand()` in `cmd/root.go` init()
- Repo function: Add query logic to appropriate repo file (e.g., `internal/repo/task.go`)
- Completion: Add ValidArgsFunction and helper to `cmd/completion.go`

**New Output Format (e.g., CSV export):**
- Extend: Add function to `internal/output/output.go` (e.g., `PrintTasksCSV`)
- Toggle: Use existing `output.JSONMode` pattern or add new global flag in `cmd/root.go`

**Utility Functions:**
- Shared task helpers: `internal/repo/task.go` (e.g., `scanTask`)
- Date/time utilities: `internal/repo/recur_scheduler.go` or extract to new `internal/time/time.go` if reused
- Output helpers: `internal/output/output.go` (e.g., `formatDate`, `colorDot`)
- Config platform logic: `internal/config/config.go` (e.g., `DefaultDBPath`)

## Special Directories

**`.planning/codebase/`:**
- Purpose: GSD mapping documents
- Generated: No (written by `/gsd:map-codebase` command)
- Committed: Yes
- Contents: ARCHITECTURE.md, STRUCTURE.md, CONVENTIONS.md, TESTING.md, STACK.md, INTEGRATIONS.md, CONCERNS.md

**`.github/workflows/`:**
- Purpose: GitHub Actions CI/CD
- Generated: No (checked in)
- Committed: Yes
- Contents: Build, test, release pipelines

**`dist/`:**
- Purpose: Build output directory
- Generated: Yes (by `make build`, `make build-all`)
- Committed: No (in `.gitignore`)
- Contents: Binary executables for all platforms

**`assets/`:**
- Purpose: Images and documentation resources
- Generated: No (manual)
- Committed: Yes
- Contents: Screenshots for README

**`skills/dtasks-cli/`:**
- Purpose: AI agent skill definition
- Generated: No (manual, for external agents)
- Committed: Yes
- Contents: Command reference, flags, JSON output schemas

---

*Structure analysis: 2026-03-06*
