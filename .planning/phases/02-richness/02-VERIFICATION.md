---
phase: 02-richness
verified: 2026-03-06T10:00:00Z
status: passed
score: 11/11 must-haves verified
re_verification: false
human_verification:
  - test: "dtasks rm --completed 2026-01-01 without --yes in a TTY should prompt for confirmation"
    expected: "Prints 'This will permanently delete N task(s). Confirm? [y/N]:' and waits for input"
    why_human: "isTerminal(os.Stdin) is true only in a real TTY — cannot assert in CI or piped test"
  - test: "dtasks task ls --sort=priority in a real binary shows tasks ordered high > medium > low > nil"
    expected: "Table output with high-priority task first, nil-priority task last"
    why_human: "Visual ordering in table output cannot be asserted without a real DB and binary invocation"
---

# Phase 2: Richness — Verification Report

**Phase Goal:** Deliver task priority (high/medium/low), bulk-delete completed tasks, and task statistics to users via the CLI.
**Verified:** 2026-03-06T10:00:00Z
**Status:** PASSED
**Re-verification:** No — initial verification.

---

## Goal Achievement

### Observable Truths

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | Tasks can be created with priority (high/medium/low) and retrieved with that value | VERIFIED | `TaskInput.Priority *string` in `repo/task.go:22`; `TestTaskCreate_WithPriority` passes |
| 2 | Task priority can be updated via TaskPatchFields | VERIFIED | `TaskPatch.Priority *string` in `repo/task.go:222`; nil clears to NULL; `TestTaskPatchFields_Priority` passes |
| 3 | TaskList with SortBy='priority' returns tasks ordered high > medium > low > nil | VERIFIED | CASE expression in `sortMap` at `repo/task.go:107`; `TestTaskList_SortPriority` passes |
| 4 | PrintTasks shows PRIO column with ! for high, ~ for medium, - for low, space for nil | VERIFIED | `output/output.go:66,89-97`; headers include "PRIO"; `TestPrintTasks_Priority` passes |
| 5 | TaskDeleteCompleted deletes completed tasks on or before a date with optional list scope | VERIFIED | `repo.TaskDeleteCompleted` at `repo/task.go:362`; `TestTaskDeleteCompleted` and `TestTaskDeleteCompleted_Scoped` pass |
| 6 | TaskDeleteCompleted with DryRun=true returns task list without deleting | VERIFIED | DryRun branch at `repo/task.go:385-387`; `TestTaskDeleteCompleted_DryRun` passes |
| 7 | Bulk delete requires confirmation unless --yes is passed; errors in non-TTY without --yes | VERIFIED | `cmd/task.go:359-371`; `isTerminal` check at line 360; error path at line 361 |
| 8 | Bulk delete emits {"deleted": N} when --json flag is set | VERIFIED | `output.PrintDeletedCount` at `output/output.go:163`; `TestPrintDeletedCount` passes for JSON mode |
| 9 | TaskStats returns per-list totals including lists with zero tasks | VERIFIED | LEFT JOIN in `TaskStats` at `repo/task.go:424`; `TestTaskStats` asserts `ByList` len=3 including empty list |
| 10 | PrintStats prints per-list breakdown; respects --json flag | VERIFIED | `output.PrintStats` at `output/output.go:171`; `TestPrintStats_Table` and `TestPrintStats_JSON` pass |
| 11 | dtasks stats and dtasks task add/edit --priority are registered CLI commands | VERIFIED | `rootCmd.AddCommand(statsCmd)` at `cmd/root.go:86`; priority flags on addCmd/editCmd in `cmd/task.go` |

**Score:** 11/11 truths verified.

---

### Required Artifacts

