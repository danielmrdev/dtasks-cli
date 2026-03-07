---
phase: 6
slug: skill-install
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-03-07
---

# Phase 6 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go testing (stdlib) |
| **Config file** | none — standard `go test` |
| **Quick run command** | `go test ./internal/skill/... -count=1` |
| **Full suite command** | `go test ./...` |
| **Estimated runtime** | ~5 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./internal/skill/... -count=1`
- **After every plan wave:** Run `go test ./...`
- **Before `/gsd:verify-work`:** Full suite must be green + `bash -n install.sh` passes
- **Max feedback latency:** 10 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 6-01-01 | 01 | 0 | SKIL-02 | unit | `go test ./cmd/... -run TestInstallSkillCmd -v` | ❌ W0 | ⬜ pending |
| 6-01-02 | 01 | 1 | SKIL-01/02/03/04 | build | `go build ./...` | ✅ | ⬜ pending |
| 6-01-03 | 01 | 1 | SKIL-01/02/03/04 | unit | `go test ./internal/skill/... -count=1` | ✅ | ⬜ pending |
| 6-01-04 | 01 | 1 | SKIL-02 | smoke | `bash -n install.sh` | ✅ | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `cmd/install_skill_test.go` — covers `install-skill` command wiring (builds, exits 0 in non-TTY environment)

*Existing infrastructure (go test) covers all other requirements.*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Consent prompt appears in TTY | SKIL-02 | Requires interactive terminal | Run `dtasks install-skill` in a real terminal with Claude installed; confirm prompt appears |
| Non-TTY install skips skill install | SKIL-02 | Requires piped shell execution | Run `echo "" \| sh install.sh` — should complete without installing skill |
| `install.sh` calls `install-skill` end-to-end | SKIL-01/02/03/04 | Requires real binary on disk | Run `sh install.sh` in a TTY; confirm skill consent prompt appears after binary install |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 10s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
