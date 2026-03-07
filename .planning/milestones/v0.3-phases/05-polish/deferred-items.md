# Deferred Items — Phase 05-polish

## Out-of-scope issues discovered during execution

### Bug: TestAutocomplete_DueTimeNotYetPassed fails

**Discovered during:** 05-01 Task 1 execution
**Status:** Pre-existing on main before 05-01 changes

**Description:**
`internal/repo/repo_test.go:773` - `TestAutocomplete_DueTimeNotYetPassed` fails because
`ProcessAutocompleteTasks` completes tasks with `due_date = today` regardless of `due_time`.
The scheduler should only complete a task when both `due_date <= today` AND `due_time <= current_time`.

**Failure message:**
```
--- FAIL: TestAutocomplete_DueTimeNotYetPassed (0.00s)
    repo_test.go:792: expected task NOT to be completed (due_time not yet passed)
```

**Files involved:** `internal/repo/recur_scheduler.go`

**Action needed:** Fix `ProcessAutocompleteTasks` SQL or Go logic to include `due_time` check
when deciding whether to auto-complete a task. Candidate fix: add `AND (due_time IS NULL OR due_time <= strftime('%H:%M', 'now', 'localtime'))` to the autocomplete query.
