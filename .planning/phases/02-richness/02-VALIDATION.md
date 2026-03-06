---
phase: 2
slug: richness
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-03-06
---

# Phase 2 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go testing stdlib (go test) |
| **Config file** | none — standard `go test ./...` |
| **Quick run command** | `go test ./internal/repo/... -run TestTask -v` |
| **Full suite command** | `go test ./...` |
| **Estimated runtime** | ~5 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./internal/repo/... -v`
- **After every plan wave:** Run `go test ./...`
- **Before `/gsd:verify-work`:** Full suite must be green
- **Max feedback latency:** ~5 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 2-01-01 | 01 | 0 | PRIO-01, PRIO-02, PRIO-04 | unit | `go test ./internal/repo/... -run TestTaskCreate_WithPriority -v` | ❌ Wave 0 | ⬜ pending |
| 2-01-02 | 01 | 0 | PRIO-03 | unit | `go test ./internal/output/... -run TestPrintTasks_Priority -v` | ❌ Wave 0 | ⬜ pending |
| 2-01-03 | 01 | 1 | PRIO-01 | unit | `go test ./internal/repo/... -run TestTaskCreate_WithPriority -v` | ✅ W0 | ⬜ pending |
| 2-01-04 | 01 | 1 | PRIO-02 | unit | `go test ./internal/repo/... -run TestTaskPatchFields_Priority -v` | ✅ W0 | ⬜ pending |
| 2-01-05 | 01 | 1 | PRIO-03 | unit | `go test ./internal/output/... -run TestPrintTasks_Priority -v` | ✅ W0 | ⬜ pending |
| 2-01-06 | 01 | 1 | PRIO-04 | unit | `go test ./internal/repo/... -run TestTaskList_SortPriority -v` | ✅ W0 | ⬜ pending |
| 2-02-01 | 02 | 0 | MAINT-01, MAINT-02, MAINT-05 | unit | `go test ./internal/repo/... -run TestTaskDeleteCompleted -v` | ❌ Wave 0 | ⬜ pending |
| 2-02-02 | 02 | 1 | MAINT-01 | unit | `go test ./internal/repo/... -run TestTaskDeleteCompleted -v` | ✅ W0 | ⬜ pending |
| 2-02-03 | 02 | 1 | MAINT-02 | unit | `go test ./internal/repo/... -run TestTaskDeleteCompleted_DryRun -v` | ✅ W0 | ⬜ pending |
| 2-02-04 | 02 | 1 | MAINT-03 | manual | `dtasks rm --completed 2026-01-01 --yes` | N/A | ⬜ pending |
| 2-02-05 | 02 | 1 | MAINT-04 | unit | `go test ./internal/repo/... -run TestTaskDeleteCompleted -v` | ✅ W0 | ⬜ pending |
| 2-02-06 | 02 | 1 | MAINT-05 | unit | `go test ./internal/repo/... -run TestTaskDeleteCompleted_Scoped -v` | ✅ W0 | ⬜ pending |
| 2-03-01 | 03 | 0 | STAT-01, STAT-02 | unit | `go test ./internal/repo/... -run TestTaskStats -v` | ❌ Wave 0 | ⬜ pending |
| 2-03-02 | 03 | 1 | STAT-01 | unit | `go test ./internal/repo/... -run TestTaskStats -v` | ✅ W0 | ⬜ pending |
| 2-03-03 | 03 | 1 | STAT-02 | unit | `go test ./internal/output/... -run TestPrintStats_JSON -v` | ✅ W0 | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `internal/repo/repo_test.go` — add `TestTaskCreate_WithPriority`, `TestTaskPatchFields_Priority`, `TestTaskList_SortPriority` stubs (PRIO-01, PRIO-02, PRIO-04)
- [ ] `internal/repo/repo_test.go` — add `TestTaskDeleteCompleted`, `TestTaskDeleteCompleted_DryRun`, `TestTaskDeleteCompleted_Scoped` stubs (MAINT-01, MAINT-02, MAINT-05)
- [ ] `internal/repo/repo_test.go` — add `TestTaskStats` stub with multi-list scenario (STAT-01)
- [ ] `internal/output/output_test.go` — add `TestPrintTasks_Priority` stub (PRIO-03)
- [ ] `internal/output/output_test.go` — add `TestPrintStats_JSON` and `TestPrintStats_Table` stubs (STAT-02)

*All new repo tests follow the `openTestDB` pattern established in the existing `repo_test.go`.*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Confirmation prompt requires `y`/`yes` from terminal | MAINT-03 | Reads from stdin/TTY; cannot be automated without test harness | Run `dtasks rm --completed 2026-01-01`, verify prompt appears; type `n`, verify no deletion; repeat with `y` |
| TTY detection aborts without `--yes` in pipe | MAINT-03 | Requires piping stdin to verify behavior | `echo "" \| dtasks rm --completed 2026-01-01`; verify error message |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 5s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
