---
phase: 02-richness
plan: "05"
subsystem: cli
tags: [cobra, bulk-delete, boolvar, sqlite, tdd]

requires:
  - phase: 02-richness-04
    provides: TaskDeleteCompleted, DeleteCompletedOptions, rmCmd bulk-delete path

provides:
  - "--completed as BoolVar: dtasks task rm --completed works without date argument"
  - "TaskDeleteCompleted with empty Before: deletes all completed tasks (no date cutoff)"
  - "TestTaskDeleteCompleted_NoBefore: MAINT-04 coverage for no-cutoff bulk delete"

affects: [02-richness-UAT, release]

tech-stack:
  added: []
  patterns:
    - "Conditional WHERE builder: check Before != '' before adding date filter to SELECT and DELETE"

key-files:
  created: []
  modified:
    - internal/repo/task.go
    - internal/repo/repo_test.go
    - cmd/task.go

key-decisions:
  - "BoolVar for --completed: no date argument needed — the flag alone signals bulk-delete-all"
  - "Before='' means no date filter: consistent with Go zero-value semantics; no sentinel needed"
  - "Two separate WHERE builders (select/delete): symmetry required since SELECT uses JOIN alias t. and DELETE uses bare column names"

patterns-established:
  - "Conditional clause builder: if field != '' { append clause+arg } else { bare clause, empty args }"

requirements-completed: [MAINT-01, MAINT-02, MAINT-03, MAINT-04, MAINT-05]

duration: 8min
completed: 2026-03-06
---

# Phase 02 Plan 05: Fix --completed BoolVar Summary

**`--completed` flag changed from StringVar to BoolVar; `TaskDeleteCompleted` updated to delete all completed tasks when `Before` is empty, closing UAT failures on bulk-delete paths.**

## Performance

- **Duration:** 8 min
- **Started:** 2026-03-06T10:22:00Z
- **Completed:** 2026-03-06T10:30:14Z
- **Tasks:** 2
- **Files modified:** 3

## Accomplishments

- Fixed root cause of UAT-6/7/8 failures: Cobra was consuming `--dry-run`, `--yes`, `--list` as the value of `--completed` (StringVar)
- Updated `TaskDeleteCompleted` to support `Before=""` (no date cutoff) — SELECT and DELETE now build WHERE clauses conditionally
- Added `TestTaskDeleteCompleted_NoBefore` (MAINT-04 coverage): DryRun:true returns task list, DryRun:false deletes all completed tasks

## Task Commits

Each task was committed atomically:

1. **Task 1 RED: add failing test TestTaskDeleteCompleted_NoBefore** - `78f69b1` (test)
2. **Task 1 GREEN: support empty Before in TaskDeleteCompleted** - `a52f525` (feat)
3. **Task 2: change --completed to BoolVar in rmCmd** - `c2e782c` (fix)

## Files Created/Modified

- `internal/repo/task.go` - Conditional WHERE builder for SELECT and DELETE branches in `TaskDeleteCompleted`
- `internal/repo/repo_test.go` - Added `TestTaskDeleteCompleted_NoBefore` (DryRun:true and DryRun:false sub-cases)
- `cmd/task.go` - `var rmCompleted bool`; `BoolVar` registration; removed `Before: rmCompleted` from opts

## Decisions Made

- `BoolVar` for `--completed`: the flag alone signals intent — no date argument needed or meaningful after dropping date-cutoff UX
- `Before=""` as zero-value sentinel: consistent with Go zero-value, no special constant needed
- Two independent WHERE builders (one for `taskSelectSQL` JOIN alias `t.`, one for bare `DELETE FROM tasks`): symmetry required because column references differ between the SELECT and DELETE forms

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- UAT tests 6/7/8 (bulk-delete paths) should now pass
- All existing MAINT-01/02/05 tests remain GREEN
- Phase 02-richness fully implemented — ready for final UAT run and release prep

## Self-Check: PASSED

- SUMMARY.md: FOUND
- Commit 78f69b1 (test RED): FOUND
- Commit a52f525 (feat GREEN): FOUND
- Commit c2e782c (fix Task 2): FOUND

---
*Phase: 02-richness*
*Completed: 2026-03-06*
