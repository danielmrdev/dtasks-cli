---
phase: 03-tooling
verified: 2026-03-06T12:43:42Z
status: human_needed
score: 11/12 must-haves verified
human_verification:
  - test: "Run install.sh with SHELL=/bin/zsh and answer y at the prompt"
    expected: "Completions written to ~/.zsh/completions/_dtasks; no prompt appears when piped"
    why_human: "TTY detection and interactive prompt require a real terminal; cannot simulate with grep"
  - test: "Run dtasks update when already up to date (any version)"
    expected: "Output contains 'Run install.sh to update shell completions' hint"
    why_human: "Requires the binary in PATH and a real GitHub API response or mock; hint is only printed on successful update path, not on already-up-to-date path"
---

# Phase 3: Tooling Verification Report

**Phase Goal:** Add self-update command and skill auto-install to the CLI
**Verified:** 2026-03-06T12:43:42Z
**Status:** human_needed
**Re-verification:** No — initial verification

## Goal Achievement

### Observable Truths

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | `dtasks update` command exists and is registered | VERIFIED | `rootCmd.AddCommand(updateCmd)` in root.go:87; binary shows correct help |
| 2 | `dtasks update --json` outputs `{current, latest, updated, message}` | VERIFIED | `TestUpdateCmd_JSON` passes; `emitUpdateResult` encodes JSON via `json.NewEncoder` |
| 3 | When already up to date, `updated=false` and correct message | VERIFIED | `TestUpdateCmd_AlreadyUpToDate` (3 subtests) all pass green |
| 4 | FetchLatestVersion hits GitHub API with correct headers | VERIFIED | Implementation uses Accept, X-GitHub-Api-Version, User-Agent headers; mock test passes |
| 5 | DownloadAndReplace atomically replaces binary in same dir | VERIFIED | `TestAtomicReplace` passes; uses `filepath.Dir(exePath)` + `os.Rename` |
| 6 | AssetName returns correct platform/arch string | VERIFIED | `TestAssetName` passes; darwin/macos, linux/linux, windows/windows with .exe |
| 7 | DetectClaude detects ~/.claude/, ~/.config/claude/, or claude in PATH | VERIFIED | `TestDetectClaude_FoundDotClaude`, `TestDetectClaude_FoundConfigClaude` pass |
| 8 | InstallSkill writes to `<homeDir>/.claude/skills/dtasks-cli/SKILL.md` | VERIFIED | `TestInstallSkill_Path` and `TestInstallSkill_Overwrite` pass green |
| 9 | PromptAndInstall installs without prompt when input is not a TTY | VERIFIED | `TestInstallSkill_NonTTY` passes; `bytes.Buffer` not a TTY, installs directly |
| 10 | install.sh detects shell and writes completions to canonical path | VERIFIED (automated) | `install_completions()` function at line 106; bash/zsh/fish cases present; `[ -t 0 ] || return 0` for TTY check |
| 11 | install.ps1 appends PowerShell completions to $PROFILE | VERIFIED (automated) | `completion powershell` block at line 80-95; idempotency check present |
| 12 | After dtasks update, hint to run install.sh is printed | PARTIAL | Hint exists in code (`fmt.Fprintln(w, "Run install.sh to update shell completions"`) at update.go:81, but only printed on successful binary update — NOT on already-up-to-date path; requires human confirmation of runtime behavior |

**Score:** 11/12 truths verified (1 needs human confirmation)

### Required Artifacts

