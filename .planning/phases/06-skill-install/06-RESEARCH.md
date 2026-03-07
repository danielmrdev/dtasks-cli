# Phase 6: Skill Install - Research

**Researched:** 2026-03-07
**Domain:** Shell scripting (POSIX sh), Go CLI (Cobra), skill auto-install integration
**Confidence:** HIGH

---

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|-----------------|
| SKIL-01 | CLI detects whether Claude is installed (`~/.claude/` or `claude` command) | `DetectClaude()` already implements this — no new code needed in internal/skill |
| SKIL-02 | User is prompted for consent before copying the skill | `PromptAndInstall()` already implements TTY-gated prompt — exposed via new `install-skill` command |
| SKIL-03 | Skill is copied to `~/.claude/skills/dtasks-cli/` | `InstallSkill()` already implements this — no new code needed |
| SKIL-04 | If skill already exists it is overwritten silently; if platform not found, install is skipped gracefully | Already implemented in `InstallSkill()` (overwrite) and `PromptAndInstall()` (skip when no Claude) |
</phase_requirements>

---

## Summary

All four SKIL requirements are already implemented in `internal/skill` (Phase 3). The gap is purely at the **integration layer**: `install.sh` downloads and installs the binary but never invokes the skill install flow. `cmd/update.go` calls `skill.PromptAndInstall` post-update, but the first-install path via `install.sh` has no equivalent hook.

The fix requires two changes: (1) add a `dtasks install-skill` Cobra subcommand that calls `skill.PromptAndInstall` using `os.Stdin` and `os.Stdout`, and (2) add a function call at the end of `install.sh` (after the binary is installed, mirroring the existing `install_completions` pattern).

**Primary recommendation:** Add `cmd/install_skill.go` with a single `install-skill` subcommand, wire it into `root.go`, then call `"${dest}" install-skill` at the end of `install.sh` inside a function that mirrors `install_completions`.

---

## Standard Stack

### Core (already present — no new dependencies)

| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| `internal/skill` | (project) | DetectClaude, InstallSkill, PromptAndInstall | Already fully implemented and tested |
| `skills/dtasks-cli/skilldata` | (project) | `//go:embed SKILL.md` wrapper | Already used by updateCmd |
| `github.com/spf13/cobra` | v1.8.0 | New subcommand registration | Already the project CLI framework |
| `golang.org/x/term` | (project dep) | TTY detection inside PromptAndInstall | Already a direct dependency |

### No new dependencies needed

The entire skill install stack is in place. Phase 6 is a wiring phase, not an implementation phase.

---

## Architecture Patterns

### Recommended Project Structure (additions only)

```
cmd/
├── install_skill.go     # NEW: install-skill subcommand
└── root.go              # MODIFIED: AddCommand(installSkillCmd)
install.sh               # MODIFIED: install_skill() function + call
```

### Pattern 1: New Cobra subcommand (mirrors updateCmd skill block)

**What:** Thin Cobra command that delegates to `skill.PromptAndInstall`.
**When to use:** Any CLI-exposed operation that needs interactive consent.

```go
// cmd/install_skill.go
// Source: existing cmd/update.go skill block (lines 74-79)
package cmd

import (
    "fmt"
    "os"

    "github.com/danielmrdev/dtasks-cli/internal/skill"
    skilldata "github.com/danielmrdev/dtasks-cli/skills/dtasks-cli"
    "github.com/spf13/cobra"
)

var installSkillCmd = &cobra.Command{
    Use:   "install-skill",
    Short: "Install the dtasks skill for Claude Code",
    RunE: func(cmd *cobra.Command, args []string) error {
        homeDir, err := os.UserHomeDir()
        if err != nil {
            return fmt.Errorf("resolve home dir: %w", err)
        }
        return skill.PromptAndInstall(homeDir, skilldata.Content, os.Stdin, cmd.OutOrStdout())
    },
}
```

**root.go addition:**
```go
rootCmd.AddCommand(installSkillCmd)
```

**PersistentPreRunE skip:** The `install-skill` command does not need DB access. It must be added to the skip list in `root.go`, the same way `update` is already skipped:

```go
// root.go line 28 — extend the existing skip condition
if isCompletionScript(cmd) || cmd.Name() == "help" || cmd.Name() == "update" || cmd.Name() == "install-skill" {
    return nil
}
```

### Pattern 2: install.sh integration (mirrors install_completions)

**What:** Shell function `install_skill` called at end of `install.sh`.
**Key constraint:** Use `"${dest}"` (the just-installed binary path), not bare `dtasks`, identical to how `install_completions` uses `"${install_dir}/${BINARY}"`.

```sh
# ── Skill auto-install ────────────────────────────────────────────────────────
install_skill() {
    # Skip in non-interactive (pipe/CI) environments
    [ -t 0 ] || return 0

    "${dest}" install-skill
}

install_skill
```

**Why `[ -t 0 ] || return 0`:** Non-TTY environments (CI, piped installs like `curl | sh`) must skip silently. This mirrors the existing `install_completions` guard. The `PromptAndInstall` function also has its own TTY check, but the shell-level guard prevents even launching the binary for a no-op.

