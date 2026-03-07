---
phase: 01
slug: querying
status: validated
nyquist_compliant: true
wave_0_complete: true
created: 2026-03-07
---

# Phase 01 — Validation Strategy

> Per-phase validation contract. Reconstructed from Plan/Summary artifacts (State B).

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test (stdlib) |
| **Config file** | none — standard Go test runner |
| **Quick run command** | `go test ./internal/repo/... -run "TestTaskList_Filter\|TestTaskList_Sort\|TestTaskSearch" -v` |
| **Full suite command** | `go test ./...` |
| **Estimated runtime** | ~1.2s |

---

## Sampling Rate

- **After every task commit:** Run `go test ./internal/repo/... -run "TestTaskList_Filter|TestTaskList_Sort|TestTaskSearch"`
- **After every plan wave:** Run `go test ./...`
- **Before `/gsd:verify-work`:** Full suite must be green
- **Max feedback latency:** ~1.2 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 01-01-01 | 01 | 1 | FILT-01 | unit | `go test ./internal/repo/... -run TestTaskList_FilterToday -v` | ✅ | ✅ green |
| 01-01-02 | 01 | 1 | FILT-02 | unit | `go test ./internal/repo/... -run TestTaskList_FilterOverdue -v` | ✅ | ✅ green |
| 01-01-03 | 01 | 1 | FILT-03 | unit | `go test ./internal/repo/... -run TestTaskList_FilterTomorrow -v` | ✅ | ✅ green |
| 01-01-04 | 01 | 1 | FILT-04 | unit | `go test ./internal/repo/... -run TestTaskList_FilterWeek -v` | ✅ | ✅ green |
| 01-01-05 | 01 | 1 | SORT-01 | unit | `go test ./internal/repo/... -run TestTaskList_Sort -v` | ✅ | ✅ green |
| 01-01-06 | 01 | 1 | SORT-02 | unit | `go test ./internal/repo/... -run TestTaskList_SortReverse -v` | ✅ | ✅ green |
| 01-01-07 | 01 | 1 | SRCH-01 | unit | `go test ./internal/repo/... -run TestTaskSearch_Keyword -v` | ✅ | ✅ green |
| 01-01-08 | 01 | 1 | SRCH-02 | unit | `go test ./internal/repo/... -run TestTaskSearch_List -v` | ✅ | ✅ green |
| 01-01-09 | 01 | 1 | SRCH-03 | unit | `go test ./internal/repo/... -run TestTaskSearch_Regex -v` | ✅ | ✅ green |
| 01-02-01 | 02 | 2 | FILT-01..04, SORT-01..02 | unit | `go test ./internal/repo/... -run "TestTaskList_Filter\|TestTaskList_Sort" -v` | ✅ | ✅ green |
| 01-02-02 | 02 | 2 | SRCH-01..03 | unit | `go test ./internal/repo/... -run TestTaskSearch -v` | ✅ | ✅ green |
| 01-03-01 | 03 | 3 | FILT-01..04, SORT-01..02 | build | `go build ./...` | ✅ | ✅ green |
| 01-03-02 | 03 | 3 | SRCH-01..03 | build | `go build ./...` | ✅ | ✅ green |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

Existing infrastructure covers all phase requirements. No scaffolding was needed:
- `internal/repo/repo_test.go` already existed with `openTestDB(t)` pattern
- Tests were written in Plan 01 (TDD RED phase) and passed after Plan 02 implementation

---

## Manual-Only Verifications

All phase behaviors have automated verification.

CLI smoke checks were used during development but are not part of the automated suite:
- `dtasks task ls --today` — requires configured DB
- `dtasks find <keyword>` — requires configured DB

These are integration-level checks; the repo-layer unit tests fully cover the underlying logic.

---

## Validation Sign-Off

- [x] All tasks have `<automated>` verify or Wave 0 dependencies
- [x] Sampling continuity: no 3 consecutive tasks without automated verify
- [x] Wave 0 covers all MISSING references
- [x] No watch-mode flags
- [x] Feedback latency < 2s
- [x] `nyquist_compliant: true` set in frontmatter

**Approval:** approved 2026-03-07
