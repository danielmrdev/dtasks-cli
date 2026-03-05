# Codebase Concerns

**Analysis Date:** 2026-03-06

## Ignored Error Returns

**Silent error drops on database operations:**
- Issue: `LastInsertId()` and `RowsAffected()` errors are silently ignored with underscore blanks, then the code proceeds assuming success
- Files: `internal/repo/list.go` (lines 18, 78, 90), `internal/repo/task.go` (lines 32, 111, 168, 190)
- Example: `id, _ := res.LastInsertId()` at line 32 in `task.go` — if this fails, `id` is 0, then `TaskGet(db, 0)` will fail with unclear error
- Impact: Real database errors (constraint violations, disk full, permissions) are masked, leading to confusing "not found" errors downstream
- Fix approach: Check and propagate these errors, or at minimum wrap them with context about the operation that failed

**Unhandled errors in migration:**
- Issue: `QueryRow().Scan()` in `internal/db/db.go` (lines 100, 109, 119) do not check for errors
- Impact: If pragma_table_info queries fail (rare but possible), column existence checks silently fail, and subsequent ALTER TABLE operations may fail with cryptic errors
- Fix approach: Add error checks for all `Scan()` operations in the `migrate()` function

## Date Validation Gaps

**No format validation for due dates and due times:**
- Issue: CLI accepts any string for `--due` and `--due-time`, passes directly to DB without validation
- Files: `cmd/task.go` (lines 28-29), `cmd/recur.go`
- Validation: `--due-time requires --due` is checked, but not whether dates are actually valid YYYY-MM-DD or times are HH:MM
- Risk: Invalid dates (e.g., `2026-13-45` or `99:99`) are stored in the database, breaking sorting and scheduler logic
- Current behavior: Queries like `due_date <= ?` (line 76 in `task.go`) will silently fail to match malformed dates
- Fix approach: Validate date format with `time.Parse("2006-01-02", input)` in CLI before inserting; reject invalid formats

**Scheduler uses local time, no timezone awareness:**
- Issue: `time.Now()` in `recur_scheduler.go` (lines 142, 144, 16) and `task.go` (line 74) returns local system time; no timezone context
- Impact: If user's system time is wrong or they run the CLI from a different timezone, autocomplete and "due today" logic produces incorrect results
- Risk: Recurring tasks may auto-complete at wrong times; "due-today" filter is timezone-dependent
- Fix approach: Document this as a known limitation or add timezone configuration to `.env`

## Error Handling Fragility

**QueryRow errors not distinguished:**
- Issue: `ListGet` (line 25 in `list.go`) and `TaskGet` (line 37 in `task.go`) return generic `"list not found: %w"` error for ANY Scan failure
- Impact: Database corruption, permission errors, and NULL handling bugs all surface as "not found", making debugging hard
- Fix approach: Use `errors.Is(err, sql.ErrNoRows)` to distinguish "missing row" from "scan error"

**Cascade deletes with no constraint validation:**
- Issue: `ListDelete` (line 85 in `list.go`) relies on `ON DELETE CASCADE` in the schema, but the code doesn't validate this constraint exists
- Risk: If schema is manually altered and CASCADE is removed, task deletion becomes orphaning instead of cascading
- Fix approach: Add a comment documenting the CASCADE requirement, or verify constraints on startup

## Potential Race Conditions

**No pessimistic locking on task recurrence:**
- Issue: `TaskScheduleNext` (line 90 in `recur_scheduler.go`) uses a transaction, but between `TaskGet` and the `INSERT...UPDATE` sequence, another process could delete or modify the task
- Impact: Low risk in single-binary CLI context, but if multiple instances run against the same DB:
  - Second instance could read stale task data and create wrong next occurrence
  - Delete inside transaction could succeed then fail on UPDATE, leaving orphaned new task
- Current mitigation: `MaxOpenConns = 1` (line 24 in `db.go`) ensures single writer, but multiple readers can race
- Fix approach: Document this as a single-process limitation, or add row-level locking with `SELECT...FOR UPDATE`

**Autocomplete processor doesn't lock read-then-modify sequence:**
- Issue: `ProcessAutocompleteTasks` (line 145 in `recur_scheduler.go`) queries overdue tasks, then loops calling `TaskDone` and `TaskScheduleNext`
- Impact: Between query and processing, another process could complete the same task, creating duplicate next occurrences
- Risk: Mitigated by `MaxOpenConns=1`, but still fragile if that is ever changed
- Fix approach: Either move entire autocomplete logic into a single transaction, or add row-level locking

## Incomplete Features & Known Gaps

**Week-based recurrence logic not implemented:**
- Issue: `recur_day_of_week` field exists in schema and model, but `calcNextDate()` in `recur_scheduler.go` never uses it
- Files: `recur_scheduler.go` line 28-32 only has cases for "daily" and "weekly" (which increments by 7*interval days, ignoring which day of week)
- Impact: Setting `--day mon` on weekly recurrence has no effect; recurrence advances 7 days from the same weekday, not to the same day next week
- Workaround: None — weekly recurrence is currently broken for specific day patterns
- Fix approach: Implement proper day-of-week logic in `calcNextDate()` to find next occurrence on the specified day