**Why the shell guard is sufficient:** When stdin is not a TTY, `PromptAndInstall` takes the non-TTY path and calls `InstallSkill` directly — but `install.sh` should not call the binary at all in non-TTY mode (matches the completions pattern and the success criteria: "skill install is skipped gracefully on non-TTY install").

### Anti-Patterns to Avoid

- **Reimplementing DetectClaude in shell:** The Go package already handles `~/.claude/`, `~/.config/claude/`, and `exec.LookPath("claude")`. Duplicating this in sh is unnecessary and will diverge.
- **Calling `dtasks update` to trigger skill install:** Semantically wrong; `update` fetches from GitHub, downloads a binary, and replaces the executable. Wrong tool for fresh install.
- **Installing skill unconditionally in install.sh:** Must respect non-TTY and must go through `PromptAndInstall` to check for Claude and get user consent.
- **Using bare `dtasks` instead of `"${dest}"`:** The binary may not be in PATH yet (install.sh prints a PATH hint for this case). Always use the explicit path.

---

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Claude detection | Shell `test -d ~/.claude` logic | `skill.DetectClaude()` via `dtasks install-skill` | Already handles 3 detection paths; Go version is tested |
| TTY detection | `tty`, `stty` or additional shell logic | `golang.org/x/term.IsTerminal` inside `PromptAndInstall` | Already implemented, tested; shell guard is just for early exit |
| Skill file write | `cp` or `cat >` in shell | `skill.InstallSkill()` via `dtasks install-skill` | Handles `MkdirAll`, overwrite, and error reporting |

**Key insight:** `internal/skill` is a complete, tested implementation. Phase 6 is pure wiring — the right answer is to expose it via a CLI command and invoke that command from the installer.

---

## Common Pitfalls

### Pitfall 1: Non-TTY install installs skill silently instead of skipping

**What goes wrong:** If `install.sh` calls `"${dest}" install-skill` without a TTY guard, on `curl | sh` installs `PromptAndInstall` receives a non-TTY stdin and calls `InstallSkill` directly (the non-TTY path). This installs the skill without consent.

**Why it happens:** The non-TTY path in `PromptAndInstall` was designed for the `dtasks update` use case — programmatic invocation where the caller (a human running `dtasks update`) already consented by running the command. The `install.sh` use case is different: the user piped a shell script and may not have expected skill install.

**How to avoid:** Guard in `install.sh` with `[ -t 0 ] || return 0` before calling `"${dest}" install-skill`. This ensures skill install only runs in interactive TTY sessions.

**Warning signs:** Test with `echo "" | sh install.sh` — should complete without installing skill.

### Pitfall 2: Binary not yet in PATH when skill command runs

**What goes wrong:** Using `dtasks install-skill` (bare name) fails when `install_dir` is not in PATH yet (common on first install — install.sh prints a PATH hint for this reason).

**How to avoid:** Use `"${dest}" install-skill` where `dest="${install_dir}/${BINARY}"` — the absolute path is already computed at the top of the script.

### Pitfall 3: install.sh syntax check failure

**What goes wrong:** The file uses `#!/usr/bin/env sh` with `set -e`. Any bash-only syntax (e.g., `[[`, `function`, process substitution) fails `bash -n` checks and also fails on strict POSIX sh interpreters.

**How to avoid:** Keep new shell code POSIX-compliant. The existing `install_completions` function is the style reference. Use `[ -t 0 ]`, `case`, and single-bracket `[`.

### Pitfall 4: PersistentPreRunE tries to open DB for install-skill

**What goes wrong:** On a fresh install with no config file, `PersistentPreRunE` calls `config.Load()` which launches the interactive wizard. Running `dtasks install-skill` before any DB config exists would trigger the config wizard unexpectedly.

**How to avoid:** Add `"install-skill"` to the skip list in `root.go`'s `PersistentPreRunE`, the same way `"update"` is already skipped (line 28 of root.go).

---

## Code Examples

### Verified: existing PromptAndInstall signature

```go
// Source: internal/skill/skill.go
// PromptAndInstall checks whether in is a real TTY.
// If it is not a TTY, it calls InstallSkill directly without prompting.
// If it is a TTY and DetectClaude returns false, it returns nil (graceful skip).
// If it is a TTY and Claude is detected, it prompts and installs on y/Y.
func PromptAndInstall(homeDir string, content []byte, in io.Reader, out io.Writer) error
```

### Verified: existing updateCmd skill block (reference for installSkillCmd)

```go
// Source: cmd/update.go lines 74-79
homeDir, err := os.UserHomeDir()
if err == nil {
    if skillErr := skill.PromptAndInstall(homeDir, skilldata.Content, os.Stdin, w); skillErr != nil {
        fmt.Fprintln(cmd.ErrOrStderr(), "warning: skill install:", skillErr)
    }
}
```

### Verified: existing PersistentPreRunE skip list (root.go line 28)

```go
// Source: cmd/root.go
if isCompletionScript(cmd) || cmd.Name() == "help" || cmd.Name() == "update" {
    return nil
}
// Extend to:
if isCompletionScript(cmd) || cmd.Name() == "help" || cmd.Name() == "update" || cmd.Name() == "install-skill" {
    return nil
}
```

