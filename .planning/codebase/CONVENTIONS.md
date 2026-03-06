# Coding Conventions

**Analysis Date:** 2026-03-06

## Naming Patterns

**Files:**
- Package directory names: lowercase, no underscores (e.g. `cmd`, `internal/repo`, `internal/models`)
- Test files: `*_test.go` suffix (e.g. `repo_test.go`, `db_test.go`)
- Command handlers in cmd package: suffixed with `Cmd` (e.g. `addCmd`, `doneCmd`, `listCreateCmd`)

**Functions:**
- Public API functions: PascalCase (e.g. `ListCreate`, `TaskGet`, `TaskList`, `PrintTasks`)
- Private helper functions: camelCase (e.g. `openTestDB`, `scanTask`, `parseID`)
- Test functions: `TestFunctionName` or `TestFunctionName_Scenario` (e.g. `TestTaskCreate`, `TestListEdit_NotFound`)

**Variables:**
- Cobra command flags: camelCase (e.g. `addListID`, `dueDate`, `jsonFlag`)
- Local variables: camelCase (e.g. `dbPath`, `result`, `lists`)
- Boolean flags: descriptive names without `is`/`has` prefix (e.g. `completed`, `autocomplete`, `lsAll`)
- Temporary/short-lived: short names acceptable (e.g. `err`, `d`, `l`, `t`, `p` for patches)

**Types:**
- Struct names: PascalCase (e.g. `List`, `Task`, `Config`, `TaskInput`, `ListPatch`)
- Interface names: typically suffixed with `er` pattern (e.g. `scanner` interface at `internal/repo/task.go:268`)
- Input structs: suffixed with `Input` (e.g. `TaskInput` at `internal/repo/task.go:12`)
- Patch/update structs: suffixed with `Patch` (e.g. `TaskPatch`, `ListPatch`)
- Options structs: suffixed with `Options` (e.g. `TaskListOptions`)

## Code Style

**Formatting:**
- Go standard formatting (managed by `gofmt`, implicit in `go build`)
- No explicit prettier/eslint config; Go community standards apply
- 100% standard Go idioms — no custom formatters

**Linting:**
- `go vet` enforced (run via `go test ./...`)
- No external linter config (golangci-lint not configured)
- Code passes `go vet` without errors

## Import Organization

**Order:**
1. Standard library imports (`import (`)
2. Empty line
3. Third-party imports (github.com, modernc.org, golang.org, etc.)

**Example from `cmd/root.go`:**
```go
import (
	"database/sql"
	"fmt"
	"os"

	"github.com/danielmrdev/dtasks-cli/internal/config"
	"github.com/danielmrdev/dtasks-cli/internal/db"
	"github.com/danielmrdev/dtasks-cli/internal/output"
	"github.com/danielmrdev/dtasks-cli/internal/repo"
	"github.com/spf13/cobra"
)
```

**Path Aliases:**
- No aliases used; full module paths only (e.g. `github.com/danielmrdev/dtasks-cli/internal/repo`)

## Error Handling

**Pattern:**
- Use `fmt.Errorf` with `%w` for error wrapping (enables `errors.Unwrap`)
- Format: `fmt.Errorf("operation: %w", err)` or `fmt.Errorf("specific message: %w", err)`
- Include operation name for context

**Examples from codebase:**
```go
// From internal/repo/list.go:16
return nil, fmt.Errorf("create list: %w", err)

// From cmd/root.go:54
return fmt.Errorf("config: %w", err)

// From internal/db/db.go:15
return nil, fmt.Errorf("cannot create db directory: %w", err)
```

**Error messages:**
- Lowercase start (e.g. "list not found", "invalid id")
- Include the ID/identifier when relevant (e.g. `fmt.Errorf("list %d not found", id)`)
- For validation: clear, actionable messages (e.g. `fmt.Errorf("--due-time requires --due")`)

**Not-found pattern:**
```go
// From internal/repo/list.go:79-80
if n == 0 {
    return nil, fmt.Errorf("list %d not found", id)
}
```

