---
phase: 7
slug: json-update-fix
status: complete
nyquist_compliant: true
wave_0_complete: true
created: 2026-03-07
validated: 2026-03-07
---

# Phase 7 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go testing (stdlib) |
| **Config file** | none |
| **Quick run command** | `go test ./cmd/ -run TestUpdateCmd -v` |
| **Full suite command** | `go test ./...` |
| **Estimated runtime** | ~5 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./cmd/ -run TestUpdateCmd -v`
- **After every plan wave:** Run `go test ./...`
- **Before `/gsd:verify-work`:** Full suite must be green
- **Max feedback latency:** ~5 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 07-01-01 | 01 | 1 | UPDT-04 | unit | `go test ./cmd/ -run TestUpdateCmd -v` | ✅ | ✅ green |
| 07-01-02 | 01 | 1 | UPDT-04 | unit | `go test ./cmd/ -run TestUpdateCmd -v` | ✅ | ✅ green |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

*Existing infrastructure covers all phase requirements.*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| `--json update` (successful binary update) emits no plain text | UPDT-04 | `updater.DownloadAndReplace` not injectable; structural check sufficient | Code review: verify hint print and `PromptAndInstall` call are gated on `!output.JSONMode` |

---

## Validation Sign-Off

- [x] All tasks have `<automated>` verify or Wave 0 dependencies
- [x] Sampling continuity: no 3 consecutive tasks without automated verify
- [x] Wave 0 covers all MISSING references
- [x] No watch-mode flags
- [x] Feedback latency < 5s
- [x] `nyquist_compliant: true` set in frontmatter

**Approval:** 2026-03-07