| Artifact | Expected | Status | Details |
|----------|----------|--------|---------|
| `internal/db/db.go` | Idempotent migration for priority TEXT column | VERIFIED | Lines 107-114: pragma_table_info guard + ALTER TABLE |
| `internal/models/models.go` | `Priority *string` field on Task struct | VERIFIED | Line 35: `Priority *string \`json:"priority,omitempty"\`` |
| `internal/repo/task.go` | Extended TaskInput, TaskPatch, TaskListOptions; new TaskDeleteCompleted, TaskStats | VERIFIED | All exported; TaskInput.Priority:22, TaskPatch.Priority:222, sortMap priority:107, TaskDeleteCompleted:362, TaskStats:424 |
| `internal/output/output.go` | PrintTasks PRIO column; PrintDeletedCount; PrintStats | VERIFIED | PRIO headers:66, prio switch:89-97, PrintDeletedCount:163, PrintStats:171 |
| `cmd/task.go` | --priority flag on addCmd/editCmd; rmCmd --completed/--dry-run/--yes/--list | VERIFIED | addPriority:24, editPriority:197, rmCmd bulk path:329-380, flags registered in init() |
| `cmd/stats.go` | statsCmd calling repo.TaskStats + output.PrintStats | VERIFIED | Full file: 21 lines, calls repo.TaskStats(DB) and output.PrintStats(s) |
| `cmd/root.go` | rootCmd.AddCommand(statsCmd) | VERIFIED | Line 86 |

---

### Key Link Verification

| From | To | Via | Status | Details |
|------|----|-----|--------|---------|
| `repo/task.go (taskSelectSQL)` | `models.Task.Priority` | `scanTaskRow` scans `&t.Priority` | VERIFIED | `taskSelectSQL` includes `t.priority` at line 482; `scanTaskRow` scans `&t.Priority` at line 506 |
| `repo/task.go (TaskList)` | `sortMap priority entry` | CASE expression in ORDER BY | VERIFIED | `repo/task.go:107`: `"CASE t.priority WHEN 'high' THEN 1 WHEN 'medium' THEN 2 WHEN 'low' THEN 3 ELSE 4 END ASC, t.created_at ASC"` |
| `output/output.go (PrintStats)` | `repo.StatsSummary` | parameter type `*repo.StatsSummary` | VERIFIED | `output/output.go:171`: `func PrintStats(s *repo.StatsSummary)` |
| `output/output.go (PrintTasks)` | `models.Task.Priority` | `t.Priority != nil` check | VERIFIED | `output/output.go:91-96`: switch on `t.Priority` value |
| `cmd/task.go (addCmd)` | `repo.TaskInput.Priority` | `cmd.Flags().Changed("priority")` guard | VERIFIED | `cmd/task.go:54-59`: Changed guard before `in.Priority = &addPriority` |
| `cmd/task.go (rmCmd)` | `repo.TaskDeleteCompleted` | `--completed` flag triggers bulk delete path | VERIFIED | `cmd/task.go:329-381`: Changed("completed") gate, two-step DryRun:true then DryRun:false |
| `cmd/task.go (rmCmd bulk delete)` | `output.PrintDeletedCount` | `output.PrintDeletedCount(result.Deleted)` | VERIFIED | `cmd/task.go:379` |
| `cmd/stats.go (statsCmd)` | `output.PrintStats` | direct call with `*repo.StatsSummary` | VERIFIED | `cmd/stats.go:18`: `output.PrintStats(s)` |

---

### Requirements Coverage

| Requirement | Source Plan | Description | Status | Evidence |
|-------------|------------|-------------|--------|----------|
| PRIO-01 | 02-01, 02-02, 02-04 | Set priority when adding a task with `add --priority <level>` | SATISFIED | `addCmd` priority flag + validation; `TaskInput.Priority`; `TestTaskCreate_WithPriority` passes |
| PRIO-02 | 02-01, 02-02, 02-04 | Set priority when editing with `edit --priority <level>` | SATISFIED | `editCmd` priority flag; `TaskPatch.Priority`; empty string clears to NULL; `TestTaskPatchFields_Priority` passes |
| PRIO-03 | 02-01, 02-03 | Priority shown as visual indicator in table output | SATISFIED | PRIO column in PrintTasks with !/~/- indicators; `TestPrintTasks_Priority` passes |
| PRIO-04 | 02-01, 02-02, 02-04 | Task listing sortable by priority | SATISFIED | `SortBy="priority"` in sortMap; lsCmd wires `--sort`; `TestTaskList_SortPriority` passes |
| MAINT-01 | 02-01, 02-02, 02-04 | Bulk-delete completed tasks on or before a date | SATISFIED | `TaskDeleteCompleted` with `Before` param; `rmCmd --completed`; `TestTaskDeleteCompleted` passes |
| MAINT-02 | 02-01, 02-02, 02-04 | Preview what would be deleted without committing | SATISFIED | `DryRun=true` path in `TaskDeleteCompleted`; `rmCmd --dry-run`; `TestTaskDeleteCompleted_DryRun` passes |
| MAINT-03 | 02-04 | Bulk delete requires explicit confirmation unless --yes | SATISFIED | `isTerminal` check + prompt in `cmd/task.go:359-371`; error in non-TTY without `--yes` |
| MAINT-04 | 02-01, 02-03, 02-04 | Bulk delete respects --json flag (emits `{"deleted": N}`) | SATISFIED | `output.PrintDeletedCount` checks `JSONMode`; `TestPrintDeletedCount` verifies JSON output |
| MAINT-05 | 02-01, 02-02, 02-04 | Bulk delete scoped to a specific list with --list | SATISFIED | `DeleteCompletedOptions.ListID`; `rmCmd --list` flag; `TestTaskDeleteCompleted_Scoped` passes |
| STAT-01 | 02-01, 02-02, 02-04 | Task summary with `dtasks stats` (total, pending, done, % by list) | SATISFIED | `TaskStats` with LEFT JOIN includes empty lists; `statsCmd` registered; `TestTaskStats` passes |
| STAT-02 | 02-01, 02-03 | Stats command respects --json flag | SATISFIED | `PrintStats` calls `printJSON(s)` when `JSONMode`; `TestPrintStats_JSON` passes |

