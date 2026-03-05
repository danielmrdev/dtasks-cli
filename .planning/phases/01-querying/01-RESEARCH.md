# Phase 1: Querying - Research

**Researched:** 2026-03-06
**Domain:** Go CLI (Cobra), SQLite filtering/sorting/search, pure-Go regex
**Confidence:** HIGH

## Summary

This phase adds three orthogonal capabilities to the existing `dtasks task ls` command and introduces a new top-level `dtasks find` command. All required functionality maps directly to SQL query construction and Cobra flag additions — no new dependencies are needed.

The codebase already has a working `TaskListOptions` struct with `DueToday bool` and a dynamic WHERE-clause builder in `repo.TaskList`. The pattern for this phase is simply extending that struct and its builder, then wiring new flags in `cmd/task.go`. For search, a new `cmd/find.go` and `repo.TaskSearch` function follow the identical pattern used by `TaskList`.

Sorting is the most significant design decision: the current `ORDER BY` clause is hardcoded in `taskSelectSQL`. It must move to runtime construction inside `TaskList`, driven by a `SortBy` + `Reverse` field in `TaskListOptions`.

**Primary recommendation:** Extend `TaskListOptions` for filters and sorting; add `TaskSearch` in repo for keyword/regex search; wire everything via Cobra flags — no new dependencies required.

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|-----------------|
| FILT-01 | `ls --today`: tasks due today or earlier | Extend `DueToday` semantics already in `TaskListOptions.DueToday`; rename flag from `--due-today` to `--today` |
| FILT-02 | `ls --overdue`: tasks past due date (strictly before today) | New `Overdue bool` field in `TaskListOptions`; SQL: `due_date < date('now','localtime')` |
| FILT-03 | `ls --tomorrow`: tasks due exactly tomorrow | New `DueTomorrow bool`; SQL: `due_date = date('now','localtime','+1 day')` |
| FILT-04 | `ls --week`: tasks due within next 7 days | New `DueWeek bool`; SQL: `due_date >= date('now','localtime') AND due_date <= date('now','localtime','+6 days')` |
| SORT-01 | `ls --sort=<field>`: sort by due, priority, created, completed | Add `SortBy string` to `TaskListOptions`; build ORDER BY dynamically; note: `priority` column doesn't exist yet — handle gracefully or add stub |
| SORT-02 | `ls --reverse`: reverse sort order | Add `Reverse bool`; append `DESC` to ORDER BY |
| SRCH-01 | `find <keyword>`: case-insensitive title/notes search | New `repo.TaskSearch`, new `cmd/find.go`; use SQLite `LIKE` with `%keyword%`; SQLite LIKE is case-insensitive for ASCII by default |
| SRCH-02 | `find --list <id>`: scope search to list | Add `ListID *int64` to `TaskSearchOptions` |
| SRCH-03 | `find --regex`: search with regex pattern | Use Go `regexp` stdlib in application layer after fetching candidates; SQLite pure-Go driver does not expose REGEXP function by default |
</phase_requirements>

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| `modernc.org/sqlite` | v1.29.0 | SQLite driver (pure Go) | Already in project; CGO_ENABLED=0 |
| `github.com/spf13/cobra` | v1.8.0 | CLI flags and subcommands | Already in project |
| `regexp` (stdlib) | Go 1.22 | Regex matching for SRCH-03 | stdlib; no external dep needed |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `strings` (stdlib) | Go 1.22 | Case-insensitive keyword fallback | Always available |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| SQLite `LIKE` for text search | SQLite FTS5 | FTS5 requires more schema work; LIKE is sufficient for small personal datasets |
| Go `regexp` post-fetch | SQLite REGEXP via custom function | Custom function requires C binding or runtime registration — contradicts CGO_ENABLED=0 goal; Go-side regex is simpler |
| New `--sort` flag | Rename existing ORDER BY | The hardcoded ORDER BY must be removed from `taskSelectSQL` const and moved to runtime — no alternative |

**Installation:** No new dependencies. All capabilities use existing libraries or stdlib.

## Architecture Patterns

### Recommended Project Structure