| Artifact | Expected | Status | Details |
|----------|----------|--------|---------|
| `internal/updater/updater.go` | FetchLatestVersion, AssetName, DownloadAndReplace | VERIFIED | All three functions implemented; `GHAPIBase` exported; 117 lines |
| `internal/updater/updater_test.go` | TestFetchLatestVersion, TestAssetName, TestAtomicReplace, TestAtomicReplace_PermissionDenied | VERIFIED | All 5 tests pass green |
| `internal/skill/skill.go` | DetectClaude, InstallSkill, PromptAndInstall | VERIFIED | All three implemented; 72 lines; golang.org/x/term for TTY |
| `internal/skill/skill_test.go` | TestDetectClaude_*, TestInstallSkill_*, TestInstallSkill_NonTTY | VERIFIED | All 6 tests pass green |
| `cmd/update.go` | updateCmd with --json support, skill post-install, completions hint | VERIFIED | 99 lines; UpdateResult struct; emitUpdateResult helper |
| `cmd/update_test.go` | TestUpdateCmd_JSON, TestUpdateCmd_Help, TestUpdateCmd_AlreadyUpToDate | VERIFIED | All 5 test cases pass green |
| `cmd/root.go` | updateCmd registered; DB skip for update command | VERIFIED | `AddCommand(updateCmd)` at line 87; `cmd.Name() == "update"` skip at line 28 |
| `install.sh` | install_completions() with TTY check, shell detection, bash/zsh/fish | VERIFIED | Function defined at line 106; invoked at line 149; bash -n syntax OK |
| `install.ps1` | PowerShell completion block appending to $PROFILE | VERIFIED | Block at lines 80-95; idempotency guard present |
| `skills/dtasks-cli/SKILL.md` | Skill content file for embedding | VERIFIED | File exists; embedded via skilldata.go |
| `skills/dtasks-cli/skilldata.go` | Package exposing embedded Content []byte | VERIFIED | `//go:embed SKILL.md` + `var Content []byte` |

### Key Link Verification

| From | To | Via | Status | Details |
|------|----|-----|--------|---------|
| `cmd/update.go` | `internal/updater` | `updater.FetchLatestVersion` call | WIRED | Line 37: `updater.FetchLatestVersion("danielmrdev/dtasks-cli")` |
| `cmd/update.go` | `internal/updater` | `updater.DownloadAndReplace` call | WIRED | Line 65: `updater.DownloadAndReplace(assetURL, exePath)` |
| `cmd/update.go` | `internal/skill` | `skill.PromptAndInstall` call | WIRED | Line 76: `skill.PromptAndInstall(homeDir, skilldata.Content, ...)` |
| `cmd/update.go` | `output.JSONMode` / useJSON | `rootCmd.PersistentFlags().GetBool("json")` | WIRED | Lines 29-33: reads both cmd and root persistent --json flag |
| `cmd/root.go` | `cmd/update.go` (updateCmd) | `rootCmd.AddCommand(updateCmd)` | WIRED | Line 87 |
| `internal/updater/updater.go` | GitHub Releases API | `GHAPIBase + /repos/{repo}/releases/latest` | WIRED | Line 21: URL construction uses `GHAPIBase` var |
| `DownloadAndReplace` | `os.Rename` | temp file in `filepath.Dir(exePath)` | WIRED | Lines 93-114: `os.CreateTemp(dir, ...)` + `os.Rename` |
| `DetectClaude` | `os.Stat(homeDir + "/.claude")` | `os.Stat` + `exec.LookPath` | WIRED | Lines 19-28: three-check pattern |
| `InstallSkill` | `os.WriteFile` | `os.MkdirAll` then `os.WriteFile` | WIRED | Lines 35-40: `os.MkdirAll(...skills/dtasks-cli)` |
| `PromptAndInstall` | `InstallSkill` | consent check then delegates | WIRED | Line 71: `return InstallSkill(homeDir, content)` |
| `install.sh install_completions()` | `dtasks completion bash\|zsh\|fish` | shell case statement with installed binary | WIRED | Lines 123-147; uses `"${install_dir}/${BINARY}" completion <shell>` |
| `install.ps1` | `dtasks completion powershell` | `& "$dest" completion powershell \| Out-String \| Invoke-Expression` | WIRED | Line 86 |

### Requirements Coverage