## Logging

**Framework:** `fmt` package (standard library)

**Output destinations:**
- Normal output: `fmt.Println()` or `fmt.Printf()`
- Errors/warnings: `fmt.Fprintf(os.Stderr, ...)` (example from `cmd/root.go:66`)
- User feedback: via output package functions (`output.PrintSuccess`, `output.PrintTask`)

**Patterns:**
- Success messages via `output.PrintSuccess()` at `internal/output/output.go`
- Warnings to stderr (e.g. autocomplete warnings)
- No debug logging in current codebase

## Comments

**When to Comment:**
- Function comments for public functions (not consistently present; add when adding new functions)
- SQL comments for clarity on column meanings (e.g. `-- YYYY-MM-DD` inline in schema)
- Operational comments for non-obvious code paths (e.g. "Single writer to avoid WAL conflicts" at `internal/db/db.go:23`)

**JSDoc/TSDoc:**
- Not applicable (Go codebase)
- Package documentation comments above `package` statement (minimal usage)

**Example from `internal/db/db.go`:**
```go
// Single writer to avoid WAL conflicts; reads are concurrent
db.SetMaxOpenConns(1)
```

## Function Design

**Size:**
- Short, focused functions (most under 30 lines)
- Larger functions (50+ lines) are composite operations (e.g. task scheduling with nested logic)

**Parameters:**
- Database passed as first parameter: `func TaskCreate(db *sql.DB, in TaskInput) (*models.Task, error)`
- Input structs for multiple related parameters (e.g. `TaskInput`, `TaskListOptions`)
- Use pointers for optional/nullable values (e.g. `color *string`, `ParentTaskID *int64`)

**Return Values:**
- Primary return first, error last: `(*models.Task, error)` or `([]models.List, error)`
- Omit error wrapping on row errors (propagate raw) unless in command layer
- Single return type for simple getters (e.g. `ListAll() ([]models.List, error)`)

**Named vs Unnamed:**
- No named return values used in codebase; use bare returns minimally

## Module Design

**Exports:**
- All public repo functions exported: `ListCreate`, `TaskGet`, `TaskList`, etc.
- Internal models/types exported for use by other packages
- No barrel exports (cmd, internal packages don't re-export)

**Package responsibilities:**
- `cmd/` — CLI parsing and command handlers (uses global `DB`)
- `internal/db/` — SQLite schema and connection management
- `internal/repo/` — Pure CRUD operations (no side effects, no globals)
- `internal/models/` — Data structures (List, Task)
- `internal/config/` — Config file loading and platform paths
- `internal/output/` — Formatting and printing (respects `JSONMode` global)

**Cobra patterns:**
- Global `DB *sql.DB` declared in `cmd/root.go:16`
- `PersistentPreRunE` for setup (opens DB, runs migrations, processes autocompletion)
- Flag checking with `cmd.Flags().Changed("flag")` to distinguish "not provided" vs "provided with empty value"

**Optional flags pattern:**
```go
// From cmd/task.go:38-39
if cmd.Flags().Changed("notes") {
    in.Notes = &addNotes
}
```

## Type System

**Pointer usage:**
- Pointers for nullable fields: `Color *string`, `Notes *string`, `DueDate *string`
- Empty string `""` is different from `nil` for optional fields
- Clear nil semantics (e.g. empty string color clears it; nil means unchanged in patches)

**JSON marshaling:**
- Struct tags include `json:"field_name"` and `omitempty` for nullable fields
- Example from `internal/models/models.go:8`:
```go
Color     *string   `json:"color,omitempty"`
```

## Convention Summary

Go standard library conventions applied throughout:
- Package layout: `cmd/`, `internal/{db,config,repo,models,output}/`
- Error handling: `fmt.Errorf` with `%w` wrapping
- Naming: PascalCase for exports, camelCase for private
- Functions: short, focused, database-first parameter, error-last return
- Input validation: at CLI boundary (`cmd/`), not in repo
- No configuration files for formatting/linting (Go defaults)
