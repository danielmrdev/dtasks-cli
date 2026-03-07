# Phase 2: Richness - Research

**Researched:** 2026-03-06
**Domain:** Go CLI (Cobra), SQLite schema migration, priority ordering, bulk DELETE, aggregate stats queries
**Confidence:** HIGH

## Summary

Phase 2 adds three independent capabilities on top of the Phase 1 foundation: task priorities, bulk maintenance (bulk delete of completed tasks), and usage statistics. All three map to well-defined SQL operations and Cobra flag additions. No new external dependencies are required.

Priority requires a schema migration (new `priority` column via `ALTER TABLE`), model extension, and changes to `TaskInput`, `TaskPatch`, `TaskListOptions` (for `--sort=priority`), and `output.PrintTasks`. The existing idempotent migration pattern (check with `pragma_table_info`, then `ALTER TABLE`) is established and must be reused exactly.

Bulk delete (`dtasks rm --completed <date>`) extends the existing `rmCmd` with new flags (`--completed`, `--dry-run`, `--yes`, `--list`). It requires a new repo function `TaskDeleteCompleted` that accepts the date and optional list scope. Dry-run preview lists affected tasks without deleting; confirmation prompt reads from stdin; `--json` emits `{"deleted": N}`.

Stats (`dtasks stats`) is a new top-level command with a new repo function `TaskStats` that runs a single GROUP BY query per list. JSON output follows the established `output.JSONMode` pattern.

**Primary recommendation:** Extend schema and repo with additive migrations only; extend existing command structs for priority and bulk-rm; add `statsCmd` as a new top-level command — no new dependencies required.

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|-----------------|
| PRIO-01 | `add --priority high\|medium\|low` sets task priority | Add `Priority *string` to `TaskInput`; add `priority` column via migration; validate enum at CLI layer |
| PRIO-02 | `edit --priority high\|medium\|low` sets task priority | Add `Priority *string` to `TaskPatch`; extend `TaskPatchFields` dynamic builder |
| PRIO-03 | Priority shown as visual indicator in table output | Add `Priority` field to `models.Task`; extend `PrintTasks` and `PrintTask` with indicator symbol |
| PRIO-04 | `ls --sort=priority` sorts tasks by priority | Add `"priority"` to `sortMap` in `TaskList` using `CASE` expression for ordering |
| MAINT-01 | `rm --completed <date>` bulk-deletes completed tasks on or before date | New `TaskDeleteCompleted(db, date, listID)` repo function; new flags on `rmCmd` |
| MAINT-02 | `--dry-run` previews without deleting | `TaskDeleteCompleted` returns affected tasks without executing DELETE when dryRun=true |
| MAINT-03 | Bulk delete requires confirmation unless `--yes` | Stdin prompt in `rmCmd` handler; skip if `--yes` flag set |
| MAINT-04 | Bulk delete respects `--json` flag | Emit `{"deleted": N}` via existing `output.JSONMode` pattern |
| MAINT-05 | Bulk delete can be scoped with `--list <id>` | Pass optional `*int64` listID to `TaskDeleteCompleted` |
| STAT-01 | `dtasks stats` shows total, pending, done, % per list | New `statsCmd`; new `TaskStats(db)` repo function using GROUP BY |
| STAT-02 | `stats --json` outputs structured JSON | Reuse `output.JSONMode` / `output.printJSON` pattern |
</phase_requirements>

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| `modernc.org/sqlite` | v1.29.0 | SQLite driver (pure Go) | Already in project; all SQL operations |
| `github.com/spf13/cobra` | v1.8.0 | CLI flags and subcommands | Already in project |
| `bufio` + `os.Stdin` (stdlib) | Go 1.22 | Confirmation prompt for MAINT-03 | stdlib; no external dep needed |
| `fmt` (stdlib) | Go 1.22 | Table output of stats | Already used throughout |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `strings` (stdlib) | Go 1.22 | String building for SQL CASE expression | Minimal use |
| `strconv` (stdlib) | Go 1.22 | Format stats percentages | Already used in output.go |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| `priority TEXT` column (high/medium/low) | `priority INTEGER` (1/2/3) | TEXT is readable in SQLite Browser; ordering uses CASE expression; INTEGER is simpler to sort but less debuggable |
| Confirmation prompt via `bufio.Scanner` | `survey` or `promptui` library | External deps for a single yes/no question is over-engineering; stdlib is sufficient |
| Single `TaskStats` query | Multiple queries (one per list) | Single GROUP BY query is more efficient and simpler to maintain |