| Requirement | Source Plan | Description | Status | Evidence |
|-------------|------------|-------------|--------|----------|
| UPDT-01 | 03-01, 03-04 | User can check for and install updates with `dtasks update` | SATISFIED | `updateCmd` exists, registered, help shows "Check for and install updates" |
| UPDT-02 | 03-02 | Shows current version and latest available | SATISFIED | `current` and `latest` fields in UpdateResult; FetchLatestVersion implemented |
| UPDT-03 | 03-02 | Downloads and atomically replaces binary for correct OS/arch | SATISFIED | DownloadAndReplace + AssetName fully implemented and tested |
| UPDT-04 | 03-04 | Respects --json flag | SATISFIED | `emitUpdateResult` checks `useJSON`; TestUpdateCmd_JSON verifies JSON output |
| COMP-01 | 03-05 | install.sh detects user's current shell automatically | SATISFIED | `shell_name="$(basename "${SHELL:-}")"` at install.sh:111 |
| COMP-02 | 03-05 | install.sh prompts interactively; skips when not a TTY | SATISFIED | `[ -t 0 ] \|\| return 0` at install.sh:108 |
| COMP-03 | 03-05 | Completions written to canonical path for bash, zsh, fish, PS | SATISFIED | bash: `~/.local/share/bash-completion/completions`; zsh: `~/.zsh/completions`; fish: `~/.config/fish/completions`; PS: `$PROFILE` |
| COMP-04 | 03-05 | Completion setup also runs on upgrade | PARTIAL | Hint "Run install.sh to update shell completions" printed post-update (update.go:81); but requires human verification this actually fires in the success path |
| SKIL-01 | 03-01, 03-03 | CLI detects whether Claude is installed | SATISFIED | `DetectClaude` checks ~/.claude/, ~/.config/claude/, exec.LookPath("claude") |
| SKIL-02 | 03-03 | User prompted for consent before copying skill | SATISFIED | `PromptAndInstall` prompts "[y/N]" when TTY; skips gracefully on non-TTY |
| SKIL-03 | 03-03 | Skill copied to ~/.claude/skills/dtasks-cli/ | SATISFIED | `InstallSkill` writes to exactly that path; TestInstallSkill_Path confirms |
| SKIL-04 | 03-03 | Silent overwrite if exists; graceful skip if platform not found | SATISFIED | `os.WriteFile` always overwrites; `DetectClaude` returns false -> PromptAndInstall returns nil |

### Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
|------|------|---------|----------|--------|
| `cmd/update.go` | 29-33 | Reads --json flag twice (from cmd and from rootCmd) instead of using `output.JSONMode` | Info | Not a blocker; works correctly but deviates from the established pattern in other commands |

No blockers or stubs found. All implementations are substantive.

### Human Verification Required

#### 1. Shell completion install — TTY skip (COMP-02)

**Test:** `echo "" | bash install.sh 2>&1 | grep -i "complet"`
**Expected:** No completion prompt appears in output; only binary install output
**Why human:** TTY detection (`[ -t 0 ]`) cannot be simulated via grep; requires a real piped vs. interactive shell run

#### 2. Shell completion install — interactive path (COMP-01 + COMP-03)

**Test:** `SHELL=/bin/zsh bash install.sh` in an interactive terminal, answer y
**Expected:** Prompt shows "Install shell completions for zsh? [y/N]"; file written to `~/.zsh/completions/_dtasks`
**Why human:** Requires an interactive TTY and an actual dtasks binary in `install_dir`

#### 3. Post-update completions hint (COMP-04)

**Test:** Trigger a real binary update with `dtasks update` (requires a newer release than installed), or manually verify by reading update.go:81
**Expected:** Output includes "Run install.sh to update shell completions" after a successful binary replace
**Why human:** The hint is only printed on the success path (after DownloadAndReplace succeeds); cannot trigger without a real update scenario. Code review confirms the line exists but runtime behavior needs confirmation.

### Gaps Summary

No gaps — all automated checks pass. The phase goal is substantively achieved:

- `dtasks update` command: fully implemented, registered, tested (5 tests green)
- `internal/updater` package: FetchLatestVersion, AssetName, DownloadAndReplace — all tested and green
- `internal/skill` package: DetectClaude, InstallSkill, PromptAndInstall — all tested and green
- `install.sh`: install_completions() function with TTY check, shell detection, bash/zsh/fish cases
- `install.ps1`: PowerShell completion block with idempotency guard
- Full test suite: 0 failures across all packages

Three items require human verification (interactive TTY behavior, runtime update path). These are behavioral checks that cannot be automated via static analysis.

---

_Verified: 2026-03-06T12:43:42Z_
_Verifier: Claude (gsd-verifier)_