No orphaned requirements. All 11 Phase 2 requirements (PRIO-01..04, MAINT-01..05, STAT-01..02) are satisfied and covered by the 4 plans. MAINT-03 is only in plan 02-04 (not in 02-01 as it has no direct test — it is a CLI interaction requirement confirmed by code inspection).

---

### Anti-Patterns Found

No TODO/FIXME/HACK/PLACEHOLDER comments found in any modified Go file.
No stub implementations (empty returns, console-log-only handlers).
No dead exports.

One minor documentation gap (non-blocking):

| File | Line | Pattern | Severity | Impact |
|------|------|---------|----------|--------|
| `cmd/task.go` | 149 | `--sort` flag help text says "Sort by: due, created, completed" — omits "priority" | INFO | Shell completion correctly returns "priority"; user running `--help` would not see it listed |

---

### Human Verification Required

#### 1. TTY confirmation prompt for bulk delete

**Test:** Run `dtasks rm --completed 2026-01-01` (without `--yes`) in a real terminal with at least one completed task before that date.
**Expected:** Prints "This will permanently delete N task(s). Confirm? [y/N]:" and waits. Answering "y" deletes; anything else prints "Aborted." and returns.
**Why human:** `isTerminal(os.Stdin)` returns true only in a real TTY. The non-interactive code path (error on pipe) is verifiable by code inspection but the interactive prompt path requires manual exercise.

#### 2. Priority sort visible in table output

**Test:** Add tasks with different priorities via `dtasks task add`, then run `dtasks task ls --sort=priority`.
**Expected:** Table shows tasks ordered high (!) first, then medium (~), then low (-), then unprioritized (space) last. PRIO column present and aligned.
**Why human:** Visual column alignment and symbol rendering depend on terminal width and runewidth library behavior — cannot assert without a real display.

---

### Gaps Summary

No gaps. All 11 requirements are satisfied at every layer:

- **Data layer:** schema migration, `models.Task.Priority`, `TaskInput.Priority`, `TaskPatch.Priority`, `TaskDeleteCompleted`, `TaskStats`, `DeleteCompletedOptions`, `DeleteCompletedResult`, `ListStat`, `StatsSummary` — all present and exercised by passing tests.
- **Output layer:** PRIO column in `PrintTasks`, `PrintDeletedCount` (JSON + text modes), `PrintStats` (JSON + table modes) — all present and exercised by passing tests.
- **CLI layer:** `--priority` on `addCmd`/`editCmd` with validation, `rmCmd` with `--completed`/`--dry-run`/`--yes`/`--list`, `statsCmd` registered in root — all wired and confirmed by `go build ./...` passing cleanly.

Full test suite: `go test ./...` — all packages GREEN. No compilation errors. No anti-patterns.

The one documentation gap (missing "priority" in `--sort` flag description) is informational and does not block any requirement.

---

_Verified: 2026-03-06T10:00:00Z_
_Verifier: Claude (gsd-verifier)_