**No validation of recurrence end dates:**
- Issue: When setting `recur ends on-date`, no validation that `EndsDate >= current due date`
- Risk: Can create impossible configs like "ends on 2020-01-01" for a task due 2026-03-06
- Fix approach: Add validation in `cmd/recur.go` functions before calling `TaskSetRecur`

## SQL Injection Risks (Low Risk)

**SQL construction in TaskPatchFields:**
- Issue: `internal/repo/task.go` line 164 builds SQL dynamically: `"UPDATE tasks SET " + set + " WHERE id = ?"`
- Mitigation: `set` is constructed by the code, not user input; values are parameterized
- Risk level: Low (not user-controlled), but pattern is fragile
- Fix approach: Refactor to use a pre-defined statement or query builder to avoid string concatenation

**SQL construction in ListPatchFields:**
- Issue: `internal/repo/list.go` line 73 uses `strings.Join(setClauses, ", ")`
- Same mitigation and risk as above

## Connection & Resource Management

**No connection pooling tuning:**
- Issue: `SetMaxOpenConns(1)` (line 24 in `db.go`) forces single writer to prevent WAL conflicts
- Impact: Serializes all writes, which is safe but slow under high concurrency
- Risk: If code is refactored to use goroutines, single-threaded writes become a bottleneck
- Fix approach: Document this design decision and test concurrent access behavior before changing

**DB connection never closed in CLI:**
- Issue: `cmd/root.go` opens DB but never explicitly closes it
- Impact: Minor (process exits immediately), but not idiomatic Go
- Fix approach: Add `defer DB.Close()` in `Execute()` function

## Missing Input Validation

**No bounds checking on recurrence intervals:**
- Issue: `--every` flag accepts any int; no checks for <= 0 or > 999
- Risk: User can set `--every 0` or `--every -1`, creating invalid recurrence
- Fix approach: Add flag validation in `cmd/recur.go` to reject invalid intervals

**No list existence check before moving task:**
- Issue: `edit` command accepts `--list` ID without verifying the list exists (line 184 in `cmd/task.go`)
- Risk: Can move task to non-existent list; DB constraint will fail with obscure error
- Fix approach: Validate list exists in CLI layer before calling `TaskPatchFields`

## Missing Nil Checks

**Potential nil dereference in time parsing:**
- Issue: `scanTaskRow` (line 293 in `task.go`) calls `time.Parse()` on `completedAt.String` without checking if it's valid RFC3339
- Risk: If `completed_at` is stored in wrong format, `Parse` fails silently and `t.CompletedAt` remains nil
- Fix approach: Check error from `Parse` and return it; validate timestamp format on insert

**Recur fields not initialized on creation:**
- Issue: `TaskCreate` doesn't set `recurring = 0` explicitly, relies on DEFAULT
- Risk: If schema changes or migration fails to set DEFAULT, new tasks get NULL recurring flag
- Fix approach: Explicitly initialize recurrence fields in INSERT statement

## Test Coverage Gaps

**No error path testing for database failures:**
- Issue: Tests in `repo_test.go` don't cover disk full, permission denied, constraint violations, or connection failures
- Impact: Error handling code is untested, bugs only surface in production
- Fix approach: Add tests that inject SQL errors or use a mock database

**No validation of edge case dates:**
- Issue: No tests for leap years, month boundaries, invalid date strings
- Impact: `calcNextDate` logic could have off-by-one errors undetected
- Fix approach: Add parametric tests for Feb 28/29, Jan 31 → Feb, month roll-overs

**Autocomplete time precision issues not covered:**
- Issue: `ProcessAutocompleteTasks` compares `due_time` as strings (line 149): `"due_time <= ?"`
- Risk: String comparison of HH:MM works but is fragile (will break with timezone logic)
- Fix approach: Test edge cases like task due at exactly "now"

## Documentation Debt

**Scheduler behavior not documented:**
- Issue: `calcNextDate` logic for monthly recurrence (clamping to last day of month) is only explained in comments
- Impact: Users can't predict behavior; developers unfamiliar with the code don't understand why Jan 31 → Feb 28
- Fix approach: Add unit test names and comments explaining the clamping behavior

**Global DB variable and PersistentPreRunE pattern:**
- Issue: Global `var DB *sql.DB` in `cmd/root.go` is not idiomatic; initialization in `PersistentPreRunE` is implicit
- Impact: Hard to test, unclear lifecycle, requires reading Cobra docs to understand
- Fix approach: Consider refactoring to explicit initialization or dependency injection

## Known Limitations (By Design)

These are not bugs but architectural constraints:

- **Single-writer limitation:** `MaxOpenConns=1` means dtasks can't scale to multi-process usage without redesign
- **No sync backend:** Database path is local only; cross-device sync relies on external tools (Dropbox, Syncthing)
- **No transaction isolation for multi-instance:** If two dtasks processes run simultaneously, they may corrupt shared task state
- **Timezone-naive scheduling:** Autocomplete and "due today" filters use local system time with no timezone support

---

*Concerns audit: 2026-03-06*
