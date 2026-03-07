---
phase: 03-tooling
plan: 02
subsystem: tooling
tags: [updater, github-api, binary-replacement, net/http, httptest, tdd, golang]

requires:
  - phase: 03-01
    provides: "internal/updater stub with ghAPIBase var, failing tests (TestFetchLatestVersion, TestAssetName, TestAtomicReplace, TestAtomicReplace_PermissionDenied)"

provides:
  - "internal/updater package fully implemented: FetchLatestVersion, AssetName, DownloadAndReplace"
  - "All 5 updater tests passing green"
  - "Atomic binary replacement via temp-file-in-same-dir + os.Rename"

affects: [03-04]

tech-stack:
  added: []
  patterns:
    - "HTTP dependency injection via ghAPIBase package var — same-package test override without exported setter"
    - "Atomic binary replace: CreateTemp in filepath.Dir(exePath), io.Copy stream, chmod 0755, os.Rename — no cross-device rename risk"
    - "context.Background() with http.NewRequestWithContext for cancellable HTTP requests"

key-files:
  created: []
  modified:
    - internal/updater/updater.go

key-decisions:
  - "DownloadAndReplace uses filepath.Dir(exePath) for temp dir — avoids cross-device rename failure that os.TempDir() would cause"
  - "io.Copy streams download to temp file — no io.ReadAll buffering, safe for large binaries"
  - "Defer os.Remove(tmpName) is no-op after successful os.Rename — correct cleanup for both success and error paths"

patterns-established:
  - "Atomic file replace pattern: CreateTemp same-dir + io.Copy + chmod + Rename"
  - "GitHub API client: Accept + X-GitHub-Api-Version + User-Agent headers; non-200 returns status code error"

requirements-completed: [UPDT-02, UPDT-03]

duration: 1min
completed: 2026-03-06
---

# Phase 3 Plan 02: Updater Package Implementation Summary

**GitHub Releases API client with atomic binary self-replacement via streaming temp-file rename — all 5 updater tests green**

## Performance

- **Duration:** 1 min
- **Started:** 2026-03-06T12:18:32Z
- **Completed:** 2026-03-06T12:19:25Z
- **Tasks:** 2
- **Files modified:** 1

## Accomplishments
- Implemented `FetchLatestVersion` with correct GitHub API headers (Accept, X-GitHub-Api-Version, User-Agent), mock-server-based test injection via `ghAPIBase`, and non-200 status error handling
- Implemented `AssetName` mapping darwin->macos, linux->linux, windows->windows with .exe suffix for correct cross-platform asset resolution
- Implemented `DownloadAndReplace` with streaming io.Copy, atomic os.Rename using temp file in same directory as target binary, and deferred cleanup

## Task Commits

Each task was committed atomically:

1. **Task 1: Implement FetchLatestVersion and AssetName** - `e25fba2` (feat)
2. **Task 2: Implement DownloadAndReplace** - included in `e25fba2` (same file, single Write operation)

**Plan metadata:** (pending final commit)

_Note: Both tasks modify the same file. Task 2 implementation was included in Task 1's commit as the full file was written in a single operation. All 5 tests (including TestAtomicReplace and TestAtomicReplace_PermissionDenied) pass green._

## Files Created/Modified
- `internal/updater/updater.go` - Full implementation: FetchLatestVersion (GitHub API client), AssetName (OS/arch mapping), DownloadAndReplace (atomic streaming replace)

## Decisions Made
- `filepath.Dir(exePath)` for temp dir instead of `os.TempDir()` — same filesystem guarantees atomic rename works; cross-filesystem rename would fail on most Unix systems
- `io.Copy` streaming instead of `io.ReadAll` — avoids loading entire binary into memory, essential for large binaries (20+ MB)
- Defer `os.Remove(tmpName)` handles cleanup on both success and failure paths — no-op after `os.Rename` succeeds since the temp name no longer exists

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness
- `internal/updater` package fully implemented and tested — Plan 03-04 can call these functions from `cmd/update.go`
- `internal/skill` package still has stub implementations (pending Plan 03-03)
- All pre-existing tests (config, db, output, repo) continue passing

## Self-Check: PASSED

---
*Phase: 03-tooling*
*Completed: 2026-03-06*
