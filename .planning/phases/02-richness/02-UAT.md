---
status: diagnosed
phase: 02-richness
source: 02-01-SUMMARY.md, 02-02-SUMMARY.md, 02-03-SUMMARY.md, 02-04-SUMMARY.md
started: 2026-03-06T09:26:57Z
updated: 2026-03-06T09:35:00Z
---

## Current Test

[testing complete]

## Tests

### 1. Cold Start Smoke Test
expected: Run `make build` then `dist/dtasks task ls` (or any list). Binary builds without errors. Command returns output (even if empty) without crashing. Database migration for priority column applies silently on first run.
result: pass

### 2. Add task with priority
expected: Run `dist/dtasks task add "Test priority task" --priority high` (adjust list if needed). Command succeeds. Running `dist/dtasks task ls` shows `!` in the PRIO column for that task.
result: pass

### 3. Edit task priority
expected: Run `dist/dtasks task edit <id> --priority low`. Task priority updates to low. Running `dist/dtasks task ls` shows `-` in the PRIO column. Then run `dist/dtasks task edit <id> --priority ""` — priority clears, PRIO column shows blank.
result: pass
note: "command is `dtasks edit` not `dtasks task edit` — user noted naming inconsistency (out of scope for phase 2)"

### 4. Task list shows PRIO column
expected: Run `dist/dtasks task ls`. Output table has a PRIO column. Tasks with `high` show `!`, `medium` show `~`, `low` show `-`, tasks with no priority show blank/space.
result: pass

### 5. Sort tasks by priority
expected: Create tasks with different priorities (high/medium/low/none). Run `dist/dtasks task ls --sort priority`. Tasks appear ordered: high first, then medium, then low, then those with no priority.
result: pass

### 6. Bulk delete completed — dry-run preview
expected: Mark some tasks as done. Run `dist/dtasks task rm --completed --dry-run`. Command shows how many completed tasks would be deleted but does NOT actually delete them. Running `dist/dtasks task ls` confirms tasks still exist.
result: issue
reported: "dry-run muestra mensaje: This will permanently delete 0 task(s). Confirm? [y/N]: y\nDeleted 0 task(s).\n\nincorrecto. además hay tareas completadas y muestra 0 -> mal"
severity: major

### 7. Bulk delete completed — with confirmation
expected: Run `dist/dtasks task rm --completed`. Command shows count of tasks to delete and prompts for confirmation. Typing `y` deletes them. Running `dist/dtasks task ls` confirms they are gone. With `--yes` flag, skips prompt and deletes directly.
result: issue
reported: "error de concepto: `dist/dtasks rm --completed` → Error: flag needs an argument (--completed string: on or before YYYY-MM-DD). Con --dry-run salta directo al prompt de confirmación. Con --yes muestra 0 tareas y sigue mostrando el prompt."
severity: major

### 8. Bulk delete scoped to a list
expected: Have completed tasks in multiple lists. Run `dist/dtasks task rm --completed --list <listname>`. Only completed tasks from that specific list are deleted. Completed tasks in other lists remain.
result: issue
reported: "dist/dtasks rm --completed --list 2 → Error: --completed cannot be used with a task ID argument"
severity: major

### 9. Stats command — table output
expected: Run `dist/dtasks stats`. Output shows a table with columns: LIST, Total, Pending, Done, Done%. Each list appears as a row with correct counts. A summary totals row appears.
result: pass

### 10. Stats command — JSON output
expected: Run `dist/dtasks stats --json`. Output is valid JSON with stats data (total, pending, done counts per list). No table formatting, pure JSON.
result: pass

## Summary

total: 10
passed: 7
issues: 3
pending: 0
skipped: 0

## Gaps

- truth: "dry-run shows count of completed tasks that would be deleted without prompting or deleting"
  status: failed
  reason: "User reported: dry-run muestra mensaje: This will permanently delete 0 task(s). Confirm? [y/N]: y\nDeleted 0 task(s).\n\nincorrecto. además hay tareas completadas y muestra 0 -> mal"
  severity: major
  test: 6
  root_cause: "cmd/task.go:399 declara --completed como StringVar en lugar de BoolVar. Cobra consume '--dry-run' como valor string de --completed, rmDryRun queda false, y la query recibe '--dry-run' como fecha cutoff → date('--dry-run') = NULL → 0 filas"
  artifacts:
    - path: "cmd/task.go"
      issue: "line 316: var rmCompleted string; line 399: StringVar instead of BoolVar"
    - path: "internal/repo/task.go"
      issue: "line 363: query requires Before date string; needs to handle boolean/no-date case"
  missing:
    - "Change --completed to BoolVar in cmd/task.go"
    - "Update DeleteCompletedOptions and TaskDeleteCompleted to work without mandatory date cutoff"
    - "Update rmCmd logic to call TaskDeleteCompleted with DryRun=true first, show count, then prompt"

- truth: "rm --completed is a boolean flag that deletes all completed tasks; --yes skips confirmation prompt; count is accurate"
  status: failed
  reason: "User reported: error de concepto: `dist/dtasks rm --completed` → Error: flag needs an argument (--completed string: on or before YYYY-MM-DD). Con --dry-run salta directo al prompt de confirmación. Con --yes muestra 0 tareas y sigue mostrando el prompt."
  severity: major
  test: 7
  root_cause: "Same root: --completed StringVar causes Cobra to consume '--yes' as its value. rmYes stays false → prompt always appears. Query gets '--yes' as date → 0 matches."
  artifacts:
    - path: "cmd/task.go"
      issue: "line 399: StringVar instead of BoolVar — cascades to --yes and --dry-run not being parsed"
  missing:
    - "Change --completed to BoolVar; --yes and --dry-run will parse correctly after that fix"

- truth: "rm --completed --list <name> scopes bulk delete to tasks in a specific list"
  status: failed
  reason: "User reported: dist/dtasks rm --completed --list 2 → Error: --completed cannot be used with a task ID argument"
  severity: major
  test: 8
  root_cause: "Same root: --completed StringVar causes Cobra to consume '--list' as its value, leaving '2' as positional arg[0]. Guard at cmd/task.go:329-332 detects len(args)>0 and errors."
  artifacts:
    - path: "cmd/task.go"
      issue: "lines 329-332: guard fires because --list value is consumed by --completed StringVar, leaving numeric arg"
  missing:
    - "Change --completed to BoolVar; --list will parse correctly after that fix"
