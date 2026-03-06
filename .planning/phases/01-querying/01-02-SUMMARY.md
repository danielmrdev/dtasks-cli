---
phase: 01-querying
plan: 02
subsystem: database
tags: [go, tdd, repo, sqlite, filtering, sorting, search, regexp]

# Dependency graph
requires:
  - phase: 01-01
    provides: "9 failing tests defining the contract for TaskListOptions extensions and TaskSearch"
provides:
  - Extended TaskListOptions with Overdue, DueTomorrow, DueWeek, SortBy, Reverse fields
  - Dynamic ORDER BY builder in TaskList using sortMap lookup and Reverse toggle
  - TaskSearch function and TaskSearchOptions struct with LIKE and Go regexp modes
affects:
  - 01-03 (CLI plan reads these repo functions to wire up --overdue, --tomorrow, --week, --sort, --reverse, --search, --regex flags)

# Tech tracking
tech-stack:
  added:
    - "strings (stdlib) — ReplaceAll for ASC->DESC reversal"
    - "regexp (stdlib) — Go regexp filter for TaskSearch Regex mode"
  patterns:
    - "Dynamic query builder: append WHERE conditions and ORDER BY at runtime; base const has no ORDER BY"
    - "Dual-mode search: SQL LIKE for keyword mode, Go regexp.Compile post-fetch for regex mode"
    - "sortMap lookup with fallback to default order expression"

key-files:
  created: []
  modified:
    - internal/repo/task.go
    - internal/repo/repo_test.go

key-decisions:
  - "Regex mode compiles opts.Keyword directly (not wrapped with (?i)) — user controls the full regexp pattern"
  - "Test bug fixed: (?i)grocery does not match 'groceries', corrected to (?i)groceri"
  - "taskSelectSQL const has no ORDER BY — TaskList appends it dynamically, TaskGet uses it as-is for single-row"

patterns-established:
  - "Query builder pattern: base SQL in const, append WHERE/ORDER BY in function — never in const"
  - "Regex search: fetch broader set via SQL, then filter in Go for full regexp power without SQLite extension"

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

# Phase 1 Plan 02: Querying Repo Layer Summary

**Dynamic WHERE/ORDER BY builder in TaskList with Overdue/DueTomorrow/DueWeek filters, SortBy/Reverse sorting, and TaskSearch with SQLite LIKE and Go regexp dual-mode support**

## Performance

- **Duration:** 2 min
- **Started:** 2026-03-06T10:42:50Z
- **Completed:** 2026-03-06T10:44:50Z
- **Tasks:** 2
- **Files modified:** 2

## Accomplishments

- Extended `TaskListOptions` with 5 new fields: `Overdue`, `DueTomorrow`, `DueWeek`, `SortBy`, `Reverse`
- Removed hardcoded `ORDER BY` from line 79 of `task.go`; replaced with dynamic builder using `sortMap`
- Added `TaskSearchOptions` and `TaskSearch` function with LIKE-based keyword search and Go `regexp.Compile` mode
- All 9 Phase 1 querying tests now pass; 0 regressions in existing passing tests

## Task Commits

Each task was committed atomically:

1. **Task 1 + Task 2: Extend TaskListOptions, dynamic ORDER BY, TaskSearch** - `2f019a3` (feat)

**Plan metadata:** _(docs commit follows)_

## Files Created/Modified

- `internal/repo/task.go` — Added Overdue/DueTomorrow/DueWeek/SortBy/Reverse to TaskListOptions, dynamic sortMap ORDER BY, TaskSearchOptions struct, TaskSearch function; added strings and regexp imports
- `internal/repo/repo_test.go` — Fixed typo in TestTaskSearch_Regex: `(?i)grocery` → `(?i)groceri`

## Decisions Made

- `Regex=true` compiles `opts.Keyword` directly via `regexp.Compile(opts.Keyword)` without adding `(?i)` prefix — user is responsible for the full pattern, including case flags. Adding `(?i)` automatically caused double-flag `(?i)(?i)...` which returns no matches in Go's regexp engine.
- `taskSelectSQL` const remains without `ORDER BY` — `TaskGet` uses it directly for single-row queries; `TaskList` appends ORDER BY dynamically after all WHERE clauses.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Fixed incorrect regex pattern in TestTaskSearch_Regex**
- **Found during:** Task 2 (TaskSearch implementation)
- **Issue:** Test used `(?i)grocery` expecting to match title "Buy groceries". The regex `(?i)grocery` matches the string "grocery" but not "groceries" (different substring). Test expectation was impossible to satisfy.
- **Fix:** Changed test keyword from `(?i)grocery` to `(?i)groceri` which is a valid substring of "groceries" and correctly verifies case-insensitive regex matching.
- **Files modified:** `internal/repo/repo_test.go`
- **Verification:** `TestTaskSearch_Regex` passes after fix
- **Committed in:** `2f019a3`

**2. [Rule 1 - Bug] Imported regexp before TaskSearch was implemented**
- **Found during:** Task 1 (filter implementation)
- **Issue:** Adding `regexp` import before `TaskSearch` was written caused `"regexp" imported and not used` compile error.
- **Fix:** Implemented both tasks together in a single commit — import is used by `TaskSearch`.
- **Files modified:** `internal/repo/task.go`
- **Committed in:** `2f019a3`

---

**Total deviations:** 2 auto-fixed (both Rule 1 - bugs)
**Impact on plan:** Both fixes required for correctness. No scope creep.

## Issues Encountered

Two pre-existing test failures in the repo test suite were uncovered:
- `TestAutocomplete_NotYetDue` — uses hardcoded date `"2026-02-27"` that is now in the past
- `TestAutocomplete_DueTimePassed` — autocomplete scheduler not completing task with past due_time

These failures existed before Plan 01 (confirmed by checking out `820ae52`). They are out of scope for this plan and documented in `deferred-items.md`.

## User Setup Required

None — no external service configuration required.

## Next Phase Readiness

- Repo layer complete for all Phase 1 requirements
- Plan 03 (CLI) can now wire `--overdue`, `--tomorrow`, `--week`, `--sort`, `--reverse`, `--search`, `--regex` flags to the repo functions implemented here
- `TaskSearch` exported and ready to be called from new `task search` subcommand

---
*Phase: 01-querying*
*Completed: 2026-03-06*

## Self-Check: PASSED

- FOUND: `.planning/phases/01-querying/01-02-SUMMARY.md`
- FOUND: commit `2f019a3` (feat - repo layer implementation)
- FOUND: `internal/repo/task.go` with extended TaskListOptions and TaskSearch
- VERIFIED: 9 Phase 1 querying tests pass (`go test ./internal/repo/... -run "TestTaskList_Filter|TestTaskList_Sort|TestTaskSearch"` → ok)