**Installation:** No new dependencies. All capabilities use existing libraries or stdlib.

## Architecture Patterns

### Recommended Project Structure

No new packages. Changes are additive within existing packages:

```
cmd/
├── task.go          # extend addCmd/editCmd with --priority; extend rmCmd with --completed/--dry-run/--yes
├── stats.go         # NEW: statsCmd top-level command
└── root.go          # register statsCmd

internal/
├── models/models.go        # add Priority *string field to Task
├── repo/task.go            # extend TaskInput, TaskPatch, TaskListOptions; add TaskDeleteCompleted, TaskStats
├── output/output.go        # extend PrintTasks/PrintTask with priority indicator; add PrintStats
└── db/db.go                # add migration for priority column
```

### Pattern 1: Idempotent schema migration for priority column

**What:** Add `priority TEXT` column to the `tasks` table using the established `pragma_table_info` guard.
**When to use:** Required for PRIO-01 through PRIO-04.

```go
// Source: existing internal/db/db.go migrate() pattern
var pCount int
db.QueryRow(`SELECT COUNT(*) FROM pragma_table_info('tasks') WHERE name='priority'`).Scan(&pCount)
if pCount == 0 {
    if _, err := db.Exec(`ALTER TABLE tasks ADD COLUMN priority TEXT`); err != nil {
        return fmt.Errorf("migrate priority column: %w", err)
    }
}
```

**Note:** `priority` has no NOT NULL constraint and no DEFAULT — existing rows will have `NULL` priority, which is the correct "unset" state and sorts last when using `CASE` ordering.

### Pattern 2: Priority sort using CASE expression

**What:** Priority is a TEXT enum (`high`, `medium`, `low`, `NULL`). SQL cannot sort these lexicographically in the desired order. Use a CASE expression.
**When to use:** Required for PRIO-04 (`ls --sort=priority`).

```go
// Source: adaptation of existing sortMap pattern in internal/repo/task.go
sortMap := map[string]string{
    "due":      "t.due_date ASC, t.due_time ASC, t.created_at ASC",
    "created":  "t.created_at ASC",
    "completed":"t.completed_at ASC",
    "priority": "CASE t.priority WHEN 'high' THEN 1 WHEN 'medium' THEN 2 WHEN 'low' THEN 3 ELSE 4 END ASC, t.created_at ASC",
}
```

NULL priority sorts last (ELSE 4) when ascending. `--reverse` applies `strings.ReplaceAll(expr, " ASC", " DESC")` as already done, but CASE + DESC puts NULL first — this is acceptable behavior.

### Pattern 3: Priority visual indicator in table output

**What:** Add a PRIO column to the task table with a short symbol.
**When to use:** Required for PRIO-03.

Proposed symbols:
- `high` → `!` (or `H`, or `▲`) — ASCII-safe, unambiguous
- `medium` → `~` (or `M`, or `▼`)
- `low` → `-` (or `L`)
- `NULL` → ` ` (empty)

The symbol must be computed from the `plain` string (for column width) and the `styled` string can use ANSI color if desired. Given the project already uses ANSI for `colorDot`, color-coding priority is viable but optional for PRIO-03.

```go
// Source: mirrors existing done/ac indicator pattern in output.PrintTasks
prio := " "
switch {
case t.Priority != nil && *t.Priority == "high":   prio = "!"
case t.Priority != nil && *t.Priority == "medium":  prio = "~"
case t.Priority != nil && *t.Priority == "low":     prio = "-"
}
```

Add `"PRIO"` to headers and `prio` to each row in `PrintTasks`.

### Pattern 4: TaskDeleteCompleted repo function

**What:** A repo function that finds and optionally deletes completed tasks on or before a given date, optionally scoped to a list.
**When to use:** Required for MAINT-01 through MAINT-05.

