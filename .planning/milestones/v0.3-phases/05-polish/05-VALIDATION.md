---
phase: 5
slug: polish
status: complete
nyquist_compliant: true
wave_0_complete: true
created: 2026-03-06
audited: 2026-03-07
---

# Phase 5 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go testing package (stdlib) |
| **Config file** | none — go test ./... is self-discovering |
| **Quick run command** | `go test ./cmd/... -count=1` |
| **Full suite command** | `go test ./... -count=1` |
| **Estimated runtime** | ~5 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./cmd/... -count=1`
- **After every plan wave:** Run `go test ./... -count=1`
- **Before `/gsd:verify-work`:** Full suite must be green
- **Max feedback latency:** 10 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 5-01-01 | 01 | 1 | SORT-01 | automated | `go test ./cmd/... -run TestLsCmd_SortFlagIncludesPriority -count=1` | ✅ | ✅ green |
| 5-01-01 | 01 | 1 | PRIO-04 | automated | `go test ./internal/repo/... -run TestTaskList_SortPriority -count=1` | ✅ | ✅ green |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

*Existing infrastructure covers all phase requirements. No new tests needed — the fix is to a string literal; existing `TestTaskList_SortPriority` already validates the sort functionality.*

---

## Manual-Only Verifications

*All phase behaviors have automated verification.*

---

## Validation Sign-Off

- [x] All tasks have `<automated>` verify or Wave 0 dependencies
- [x] Sampling continuity: no 3 consecutive tasks without automated verify
- [x] Wave 0 covers all MISSING references
- [x] No watch-mode flags
- [x] Feedback latency < 10s
- [x] `nyquist_compliant: true` set in frontmatter

**Approval:** 2026-03-07

---

## Validation Audit 2026-03-07

| Metric | Count |
|--------|-------|
| Gaps found | 1 |
| Resolved | 1 |
| Escalated | 0 |