No new packages. Changes are additive within existing packages:

```
cmd/
├── task.go          # extend lsCmd flags; add --today, --overdue, --tomorrow, --week, --sort, --reverse
├── find.go          # NEW: findCmd top-level command (dtasks find <keyword>)
└── root.go          # register findCmd

internal/repo/
└── task.go          # extend TaskListOptions; extend TaskList(); add TaskSearch() + TaskSearchOptions
```

### Pattern 1: Extend TaskListOptions with filter and sort fields

**What:** Add fields to the existing options struct and extend the `WHERE 1=1` builder.
**When to use:** Always — the existing pattern is the right one; just extend it.

```go
// Source: existing internal/repo/task.go pattern
type TaskListOptions struct {
    ListID      *int64
    ParentID    *int64
    OnlyRoot    bool
    Completed   *bool
    DueToday    bool // FILT-01: due_date <= today
    Overdue     bool // FILT-02: due_date < today AND completed = 0
    DueTomorrow bool // FILT-03: due_date = tomorrow
    DueWeek     bool // FILT-04: due_date BETWEEN today AND today+6
    SortBy      string // SORT-01: "due" | "created" | "completed" — "priority" deferred
    Reverse     bool   // SORT-02
}
```

### Pattern 2: Dynamic ORDER BY construction

**What:** Remove the hardcoded `ORDER BY` from `taskSelectSQL` and build it in `TaskList`.
**When to use:** Required for SORT-01/SORT-02.

```go
// Source: adaptation of existing WHERE builder pattern
orderCol := map[string]string{
    "due":       "t.due_date, t.due_time",
    "created":   "t.created_at",
    "completed": "t.completed_at",
}

col, ok := orderCol[opts.SortBy]
if !ok {
    col = "t.due_date, t.due_time, t.created_at" // default
}
dir := "ASC"
if opts.Reverse {
    dir = "DESC"
}
query += " ORDER BY " + col + " " + dir
```

**Critical:** `taskSelectSQL` is a `const` — it must become a `var` or the ORDER BY must be appended by `TaskList`, not embedded in the constant.

### Pattern 3: TaskSearch with LIKE + optional Go-side regex

**What:** New function `TaskSearch(db, opts TaskSearchOptions)` using `LIKE '%keyword%'`.
**When to use:** SRCH-01, SRCH-02, SRCH-03.

```go
// Source: mirrors TaskList pattern
type TaskSearchOptions struct {
    Keyword string
    ListID  *int64
    Regex   bool
}

func TaskSearch(db *sql.DB, opts TaskSearchOptions) ([]models.Task, error) {
    query := taskSelectSQL + ` WHERE 1=1`
    args := []any{}

    if opts.ListID != nil {
        query += ` AND t.list_id = ?`
        args = append(args, *opts.ListID)
    }

    if !opts.Regex {
        // SQLite LIKE is case-insensitive for ASCII by default
        pattern := "%" + opts.Keyword + "%"
        query += ` AND (t.title LIKE ? OR t.notes LIKE ?)`
        args = append(args, pattern, pattern)
    }
    // If Regex: fetch with no keyword filter (or pre-filter with LIKE for perf),
    // then apply regexp.MatchString in Go

    query += ` ORDER BY t.due_date ASC, t.created_at ASC`
    // ... rest of query execution
}
```

### Pattern 4: New top-level `find` command

**What:** A root-level command `dtasks find <keyword>` (not a subcommand of `task`).
**When to use:** SRCH-01, SRCH-02, SRCH-03 — the spec says `dtasks find`, not `dtasks task find`.

```go
// Source: mirrors cmd/task.go command structure
var findCmd = &cobra.Command{
    Use:   "find <keyword>",
    Short: "Search tasks by keyword",
    Args:  cobra.ExactArgs(1),
    RunE: func(cmd *cobra.Command, args []string) error {
        // ...
    },
}
```

Register in `root.go`: `rootCmd.AddCommand(findCmd)`.

### Anti-Patterns to Avoid

