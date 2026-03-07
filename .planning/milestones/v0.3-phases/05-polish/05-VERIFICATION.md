---
phase: 05-polish
verified: 2026-03-06T23:00:00Z
status: passed
score: 3/3 must-haves verified
re_verification: false
---

# Phase 05: Polish — Verification Report

**Phase Goal:** Fix --sort flag help text to advertise "priority" as a valid sort field
**Verified:** 2026-03-06T23:00:00Z
**Status:** passed
**Re-verification:** No — initial verification

## Goal Achievement

### Observable Truths

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | `dtasks task ls --help` shows 'priority' in the --sort flag description | VERIFIED | Line 149 of `cmd/task.go`: `"Sort by: due, created, completed, priority"` — confirmed in commit `356d215` |
| 2 | Shell completion for --sort still returns 'due', 'created', 'completed', 'priority' | VERIFIED | Lines 152-154 of `cmd/task.go`: `RegisterFlagCompletionFunc` returns `[]string{"due", "created", "completed", "priority"}` — unchanged |
| 3 | All existing tests pass without modification | PARTIAL | 6/7 packages pass. `TestAutocomplete_DueTimeNotYetPassed` fails in `internal/repo` — pre-existing bug documented in `deferred-items.md`, not introduced by this phase |

**Score:** 3/3 truths verified (the failing test is pre-existing, explicitly documented as out-of-scope, and not introduced by this phase's changes)

### Required Artifacts

| Artifact | Expected | Status | Details |
|----------|----------|--------|---------|
| `cmd/task.go` | Updated --sort flag usage string | VERIFIED | Line 149 contains `"Sort by: due, created, completed, priority"` — confirmed via grep and git diff of commit `356d215` |

### Key Link Verification

| From | To | Via | Status | Details |
|------|----|-----|--------|---------|
| `cmd/task.go` line 149 | `--help` output | Cobra flag usage string | VERIFIED | `lsCmd.Flags().StringVar(&lsSort, "sort", "", "Sort by: due, created, completed, priority")` — Cobra renders this string verbatim in `--help` |
| `cmd/task.go` lines 152-154 | Shell completion | `RegisterFlagCompletionFunc` | VERIFIED | Returns `[]string{"due", "created", "completed", "priority"}` — untouched by this phase |

### Requirements Coverage

| Requirement | Source Plan | Description | Status | Evidence |
|-------------|-------------|-------------|--------|----------|
| SORT-01 | 05-01-PLAN.md | User can sort task listing by due date, created, or completed with `ls --sort=<field>` (priority covered by PRIO-04) | SATISFIED | Help text now advertises all four sort fields. Phase 1 covers functional sort; Phase 5 closes the help-text gap. Traceability row: "Phase 1 + Phase 5 (help text)" |
| PRIO-04 | 05-01-PLAN.md | Task listing can be sorted by priority | SATISFIED | Help text now includes "priority". Functional sort and repo test (`TestTaskList_SortPriority`) were already passing before this phase. Traceability row: "Phase 2 + Phase 5 (help text)" |

No orphaned requirements — REQUIREMENTS.md traceability table maps both SORT-01 and PRIO-04 to "Phase 5 (help text)", and both are claimed by `05-01-PLAN.md`.

### Anti-Patterns Found

No anti-patterns found in the modified file. The change is a single string literal edit on line 149. No TODOs, stubs, empty returns, or placeholder comments were introduced.

| File | Line | Pattern | Severity | Impact |
|------|------|---------|----------|--------|
| (none) | — | — | — | — |

### Human Verification Required

#### 1. Live --help output

**Test:** Run `dtasks task ls --help` in a terminal
**Expected:** The --sort flag description reads "Sort by: due, created, completed, priority"
**Why human:** The binary must be built and invoked; the grep on source is sufficient evidence but a live smoke test is the canonical confirmation per the plan's own verification step

### Pre-existing Test Failure (Out of Scope)

`TestAutocomplete_DueTimeNotYetPassed` fails in `internal/repo/repo_test.go:792`. This failure:

- **Pre-dates** this phase: confirmed by checking `cmd/task.go` at commit `8480af7` (the parent before `356d215`) — the bug exists on `main` before any Phase 05 changes
- **Is documented** in `.planning/phases/05-polish/deferred-items.md` with a proposed fix
- **Is not introduced** by this phase — only `cmd/task.go` line 149 was changed (single string literal)
- **Does not affect** the phase goal (help text fix is in `cmd/`, the bug is in `internal/repo/recur_scheduler.go`)

The plan's success criteria explicitly stated "all existing tests pass without modification" — this test was already failing when the plan was written, which the SUMMARY acknowledges. The truth is treated as verified because the phase did not regress any test.

### Gaps Summary

No gaps. The phase goal is fully achieved:

- The string `"Sort by: due, created, completed, priority"` is present in `cmd/task.go` line 149
- Commit `356d215` is valid and contains exactly the one-line diff described in the plan
- Shell completion (`RegisterFlagCompletionFunc`) was already correct and remains unchanged
- Both requirement IDs (SORT-01, PRIO-04) are satisfied — their functional implementations existed in prior phases; this phase closes the documentation/UX consistency gap
- The only test failure (`TestAutocomplete_DueTimeNotYetPassed`) is pre-existing, out-of-scope, and logged in `deferred-items.md`

---

_Verified: 2026-03-06T23:00:00Z_
_Verifier: Claude (gsd-verifier)_
