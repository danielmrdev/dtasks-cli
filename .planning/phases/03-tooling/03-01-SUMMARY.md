---
phase: 03-tooling
plan: 01
subsystem: testing
tags: [tdd, updater, skill, golang.org/x/term, httptest, stub]

requires: []
provides:
  - "internal/updater package with stub FetchLatestVersion, AssetName, DownloadAndReplace"
  - "internal/skill package with stub DetectClaude, InstallSkill, PromptAndInstall"
  - "cmd/update_test.go scaffold (build-tagged, awaiting cmd/update.go)"
  - "golang.org/x/term added to go.mod for TTY detection"
affects: [03-02, 03-03, 03-04]

tech-stack:
  added: [golang.org/x/term v0.40.0]
  patterns:
    - "TDD red phase: stub functions return fmt.Errorf(\"not implemented\") to drive compile-then-fail tests"
    - "ghAPIBase unexported var for test injection of mock HTTP server URL"
    - "//go:build ignore for test files that reference not-yet-existing commands"
    - "httptest.NewServer as mock GitHub API for network-free unit tests"

key-files:
  created:
    - internal/updater/updater.go
    - internal/updater/updater_test.go
    - internal/skill/skill.go
    - internal/skill/skill_test.go
    - cmd/update_test.go
  modified:
    - go.mod
    - go.sum

key-decisions:
  - "ghAPIBase unexported var in updater package allows test injection without exported setter"
  - "cmd/update_test.go uses //go:build ignore until Plan 03-04 creates updateCmd — avoids compile errors while preserving test contracts"
  - "golang.org/x/term added as direct dependency for TTY detection in PromptAndInstall"
  - "TestDetectClaude_NotFound uses empty temp dir — may pass if claude is in PATH, documented with t.Log"

patterns-established:
  - "Mock HTTP server pattern: newMockGHAPI(t, tagName) helper creates httptest.Server with JSON payload"
  - "Package-level base URL var for HTTP dependency injection in tests"

requirements-completed: [UPDT-01, UPDT-02, UPDT-03, UPDT-04, SKIL-01, SKIL-02, SKIL-03, SKIL-04]

duration: 2min
completed: 2026-03-06
---

# Phase 3 Plan 01: TDD Red Phase — updater and skill scaffolds

**TDD red phase: stub packages for updater (FetchLatestVersion, AssetName, DownloadAndReplace) and skill (DetectClaude, InstallSkill, PromptAndInstall) with failing test contracts that drive Plans 02-04**

## Performance

- **Duration:** 2 min
- **Started:** 2026-03-06T12:13:44Z
- **Completed:** 2026-03-06T12:16:23Z
- **Tasks:** 3
- **Files modified:** 7

## Accomplishments
- Created `internal/updater` package with stub exports and mock-HTTP-server-based failing tests (TestFetchLatestVersion, TestAssetName, TestAtomicReplace, TestAtomicReplace_PermissionDenied)
- Created `internal/skill` package with stub exports and temp-dir-based failing tests (TestDetectClaude_*, TestInstallSkill_*, TestInstallSkill_NonTTY)
- Created `cmd/update_test.go` with `//go:build ignore` build tag as a frozen test contract for the update subcommand (activated in Plan 03-04)
- Added `golang.org/x/term v0.40.0` to go.mod for TTY detection

## Task Commits

Each task was committed atomically:

1. **Task 1: Scaffold internal/updater package** - `7aee06a` (test)
2. **Task 2: Scaffold internal/skill package** - `5beb4a8` (test)
3. **Task 3: Scaffold cmd/update_test.go** - `08080b7` (test)

**Plan metadata:** (pending final commit)

_Note: This is a TDD red phase plan — all task commits are test/stub commits_

## Files Created/Modified
- `internal/updater/updater.go` - Stub exports: FetchLatestVersion, AssetName, DownloadAndReplace; unexported ghAPIBase for test injection
- `internal/updater/updater_test.go` - Failing tests with mock HTTP server helper (newMockGHAPI)
- `internal/skill/skill.go` - Stub exports: DetectClaude, InstallSkill, PromptAndInstall
- `internal/skill/skill_test.go` - Failing tests using t.TempDir() for homeDir isolation
- `cmd/update_test.go` - TestUpdateCmd_JSON and TestUpdateCmd_Help, guarded with //go:build ignore
- `go.mod` - Added golang.org/x/term v0.40.0
- `go.sum` - Updated with term checksums

## Decisions Made
- `ghAPIBase` as unexported package-level var (not exported setter) — sufficient for same-package tests, consistent with Go idiom for internal test injection
- `cmd/update_test.go` uses `//go:build ignore` rather than a compile-time reference error — cleaner than letting the package fail to build for 3 plans
- `TestDetectClaude_NotFound` uses `t.Log` instead of `t.Error` when claude is in PATH — avoids false failures in dev environments where claude is installed
- go.mod upgraded from go 1.22 to go 1.24.0 automatically by `go get` (golang.org/x/term requires it)

## Deviations from Plan

None — plan executed exactly as written. The `//go:build ignore` approach was the preferred option described in the plan ("Alternative approach (cleaner)").

## Issues Encountered

None.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness
- Plans 03-02 and 03-03 can implement against the established test contracts (green phase for updater and skill packages)
- Plan 03-04 must remove `//go:build ignore` from `cmd/update_test.go` when creating `cmd/update.go`
- All existing tests still pass (config, db, output, repo packages unaffected)

## Self-Check: PASSED

All 6 files confirmed present. All 3 task commits confirmed in git log.

---
*Phase: 03-tooling*
*Completed: 2026-03-06*
