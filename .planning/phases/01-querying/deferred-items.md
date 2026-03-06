---
phase: 01-querying
created: 2026-03-06
---

# Deferred Items — Phase 01 Querying

Items discovered during plan execution that are out of scope for the current plan.

## Pre-existing Test Failures (out of scope for Plan 02)

### TestAutocomplete_NotYetDue

**File:** `internal/repo/repo_test.go` line 668
**Status:** Failing before Plan 01 was started
**Cause:** Test uses hardcoded date `"2026-02-27"` as "tomorrow". As of 2026-03-06 this date is in the past, so the scheduler autocompletes it. The test expects it NOT to be autocompleted.
**Fix:** Change hardcoded dates to use `time.Now().AddDate(0, 0, 1)` for tomorrow.

### TestAutocomplete_DueTimePassed

**File:** `internal/repo/repo_test.go` line 796
**Status:** Failing before Plan 01 was started
**Cause:** Test expects a task with `DueDate=today, DueTime=pastTime (-2h)` to be autocompleted. The scheduler is not completing it.
**Fix:** Investigate the autocomplete scheduler logic for due_time comparison.
