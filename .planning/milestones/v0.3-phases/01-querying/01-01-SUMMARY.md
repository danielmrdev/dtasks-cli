---
phase: 01-querying
plan: 01
subsystem: testing
tags: [go, tdd, repo, sqlite, filtering, sorting, search]

# Dependency graph
requires: []
provides:
  - Failing test scaffold for all 9 Phase 1 querying requirements (RED phase)
  - Defined contract for TaskListOptions new fields: Overdue, DueTomorrow, DueWeek, SortBy, Reverse
  - Defined contract for TaskSearch function and TaskSearchOptions struct
affects:
  - 01-02 (implementation plan reads these tests to know exactly what to build)

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "TDD RED phase: tests reference undefined fields/functions to define the implementation contract"
    - "Relative date construction with time.Now().AddDate for filter boundary tests"

key-files:
  created: []
  modified:
    - internal/repo/repo_test.go

key-decisions:
  - "SRCH-03 invalid regex test asserts err != nil and tasks == nil (nil slice, not empty)"
  - "SORT-01 sort by created uses insertion order (SQLite auto-increments ID and created_at)"
  - "FILT-04 week range is [today, today+6] inclusive — 7 days starting from today"

patterns-established:
  - "Filter tests: create boundary tasks (just inside, just outside) to verify exact semantics"
  - "Search tests: test case-insensitive match on both title and notes fields"

requirements-completed:
  - FILT-01
  - FILT-02
  - FILT-03
  - FILT-04
  - SORT-01
  - SORT-02
  - SRCH-01
  - SRCH-02
  - SRCH-03

# Metrics
duration: 2min
completed: 2026-03-06
---

# Phase 1 Plan 01: Querying Test Scaffold Summary

**TDD RED phase: 9 failing tests define the exact contract for TaskListOptions filters (Overdue, DueTomorrow, DueWeek), sorting (SortBy, Reverse), and TaskSearch with keyword/list/regex options**

## Performance

- **Duration:** 2 min
- **Started:** 2026-03-06T10:39:41Z
- **Completed:** 2026-03-06T10:41:07Z
- **Tasks:** 1
- **Files modified:** 1

## Accomplishments

- Added 9 test functions to `internal/repo/repo_test.go` covering all Phase 1 requirements
- Tests reference undefined `TaskListOptions` fields and `TaskSearch` function — compile errors confirm RED phase
- All non-repo package tests continue to pass (db, config, output)

## Task Commits

Each task was committed atomically:

1. **Task 1: Add 9 failing tests for Phase 1 querying requirements** - `b8b7664` (test)

**Plan metadata:** _(docs commit follows)_

## Files Created/Modified

- `internal/repo/repo_test.go` - Added 9 test functions: TestTaskList_FilterToday, TestTaskList_FilterOverdue, TestTaskList_FilterTomorrow, TestTaskList_FilterWeek, TestTaskList_Sort, TestTaskList_SortReverse, TestTaskSearch_Keyword, TestTaskSearch_List, TestTaskSearch_Regex

## Decisions Made

- `SRCH-03` invalid regex test asserts `err != nil` and `tasks == nil` (not empty slice) to match Go convention
- `SORT-01` sort by created relies on SQLite insertion order (auto-incremented `created_at`), no sleep needed
- Week filter boundary: `[today, today+6]` inclusive — 7 tasks due in next 7 days starting today

## Deviations from Plan

None — plan executed exactly as written. Compile errors are intentional (RED phase), confirmed with `go test ./internal/repo/...`.

## Issues Encountered

None.

## User Setup Required

None — no external service configuration required.

## Next Phase Readiness

- Test scaffold ready for Plan 02 implementation
- Plan 02 must add to `TaskListOptions`: `Overdue bool`, `DueTomorrow bool`, `DueWeek bool`, `SortBy string`, `Reverse bool`
- Plan 02 must add `TaskSearch(db, TaskSearchOptions)` function and `TaskSearchOptions` struct
- All 9 tests should pass after Plan 02 completes

---
*Phase: 01-querying*
*Completed: 2026-03-06*
