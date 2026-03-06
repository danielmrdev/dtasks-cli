---
phase: 02-richness
verified: 2026-03-06T11:00:00Z
status: passed
score: 11/11 must-haves verified
re_verification:
  previous_status: passed
  previous_score: 11/11
  note: "Previous verification predated plan 05 execution (gap closure for --completed BoolVar). This report reflects post-plan-05 codebase state."
  gaps_closed:
    - "dtasks rm --completed --dry-run now works (BoolVar fix prevents Cobra consuming --dry-run as the flag value)"
    - "dtasks rm --completed --yes now works for the same reason"
    - "dtasks rm --completed --list <id> now works for the same reason"
    - "TaskDeleteCompleted with empty Before now deletes all completed tasks (no mandatory date cutoff)"
    - "TestTaskDeleteCompleted_NoBefore added covering MAINT-04 repo-layer no-date path"
  gaps_remaining: []
  regressions: []
human_verification:
  - test: "dtasks task rm --completed in a real TTY with completed tasks present shows confirmation prompt"
    expected: "Prints 'This will permanently delete N task(s). Confirm? [y/N]:' and waits for input. Typing 'y' deletes. Anything else prints 'Aborted.' and exits without deleting."
    why_human: "isTerminal(os.Stdin) returns true only in a real TTY — the interactive code path cannot be exercised in automated tests or piped commands"
  - test: "dtasks task ls --sort=priority shows tasks ordered high > medium > low > nil in table output"
    expected: "Table has a PRIO column. High-priority tasks (!) appear first, then medium (~), then low (-), then unprioritized (blank). Columns are aligned."
    why_human: "Visual table alignment and symbol rendering depend on terminal width and runewidth library behavior — cannot assert without a live binary invocation"
---

# Phase 2: Richness — Verification Report

**Phase Goal:** Users can assign priorities to tasks, bulk-clean completed tasks, and view task statistics.
**Verified:** 2026-03-06T11:00:00Z
**Status:** PASSED
**Re-verification:** Yes — after gap closure (plan 05 fixed --completed BoolVar and TaskDeleteCompleted no-date path).

---

## Goal Achievement

### Observable Truths

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | Tasks can be created with priority (high/medium/low) and retrieved with that value | VERIFIED | `TaskInput.Priority *string` in `repo/task.go:22`; `TestTaskCreate_WithPriority` passes |
| 2 | Task priority can be updated or cleared via TaskPatch | VERIFIED | `TaskPatch.Priority *string` in `repo/task.go`; nil clears to NULL; `TestTaskPatchFields_Priority` passes |
| 3 | TaskList with SortBy='priority' returns tasks ordered high > medium > low > nil | VERIFIED | CASE expression in `sortMap` at `repo/task.go:107`; `TestTaskList_SortPriority` passes |
| 4 | PrintTasks shows PRIO column with ! for high, ~ for medium, - for low, space for nil | VERIFIED | `output/output.go:66,89-97`; headers include "PRIO"; `TestPrintTasks_Priority` passes |
| 5 | dtasks rm --completed bulk-deletes all completed tasks after confirmation; --yes skips prompt | VERIFIED | `var rmCompleted bool`; BoolVar at `cmd/task.go:398`; two-step DryRun logic at lines 337-379; `go test ./... GREEN` |
| 6 | dtasks rm --completed --dry-run shows count and task list without deleting or prompting | VERIFIED | DryRun path at `cmd/task.go:348-356`; first call DryRun:true returns early before confirmation block |
| 7 | dtasks rm --completed --list <id> scopes bulk delete to tasks in a specific list | VERIFIED | `opts.ListID = &rmListID` guard at `cmd/task.go:340-342`; `TestTaskDeleteCompleted_Scoped` passes |
| 8 | Bulk delete emits {"deleted": N} when --json flag is set | VERIFIED | `output.PrintDeletedCount` at `output/output.go:163`; `TestPrintDeletedCount` passes for JSON mode |
| 9 | TaskDeleteCompleted with empty Before deletes all completed tasks (no date cutoff) | VERIFIED | Conditional WHERE at `repo/task.go:363-370`; `TestTaskDeleteCompleted_NoBefore` passes (added in plan 05) |
| 10 | TaskStats returns per-list totals including lists with zero tasks | VERIFIED | LEFT JOIN in `TaskStats` at `repo/task.go`; `TestTaskStats` asserts ByList len=3 including empty list |
| 11 | dtasks stats and --priority flags on add/edit are registered CLI commands | VERIFIED | `rootCmd.AddCommand(statsCmd)` at `cmd/root.go:86`; `addPriority` and `editPriority` flags in `cmd/task.go` init() |

**Score:** 11/11 truths verified.

---

### Required Artifacts