```go
// Source: mirrors TaskDelete pattern; new function in internal/repo/task.go
type DeleteCompletedOptions struct {
    Before string  // YYYY-MM-DD, inclusive
    ListID *int64  // optional scope
    DryRun bool    // if true, return tasks without deleting
}

type DeleteCompletedResult struct {
    Tasks   []models.Task // tasks that would be/were deleted
    Deleted int           // count of deleted rows (0 if DryRun)
}

func TaskDeleteCompleted(db *sql.DB, opts DeleteCompletedOptions) (DeleteCompletedResult, error) {
    // 1. Build SELECT query (same WHERE regardless of DryRun)
    // 2. If DryRun: return tasks, Deleted=0
    // 3. Else: execute DELETE, return count
}
```

The SELECT and DELETE share the same WHERE clause: `completed = 1 AND completed_at <= ?` with optional `AND list_id = ?`. Return the task list for the dry-run preview and the count for the actual delete confirmation output.

### Pattern 5: Confirmation prompt for MAINT-03

**What:** Read a single yes/no from stdin before executing the delete.
**When to use:** Required for MAINT-03. Skipped when `--yes` flag is set or when `--dry-run`.

```go
// Source: stdlib bufio pattern
func confirmPrompt(msg string) (bool, error) {
    fmt.Print(msg + " [y/N]: ")
    scanner := bufio.NewScanner(os.Stdin)
    if scanner.Scan() {
        answer := strings.TrimSpace(strings.ToLower(scanner.Text()))
        return answer == "y" || answer == "yes", nil
    }
    return false, scanner.Err()
}
```

Default answer is NO (user must explicitly type `y` or `yes`). This is the standard CLI convention for destructive operations.

### Pattern 6: TaskStats repo function

**What:** A single GROUP BY query that returns per-list counts of total, pending, done tasks.
**When to use:** Required for STAT-01 and STAT-02.

```go
// Source: new function; follows repo pure-function pattern
type ListStat struct {
    ListID   int64   `json:"list_id"`
    ListName string  `json:"list_name"`
    Total    int     `json:"total"`
    Pending  int     `json:"pending"`
    Done     int     `json:"done"`
    PctDone  float64 `json:"pct_done"`
}

type StatsSummary struct {
    Total   int        `json:"total"`
    Pending int        `json:"pending"`
    Done    int        `json:"done"`
    PctDone float64    `json:"pct_done"`
    ByList  []ListStat `json:"by_list"`
}

func TaskStats(db *sql.DB) (*StatsSummary, error) {
    rows, err := db.Query(`
        SELECT
            l.id, l.name,
            COUNT(*) AS total,
            SUM(CASE WHEN t.completed = 0 THEN 1 ELSE 0 END) AS pending,
            SUM(CASE WHEN t.completed = 1 THEN 1 ELSE 0 END) AS done
        FROM tasks t
        JOIN lists l ON t.list_id = l.id
        WHERE t.parent_task_id IS NULL
        GROUP BY l.id, l.name
        ORDER BY l.name ASC
    `)
    // ... scan, compute PctDone = Done/Total*100, aggregate totals
}
```

**Note:** `WHERE t.parent_task_id IS NULL` restricts stats to root tasks (same as `OnlyRoot: true` in `TaskList`). This is consistent with the existing default listing behavior. Including subtasks in stats would inflate counts and is misleading for a "task summary" view.

### Anti-Patterns to Avoid

- **Storing priority as INTEGER:** TEXT enum is more readable and already validated at the CLI layer; INTEGER would require a lookup table or magic numbers in every display path.
- **New command `task bulk-rm`:** MAINT-01 extends the existing `rm` command with `--completed` flag, not a new subcommand. The spec says `dtasks rm --completed <date>`.
- **Interactive confirmation in JSON mode:** When `--json` is active, skip the confirmation prompt entirely and execute the delete (or honor `--yes`). Mixing interactive prompts with JSON output breaks scripting.
- **Computing stats in Go by fetching all tasks:** Use SQL aggregates. Fetching all tasks to count in Go wastes memory and is O(n) in the application layer for something SQL handles natively.
- **NULL priority sorting without CASE:** `ORDER BY t.priority ASC` sorts alphabetically: `high` > `low` > `medium` > NULL — completely wrong. Always use the CASE expression.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Priority ordering | Custom sort in Go after fetch | SQL `CASE` expression in ORDER BY | DB sorts before sending data; avoids loading all rows |
| Count/percentage computation | Multiple queries, one per list | Single `GROUP BY` query with `SUM(CASE ...)` | One round-trip; DB aggregation is correct and fast |
| Yes/no prompt | External library (`survey`, `promptui`) | `bufio.Scanner` on `os.Stdin` | Single use case; no dep needed |
| Schema migration for new column | Drop and recreate table | `ALTER TABLE ADD COLUMN` guarded by `pragma_table_info` | Safe for existing data; established pattern in codebase |
| Priority validation | Database CHECK constraint | CLI-layer validation before insert | Consistent with project convention (validate at boundary); pure-Go SQLite driver supports CHECK but adds complexity |

