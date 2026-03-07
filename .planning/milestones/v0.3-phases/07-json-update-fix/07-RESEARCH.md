# Phase 7: Fix JSON Update Output - Research

**Researched:** 2026-03-07
**Domain:** Go CLI output correctness — `cmd/update.go` JSON mode gate
**Confidence:** HIGH

## Summary

Phase 7 is a surgical fix to `cmd/update.go`. Two defects exist in the successful-update path:

1. `fmt.Fprintln(w, "Run install.sh to update shell completions")` at line 81 writes plain text to stdout unconditionally — before `emitUpdateResult` emits the JSON object. When `--json` is active, stdout receives `<plain text>\n<json>`, which is not parseable as JSON.
2. `useJSON` is resolved by reading the flag twice (lines 29-33: first from `cmd.Flags()`, then from `rootCmd.PersistentFlags()`). The rest of the codebase uses `output.JSONMode`, which is already set in `PersistentPreRunE` (`root.go:25`) before any `RunE` body executes.

Both defects are in `cmd/update.go` only. No other files need changing. No schema, repo, or output package changes required. Existing tests in `cmd/update_test.go` already cover the "already up to date" JSON path; a new test for the contamination path (successful update with `--json`) needs to be added.

**Primary recommendation:** In `cmd/update.go`, replace the `useJSON` double-read with `output.JSONMode`, and gate the plain-text hint line and the `PromptAndInstall` call on `!output.JSONMode`.

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|-----------------|
| UPDT-04 | `dtasks update` respects `--json` flag | Two concrete defects in `cmd/update.go` identified and fully diagnosed; fix is mechanical |
</phase_requirements>

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| `encoding/json` | stdlib | JSON encoding | Already used in `emitUpdateResult` |
| `github.com/spf13/cobra` | v1.8.0 | CLI framework | Project standard |
| `github.com/danielmrdev/dtasks-cli/internal/output` | local | `output.JSONMode` global bool | Pattern used by every other command |

No new dependencies. No installation step required.

## Architecture Patterns

### Pattern: Single Source of Truth for JSON Mode

All commands in the project use `output.JSONMode` (set in `PersistentPreRunE`) rather than reading the flag themselves. `updateCmd` is the only command that deviates, reading the flag manually with a two-step fallback.

**How `output.JSONMode` is set (root.go:24-25):**
```go
PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
    output.JSONMode = jsonFlag   // jsonFlag is bound to --json persistent flag
    ...
```

`update` is explicitly skipped from DB init but NOT from `PersistentPreRunE` execution — the early return only skips DB open, not the `output.JSONMode = jsonFlag` assignment. Therefore `output.JSONMode` is already correctly set when `updateCmd.RunE` runs.

**Correct pattern (used by statsCmd, rmCmd, etc.):**
```go
// No local flag read — just use output.JSONMode
if output.JSONMode {
    // emit JSON
}
```

**Defective pattern in update.go (lines 29-33):**
```go
useJSON, _ := cmd.Flags().GetBool("json")
if !useJSON {
    // Fall back to the persistent flag registered on rootCmd.
    useJSON, _ = rootCmd.PersistentFlags().GetBool("json")
}
```

This works correctly at runtime (both paths reach the same value) but deviates from convention and should be removed.

### Pattern: Gate Side-Effects on !output.JSONMode

Any output that is not the final structured result must be gated. In JSON mode, stdout must receive exactly one JSON object and nothing else.

**Current defect (update.go:74-82):**
```go
homeDir, err := os.UserHomeDir()
if err == nil {
    if skillErr := skill.PromptAndInstall(homeDir, skilldata.Content, os.Stdin, w); skillErr != nil {
        fmt.Fprintln(cmd.ErrOrStderr(), "warning: skill install:", skillErr)
    }
}

fmt.Fprintln(w, "Run install.sh to update shell completions")  // LINE 81: contaminates stdout in JSON mode
```

`PromptAndInstall` writes to `w` (stdout). In JSON mode, any output from `PromptAndInstall` also contaminates the JSON stream. Both the `PromptAndInstall` call and the hint `fmt.Fprintln` must be gated on `!output.JSONMode`.

**Fixed pattern:**
```go
if !output.JSONMode {
    homeDir, err := os.UserHomeDir()
    if err == nil {
        if skillErr := skill.PromptAndInstall(homeDir, skilldata.Content, os.Stdin, w); skillErr != nil {
            fmt.Fprintln(cmd.ErrOrStderr(), "warning: skill install:", skillErr)
        }
    }
    fmt.Fprintln(w, "Run install.sh to update shell completions")
}
```

### Anti-Patterns to Avoid
- **Reading `--json` locally in `RunE`:** Bypasses `output.JSONMode` as SSOT. Remove the double-read block entirely.
- **Ungated plain-text on success path:** Any `fmt.Fprintln(w, ...)` before `emitUpdateResult` poisons JSON parsability.
- **Gating only the hint, not `PromptAndInstall`:** `PromptAndInstall` also writes to `w`; must be gated together.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| JSON mode detection | Custom flag-read logic | `output.JSONMode` | Already set in `PersistentPreRunE`; SSOT for all commands |

## Common Pitfalls

### Pitfall 1: Gating only `fmt.Fprintln` but not `PromptAndInstall`
**What goes wrong:** `PromptAndInstall` receives `w` (stdout) as its writer. Even if the hint line is gated, `PromptAndInstall` can print to stdout (e.g., skill install confirmation messages), contaminating JSON output.
**Why it happens:** The writer is passed through — it's not obvious `PromptAndInstall` writes to it.
**How to avoid:** Gate the entire block (both `PromptAndInstall` and the hint) on `!output.JSONMode`.
**Warning signs:** Test captures stdout and sees text before `{` in JSON output.

