---
phase: 4
slug: release
status: complete
nyquist_compliant: true
wave_0_complete: true
created: 2026-03-06
audited: 2026-03-07
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
| 4-01-01 | 01 | 1 | pre-PR | automated | `go test ./... -count=1` | ✅ | ✅ green |
| 4-01-02 | 01 | 1 | pre-PR | automated | `gofmt -l .` | ✅ | ✅ green |
| 4-01-03 | 01 | 1 | pre-PR | automated | `go vet ./...` | ✅ | ✅ green |
| 4-01-04 | 01 | 2 | CI gate | manual | `gh pr view --json statusCheckRollup` | ✅ | ✅ green |
| 4-01-05 | 01 | 3 | release gate | manual | `gh release view v0.3.0` | ✅ | ✅ green |

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

- [x] All tasks have `<automated>` verify or Wave 0 dependencies
- [x] Sampling continuity: no 3 consecutive tasks without automated verify
- [x] Wave 0 covers all MISSING references
- [x] No watch-mode flags
- [x] Feedback latency < 15s
- [x] `nyquist_compliant: true` set in frontmatter

**Approval:** 2026-03-07

---

## Validation Audit 2026-03-07

| Metric | Count |
|--------|-------|
| Gaps found | 0 |
| Resolved | 0 |
| Escalated | 0 |
| Status corrected | 5 (pending → green, confirmed by SUMMARY files) |

**Notes:** No test gaps found. Phase 4 is a release-only phase — no new code paths requiring new tests. The `fix(rm)` early-exit guard committed in 4-01-T1 is covered by the existing `go test ./...` suite passing green. CI gate (4-01-04) and release gate (4-01-05) are inherently manual (GitHub Actions) and properly documented in Manual-Only. All 5 tasks confirmed green via SUMMARY files and live verification (`go test`, `gofmt`, `go vet`, `gh release view v0.3.0` → 7 assets, PR #19 MERGED).