**Key insight:** SQL aggregates and CASE expressions handle all priority ordering and stats computation correctly at zero cost. Only confirmation prompts and priority validation require application-layer code.

## Common Pitfalls

### Pitfall 1: completed_at vs completed_at date comparison

**What goes wrong:** `completed_at` is stored as `DATETIME` (e.g., `2026-03-06T14:30:00Z` or `2026-03-06 14:30:00`). Comparing `completed_at <= '2026-03-06'` with a date string may exclude tasks completed on that date due to the time component.
**Why it happens:** SQLite DATETIME string comparison is lexicographic. `'2026-03-06 23:59:59' > '2026-03-06'` is true, so the task would be excluded.
**How to avoid:** Use `date(completed_at) <= ?` in the WHERE clause to strip the time component before comparison. This is safe for both ISO8601 and SQLite datetime formats.

```sql
WHERE completed = 1 AND date(completed_at) <= ?
```

**Warning signs:** Tasks completed late in the day on the cutoff date are not shown in dry-run preview.

### Pitfall 2: Priority validation must be at CLI layer, not repo layer

**What goes wrong:** An invalid priority value (e.g., `"urgent"`) gets inserted into the DB silently if validation is in the repo. Later queries using CASE expressions return 4 (other) for unknown values — no error, but incorrect behavior.
**Why it happens:** The repo layer in this project trusts internal contracts (see CLAUDE.md: "Validate only at system boundaries").
**How to avoid:** Validate `--priority` values in `addCmd` and `editCmd` `RunE` functions before calling repo functions. Return an error for any value not in `{"high", "medium", "low"}`.

```go
validPriorities := map[string]bool{"high": true, "medium": true, "low": true}
if !validPriorities[priority] {
    return fmt.Errorf("invalid priority %q: must be high, medium, or low", priority)
}
```

### Pitfall 3: Dry-run in JSON mode shows tasks, not count

**What goes wrong:** `--dry-run --json` should emit the list of tasks that would be deleted, not `{"deleted": 0}`. The requirement says "previews without deleting" — the useful JSON output is the task list.
**Why it happens:** Mixing dry-run and JSON output modes creates an ambiguous contract.
**How to avoid:** When `--dry-run` is active: output the tasks list (same as `output.PrintTasks`). When deleting (no `--dry-run`): output `{"deleted": N}`. Document this in command `--help`.

### Pitfall 4: Confirmation prompt breaks piping / scripting

**What goes wrong:** `dtasks rm --completed 2026-03-01 | other-command` hangs waiting for stdin confirmation.
**Why it happens:** The confirmation prompt reads from `os.Stdin` which may be a pipe.
**How to avoid:** Check if stdin is a TTY before prompting. If not a TTY and `--yes` is not set, abort with an error: `"use --yes to confirm non-interactive bulk delete"`.

```go
// Detection pattern
func isTerminal() bool {
    info, err := os.Stdin.Stat()
    if err != nil { return false }
    return (info.Mode() & os.ModeCharDevice) != 0
}
```

### Pitfall 5: taskSelectSQL does not include priority column

**What goes wrong:** `scanTaskRow` scans a fixed number of columns from `taskSelectSQL`. Adding `priority` to the schema without updating the SELECT and Scan causes a panic or incorrect column mapping.
**Why it happens:** The SELECT column list and the `scanTaskRow` Scan call are coupled. Adding a DB column is not sufficient — both must be updated together.
**How to avoid:** Update `taskSelectSQL` to include `t.priority`, add `Priority *string` to `models.Task`, and add `&t.Priority` to the `Scan` call in `scanTaskRow`. All three changes must happen atomically in the same commit.

### Pitfall 6: Stats query excludes lists with zero tasks

