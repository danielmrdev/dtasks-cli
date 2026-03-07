---
phase: 03-tooling
plan: "04"
subsystem: tooling
tags: [update-command, cobra, json-output, embed, skill-integration, tdd, golang]

requires:
  - phase: 03-02
    provides: "internal/updater fully implemented: FetchLatestVersion, AssetName, DownloadAndReplace"
  - phase: 03-03
    provides: "internal/skill fully implemented: DetectClaude, InstallSkill, PromptAndInstall"

provides:
  - "cmd/update.go: updateCmd Cobra command with --json support and skill post-install"
  - "skills/dtasks-cli/skilldata.go: embed wrapper for SKILL.md (workaround for Go //go:embed .. restriction)"
  - "cmd/update_test.go: TestUpdateCmd_JSON, TestUpdateCmd_Help, TestUpdateCmd_AlreadyUpToDate all green"
  - "GHAPIBase exported in internal/updater for cross-package test injection"

affects: [03-05]

tech-stack:
  added: []
  patterns:
    - "cmd.OutOrStdout() for testable output — never os.Stdout directly in RunE"
    - "Cobra help flag reset between tests via updateCmd.Flags().Lookup('help').Value.Set('false')"
    - "skilldata wrapper package at skills/dtasks-cli/ to embed SKILL.md without .. traversal"
    - "PersistentPreRunE early-return for 'update' command — skips DB init on fresh-install path"

key-files:
  created:
    - cmd/update.go
    - skills/dtasks-cli/skilldata.go
  modified:
    - cmd/root.go
    - cmd/update_test.go
    - internal/updater/updater.go
    - internal/updater/updater_test.go

key-decisions:
  - "GHAPIBase exported (was ghAPIBase) to allow cross-package test injection without reflection"
  - "skilldata wrapper package in skills/dtasks-cli/ — Go //go:embed prohibits .. path traversal"
  - "cmd.OutOrStdout() instead of os.Stdout — required for Cobra test output capture via SetOut"
  - "PersistentPreRunE skips DB init for 'update' command — update has no DB dependency, works on fresh install"
  - "Cobra help flag reset in table-driven tests — prevents state leaking between test runs"

patterns-established:
  - "UpdateResult struct: {current, latest, updated, message} JSON shape for update command"
  - "emitUpdateResult(w io.Writer, r UpdateResult, useJSON bool) — clean separation of JSON vs human output"

requirements-completed: [UPDT-01, UPDT-04]

duration: ~8min
completed: 2026-03-06
---

# Phase 3 Plan 04: updateCmd Wire-Up Summary

**Cobra updateCmd with GitHub version check, atomic binary replace, skill post-install, and JSON output via cmd.OutOrStdout() — all tests green**

## Performance

- **Duration:** ~8 min
- **Started:** 2026-03-06T12:23:13Z
- **Completed:** 2026-03-06T12:30:59Z
- **Tasks:** 2
- **Files modified:** 6 (2 created, 4 modified)

## Accomplishments

- Created `cmd/update.go` with `updateCmd` Cobra command:
  - Fetches latest version from GitHub via `updater.FetchLatestVersion`
  - Skips download when current == latest (normalizes `v` prefix)
  - Calls `updater.DownloadAndReplace` for atomic binary self-update
  - Calls `skill.PromptAndInstall` post-update (non-fatal on error)
  - Prints "Run install.sh to update shell completions" after update
  - Emits `UpdateResult` JSON when `--json` is passed
- Created `skills/dtasks-cli/skilldata.go` as embed wrapper (Go prohibits `..` in `//go:embed` paths)
- Exported `GHAPIBase` in `internal/updater` for cross-package test injection
- Registered `updateCmd` on `rootCmd` and added DB-skip for `update` in `PersistentPreRunE`
- Activated `cmd/update_test.go` (removed `//go:build ignore`), all 3 test functions pass

## Task Commits

Each task was committed atomically:

1. **Task 1: Create cmd/update.go with updateCmd** - `b8853ae` (feat)
2. **Task 2: Enable cmd/update_test.go and verify JSON output** - `9c7f6ca` (test)

## Files Created/Modified

- `cmd/update.go` — updateCmd implementation: version check, download, skill post-install, JSON output
- `skills/dtasks-cli/skilldata.go` — embed wrapper for SKILL.md
- `cmd/root.go` — added `updateCmd` to init(), skip DB init for `update` in PersistentPreRunE
- `cmd/update_test.go` — activated tests, added GHAPIBase mock, table-driven AlreadyUpToDate tests
- `internal/updater/updater.go` — exported `GHAPIBase` (was `ghAPIBase`)
- `internal/updater/updater_test.go` — updated references from `ghAPIBase` to `GHAPIBase`

## Decisions Made

- `GHAPIBase` exported from `internal/updater` — the unexported variable could only be overridden from within the same package; cmd tests need cross-package injection without reflection hacks
- `skilldata` wrapper package at `skills/dtasks-cli/skilldata.go` — Go's `//go:embed` spec prohibits path traversal with `..`; placing the file in the same directory as SKILL.md is the only valid approach
- `cmd.OutOrStdout()` instead of `os.Stdout` — Cobra's `SetOut` only redirects through `OutOrStdout()`; using `os.Stdout` directly made tests unable to capture output
- DB skip for `update` in `PersistentPreRunE` — the update command has no DB dependency and must work on fresh installs before config wizard runs
- Help flag reset between table-driven subtests — Cobra stores `--help` flag state on the command object; resetting via `Flags().Lookup("help").Value.Set("false")` prevents state leaking across tests

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Go //go:embed prohibits .. path traversal**
- **Found during:** Task 1
- **Issue:** Plan specified `//go:embed ../skills/dtasks-cli/SKILL.md` in `cmd/update.go`; Go rejects this with "invalid pattern syntax"
- **Fix:** Created `skills/dtasks-cli/skilldata.go` as a dedicated embed wrapper package with `//go:embed SKILL.md` (same directory — valid). `cmd/update.go` imports `skilldata.Content` instead.
- **Files modified:** `skills/dtasks-cli/skilldata.go` (created), `cmd/update.go`
- **Commit:** `b8853ae`

**2. [Rule 1 - Bug] os.Stdout bypasses Cobra SetOut in tests**
- **Found during:** Task 2
- **Issue:** Original implementation used `os.Stdout` directly; Cobra's `rootCmd.SetOut(&buf)` only redirects through `cmd.OutOrStdout()`, so test buffer was always empty
- **Fix:** Refactored `cmd/update.go` to use `cmd.OutOrStdout()` and pass `io.Writer` to `emitUpdateResult`
- **Files modified:** `cmd/update.go`
- **Commit:** `9c7f6ca`

**3. [Rule 1 - Bug] Cobra help flag state leaks between tests**
- **Found during:** Task 2
- **Issue:** After `TestUpdateCmd_Help` runs, `updateCmd` retains `--help=true`; subsequent `Execute()` calls render help instead of running the command
- **Fix:** Added `updateCmd.Flags().Lookup("help").Value.Set("false")` reset at the start of each subtest
- **Files modified:** `cmd/update_test.go`
- **Commit:** `9c7f6ca`

## Issues Encountered

None beyond the auto-fixed deviations above.

## User Setup Required

None — no external service configuration required.

## Next Phase Readiness

- `cmd/update.go` fully implemented and tested — `dtasks update` and `dtasks update --json` both work
- `UPDT-01` and `UPDT-04` requirements are complete
- Phase 03-tooling plan 05 (completion script) can proceed independently

## Self-Check: PASSED
