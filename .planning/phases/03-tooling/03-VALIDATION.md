---
phase: 3
slug: tooling
status: complete
nyquist_compliant: true
wave_0_complete: true
created: 2026-03-06
audited: 2026-03-07
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
| 3-01-01 | 01 | 0 | UPDT-02, UPDT-03 | unit | `go test ./internal/updater/... -run TestFetchLatestVersion -v` | ✅ | ✅ green |
| 3-01-02 | 01 | 0 | UPDT-01, UPDT-04 | unit | `go test ./cmd/... -run TestUpdateCmd -v` | ✅ | ✅ green |
| 3-01-03 | 01 | 0 | SKIL-01 | unit | `go test ./internal/skill/... -run TestDetectClaude -v` | ✅ | ✅ green |
| 3-01-04 | 01 | 0 | SKIL-02, SKIL-03, SKIL-04 | unit | `go test ./internal/skill/... -run TestInstallSkill -v` | ✅ | ✅ green |
| 3-02-01 | 02 | 1 | UPDT-02 | unit | `go test ./internal/updater/... -run TestFetchLatestVersion -v` | ✅ | ✅ green |
| 3-02-02 | 02 | 1 | UPDT-03 | unit | `go test ./internal/updater/... -run TestAtomicReplace -v` | ✅ | ✅ green |
| 3-03-01 | 03 | 1 | SKIL-01 | unit | `go test ./internal/skill/... -run TestDetectClaude -v` | ✅ | ✅ green |
| 3-03-02 | 03 | 1 | SKIL-02 | unit | `go test ./internal/skill/... -run TestSkillInstall_NonTTY -v` | ✅ | ✅ green |
| 3-03-03 | 03 | 1 | SKIL-03 | unit | `go test ./internal/skill/... -run TestInstallSkill_Path -v` | ✅ | ✅ green |
| 3-03-04 | 03 | 1 | SKIL-04 | unit | `go test ./internal/skill/... -run TestInstallSkill_Overwrite -v` | ✅ | ✅ green |
| 3-04-01 | 04 | 2 | UPDT-01 | smoke | `go build ./... && ./dist/dtasks update --help` | ✅ | ✅ green |
| 3-04-02 | 04 | 2 | UPDT-04 | unit | `go test ./cmd/... -run TestUpdateCmd_JSON -v` | ✅ | ✅ green |
| 3-05-01 | 05 | 2 | COMP-01 | manual | `SHELL=/bin/zsh bash install.sh` (inspect output) | manual-only | ⬜ pending |
| 3-05-02 | 05 | 2 | COMP-02 | manual | `echo "" \| bash install.sh` (should not prompt) | manual-only | ⬜ pending |
| 3-05-03 | 05 | 2 | COMP-03 | manual | `bash install.sh` then check target files exist | manual-only | ⬜ pending |
| 3-05-04 | 05 | 2 | COMP-04 | manual | Run update command, verify completion hint printed | manual-only | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [x] `internal/updater/updater_test.go` — `TestFetchLatestVersion`, `TestAtomicReplace` (UPDT-02, UPDT-03)
- [x] `internal/skill/skill_test.go` — `TestDetectClaude`, `TestInstallSkill_NonTTY`, `TestInstallSkill_Path`, `TestInstallSkill_Overwrite` (SKIL-01..04)
- [x] `cmd/update_test.go` — `TestUpdateCmd_JSON`, `TestUpdateCmd_AlreadyUpToDate` (UPDT-04)
- [x] Test helper: mock HTTP server for GitHub API responses (UPDT-02 without network)
- [x] `internal/updater/updater.go` + `internal/skill/skill.go` — fully implemented

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

- [x] All tasks have automated verify or are explicitly manual-only
- [x] Sampling continuity: no 3 consecutive tasks without automated verify
- [x] Wave 0 covers all MISSING references
- [x] No watch-mode flags
- [x] Feedback latency < 5s
- [x] `nyquist_compliant: true` set in frontmatter

**Approval:** complete

---

## Validation Audit 2026-03-07

| Metric | Count |
|--------|-------|
| Gaps found | 0 |
| Resolved | 0 |
| Escalated | 0 |
| Tasks COVERED | 12 |
| Tasks manual-only | 4 |
| Full suite result | ✅ green (`go test ./...` — all packages pass) |