### Verified: install_completions TTY guard pattern (reference for install_skill)

```sh
# Source: install.sh lines 107-108
install_completions() {
    # Skip in non-interactive (pipe/CI) environments
    [ -t 0 ] || return 0
    ...
}
```

---

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Skill install only via `dtasks update` | Skill install also on first install via `install.sh` | Phase 6 (now) | Closes `skill-first-install-path` gap from v0.3 audit |

**What Phase 3 implemented:**
- `internal/skill`: `DetectClaude`, `InstallSkill`, `PromptAndInstall` — fully tested
- `skills/dtasks-cli/skilldata.go`: `//go:embed SKILL.md` wrapper
- `cmd/update.go`: calls `skill.PromptAndInstall` post-update

**What Phase 6 adds:**
- `cmd/install_skill.go`: exposes skill install as `dtasks install-skill`
- `install.sh`: calls `"${dest}" install-skill` after binary install (TTY-gated)

---

## Open Questions

1. **Should `install-skill` print a confirmation message on success?**
   - What we know: `InstallSkill` returns nil on success with no output. `updateCmd` does not print success for skill install either.
   - What's unclear: Whether the user expects feedback ("Skill installed to ~/.claude/skills/dtasks-cli/SKILL.md").
   - Recommendation: Print a success line from `installSkillCmd.RunE` after `PromptAndInstall` returns nil. Keep it brief.

2. **Should `install.sh` failure of `install-skill` abort the install?**
   - What we know: `install.sh` uses `set -e`. If `"${dest}" install-skill` exits non-zero, the whole script aborts.
   - Recommendation: Wrap the call or use `"${dest}" install-skill || true` so skill install failure is non-fatal (matching the `updateCmd` pattern that uses a warning, not a fatal error).

---

## Validation Architecture

### Test Framework

| Property | Value |
|----------|-------|
| Framework | Go testing (stdlib) |
| Config file | none — standard `go test` |
| Quick run command | `go test ./internal/skill/... ./cmd/... -run TestInstallSkill -count=1` |
| Full suite command | `go test ./...` |

### Phase Requirements -> Test Map

| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| SKIL-01 | DetectClaude finds `~/.claude/` | unit | `go test ./internal/skill/... -run TestDetectClaude -v` | Yes (skill_test.go) |
| SKIL-02 | Consent prompt shown in TTY; skipped in non-TTY | unit | `go test ./internal/skill/... -run TestInstallSkill_NonTTY -v` | Yes (skill_test.go) |
| SKIL-03 | Skill written to `~/.claude/skills/dtasks-cli/SKILL.md` | unit | `go test ./internal/skill/... -run TestInstallSkill_Path -v` | Yes (skill_test.go) |
| SKIL-04 | Overwrite existing skill silently | unit | `go test ./internal/skill/... -run TestInstallSkill_Overwrite -v` | Yes (skill_test.go) |
| SKIL-04 | Skip gracefully when Claude not found (TTY, no Claude) | unit | `go test ./internal/skill/... -run TestDetectClaude_NotFound -v` | Yes (skill_test.go) |
| install-skill cmd | install-skill command wired to root | integration | `go build ./... && ./dist/dtasks install-skill` (manual verify) | No — Wave 0 |
| install.sh syntax | bash -n syntax check | smoke | `bash -n install.sh` | Yes (install.sh) |

### Sampling Rate

- **Per task commit:** `go test ./internal/skill/... -count=1`
- **Per wave merge:** `go test ./...`
- **Phase gate:** Full suite green + `bash -n install.sh` passes before `/gsd:verify-work`

### Wave 0 Gaps

- [ ] `cmd/install_skill_test.go` — covers the `install-skill` command wiring (builds, exits 0 in non-TTY environment)
- [ ] No framework install needed — Go testing is already in place

---

## Sources

### Primary (HIGH confidence)

- `internal/skill/skill.go` — complete implementation of DetectClaude, InstallSkill, PromptAndInstall
- `internal/skill/skill_test.go` — all 5 tests passing (verified by `go test ./internal/skill/...`)
- `cmd/update.go` — reference pattern for calling skill.PromptAndInstall from a Cobra command
- `cmd/root.go` — PersistentPreRunE skip list pattern
- `install.sh` — install_completions pattern (TTY guard, binary path, function + call structure)

### Secondary (MEDIUM confidence)

- ROADMAP.md Phase 6 success criteria — clarifies "dtasks install-skill (or equivalent)"
- REQUIREMENTS.md SKIL-01..04 traceability — confirms Phase 3 + Phase 6 joint responsibility

### Tertiary (LOW confidence)

- None

---

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — all dependencies already in project, verified by reading source
- Architecture: HIGH — mirroring existing patterns (updateCmd, install_completions) directly; no new patterns
- Pitfalls: HIGH — derived from reading actual code paths and TTY behavior already in place

**Research date:** 2026-03-07
**Valid until:** 2026-04-07 (stable — no external dependencies; all internal code)
