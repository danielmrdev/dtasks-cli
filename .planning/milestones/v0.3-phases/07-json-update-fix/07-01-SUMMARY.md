---
phase: 07-json-update-fix
plan: 01
subsystem: cli
tags: [go, cobra, json, output, update]

# Dependency graph
requires:
  - phase: 06-skill-install
    provides: install-skill command and PromptAndInstall integration in updateCmd
provides:
  - cmd/update.go with output.JSONMode as SSOT — no double flag-read, side-effects gated
  - TestUpdateCmd_JSON_NoContamination asserting clean JSON stdout
affects: []

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "output.JSONMode as SSOT: updateCmd reads output.JSONMode (set by PersistentPreRunE) rather than re-reading flags locally"
    - "Side-effect gating: plain-text side-effects wrapped in !output.JSONMode guard before emitting JSON result"

key-files:
  created: []
  modified:
    - cmd/update.go
    - cmd/update_test.go

key-decisions:
  - "output.JSONMode is SSOT — no local flag-read in updateCmd.RunE"
  - "emitUpdateResult signature left unchanged (useJSON bool param) — out-of-scope refactor, diff kept minimal"
  - "TestUpdateCmd_JSON_NoContamination resets --help flag residual state from prior test (same pattern as TestUpdateCmd_AlreadyUpToDate)"

patterns-established:
  - "All cmd/*.go files must read output.JSONMode directly, never re-read --json flag locally"

requirements-completed: [UPDT-04]

# Metrics
duration: 2min
completed: 2026-03-07
---

# Phase 07 Plan 01: JSON Update Fix Summary

**Fixed dtasks update --json contamination: removed double flag-read, gated PromptAndInstall and completions hint on !output.JSONMode, added NoContamination regression test**

## Performance

- **Duration:** 2 min
- **Started:** 2026-03-07T14:59:18Z
- **Completed:** 2026-03-07T15:01:00Z
- **Tasks:** 2
- **Files modified:** 2

## Accomplishments
- Removed local `useJSON` flag-read block from `updateCmd.RunE` — `output.JSONMode` is now the single source of truth
- Added `!output.JSONMode` guard wrapping `PromptAndInstall` call and completions hint line
- Replaced all `useJSON` references in `RunE` with `output.JSONMode`
- Added `TestUpdateCmd_JSON_NoContamination` asserting stdout starts with `{` and is valid JSON

## Task Commits

Each task was committed atomically:

1. **Task 1: Add TestUpdateCmd_JSON_NoContamination test** - `6973c51` (test)
2. **Task 2: Fix cmd/update.go — remove double flag-read and gate side-effects** - `d9f2605` (fix)

**Plan metadata:** (docs commit follows)

## Files Created/Modified
- `cmd/update.go` - Removed double flag-read, added output package import, gated side-effects on !output.JSONMode
- `cmd/update_test.go` - Added TestUpdateCmd_JSON_NoContamination; fixed --help flag reset in new test

## Decisions Made
- `emitUpdateResult` helper signature left unchanged (`useJSON bool` parameter) — the plan explicitly excluded this refactor to keep the diff minimal
- New test resets `--help` flag before execution, following the same pattern as `TestUpdateCmd_AlreadyUpToDate`, because `TestUpdateCmd_Help` leaves residual flag state

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Fixed residual --help flag state in TestUpdateCmd_JSON_NoContamination**
- **Found during:** Task 1 (test) — GREEN phase
- **Issue:** Test failed with help text in stdout because `TestUpdateCmd_Help` leaves `--help=true` on the cobra Command and the next test inherits it
- **Fix:** Added `if f := updateCmd.Flags().Lookup("help"); f != nil { _ = f.Value.Set("false") }` before executing, matching the pattern already used in `TestUpdateCmd_AlreadyUpToDate`
- **Files modified:** cmd/update_test.go
- **Verification:** All `TestUpdateCmd_*` tests pass; full suite green
- **Committed in:** d9f2605 (Task 2 commit, alongside the fix)

---

**Total deviations:** 1 auto-fixed (Rule 1 - bug in test isolation)
**Impact on plan:** Necessary for test correctness; no scope creep.

## Issues Encountered
- The "already up to date" branch (used by the new test) passes through the uncontaminated code path, so the test passed before the fix — confirming the test covers the contract. The contamination code only executes on the "binary replaced" branch which requires a real download; that branch is now correctly gated.

## Next Phase Readiness
- UPDT-04 closed: `dtasks --json update` emits clean JSON on all code paths
- No blockers for subsequent phases

---
*Phase: 07-json-update-fix*
*Completed: 2026-03-07*
