---
phase: 01-querying
plan: 03
subsystem: cli
tags: [cobra, flags, search, filtering, sorting]

# Dependency graph
requires:
  - phase: 01-querying-02
    provides: TaskListOptions with Overdue/DueTomorrow/DueWeek/SortBy/Reverse, TaskSearch with TaskSearchOptions
provides:
  - lsCmd flags --today, --overdue, --tomorrow, --week, --sort, --reverse wired to repo.TaskListOptions
  - findCmd top-level command (dtasks find <keyword>) with --list and --regex flags
  - findCmd registered in rootCmd
affects: [02-richness, tooling, release]

# Tech tracking
tech-stack:
  added: []
  patterns: [flag-to-options-struct wiring via cmd.Flags().Changed(), shell completion for enum flags]

key-files:
  created: [cmd/find.go]
  modified: [cmd/task.go, cmd/root.go]

key-decisions:
  - "--due-today flag replaced by --today (pre-release, no backward compat needed)"
  - "findCmd uses cobra.ExactArgs(1) — keyword is positional, not a flag"
  - "Pre-existing TestAutocomplete_NotYetDue and TestAutocomplete_DueTimePassed failures are out of scope (exist before this plan)"

patterns-established:
  - "Filter flags: map each bool flag to opts.Field = true directly in RunE"
  - "Optional int64 flag: check cmd.Flags().Changed() before assigning pointer"
  - "Shell completions for enum flags registered via RegisterFlagCompletionFunc"

requirements-completed: [FILT-01, FILT-02, FILT-03, FILT-04, SORT-01, SORT-02, SRCH-01, SRCH-02, SRCH-03]

# Metrics
duration: 2min
completed: 2026-03-06
---

# Phase 1 Plan 03: CLI Flag Wiring Summary

**Filter/sort flags added to `dtasks ls` and `dtasks find <keyword>` top-level command created, bridging CLI to repo layer**

## Performance

- **Duration:** ~2 min
- **Started:** 2026-03-06T00:08:10Z
- **Completed:** 2026-03-06T00:09:40Z
- **Tasks:** 2
- **Files modified:** 3

## Accomplishments
- `dtasks ls` now accepts --today, --overdue, --tomorrow, --week (date filters), --sort (due|created|completed), --reverse
- `dtasks find <keyword>` added as a top-level command with --list and --regex flags
- Shell completion for --sort flag values registered
- findCmd registered in rootCmd alongside all other commands

## Task Commits

Each task was committed atomically:

1. **Task 1: Extend lsCmd with filter and sort flags** - `0016572` (feat)
2. **Task 2: Create findCmd and register in root** - `5d66769` (feat)

**Plan metadata:** (docs commit — see below)

## Files Created/Modified
- `cmd/task.go` - Replaced lsDueToday with 6 new flag vars; updated RunE and init() for lsCmd
- `cmd/find.go` - New file: findCmd with keyword arg, --list and --regex flags
- `cmd/root.go` - Added rootCmd.AddCommand(findCmd)

## Decisions Made
- `--due-today` replaced by `--today` — project is pre-release, no backward compat needed (per spec)
- `findCmd` uses `cobra.ExactArgs(1)` so keyword is positional (matches `dtasks find <keyword>` UX)
- Shell completion registered for `--sort` returning `["due", "created", "completed"]`

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
- Two pre-existing test failures (`TestAutocomplete_NotYetDue`, `TestAutocomplete_DueTimePassed`) in `internal/repo/repo_test.go` — confirmed pre-existing before any changes in this plan. Out of scope. Logged to deferred-items.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness
- All Phase 1 querying features complete at the CLI layer
- `dtasks ls` filter/sort flags functional, wired to repo layer from Plan 02
- `dtasks find` command ready for use
- Pre-existing test failures in repo_test.go (autocomplete timing) should be addressed before Phase 2

---
*Phase: 01-querying*
*Completed: 2026-03-06*
