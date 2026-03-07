---
phase: 05-polish
plan: 01
subsystem: cli
tags: [cobra, flags, help-text, sort]

requires: []
provides:
  - "--sort flag usage string in lsCmd now includes 'priority' as a valid option"
affects: []

tech-stack:
  added: []
  patterns: []

key-files:
  created: []
  modified:
    - cmd/task.go

key-decisions:
  - "Pre-existing test failure (TestAutocomplete_DueTimeNotYetPassed) is out of scope for this plan — documented in deferred-items.md"

patterns-established: []

requirements-completed: [SORT-01, PRIO-04]

duration: 1min
completed: 2026-03-06
---

# Phase 05 Plan 01: Fix --sort Flag Help Text Summary

**Single-line string fix in `cmd/task.go` so that `--sort` usage now advertises "priority" alongside the existing due, created, completed options.**

## Performance

- **Duration:** 1 min
- **Started:** 2026-03-06T22:39:10Z
- **Completed:** 2026-03-06T22:40:26Z
- **Tasks:** 1
- **Files modified:** 1

## Accomplishments
- `dtasks ls --help` now shows "Sort by: due, created, completed, priority" for the `--sort` flag
- Shell completion (`RegisterFlagCompletionFunc`) was already correct — no changes needed there
- Repo-layer sort-by-priority test (`TestTaskList_SortPriority`) was already passing — no changes needed there

## Task Commits

Each task was committed atomically:

1. **Task 1: Fix --sort flag usage string** - `356d215` (fix)

**Plan metadata:** (docs commit follows)

## Files Created/Modified
- `cmd/task.go` - Updated `--sort` flag usage string to include "priority"

## Decisions Made
- Pre-existing test failure `TestAutocomplete_DueTimeNotYetPassed` is out of scope; documented in `deferred-items.md`

## Deviations from Plan

None - plan executed exactly as written.

The pre-existing test failure (`TestAutocomplete_DueTimeNotYetPassed`) was verified to exist before this plan's changes and was documented in `deferred-items.md` without modification.

## Issues Encountered

`TestAutocomplete_DueTimeNotYetPassed` fails on main before our change. The scheduler completes tasks with `due_date = today` even when `due_time` is in the future. This is out of scope for this plan and logged in `.planning/phases/05-polish/deferred-items.md`.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness
- `--sort` flag is now fully consistent: help text, shell completion, and repo layer all advertise and support "priority"
- The pre-existing scheduler bug (`TestAutocomplete_DueTimeNotYetPassed`) should be addressed in a subsequent plan

---
*Phase: 05-polish*
*Completed: 2026-03-06*