- **Embedding ORDER BY in `taskSelectSQL` const:** The const is used by `TaskGet`, `TaskCreate`, etc. Appending ORDER BY in `TaskList` at runtime is the right approach; the const stays without ORDER BY.
- **Using SQLite REGEXP extension:** Requires CGO or complex driver configuration. Use Go stdlib `regexp` instead.
- **Applying regex before SQL LIKE filter:** For regex mode, first filter candidates with a broad LIKE (if keyword can be used as literal prefix/suffix) or fetch all matching-list tasks, then apply Go regex. Avoids full table scans while keeping correctness.
- **Exclusive filter flags without validation:** Flags `--today`, `--overdue`, `--tomorrow`, `--week` should be mutually exclusive or last-one-wins. Document the chosen behavior.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Date arithmetic ("today", "tomorrow") | Custom Go time math + format | SQLite `date('now','localtime','+N days')` | Already in DB; avoids timezone bugs in Go |
| Case-insensitive text search | `strings.ToLower` in Go on all titles | SQLite `LIKE '%x%'` | DB does it, no full scan in Go |
| Regex matching for SRCH-03 | Any custom string matching | `regexp.MustCompile` from stdlib | stdlib is well-tested, handles all edge cases |

**Key insight:** Push date filters into SQL; only pull data to Go when regex processing is unavoidable (SRCH-03).

## Common Pitfalls

### Pitfall 1: `taskSelectSQL` const ends without ORDER BY but callers append

**What goes wrong:** `TaskGet` uses `taskSelectSQL + " WHERE t.id = ?"` — if we add ORDER BY to the const, `TaskGet` gets a spurious ORDER BY that SQLite tolerates but is semantically wrong.
**Why it happens:** The const was designed for single-row fetches and multi-row fetches alike.
**How to avoid:** Keep the const without ORDER BY. Always append ORDER BY in `TaskList` and `TaskSearch` only.
**Warning signs:** Tests passing for TaskGet but linter flags unused ORDER BY.

### Pitfall 2: `--today` flag conflict with existing `--due-today`

