# Phase 5: Polish - Research

**Researched:** 2026-03-06
**Domain:** Cobra CLI flag usage strings / shell completion
**Confidence:** HIGH

## Summary

Phase 5 is a single-line fix: the `--sort` flag in `cmd/task.go` has a help text that omits `"priority"` as a valid sort field, but the shell completion function for the same flag already returns `"priority"` as a valid value. The implementation (repo layer) correctly supports `SortBy: "priority"` and is covered by `TestTaskList_SortPriority`. The gap is purely cosmetic — one string literal in one file.

No new packages, patterns, or architecture decisions are needed. The entire phase is a one-line change to `lsCmd.Flags().StringVar(...)` in `cmd/task.go` line 149, followed by verifying `go test ./...` stays green.

**Primary recommendation:** Edit line 149 of `cmd/task.go` — change `"Sort by: due, created, completed"` to `"Sort by: due, created, completed, priority"`. No other files need to change.

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|-----------------|
| SORT-01 | User can sort task listing by due date, created, or completed with `ls --sort=<field>`; help text must advertise all valid values including priority | Fix string on line 149 of `cmd/task.go`; completion func already correct |
| PRIO-04 | Task listing can be sorted by priority (`ls --sort=priority`) | Fully implemented in repo layer (`SortBy: "priority"`); help text is the only missing piece |
</phase_requirements>

## Standard Stack

### Core

| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| `github.com/spf13/cobra` | v1.8.0 | CLI framework, flag definitions, shell completion | Already in use throughout the project |

No new dependencies. No installation step needed.

## Architecture Patterns

### Cobra Flag Definition Pattern

Flag usage strings are set at registration time in `init()` via `Flags().StringVar(...)`. The fourth argument is the usage string shown in `--help` output.

```go
// cmd/task.go — current (WRONG)
lsCmd.Flags().StringVar(&lsSort, "sort", "", "Sort by: due, created, completed")

// cmd/task.go — after fix (CORRECT)
lsCmd.Flags().StringVar(&lsSort, "sort", "", "Sort by: due, created, completed, priority")
```

The shell completion function for the same flag is defined immediately below and already includes `"priority"`:

```go
// cmd/task.go lines 152-154 — already correct, do NOT change
_ = lsCmd.RegisterFlagCompletionFunc("sort", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
    return []string{"due", "created", "completed", "priority"}, cobra.ShellCompDirectiveNoFileComp
})
```

### Exact Change Location

| File | Line | Current value | New value |
|------|------|---------------|-----------|
| `cmd/task.go` | 149 | `"Sort by: due, created, completed"` | `"Sort by: due, created, completed, priority"` |

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Help text update | A separate validation map or enum | Edit the string literal directly | Cobra reads the usage string at definition time; no runtime indirection needed |

## Common Pitfalls

### Pitfall 1: Changing the completion function when it is already correct
**What goes wrong:** Modifying `RegisterFlagCompletionFunc` unnecessarily introduces diff noise and risk.
**How to avoid:** The completion function on lines 152-154 already returns `["due", "created", "completed", "priority"]`. Leave it untouched.

### Pitfall 2: Assuming the repo layer needs changes
**What goes wrong:** Wasting time searching for missing sort logic.
**Why it doesn't apply:** `TestTaskList_SortPriority` (repo_test.go line 1142) already passes. The repo layer is complete.

### Pitfall 3: Breaking existing tests
**What goes wrong:** Touching unrelated code during the fix.
**How to avoid:** Change only the one string literal. Run `go test ./...` to confirm green.

## Validation Architecture

### Test Framework

| Property | Value |
|----------|-------|
| Framework | Go testing (stdlib) |
| Config file | none |
| Quick run command | `go test ./cmd/... -run TestTaskList -v` |
| Full suite command | `go test ./...` |

### Phase Requirements → Test Map

| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| SORT-01 | `--help` output includes "priority" as sort field | manual smoke | `go build ./... && ./dist/dtasks task ls --help` | ✅ (no automated test; string is in binary) |
| PRIO-04 | `ls --sort=priority` returns tasks ordered by priority | unit | `go test ./internal/repo/... -run TestTaskList_SortPriority -v` | ✅ repo_test.go:1142 |

### Sampling Rate

- **Per task commit:** `go test ./...`
- **Per wave merge:** `go test ./...`
- **Phase gate:** Full suite green before `/gsd:verify-work`

### Wave 0 Gaps

None — existing test infrastructure covers all phase requirements. No new test files needed.

## Sources

### Primary (HIGH confidence)

- Direct code inspection: `cmd/task.go` lines 149, 152-154 — flag definition and completion func
- Direct code inspection: `internal/repo/repo_test.go` lines 1142-1171 — `TestTaskList_SortPriority` confirms PRIO-04 is implemented
- `go test ./...` output — all packages green as of 2026-03-06

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — no new dependencies; Cobra already in use
- Architecture: HIGH — single string literal change, pattern confirmed by adjacent flags
- Pitfalls: HIGH — gap identified by direct code inspection, not inference

**Research date:** 2026-03-06
**Valid until:** Not time-sensitive; stable until `cmd/task.go` is restructured