**What goes wrong:** A list with no tasks is absent from the GROUP BY result. If the user expects to see all lists (even empty ones), they'll be confused.
**Why it happens:** `JOIN` excludes rows with no match. `LEFT JOIN` with `tasks` being the driving table won't help here since we're grouping by list.
**How to avoid:** Use a LEFT JOIN with lists as the driving table: `FROM lists l LEFT JOIN tasks t ON t.list_id = l.id`. This includes lists with zero tasks (all counts = 0). This is the more informative output for a stats view.

```sql
SELECT
    l.id, l.name,
    COUNT(t.id) AS total,
    SUM(CASE WHEN t.completed = 0 AND t.id IS NOT NULL THEN 1 ELSE 0 END) AS pending,
    SUM(CASE WHEN t.completed = 1 THEN 1 ELSE 0 END) AS done
FROM lists l
LEFT JOIN tasks t ON t.list_id = l.id AND t.parent_task_id IS NULL
GROUP BY l.id, l.name
ORDER BY l.name ASC
```

### Pitfall 7: CASE DESC reversal makes NULL priority first

**What goes wrong:** `--sort=priority --reverse` with `DESC` applied to the CASE expression makes ELSE 4 sort first (DESC: 4 > 3 > 2 > 1), so NULL-priority tasks appear before low/medium/high.
**Why it happens:** The existing `strings.ReplaceAll(expr, " ASC", " DESC")` trick applies DESC to the numeric result of CASE, reversing the NULL-last behavior.
**How to avoid:** Document this as intentional: reversed priority sort puts "no priority" tasks first, which can be useful for finding unclassified tasks. Alternatively, use a separate NULLS LAST clause — but SQLite does not support `NULLS LAST` in all versions. Since NULL-priority reversed to first is a reasonable semantic, accept it and document it.

## Code Examples

Verified patterns from existing codebase:

### Migration guard for new column
```go
// Source: internal/db/db.go:99-105 (autocomplete column pattern)
var pCount int
db.QueryRow(`SELECT COUNT(*) FROM pragma_table_info('tasks') WHERE name='priority'`).Scan(&pCount)
if pCount == 0 {
    if _, err := db.Exec(`ALTER TABLE tasks ADD COLUMN priority TEXT`); err != nil {
        return fmt.Errorf("migrate priority column: %w", err)
    }
}
```

### TaskPatch extension for priority
```go
// Source: mirrors internal/repo/task.go TaskPatch pattern
type TaskPatch struct {
    Title        *string
    Notes        *string
    DueDate      *string
    DueTime      *string
    ListID       *int64
    Autocomplete *bool
    Priority     *string // NEW: PRIO-02
}

// In TaskPatchFields dynamic builder:
if p.Priority != nil {
    if *p.Priority == "" {
        add("priority", nil) // clear priority
    } else {
        add("priority", *p.Priority)
    }
}
```

### Priority sort in TaskList
```go
// Source: extends internal/repo/task.go sortMap
sortMap := map[string]string{
    "due":       "t.due_date ASC, t.due_time ASC, t.created_at ASC",
    "created":   "t.created_at ASC",
    "completed": "t.completed_at ASC",
    "priority":  "CASE t.priority WHEN 'high' THEN 1 WHEN 'medium' THEN 2 WHEN 'low' THEN 3 ELSE 4 END ASC, t.created_at ASC",
}
```

### TaskDeleteCompleted with date(completed_at) comparison
```go
// Source: new function; date() strips time component for correct comparison
query := `SELECT ` + taskSelectSQL[strings.Index(taskSelectSQL, "\n"):] // reuse select
whereClause := ` WHERE t.completed = 1 AND date(t.completed_at) <= ?`
args := []any{opts.Before}
if opts.ListID != nil {
    whereClause += ` AND t.list_id = ?`
    args = append(args, *opts.ListID)
}
// Execute SELECT for preview, then DELETE if not DryRun
```

### Stats GROUP BY query with LEFT JOIN
```go
// Source: new function following repo pure-function pattern
const statsSQL = `
    SELECT
        l.id, l.name,
        COUNT(t.id) AS total,
        SUM(CASE WHEN t.completed = 0 AND t.id IS NOT NULL THEN 1 ELSE 0 END) AS pending,
        SUM(CASE WHEN t.completed = 1 THEN 1 ELSE 0 END) AS done
    FROM lists l
    LEFT JOIN tasks t ON t.list_id = l.id AND t.parent_task_id IS NULL
    GROUP BY l.id, l.name
    ORDER BY l.name ASC
