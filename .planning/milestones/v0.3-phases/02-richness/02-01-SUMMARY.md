---
phase: 02-richness
plan: 01
subsystem: testing
tags: [go, tdd, sqlite, repo, output]

# Dependency graph
requires: []
provides:
  - Failing test contracts for PRIO-01, PRIO-02, PRIO-03, PRIO-04
  - Failing test contracts for MAINT-01, MAINT-02, MAINT-04, MAINT-05
  - Failing test contracts for STAT-01, STAT-02
affects: [02-02, 02-03, 02-04]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "TDD red phase: tests call undefined repo/output functions to establish contracts"
    - "strPtr helper in repo_test.go for readable pointer literals"
    - "Direct SQL UPDATE for controlled completed_at timestamps in tests (avoids datetime('now'))"

key-files:
  created: []
  modified:
    - internal/repo/repo_test.go
    - internal/output/output_test.go

key-decisions:
  - "Use direct SQL UPDATE to set completed_at to known date in delete-completed tests — TaskDone uses datetime('now') which is uncontrollable"
  - "Import repo package in output_test.go to reference repo.StatsSummary and repo.ListStats from test fixtures"
  - "Empty string Priority patch ('') means clear priority to nil — consistent with existing DueDate/Notes clear pattern"

patterns-established:
  - "RED phase contract: tests reference TaskInput.Priority, TaskPatch.Priority, models.Task.Priority, repo.TaskDeleteCompleted, repo.TaskStats, output.PrintStats, output.PrintDeletedCount — all undefined until Plans 02-04"

requirements-completed: [PRIO-01, PRIO-02, PRIO-03, PRIO-04, MAINT-01, MAINT-02, MAINT-04, MAINT-05, STAT-01, STAT-02]

# Metrics
duration: 2min
completed: 2026-03-06
---

# Phase 2 Plan 01: TDD Red Phase — Priority, Bulk-Delete, Stats Tests Summary

**10 failing test stubs for priority fields, bulk-delete with dry-run/scoping, and task stats — compile errors confirm RED state for Plans 02-04**

## Performance

- **Duration:** 2 min
- **Started:** 2026-03-06T09:05:07Z
- **Completed:** 2026-03-06T09:06:52Z
- **Tasks:** 1 (TDD red — single commit)
- **Files modified:** 2

## Accomplishments
- Added 7 failing tests to `internal/repo/repo_test.go`: TestTaskCreate_WithPriority, TestTaskPatchFields_Priority, TestTaskList_SortPriority, TestTaskDeleteCompleted, TestTaskDeleteCompleted_DryRun, TestTaskDeleteCompleted_Scoped, TestTaskStats
- Added 4 failing tests to `internal/output/output_test.go`: TestPrintTasks_Priority, TestPrintDeletedCount, TestPrintStats_Table, TestPrintStats_JSON
- Confirmed RED state: both packages fail to build with "undefined" and "unknown field" compile errors

## Task Commits

Each task was committed atomically:

1. **Task 1: TDD red — all Phase 2 failing tests** - `9b4e17f` (test)

**Plan metadata:** _(docs commit pending)_

_Note: TDD tasks may have multiple commits (test → feat → refactor)_

## Files Created/Modified
- `internal/repo/repo_test.go` - Added 7 new test functions for PRIO-01/02/04, MAINT-01/02/05, STAT-01
- `internal/output/output_test.go` - Added 4 new test functions for PRIO-03, MAINT-04, STAT-02; added repo import

## Decisions Made
- Use direct SQL `UPDATE tasks SET completed_at = ?` in `TestTaskDeleteCompleted*` tests instead of `TaskDone()` — `TaskDone` uses `datetime('now')` which cannot be controlled, making date-range assertions non-deterministic
- Import `repo` package in `output_test.go` to reference `repo.StatsSummary` and `repo.ListStats` in PrintStats test fixtures (these types live in repo, not models)
- Empty string `Priority` patch (`strPtr("")`) signals "clear to nil" — consistent with existing DueDate/Notes/Color clear semantics in the codebase

## Deviations from Plan
None - plan executed exactly as written.

## Issues Encountered
None.

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- All test contracts established; Plans 02-04 have clear failing targets to make green
- Plan 02 must add: `models.Task.Priority`, `repo.TaskInput.Priority`, `repo.TaskPatch.Priority`, `repo.TaskDeleteCompleted`, `repo.DeleteCompletedOptions`, `repo.DeleteCompletedResult`
- Plan 03 must add: `output.PrintDeletedCount`, `output.PrintStats`, `repo.TaskStats`, `repo.StatsSummary`, `repo.ListStats`
- Plan 04 must wire priority into `TaskList SortBy="priority"` and CLI flags

---
*Phase: 02-richness*
*Completed: 2026-03-06*