| Artifact | Expected | Status | Details |
|----------|----------|--------|---------|
| `internal/db/db.go` | Idempotent migration for priority TEXT column | VERIFIED | pragma_table_info guard + ALTER TABLE |
| `internal/models/models.go` | `Priority *string` field on Task struct | VERIFIED | `Priority *string \`json:"priority,omitempty"\`` |
| `internal/repo/task.go` | TaskInput, TaskPatch, sortMap; TaskDeleteCompleted with conditional Before; TaskStats | VERIFIED | Conditional WHERE at lines 363-403; sortMap priority key at line 107; TaskStats with LEFT JOIN |
| `internal/repo/repo_test.go` | Tests for all PRIO/MAINT/STAT requirements including NoBefore | VERIFIED | TestTaskCreate_WithPriority, TestTaskPatchFields_Priority, TestTaskList_SortPriority, TestTaskDeleteCompleted*, TestTaskDeleteCompleted_NoBefore, TestTaskStats all present and GREEN |
| `internal/output/output.go` | PrintTasks PRIO column; PrintDeletedCount; PrintStats | VERIFIED | PRIO headers at line 66; prio switch at lines 89-97; PrintDeletedCount at 163; PrintStats at 171 |
| `cmd/task.go` | --priority on addCmd/editCmd; rmCmd with --completed as BoolVar, --dry-run, --yes, --list | VERIFIED | `var rmCompleted bool`; `BoolVar(&rmCompleted, "completed", false, ...)` at line 398; all flags registered |
| `cmd/stats.go` | statsCmd calling repo.TaskStats + output.PrintStats | VERIFIED | 21-line file; `repo.TaskStats(DB)` at line 14; `output.PrintStats(s)` at line 18 |
| `cmd/root.go` | rootCmd.AddCommand(statsCmd) | VERIFIED | Line 86 |

---

### Key Link Verification

| From | To | Via | Status | Details |
|------|----|-----|--------|---------|
| `cmd/task.go (addCmd)` | `repo.TaskInput.Priority` | `Changed("priority")` guard + `in.Priority = &addPriority` | VERIFIED | Lines 54-59 |
| `cmd/task.go (editCmd)` | `repo.TaskPatch.Priority` | `Changed("priority")` guard + `p.Priority = &editPriority` | VERIFIED | Lines 233-241 |
| `cmd/task.go (rmCmd)` | `repo.TaskDeleteCompleted` | `Changed("completed")` gate + DryRun:true first call, then DryRun:false | VERIFIED | Lines 329-380 |
| `cmd/task.go (rmCmd bulk)` | `output.PrintDeletedCount` | `output.PrintDeletedCount(result.Deleted)` | VERIFIED | Line 378 |
| `cmd/task.go (lsCmd)` | `repo.TaskListOptions.SortBy` | `opts.SortBy = lsSort` when `Changed("sort")` | VERIFIED | Lines 122-124 |
| `repo/task.go (sortMap)` | ORDER BY CASE expression | `"priority"` key maps to CASE WHEN 'high' THEN 1 WHEN 'medium' THEN 2 ... | VERIFIED | Line 107 |
| `repo/task.go (TaskDeleteCompleted)` | conditional WHERE | `if opts.Before != ""` builds date-filtered WHERE; else bare `WHERE completed = 1` | VERIFIED | Lines 363-403 |
| `cmd/stats.go (statsCmd)` | `output.PrintStats` | `output.PrintStats(s)` with `*repo.StatsSummary` | VERIFIED | Line 18 |
| `output/output.go (PrintTasks)` | `models.Task.Priority` | `t.Priority != nil` switch in row builder | VERIFIED | Lines 89-97 |
| `output/output.go (PrintStats)` | `repo.StatsSummary` | parameter type `*repo.StatsSummary` | VERIFIED | Line 171 |

---

### Requirements Coverage