`
```

### Confirmation prompt with TTY detection
```go
// Source: stdlib os.Stdin.Stat() pattern
func isTerminal(f *os.File) bool {
    info, err := f.Stat()
    if err != nil {
        return false
    }
    return (info.Mode() & os.ModeCharDevice) != 0
}

// In rmCmd handler (MAINT-03):
if !rmYes && !rmDryRun {
    if !isTerminal(os.Stdin) {
        return fmt.Errorf("bulk delete requires --yes in non-interactive mode")
    }
    // Show preview, then prompt
    fmt.Printf("This will delete %d tasks. Confirm? [y/N]: ", len(result.Tasks))
    // ... read answer
}
```

### JSON output for bulk delete (MAINT-04)
```go
// Source: mirrors output.PrintSuccess pattern
if output.JSONMode {
    output.printJSON(map[string]any{"deleted": result.Deleted})
    return nil
}
fmt.Printf("Deleted %d tasks.\n", result.Deleted)
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| No priority field | `priority TEXT` column (high/medium/low/NULL) | This phase | Schema migration; affects scan, insert, update |
| No bulk delete | `rm --completed <date>` with dry-run and confirmation | This phase | New flags on existing `rmCmd` |
| No stats command | `dtasks stats` with per-list breakdown | This phase | New top-level command + new repo function |
| `--sort=priority` not in sortMap | `"priority"` added to sortMap with CASE expression | This phase | Unblocks SORT-01 promise from Phase 1 |

**Deprecated/outdated:**
- Phase 1 research noted `--sort=priority` was deferred. Phase 2 fully implements it.
- `taskSelectSQL` const: must be updated to include `t.priority` in column list.

## Open Questions

1. **Priority indicator style in table output**
   - What we know: PRIO-03 says "visual indicator" but does not prescribe symbol or color
   - What's unclear: ASCII symbol only, or ANSI colored? Single char or text?
   - Recommendation: Use single ASCII character (`!` / `~` / `-` / ` `) for portability. ANSI color (red for high, yellow for medium, default for low) is optional enhancement. Decision is Claude's discretion.

2. **Stats: include subtasks or root tasks only?**
   - What we know: `TaskList` defaults to `OnlyRoot: true` for the standard `ls` view
   - What's unclear: Should `stats` count subtasks separately or ignore them?
   - Recommendation: Count only root tasks (`parent_task_id IS NULL`) for consistency with the primary task view. Document this in `--help`.

3. **`rm --completed` collision with existing `rm <id>` positional arg**
   - What we know: `rmCmd` currently requires `cobra.ExactArgs(1)` for task ID
   - What's unclear: How to make `--completed` and the positional arg mutually exclusive
   - Recommendation: Change `Args` to `cobra.RangeArgs(0, 1)` and validate in `RunE`: if `--completed` is set, no positional arg is allowed (and vice versa). Return a clear error if both are provided.

4. **Clearing priority (setting back to NULL)**
   - What we know: `TaskPatch` uses `*string` pointers; `nil` = "not changed", `&""` could mean "clear"
   - What's unclear: Should `edit --priority ""` clear the priority?
   - Recommendation: Accept an explicit `--priority ""` (or `--priority none`) to clear priority. In `TaskPatchFields`, store `NULL` when the value is empty string. CLI flag should use `Changed()` check to detect explicit set.

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go testing stdlib (go test) |
| Config file | none — standard `go test ./...` |
| Quick run command | `go test ./internal/repo/... -run TestTask -v` |
| Full suite command | `go test ./...` |

### Phase Requirements -> Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| PRIO-01 | `TaskCreate` with priority stores correct value | unit | `go test ./internal/repo/... -run TestTaskCreate_WithPriority -v` | ❌ Wave 0 |
| PRIO-02 | `TaskPatchFields` with Priority updates correctly | unit | `go test ./internal/repo/... -run TestTaskPatchFields_Priority -v` | ❌ Wave 0 |
| PRIO-03 | `PrintTasks` shows priority indicator in output | unit | `go test ./internal/output/... -run TestPrintTasks_Priority -v` | ❌ Wave 0 |
| PRIO-04 | `TaskList` with `SortBy="priority"` orders high>medium>low>nil | unit | `go test ./internal/repo/... -run TestTaskList_SortPriority -v` | ❌ Wave 0 |
| MAINT-01 | `TaskDeleteCompleted` deletes tasks on or before date | unit | `go test ./internal/repo/... -run TestTaskDeleteCompleted -v` | ❌ Wave 0 |
| MAINT-02 | `TaskDeleteCompleted` with DryRun=true returns tasks without deleting | unit | `go test ./internal/repo/... -run TestTaskDeleteCompleted_DryRun -v` | ❌ Wave 0 |
| MAINT-03 | Confirmation prompt skipped with --yes | manual | `dtasks rm --completed 2026-01-01 --yes` | N/A |
| MAINT-04 | `--json` emits `{"deleted": N}` | unit | `go test ./internal/output/... -run TestPrintDeletedJSON -v` | ❌ Wave 0 |
| MAINT-05 | `TaskDeleteCompleted` with ListID scopes correctly | unit | `go test ./internal/repo/... -run TestTaskDeleteCompleted_Scoped -v` | ❌ Wave 0 |
| STAT-01 | `TaskStats` returns correct totals and per-list breakdown | unit | `go test ./internal/repo/... -run TestTaskStats -v` | ❌ Wave 0 |
| STAT-02 | Stats JSON output is structured correctly | unit | `go test ./internal/output/... -run TestPrintStats_JSON -v` | ❌ Wave 0 |

### Sampling Rate
- **Per task commit:** `go test ./internal/repo/... -v`
- **Per wave merge:** `go test ./...`
- **Phase gate:** `go test ./...` fully green before `/gsd:verify-work`

### Wave 0 Gaps
- [ ] `internal/repo/repo_test.go` — add `TestTaskCreate_WithPriority`, `TestTaskPatchFields_Priority`, `TestTaskList_SortPriority` (PRIO-01, PRIO-02, PRIO-04)
- [ ] `internal/repo/repo_test.go` — add `TestTaskDeleteCompleted`, `TestTaskDeleteCompleted_DryRun`, `TestTaskDeleteCompleted_Scoped` (MAINT-01, MAINT-02, MAINT-05)
- [ ] `internal/repo/repo_test.go` — add `TestTaskStats` with multi-list scenario (STAT-01)
- [ ] `internal/output/output_test.go` — add `TestPrintTasks_Priority` verifying indicator column (PRIO-03)
- [ ] `internal/output/output_test.go` — add `TestPrintStats_JSON` and `TestPrintStats_Table` (STAT-02)

*(All new repo tests follow the `openTestDB` pattern established in the existing `repo_test.go`.)*

## Sources

### Primary (HIGH confidence)
- Existing codebase: `internal/db/db.go` — migration patterns (pragma_table_info guard, ALTER TABLE)
- Existing codebase: `internal/repo/task.go` — TaskInput, TaskPatch, TaskList, sortMap, scanTaskRow patterns
- Existing codebase: `internal/output/output.go` — PrintTasks, PrintTask, printJSON, table renderer
- Existing codebase: `internal/models/models.go` — Task struct field conventions
- SQLite docs: `date()` function strips time component from DATETIME values (standard SQLite behavior)
- SQLite docs: `CASE WHEN ... THEN ... END` expressions in ORDER BY (standard SQL, supported by all SQLite versions)
- Go stdlib: `bufio.Scanner`, `os.Stdin.Stat()`, `os.ModeCharDevice` — TTY detection

### Secondary (MEDIUM confidence)
- SQLite LEFT JOIN with GROUP BY for zero-count rows: standard SQL behavior, consistent with modernc.org/sqlite
- ANSI escape codes for color in terminal output: already used in `colorDot` function in output.go; same approach applies to priority coloring

### Tertiary (LOW confidence)
- None — all findings are based on direct codebase inspection or well-established SQLite/Go stdlib behavior

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — zero new dependencies; all patterns directly extend existing code
- Architecture: HIGH — migration pattern established; repo/output/cmd patterns fully documented
- Pitfalls: HIGH — derived from direct code inspection and SQL semantics
- Validation: HIGH — existing test infrastructure is clear and reusable

**Research date:** 2026-03-06
**Valid until:** 2026-06-06 (stable stack; no external deps; SQLite and Cobra versions locked in go.mod)
