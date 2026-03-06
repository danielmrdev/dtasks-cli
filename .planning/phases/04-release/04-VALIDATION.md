---
phase: 4
slug: release
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-03-06
---

# Phase 4 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go testing package (stdlib) |
| **Config file** | none — go test ./... is self-discovering |
| **Quick run command** | `go test ./... -count=1` |
| **Full suite command** | `go test ./... -count=1 -race` |
| **Estimated runtime** | ~10 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./... -count=1`
- **After every plan wave:** Run `go test ./... -count=1 -race`
- **Before `/gsd:verify-work`:** Full suite must be green
- **Max feedback latency:** 15 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 4-01-01 | 01 | 1 | pre-PR | automated | `go test ./... -count=1` | ✅ | ⬜ pending |
| 4-01-02 | 01 | 1 | pre-PR | automated | `gofmt -l .` | ✅ | ⬜ pending |
| 4-01-03 | 01 | 1 | pre-PR | automated | `go vet ./...` | ✅ | ⬜ pending |
| 4-01-04 | 01 | 2 | CI gate | manual | `gh pr view --json statusCheckRollup` | ✅ | ⬜ pending |
| 4-01-05 | 01 | 3 | release gate | manual | `gh release view v0.3.0` | ✅ | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

*Existing infrastructure covers all phase requirements. Release phase has no new code to test.*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| CI passes on PR | All PRs must pass | Requires GitHub Actions to run | `gh pr checks <PR-number>` — all checks green |
| 7 release assets published | v0.3.0 ships binaries for all platforms | Requires tag push + GH Actions to run | `gh release view v0.3.0` — confirm 6 binaries + checksums.txt |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 15s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
