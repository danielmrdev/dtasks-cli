---
phase: 3
slug: tooling
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-03-06
---

# Phase 3 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go testing stdlib (go test) |
| **Config file** | none — standard `go test ./...` |
| **Quick run command** | `go test ./internal/updater/... ./internal/skill/... -v` |
| **Full suite command** | `go test ./...` |
| **Estimated runtime** | ~5 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./internal/updater/... ./internal/skill/... -v`
- **After every plan wave:** Run `go test ./...`
- **Before `/gsd:verify-work`:** Full suite must be green
- **Max feedback latency:** ~5 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 3-01-01 | 01 | 0 | UPDT-02, UPDT-03 | unit | `go test ./internal/updater/... -run TestFetchLatestVersion -v` | ❌ Wave 0 | ⬜ pending |
| 3-01-02 | 01 | 0 | UPDT-01, UPDT-04 | unit | `go test ./cmd/... -run TestUpdateCmd -v` | ❌ Wave 0 | ⬜ pending |
| 3-01-03 | 01 | 0 | SKIL-01 | unit | `go test ./internal/skill/... -run TestDetectClaude -v` | ❌ Wave 0 | ⬜ pending |
| 3-01-04 | 01 | 0 | SKIL-02, SKIL-03, SKIL-04 | unit | `go test ./internal/skill/... -run TestInstallSkill -v` | ❌ Wave 0 | ⬜ pending |
| 3-02-01 | 02 | 1 | UPDT-02 | unit | `go test ./internal/updater/... -run TestFetchLatestVersion -v` | ✅ W0 | ⬜ pending |
| 3-02-02 | 02 | 1 | UPDT-03 | unit | `go test ./internal/updater/... -run TestAtomicReplace -v` | ✅ W0 | ⬜ pending |
| 3-03-01 | 03 | 1 | SKIL-01 | unit | `go test ./internal/skill/... -run TestDetectClaude -v` | ✅ W0 | ⬜ pending |
| 3-03-02 | 03 | 1 | SKIL-02 | unit | `go test ./internal/skill/... -run TestSkillInstall_NonTTY -v` | ✅ W0 | ⬜ pending |
| 3-03-03 | 03 | 1 | SKIL-03 | unit | `go test ./internal/skill/... -run TestInstallSkill_Path -v` | ✅ W0 | ⬜ pending |
| 3-03-04 | 03 | 1 | SKIL-04 | unit | `go test ./internal/skill/... -run TestInstallSkill_Overwrite -v` | ✅ W0 | ⬜ pending |
| 3-04-01 | 04 | 2 | UPDT-01 | smoke | `go build ./... && ./dist/dtasks update --help` | ✅ W0 | ⬜ pending |
| 3-04-02 | 04 | 2 | UPDT-04 | unit | `go test ./cmd/... -run TestUpdateCmd_JSON -v` | ✅ W0 | ⬜ pending |
| 3-05-01 | 05 | 2 | COMP-01 | manual | `SHELL=/bin/zsh bash install.sh` (inspect output) | manual-only | ⬜ pending |
| 3-05-02 | 05 | 2 | COMP-02 | manual | `echo "" \| bash install.sh` (should not prompt) | manual-only | ⬜ pending |
| 3-05-03 | 05 | 2 | COMP-03 | manual | `bash install.sh` then check target files exist | manual-only | ⬜ pending |
| 3-05-04 | 05 | 2 | COMP-04 | manual | Run update command, verify completion hint printed | manual-only | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `internal/updater/updater_test.go` — stubs for `TestFetchLatestVersion`, `TestAtomicReplace` (UPDT-02, UPDT-03)
- [ ] `internal/skill/skill_test.go` — stubs for `TestDetectClaude`, `TestSkillInstall_NonTTY`, `TestInstallSkill_Path`, `TestInstallSkill_Overwrite` (SKIL-01..04)
- [ ] `cmd/update_test.go` (or similar) — stub for `TestUpdateCmd_JSON` (UPDT-04)
- [ ] Test helper: mock HTTP server for GitHub API responses (UPDT-02 without network)
- [ ] `internal/updater/updater.go` + `internal/skill/skill.go` — stub files so tests compile

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| `install.sh` detects user shell from `$SHELL` | COMP-01 | Shell env detection is a side effect; not testable in Go unit tests | Run `SHELL=/bin/zsh bash install.sh`, verify zsh completions path is used |
| Completion prompt skipped in non-TTY | COMP-02 | TTY detection (`[ -t 0 ]`) requires real terminal | `echo "" \| bash install.sh`; verify no completion prompt appears |
| Completions written to canonical shell path | COMP-03 | Filesystem side-effect in shell script | Run `bash install.sh` interactively and confirm file at expected path |
| Completion hint shown after update | COMP-04 | End-to-end update flow requires network and binary replacement | Run `dtasks update` when update is available; verify hint message printed |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 5s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