**What goes wrong:** The current `lsCmd` already has `--due-today` flag (mapped to `DueToday: true`). Requirements call for `--today`. If we rename the flag, existing users break (though it's pre-release).
**Why it happens:** Flag naming churn between planning and implementation.
**How to avoid:** Replace `--due-today` with `--today` in the flag registration. Since this is v0.3.0 (not yet released as a stable API), breaking the old flag is acceptable.
**Warning signs:** `lsCmd.Flags().BoolVar(&lsDueToday, "due-today", ...)` still present after implementation.

### Pitfall 3: SQLite date functions with localtime

**What goes wrong:** `date('now')` returns UTC. Tasks due "today" in a UTC-8 timezone may appear wrong.
**Why it happens:** SQLite `now` is UTC by default.
**How to avoid:** Use `date('now','localtime')` consistently — the same pattern already used in `recur_scheduler.go` and `TaskDone`.
**Warning signs:** Failing filter tests run near midnight in non-UTC timezones.

### Pitfall 4: SORT-01 references `priority` field

**What goes wrong:** Requirements specify `--sort=priority` but `priority` column does not exist in the current schema (it belongs to Phase 2).
**Why it happens:** Phase 2 adds priority; Phase 1 must handle the flag gracefully.
**How to avoid:** Accept `priority` as a valid sort value but fall back to `created_at` order (or return an error "priority not yet available"). Document the chosen behavior. Alternatively, simply omit `priority` from the allowed values in Phase 1 and add it in Phase 2.
**Warning signs:** Planner adds a `priority` column to the schema in this phase — that's Phase 2 scope.

### Pitfall 5: Regex flag `--regex` without keyword

**What goes wrong:** `find --regex` with no positional argument panics or returns all tasks.
**Why it happens:** `cobra.ExactArgs(1)` should prevent this, but regex compilation of empty string succeeds.
**How to avoid:** `cobra.ExactArgs(1)` already handles missing arg. Also validate that the regex compiles before querying: `regexp.Compile(args[0])` and return an error if it fails.
**Warning signs:** Test `find --regex '['` should return an error, not crash.

## Code Examples

Verified patterns from existing codebase:

### Date filter in SQL (existing DueToday pattern)
```go
// Source: internal/repo/task.go:73-77
if opts.DueToday {
    today := time.Now().Format("2006-01-02")
    query += ` AND t.due_date <= ?`
    args = append(args, today)
}
```

New filters follow identical pattern:
```go
// FILT-02: Overdue (strictly before today, not completed)
if opts.Overdue {
    today := time.Now().Format("2006-01-02")
    query += ` AND t.due_date < ?`
    args = append(args, today)
}

// FILT-03: Tomorrow
if opts.DueTomorrow {
    tomorrow := time.Now().AddDate(0, 0, 1).Format("2006-01-02")
    query += ` AND t.due_date = ?`
    args = append(args, tomorrow)
}

// FILT-04: This week (today through today+6)
if opts.DueWeek {
    today := time.Now().Format("2006-01-02")
    week := time.Now().AddDate(0, 0, 6).Format("2006-01-02")
    query += ` AND t.due_date >= ? AND t.due_date <= ?`
    args = append(args, today, week)
}
```

### Dynamic ORDER BY
```go
// After all WHERE clauses, before db.Query()
sortMap := map[string]string{
    "due":       "t.due_date ASC, t.due_time ASC, t.created_at ASC",
    "created":   "t.created_at",
    "completed": "t.completed_at",
}
orderExpr, ok := sortMap[opts.SortBy]
if !ok {
    orderExpr = "t.due_date ASC, t.due_time ASC, t.created_at ASC"
}
if opts.Reverse {
    // Replace trailing ASC with DESC; simpler: just use separate direction
    orderExpr = strings.ReplaceAll(orderExpr, " ASC", " DESC")
}
query += " ORDER BY " + orderExpr
```

### Regex search (Go-side, SRCH-03)
```go
// Source: stdlib regexp pattern
re, err := regexp.Compile("(?i)" + opts.Keyword) // (?i) = case-insensitive
if err != nil {
    return nil, fmt.Errorf("invalid regex: %w", err)
}
// After fetching all tasks (no SQL LIKE filter when opts.Regex):
var matched []models.Task
for _, t := range tasks {
    if re.MatchString(t.Title) || (t.Notes != nil && re.MatchString(*t.Notes)) {
        matched = append(matched, t)
    }
}
return matched, nil
```

### Cobra flag registration (find command)
```go
// Source: mirrors cmd/task.go init() pattern
func init() {
    findCmd.Flags().Int64VarP(&findListID, "list", "l", 0, "Scope to list ID")
    findCmd.Flags().BoolVar(&findRegex, "regex", false, "Treat keyword as regex pattern")
    _ = findCmd.RegisterFlagCompletionFunc("list", completeLists)
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Hardcoded `ORDER BY` in `taskSelectSQL` | Dynamic ORDER BY in `TaskList` | This phase | Enables SORT-01/SORT-02 |
| `--due-today` flag | `--today` flag (rename) | This phase | Matches spec naming |
| No search command | `dtasks find <keyword>` | This phase | SRCH-01/02/03 |

**Deprecated/outdated:**
- `--due-today` flag: replaced by `--today` in this phase (pre-release, no compatibility concern)
- Hardcoded `ORDER BY t.due_date ASC, t.due_time ASC, t.created_at ASC` at end of `taskSelectSQL` const: must be moved to runtime in `TaskList`

## Open Questions

1. **Mutual exclusivity of filter flags**
   - What we know: `--today`, `--overdue`, `--tomorrow`, `--week` all filter `due_date`
   - What's unclear: Should combining two (e.g., `--today --week`) be an error or last-one-wins?
   - Recommendation: Last-one-wins (simpler). Document in `--help` text. Cobra doesn't enforce mutual exclusivity automatically without `MarkFlagsMutuallyExclusive`.

2. **`--sort=priority` in Phase 1**
   - What we know: `priority` column is Phase 2
   - What's unclear: Accept and silently ignore, or return an error?
   - Recommendation: Omit `priority` from valid sort values in Phase 1 entirely. Phase 2 adds it to the allowed list when the column exists. Avoids silent no-ops.

3. **`--overdue` vs `--today` overlap**
   - What we know: `--today` shows `due_date <= today` (includes overdue); `--overdue` should show `due_date < today`
   - What's unclear: Whether `--overdue` should also exclude completed tasks implicitly
   - Recommendation: `--overdue` adds a date filter only; the existing `opts.Completed` filter (default: pending only) already handles excluding completed tasks. No special casing needed.

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go testing stdlib (go test) |
| Config file | none — standard `go test ./...` |
| Quick run command | `go test ./internal/repo/... -run TestTaskList -v` |
| Full suite command | `go test ./...` |

### Phase Requirements → Test Map

| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| FILT-01 | `--today` shows tasks with `due_date <= today` | unit | `go test ./internal/repo/... -run TestTaskList_FilterToday -v` | ❌ Wave 0 |
| FILT-02 | `--overdue` shows tasks with `due_date < today` | unit | `go test ./internal/repo/... -run TestTaskList_FilterOverdue -v` | ❌ Wave 0 |
| FILT-03 | `--tomorrow` shows tasks due exactly tomorrow | unit | `go test ./internal/repo/... -run TestTaskList_FilterTomorrow -v` | ❌ Wave 0 |
| FILT-04 | `--week` shows tasks due today through +6 days | unit | `go test ./internal/repo/... -run TestTaskList_FilterWeek -v` | ❌ Wave 0 |
| SORT-01 | `--sort=due/created/completed` orders correctly | unit | `go test ./internal/repo/... -run TestTaskList_Sort -v` | ❌ Wave 0 |
| SORT-02 | `--reverse` inverts sort order | unit | `go test ./internal/repo/... -run TestTaskList_SortReverse -v` | ❌ Wave 0 |
| SRCH-01 | `find <keyword>` matches title and notes, case-insensitive | unit | `go test ./internal/repo/... -run TestTaskSearch_Keyword -v` | ❌ Wave 0 |
| SRCH-02 | `find --list <id>` scopes search to list | unit | `go test ./internal/repo/... -run TestTaskSearch_List -v` | ❌ Wave 0 |
| SRCH-03 | `find --regex` matches regex, invalid regex returns error | unit | `go test ./internal/repo/... -run TestTaskSearch_Regex -v` | ❌ Wave 0 |

### Sampling Rate
- **Per task commit:** `go test ./internal/repo/... -v`
- **Per wave merge:** `go test ./...`
- **Phase gate:** `go test ./...` fully green before `/gsd:verify-work`

### Wave 0 Gaps
- [ ] `internal/repo/repo_test.go` — add tests for new `TaskListOptions` fields (FILT-01 through SORT-02)
- [ ] `internal/repo/repo_test.go` — add `TestTaskSearch_*` tests (SRCH-01 through SRCH-03)

*(All new tests go in the existing `repo_test.go` file, following the `openTestDB` pattern already established.)*

## Sources

### Primary (HIGH confidence)
- Existing codebase: `internal/repo/task.go`, `cmd/task.go`, `internal/db/db.go`, `internal/models/models.go` — direct inspection
- Go stdlib docs: `regexp` package — well-known, no external source needed
- SQLite docs (date functions): `date('now','localtime','+N days')` pattern already used in `repo/recur_scheduler.go`

### Secondary (MEDIUM confidence)
- SQLite LIKE behavior: case-insensitive for ASCII by default (well-known SQLite property, consistent with modernc.org/sqlite pure-Go implementation)
- Cobra `MarkFlagsMutuallyExclusive`: available in cobra v1.8.0 (confirmed by cobra changelog)

### Tertiary (LOW confidence)
- None — all findings are based on direct codebase inspection or stdlib

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — project already uses all needed libraries; zero new dependencies
- Architecture: HIGH — patterns are direct extensions of existing code; no speculative design
- Pitfalls: HIGH — identified from direct codebase reading, not speculation
- Validation: HIGH — existing test infrastructure and patterns are clear

**Research date:** 2026-03-06
**Valid until:** 2026-06-06 (stable stack; Cobra and SQLite driver versions locked in go.mod)