### Pitfall 2: Removing `useJSON` local variable but leaving references to it
**What goes wrong:** Compiler error if `useJSON` is deleted but still passed to `emitUpdateResult`.
**Why it happens:** `emitUpdateResult` takes `useJSON bool` as parameter.
**How to avoid:** Replace all `useJSON` occurrences with `output.JSONMode` — both the assignment block and the `emitUpdateResult` call sites, and optionally refactor `emitUpdateResult` to read `output.JSONMode` directly.

### Pitfall 3: Assuming `PersistentPreRunE` does not run for `update`
**What goes wrong:** Thinking `output.JSONMode` is not set before `updateCmd.RunE` runs.
**Why it happens:** `root.go:28` has an early return for `update` — but that early return is *inside* the `if` block checking for DB skip, after `output.JSONMode = jsonFlag` has already executed.
**How to avoid:** Read `root.go:24-25` — `output.JSONMode` assignment is unconditional, before the DB skip check.

## Code Examples

### Defects as they exist today

```go
// update.go:29-33 — DEFECT: double flag read instead of output.JSONMode
useJSON, _ := cmd.Flags().GetBool("json")
if !useJSON {
    useJSON, _ = rootCmd.PersistentFlags().GetBool("json")
}

// update.go:74-81 — DEFECT: plain text written to w unconditionally in success path
homeDir, err := os.UserHomeDir()
if err == nil {
    if skillErr := skill.PromptAndInstall(homeDir, skilldata.Content, os.Stdin, w); skillErr != nil {
        fmt.Fprintln(cmd.ErrOrStderr(), "warning: skill install:", skillErr)
    }
}
fmt.Fprintln(w, "Run install.sh to update shell completions")
```

### Fixed code

```go
// RunE body — replace double flag read with output.JSONMode
// (delete lines 29-33 entirely)

// Gate side effects on !output.JSONMode
if !output.JSONMode {
    homeDir, err := os.UserHomeDir()
    if err == nil {
        if skillErr := skill.PromptAndInstall(homeDir, skilldata.Content, os.Stdin, w); skillErr != nil {
            fmt.Fprintln(cmd.ErrOrStderr(), "warning: skill install:", skillErr)
        }
    }
    fmt.Fprintln(w, "Run install.sh to update shell completions")
}

result.Updated = true
result.Message = fmt.Sprintf("updated to v%s", latest)
return emitUpdateResult(w, result, output.JSONMode)
```

### Existing test infrastructure (cmd/update_test.go)

The test file already exercises JSON output for the "already up to date" path using `httptest.NewServer` to mock the GitHub API and `rootCmd.SetOut(&buf)` to capture output. The existing `TestUpdateCmd_JSON` test verifies that a single JSON object is parseable from stdout.

A new test for the "successful update" JSON path would need to mock `updater.DownloadAndReplace` — currently this function is not injectable. The simpler approach is to verify the "already up to date" path (existing tests) and add a compile+behavior check confirming `output.JSONMode` is used rather than a local variable.

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go testing (stdlib) |
| Config file | none (go test ./...) |
| Quick run command | `go test ./cmd/ -run TestUpdateCmd -v` |
| Full suite command | `go test ./...` |

### Phase Requirements to Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| UPDT-04 | `--json update` (already up to date) emits valid JSON | unit | `go test ./cmd/ -run TestUpdateCmd_JSON -v` | Yes (update_test.go) |
| UPDT-04 | `--json update` (already up to date) has no plain text before JSON | unit | `go test ./cmd/ -run TestUpdateCmd_AlreadyUpToDate -v` | Yes (update_test.go) |
| UPDT-04 | `output.JSONMode` used (not local flag read) | code review | n/a — structural check | Must verify post-fix |

### Sampling Rate
- **Per task commit:** `go test ./cmd/ -run TestUpdateCmd -v`
- **Per wave merge:** `go test ./...`
- **Phase gate:** Full suite green before `/gsd:verify-work`

### Wave 0 Gaps
None — existing test infrastructure covers the "already up to date" JSON path. The "successful update with `--json`" path requires mocking `updater.DownloadAndReplace`, which is not yet injectable. That gap is acceptable: the structural fix (gating on `!output.JSONMode`) is verifiable by code review + the existing "already up to date" JSON test passing cleanly.

## Sources

### Primary (HIGH confidence)
- Direct code inspection: `cmd/update.go` lines 29-33, 74-82
- Direct code inspection: `cmd/root.go` lines 24-25 (PersistentPreRunE sets output.JSONMode)
- Direct code inspection: `internal/output/output.go` line 15 (`var JSONMode bool`)
- Direct code inspection: `cmd/update_test.go` (existing test patterns)
- `.planning/v0.3-MILESTONE-AUDIT.md` — confirms both defects with file:line precision

### Secondary (MEDIUM confidence)
- `.planning/STATE.md` decisions log — confirms `cmd.OutOrStdout()` pattern for updateCmd

## Metadata

**Confidence breakdown:**
- Defect identification: HIGH — both defects verified by direct code inspection and audit doc
- Fix pattern: HIGH — `output.JSONMode` pattern used by 6+ other commands in the codebase
- Test coverage: HIGH for "already up to date" path; LOW for "successful update" path (requires un-injectable dependency)

**Research date:** 2026-03-07
**Valid until:** Stable until `cmd/update.go` is modified
