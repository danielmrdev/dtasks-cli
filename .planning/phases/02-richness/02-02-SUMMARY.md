---
phase: 02-richness
plan: "02"
subsystem: database
tags: [sqlite, repo, priority, bulk-delete, stats, migration]

requires:
  - phase: 02-richness-01
    provides: TDD red phase tests for priority, delete-completed, and stats

provides:
  - Idempotent priority TEXT column migration in db.migrate()
  - Priority *string field on models.Task
  - t.priority in taskSelectSQL; scanned by scanTaskRow
  - TaskInput.Priority *string; passed to INSERT
  - TaskPatch.Priority *string; dynamic update (nil=no-change, ""=NULL, value=set)
  - TaskListOptions.SortBy "priority" using CASE expression (high>medium>low>nil)
  - DeleteCompletedOptions, DeleteCompletedResult, TaskDeleteCompleted function
  - ListStat, ListStats (alias), StatsSummary, TaskStats function

affects:
  - 02-03 (output layer - PrintDeletedCount, PrintStats)
  - 02-04 (CLI commands - --priority flag, prune command, stats command)

tech-stack:
  added: []
  patterns:
    - "ListStats = ListStat type alias for test compatibility"
    - "DryRun pattern: collect rows first, skip DELETE if DryRun=true"
    - "Priority patch: empty string signals clear-to-NULL, consistent with DueDate/Notes/Color"

key-files:
  created: []
  modified:
    - internal/db/db.go
    - internal/models/models.go
    - internal/repo/task.go
    - internal/repo/repo_test.go

key-decisions:
  - "ListStats type alias (= ListStat) to satisfy test references without renaming the canonical type"
  - "TaskDeleteCompleted fetches rows first (SELECT), then DELETEs — enables DryRun without separate query"
  - "TestAutocomplete_NotYetDue hardcoded date fixed with time.Now().AddDate(0,0,1) — pre-existing bug"

patterns-established:
  - "DryRun: collect matching rows, return early before DELETE — zero DB side effects in dry mode"
  - "Stats: LEFT JOIN lists on tasks to include lists with zero tasks in ByList"
  - "Priority sort: CASE expression in ORDER BY, ELSE 4 ensures nil sorts last"

requirements-completed: [PRIO-01, PRIO-02, PRIO-04, MAINT-01, MAINT-02, MAINT-03, MAINT-04, MAINT-05, STAT-01, STAT-02]

duration: 4min
completed: 2026-03-06
---

# Phase 02 Plan 02: Data Layer — Priority, Bulk Delete, Stats Summary

**SQLite migration for priority column, model/repo extension with Priority field, TaskDeleteCompleted (dry-run + list-scoped), and TaskStats with per-list breakdown including empty lists**

## Performance

- **Duration:** ~4 min
- **Started:** 2026-03-06T09:08:44Z
- **Completed:** 2026-03-06T09:11:55Z
- **Tasks:** 2
- **Files modified:** 4

## Accomplishments

- Idempotent `priority TEXT` column migration added after `autocomplete` migration block
- `models.Task.Priority *string` field with correct JSON tag and scan position
- `taskSelectSQL` extended with `t.priority`; `scanTaskRow` scans it at correct position (column 23)
- `TaskInput.Priority` and `TaskPatch.Priority` fully wired — create and patch paths both handle nil/value/clear
- `TaskList` sort by "priority" using CASE expression: high=1, medium=2, low=3, nil=4
- `TaskDeleteCompleted` with DryRun and optional ListID scope — collects tasks via SELECT before DELETE
- `TaskStats` with LEFT JOIN to include empty lists; returns `StatsSummary` with `ByList []ListStat`
- All 7 new repo tests GREEN: TestTaskCreate_WithPriority, TestTaskPatchFields_Priority, TestTaskList_SortPriority, TestTaskDeleteCompleted, TestTaskDeleteCompleted_DryRun, TestTaskDeleteCompleted_Scoped, TestTaskStats

## Task Commits

Each task was committed atomically:

1. **Task 1: Schema migration + model + taskSelectSQL** - `f05be84` (feat)
2. **Task 2: Extend TaskInput/TaskPatch/TaskList; add TaskDeleteCompleted, TaskStats** - `bb949d3` (feat)

**Plan metadata:** (docs commit follows)

## Files Created/Modified

- `internal/db/db.go` - Added idempotent priority TEXT column migration
- `internal/models/models.go` - Added Priority *string field to Task struct (after Autocomplete)
- `internal/repo/task.go` - Extended TaskInput/TaskPatch/TaskListOptions; new TaskDeleteCompleted, TaskStats, and types
- `internal/repo/repo_test.go` - Fixed TestAutocomplete_NotYetDue hardcoded past date (Rule 1 auto-fix)

## Decisions Made

- `ListStats = ListStat` type alias: test files reference `repo.ListStats` but plan defines `ListStat`; alias preserves both without duplication
- `TaskDeleteCompleted` fetches matching rows before DELETE, enabling zero-side-effect DryRun without a second query path
- Priority patch: empty string (`""`) signals clear-to-NULL, consistent with DueDate/Notes/Color clear semantics (established in STATE.md decision)

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Fixed TestAutocomplete_NotYetDue hardcoded past date**
- **Found during:** Task 2 (running full test suite)
- **Issue:** Test used `"2026-02-27"` as "tomorrow" but that date is now in the past (2026-03-06), causing the task to be autocompleted and the test to fail
- **Fix:** Changed to `time.Now().AddDate(0, 0, 1).Format("2006-01-02")` for a dynamic tomorrow
- **Files modified:** `internal/repo/repo_test.go`
- **Verification:** `go test ./internal/repo/... -run TestAutocomplete_NotYetDue` PASS
- **Committed in:** `bb949d3` (Task 2 commit)

**2. [Rule 1 - Bug] Added ListStats type alias for test compatibility**
- **Found during:** Task 2 (implementing TaskStats)
- **Issue:** Plan defines type `ListStat` but repo_test.go references `repo.ListStats` (with trailing 's') — compilation would fail
- **Fix:** Added `type ListStats = ListStat` alias so both references compile to the same type
- **Files modified:** `internal/repo/task.go`
- **Verification:** Tests compile and pass
- **Committed in:** `bb949d3` (Task 2 commit)

---

**Total deviations:** 2 auto-fixed (2 Rule 1 - Bug)
**Impact on plan:** Both fixes required for tests to compile and pass. No scope creep.

## Issues Encountered

None beyond the auto-fixed deviations above.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Data layer complete: all types and functions exported and tested
- Plan 03 (output layer) can implement `PrintDeletedCount` and `PrintStats` using `DeleteCompletedResult` and `StatsSummary`
- Plan 04 (CLI) can wire `--priority` flag to `TaskInput.Priority` and `TaskPatch.Priority`
- Output tests still fail (`PrintDeletedCount`, `PrintStats` undefined) — expected, Plan 03 will fix

---
*Phase: 02-richness*
*Completed: 2026-03-06*

## Self-Check: PASSED

- SUMMARY.md: FOUND
- internal/db/db.go: FOUND
- internal/models/models.go: FOUND
- internal/repo/task.go: FOUND
- commit f05be84: FOUND
- commit bb949d3: FOUND
