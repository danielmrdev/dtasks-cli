---
phase: 02-richness
plan: 04
subsystem: cli
tags: [cobra, priority, bulk-delete, stats, flags]

# Dependency graph
requires:
  - phase: 02-richness-02
    provides: repo.TaskInput.Priority, repo.TaskPatch.Priority, repo.TaskDeleteCompleted, repo.TaskStats, repo.DeleteCompletedOptions
  - phase: 02-richness-03
    provides: output.PrintDeletedCount, output.PrintStats, output.PrintTasks with PRIO column
provides:
  - --priority flag on addCmd and editCmd with validation
  - rmCmd bulk delete via --completed/--dry-run/--yes/--list flags
  - statsCmd top-level command wired to repo.TaskStats + output.PrintStats
  - isTerminal helper for stdin TTY detection
affects: [release, integration-tests]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Two-step bulk delete: DryRun:true preview → DryRun:false execute on confirmed path"
    - "isTerminal(os.Stdin) for non-interactive mode detection"
    - "cmd.Flags().Changed guard before priority assignment (consistent with other optional flags)"

key-files:
  created:
    - cmd/stats.go
  modified:
    - cmd/task.go
    - cmd/root.go

key-decisions:
  - "Two-step delete strategy: first call always DryRun:true to display count, second call DryRun:false — never skip preview before confirmation"
  - "rmCmd bulk delete with --yes or TTY prompt: enforces MAINT-03 by erroring in non-interactive mode without --yes"
  - "statsCmd has no subcommands — flat stats command, --json via global flag"

patterns-established:
  - "Priority validation at CLI layer (high/medium/low), empty string allowed only in editCmd to clear"
  - "Bulk operations follow two-step fetch-then-delete pattern from repo layer"

requirements-completed: [PRIO-01, PRIO-02, PRIO-03, PRIO-04, MAINT-01, MAINT-02, MAINT-03, MAINT-04, MAINT-05, STAT-01, STAT-02]

# Metrics
duration: 8min
completed: 2026-03-06
---

# Phase 2 Plan 4: CLI Layer Summary

**Priority flags, bulk-delete rm extension, and statsCmd wired to repo/output layers — all 11 Phase 2 requirements testable end-to-end**

## Performance

- **Duration:** 8 min
- **Started:** 2026-03-06T09:16:10Z
- **Completed:** 2026-03-06T09:24:00Z
- **Tasks:** 2
- **Files modified:** 3

## Accomplishments

- Added `--priority` flag to `addCmd` (validates high/medium/low) and `editCmd` (empty string clears priority)
- Extended `rmCmd` with `--completed`, `--dry-run`, `--yes`, `--list` flags and full bulk-delete flow with TTY confirmation
- Created `cmd/stats.go` with `statsCmd` and registered in `rootCmd`

## Task Commits

1. **Task 1: Priority flags on addCmd/editCmd; extend rmCmd for bulk delete** - `243d117` (feat)
2. **Task 2: Create statsCmd and register in root** - `2c7fc1c` (feat)

**Plan metadata:** (next commit — docs)

## Files Created/Modified

- `cmd/task.go` - Added `addPriority`/`editPriority` vars, priority validation blocks, extended `rmCmd` with bulk delete path and new vars/flags, added `isTerminal` helper
- `cmd/stats.go` - New file: `statsCmd` calling `repo.TaskStats` + `output.PrintStats`
- `cmd/root.go` - Added `rootCmd.AddCommand(statsCmd)`

## Decisions Made

- Two-step delete strategy: first call always `DryRun:true` to display affected count, confirmation prompt, then `DryRun:false` to execute — never skips preview.
- `isTerminal` checks `os.ModeCharDevice` on stdin; bulk delete returns error in non-TTY without `--yes`.
- `statsCmd` has no subcommands; JSON emitted via global `--json` flag inherited from root.

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- All 11 Phase 2 requirements (PRIO-01..04, MAINT-01..05, STAT-01..02) are now wired end-to-end.
- Phase 3 (tooling: shell completions, manpages, release automation) can proceed.
- `go test ./...` fully GREEN; binary builds cleanly.

---
*Phase: 02-richness*
*Completed: 2026-03-06*