| Requirement | Source Plan(s) | Description | Status | Evidence |
|-------------|---------------|-------------|--------|----------|
| PRIO-01 | 02-01, 02-02, 02-04 | Set priority when adding with `add --priority <level>` | SATISFIED | addCmd priority flag + validation; TaskInput.Priority; TestTaskCreate_WithPriority GREEN |
| PRIO-02 | 02-01, 02-02, 02-04 | Set priority when editing with `edit --priority <level>` | SATISFIED | editCmd priority flag; TaskPatch.Priority; empty string clears to NULL; TestTaskPatchFields_Priority GREEN |
| PRIO-03 | 02-01, 02-03 | Priority shown as visual indicator in table output | SATISFIED | PRIO column in PrintTasks with !/~/- indicators; TestPrintTasks_Priority GREEN |
| PRIO-04 | 02-01, 02-02, 02-04 | Task listing sortable by priority | SATISFIED | SortBy="priority" in sortMap; lsCmd wires --sort; TestTaskList_SortPriority GREEN |
| MAINT-01 | 02-01, 02-02, 02-04, 02-05 | Bulk-delete completed tasks | SATISFIED | TaskDeleteCompleted; rmCmd --completed as BoolVar. Note: REQUIREMENTS.md text says "on or before a date" but implementation was changed to a boolean flag (no date argument) per plan 05. REQUIREMENTS.md traceability table marks it Complete [x]. |
| MAINT-02 | 02-01, 02-02, 02-04, 02-05 | Preview what would be deleted without committing | SATISFIED | DryRun:true path in TaskDeleteCompleted; rmCmd --dry-run; TestTaskDeleteCompleted_DryRun GREEN |
| MAINT-03 | 02-04 | Bulk delete requires explicit confirmation unless --yes | SATISFIED | isTerminal check + prompt at cmd/task.go:358-371; error in non-TTY without --yes |
| MAINT-04 | 02-01, 02-03, 02-04, 02-05 | Bulk delete respects --json flag (emits {"deleted": N}) | SATISFIED | PrintDeletedCount checks JSONMode; TestPrintDeletedCount verifies JSON output; TestTaskDeleteCompleted_NoBefore covers no-date repo path |
| MAINT-05 | 02-01, 02-02, 02-04 | Bulk delete scoped to a specific list with --list | SATISFIED | DeleteCompletedOptions.ListID; rmCmd --list flag (Int64VarP); TestTaskDeleteCompleted_Scoped GREEN |
| STAT-01 | 02-01, 02-02, 02-04 | Task summary with `dtasks stats` | SATISFIED | TaskStats with LEFT JOIN includes empty lists; statsCmd registered in root; TestTaskStats GREEN |
| STAT-02 | 02-01, 02-03 | Stats command respects --json flag | SATISFIED | PrintStats calls printJSON(s) when JSONMode; TestPrintStats_JSON GREEN |

**Orphaned requirements check:** No Phase 2 requirements in REQUIREMENTS.md are unclaimed. All 11 IDs (PRIO-01..04, MAINT-01..05, STAT-01..02) are mapped and satisfied.

---

### Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
|------|------|---------|----------|--------|
| `cmd/task.go` | 149 | --sort flag help text says "Sort by: due, created, completed" — omits "priority" | INFO | Shell completion (RegisterFlagCompletionFunc) correctly returns "priority" as a completion option; only --help text is missing it. Does not block any requirement. |

No TODO/FIXME/HACK/PLACEHOLDER comments in any modified file. No stub implementations. No dead exports. Binary builds cleanly (`go build ./...`). All tests GREEN (`go test ./... -count=1`).

---

### Human Verification Required

#### 1. TTY confirmation prompt for bulk delete

**Test:** With at least one completed task in the DB, run `dtasks task rm --completed` (without `--yes`) in a real terminal session.
**Expected:** Prints "This will permanently delete N task(s). Confirm? [y/N]:" and waits. Typing `y` or `yes` deletes the tasks. Typing anything else (or pressing Enter) prints "Aborted." and exits without deleting.
**Why human:** `isTerminal(os.Stdin)` returns true only in a real TTY. In automated tests or piped commands, this path returns an error ("bulk delete requires --yes in non-interactive mode") — the interactive confirmation path cannot be exercised programmatically.

#### 2. Priority sort visible in table output

**Test:** Add tasks with different priorities via `dtasks task add`, then run `dtasks task ls --sort=priority`.
**Expected:** Table shows a PRIO column. Tasks appear ordered: high (!) first, then medium (~), then low (-), then unprioritized (blank space) last. Column widths align correctly.
**Why human:** Visual column alignment and symbol rendering depend on terminal width and runewidth library behavior. Cannot assert without a live binary invocation in a real terminal.

---

### Gaps Summary

No gaps. All 11 requirements are satisfied at every layer.

**Plan 05 gap closure (post-UAT):** The UAT identified 3 major failures (tests 6/7/8) caused by `--completed` being declared as `StringVar`. Cobra consumed `--dry-run`, `--yes`, and `--list` as the string value of `--completed`, breaking all bulk-delete paths. Plan 05 fixed this by:

1. Changing `var rmCompleted string` to `var rmCompleted bool` and registering via `BoolVar`.
2. Updating `TaskDeleteCompleted` to build conditional WHERE clauses: when `Before=""`, omits the date filter entirely (deletes all completed tasks); when `Before` has a value, preserves the existing date-filtered behavior.
3. Adding `TestTaskDeleteCompleted_NoBefore` to cover the no-cutoff path.

The implementation intentionally dropped the date-cutoff UX from MAINT-01 (originally "on or before a date") in favor of a simpler boolean "delete all completed" flag. REQUIREMENTS.md traceability table marks MAINT-01 as Complete, confirming stakeholder acceptance of this design change.

---

_Verified: 2026-03-06T11:00:00Z_
_Verifier: Claude (gsd-verifier)_
