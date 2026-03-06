# Phase 3: Tooling - Research

**Researched:** 2026-03-06
**Domain:** Go self-update, Cobra shell completions, Claude skill distribution
**Confidence:** HIGH

## Summary

Phase 3 adds three independent capability clusters to dtasks: (1) a `dtasks update` command that checks the GitHub Releases API and atomically replaces the running binary, (2) shell completion install logic integrated into `install.sh`, and (3) first-run detection of Claude Code and consent-based skill copying.

The project already has significant infrastructure in place: `install.sh` and `install.ps1` handle binary download and installation, `cmd/completion.go` already wires up Cobra's dynamic completion helpers, and `skills/dtasks-cli/SKILL.md` is the skill file that needs to be distributed. The self-update command is the most novel piece — it duplicates part of what `install.sh` does but inside the Go binary itself. The simplest correct approach is to implement it without an external library by calling the GitHub API directly with `net/http`, using `os.Rename` for atomic replacement on Unix (temp file in same directory + rename) and the Windows-compatible two-step pattern used by `minio/selfupdate`.

The binary naming convention used by the release workflow (`dtasks-{platform}-{arch}` / `dtasks-{platform}-{arch}.exe`) is already established and must be matched exactly in the update command.

**Primary recommendation:** Implement `updateCmd` in pure Go (no library) using `net/http` + `os.Executable` + temp-file-then-rename. Add completion install to `install.sh` (already detects shell). Copy skill from `skills/dtasks-cli/` embedded in binary or from the installed location; detect Claude via `~/.claude/` directory existence.

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|-----------------|
| UPDT-01 | User can check for and install updates with `dtasks update` | GitHub Releases API (anonymous, public repo), `os.Executable()`, atomic rename pattern |
| UPDT-02 | `dtasks update` shows current version and latest available | `rootCmd.Version` already set via `-ldflags`; compare with `tag_name` from API |
| UPDT-03 | `dtasks update` downloads and atomically replaces the running binary for the correct OS/arch | `runtime.GOOS`/`runtime.GOARCH`, temp file + `os.Rename`, asset naming `dtasks-{platform}-{arch}` |
| UPDT-04 | `dtasks update` respects `--json` flag | Global `output.JSONMode` already in place; emit structured JSON result |
| COMP-01 | `install.sh` detects the user's current shell automatically | `$SHELL` env var + `ps` fallback; already done in `install.sh` for OS detection |
| COMP-02 | `install.sh` prompts interactively to install shell completions (skips when stdin is not a TTY) | `[ -t 0 ]` TTY check in sh; Cobra provides `GenBashCompletion`, `GenZshCompletion`, etc. |
| COMP-03 | Completions written to canonical location for bash, zsh, fish, PowerShell | Shell-specific paths documented below |
| COMP-04 | Completion setup also runs on upgrade (update path) | `updateCmd` calls the completion install logic post-update via exec or shell |
| SKIL-01 | CLI detects whether Claude is installed (`~/.claude/` or `claude` command) | `os.Stat(home+"/.claude")` + `exec.LookPath("claude")` |
| SKIL-02 | User is prompted for consent before copying the skill | `bufio.NewReader(os.Stdin)` prompt, same pattern as `config.runWizard` |
| SKIL-03 | Skill is copied to the correct platform path (`~/.claude/skills/dtasks-cli/`) | `os.MkdirAll` + `os.WriteFile`; source is `skills/dtasks-cli/SKILL.md` in repo |
| SKIL-04 | If skill already exists it is overwritten silently; if platform not found, install is skipped gracefully | `os.WriteFile` always overwrites; wrap in graceful error handling |
</phase_requirements>

## Standard Stack

### Core (no new dependencies needed)

| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| `net/http` | stdlib | GitHub API fetch, binary download | No CGO, already available |
| `runtime` | stdlib | `GOOS`/`GOARCH` for asset selection | Already used in `config` package |
| `os` | stdlib | `Executable()`, `Rename()`, temp files, `Stat()` | Already used throughout |
| `encoding/json` | stdlib | Parse GitHub API JSON response | Consistent with project; no deps |
| `github.com/spf13/cobra` | v1.8.0 | `GenBashCompletion`, `GenZshCompletion`, `GenFishCompletion`, `GenPowerShellCompletion` | Already in go.mod |

### No library needed for self-update

The project is CGO_ENABLED=0, cross-platform, and already implements its own checksum verification in `install.sh`. External update libraries (`creativeprojects/go-selfupdate`, `minio/selfupdate`) add dependencies and handle archive extraction — but the release assets are plain binaries, not tarballs. The update command can be implemented in ~80 lines of stdlib Go.

### Alternatives Considered

| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| stdlib net/http | `creativeprojects/go-selfupdate` | Library handles semver comparison and archive extraction, but assets are plain binaries — overhead not justified. Adds ~5 indirect deps. |
| stdlib net/http | `minio/selfupdate` | Clean atomic API, but adds dependency. The atomic pattern is trivially reimplemented with `os.Rename` on Unix. |
| manual embed | `//go:embed skills/...` | Could embed skill into binary. Simpler: copy from installed path at `install.sh` time, or always read from `skills/dtasks-cli/SKILL.md` relative to executable. |

**Installation:** No new packages required. All work uses existing `go.mod` dependencies.

## Architecture Patterns

### Recommended Project Structure

```
cmd/
├── update.go           # new: updateCmd (UPDT-01..04)
├── skill.go            # new: skill install logic (SKIL-01..04), called from update + first-run
install.sh              # modified: add completion install section (COMP-01..04)
skills/
└── dtasks-cli/
    └── SKILL.md        # existing: source for skill install
```

The skill detection/install logic lives in `cmd/skill.go` so it can be called from both `updateCmd` (post-update) and from `PersistentPreRunE` on first-run (or a separate `setupCmd`).

### Pattern 1: GitHub Releases API (no auth, public repo)

**What:** GET `https://api.github.com/repos/{owner}/{repo}/releases/latest` returns JSON with `tag_name` and `assets[]`.
**When to use:** Always. Anonymous rate limit is 60 req/hour per IP — more than enough for an update check.
**Asset URL pattern:** `https://github.com/{owner}/{repo}/releases/download/{tag}/{asset_name}`

```go
// Source: GitHub REST API docs (https://docs.github.com/en/rest/releases/releases)
type ghRelease struct {
    TagName string `json:"tag_name"`
}

func fetchLatestVersion(repo string) (string, error) {
    url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", repo)
    req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
    req.Header.Set("Accept", "application/vnd.github+json")
    req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
    req.Header.Set("User-Agent", "dtasks-cli")
    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()
    var rel ghRelease
    if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
        return "", err
    }
    return rel.TagName, nil
}
```

### Pattern 2: Atomic Binary Replacement

**What:** Download new binary to a temp file in the same directory as the current executable, then `os.Rename` (atomic on Unix, best-effort on Windows).
**Why:** `os.Rename` across the same filesystem is atomic on POSIX. Must use same directory, not `os.TempDir()` which may be on a different filesystem.
**Windows caveat:** Cannot rename/delete a running executable on Windows. Write to a `.new` file and print a message asking the user to replace manually, OR use a post-exit script. This is an edge case — the requirements say "atomic replace" but Windows fundamentally prevents this for running processes. Simplest approach: document the Windows limitation and use the same rename on Windows (works because Go processes don't lock the binary like the OS does on some configs).

```go
// Source: stdlib os package docs
func atomicReplace(exePath string, newBinary []byte) error {
    dir := filepath.Dir(exePath)
    tmp, err := os.CreateTemp(dir, ".dtasks-update-*")
    if err != nil {
        return fmt.Errorf("create temp: %w", err)
    }
    tmpName := tmp.Name()
    defer func() { os.Remove(tmpName) }() // cleanup on failure

    if _, err := tmp.Write(newBinary); err != nil {
        tmp.Close()
        return fmt.Errorf("write temp: %w", err)
    }
    tmp.Close()

    if err := os.Chmod(tmpName, 0755); err != nil {
        return fmt.Errorf("chmod: %w", err)
    }
    return os.Rename(tmpName, exePath) // atomic on Unix
}
```

### Pattern 3: Asset Name Resolution

**What:** Map `runtime.GOOS`/`runtime.GOARCH` to the existing release asset naming convention.
**Naming convention (from release.yml):** `dtasks-{platform}-{arch}` (Unix) / `dtasks-{platform}-{arch}.exe` (Windows)

```go
// Source: Makefile and release.yml in this repo
func assetName() (string, error) {
    var platform string
    switch runtime.GOOS {
    case "darwin":
        platform = "macos"
    case "linux":
        platform = "linux"
    case "windows":
        platform = "windows"
    default:
        return "", fmt.Errorf("unsupported OS: %s", runtime.GOOS)
    }

    var arch string
    switch runtime.GOARCH {
    case "amd64":
        arch = "amd64"
    case "arm64":
        arch = "arm64"
    default:
        return "", fmt.Errorf("unsupported arch: %s", runtime.GOARCH)
    }

    name := fmt.Sprintf("dtasks-%s-%s", platform, arch)
    if runtime.GOOS == "windows" {
        name += ".exe"
    }
    return name, nil
}
```

### Pattern 4: Cobra Shell Completion Generation

**What:** Cobra's built-in completion command generates scripts for bash, zsh, fish, and PowerShell. Already wired in cobra via the default `completion` subcommand — just call the right generator and write to the canonical path.
**Note:** The project already has `rootCmd` with cobra, which automatically adds a `completion` subcommand. The install.sh extension calls `dtasks completion bash > $dest` etc.

```sh
# Source: cobra.dev/docs/how-to-guides/shell-completion/
# Canonical install locations:

# bash (user-level, no sudo needed)
mkdir -p ~/.local/share/bash-completion/completions/
dtasks completion bash > ~/.local/share/bash-completion/completions/dtasks

# zsh (user-level)
mkdir -p ~/.zsh/completions/
dtasks completion zsh > ~/.zsh/completions/_dtasks
# user must have fpath=(~/.zsh/completions $fpath) in ~/.zshrc

# fish
mkdir -p ~/.config/fish/completions/
dtasks completion fish > ~/.config/fish/completions/dtasks.fish

# PowerShell (add to $PROFILE)
dtasks completion powershell >> $PROFILE
```

### Pattern 5: Claude Detection and Skill Install

**What:** Check if Claude is installed, prompt for consent (if TTY), copy `SKILL.md` to `~/.claude/skills/dtasks-cli/SKILL.md`.
**Detection:** Claude Code (the CLI) stores config in `~/.claude/`. As of v1.0.30 the canonical path changed to `~/.config/claude/` but `~/.claude/` is retained for backward compat. Check both: if either exists, Claude is installed. Also check `exec.LookPath("claude")` as a secondary signal.
**Skill destination:** `~/.claude/skills/dtasks-cli/SKILL.md` (official docs confirm personal skills live at `~/.claude/skills/<skill-name>/SKILL.md`).
**Source:** The skill file is at `skills/dtasks-cli/SKILL.md` in the repo. At install time, `install.sh` can copy it. For the `updateCmd` path, the binary needs access to the skill content — use `//go:embed skills/dtasks-cli/SKILL.md` to embed it, OR have `install.sh` also handle skill installation (simpler).

```go
// Source: os package stdlib, exec.LookPath stdlib
func detectClaude() bool {
    home, _ := os.UserHomeDir()
    // Check ~/.claude (legacy and current)
    if _, err := os.Stat(filepath.Join(home, ".claude")); err == nil {
        return true
    }
    // Check ~/.config/claude (v1.0.30+)
    if _, err := os.Stat(filepath.Join(home, ".config", "claude")); err == nil {
        return true
    }
    // Check if claude binary is in PATH
    if _, err := exec.LookPath("claude"); err == nil {
        return true
    }
    return false
}

func installSkill(skillContent []byte) error {
    home, _ := os.UserHomeDir()
    dest := filepath.Join(home, ".claude", "skills", "dtasks-cli")
    if err := os.MkdirAll(dest, 0755); err != nil {
        return err
    }
    return os.WriteFile(filepath.Join(dest, "SKILL.md"), skillContent, 0644)
}
```

### Pattern 6: JSON Output for Update Command (UPDT-04)

**What:** Follow existing pattern — check `output.JSONMode`, emit structured JSON.

```go
// Consistent with output.PrintSuccess / output.PrintError pattern
type UpdateResult struct {
    Current string `json:"current"`
    Latest  string `json:"latest"`
    Updated bool   `json:"updated"`
    Message string `json:"message,omitempty"`
}
```

### Anti-Patterns to Avoid

- **Cross-filesystem temp file:** Using `os.TempDir()` for the download temp file can cause `os.Rename` to fail with "invalid cross-device link" — always use `filepath.Dir(exePath)` as the temp dir.
- **Blocking on TTY detection for skill install:** Skill install should be skipped silently (not error) when stdin is not a TTY, matching the COMP-02 pattern.
- **Downloading to memory for large binaries:** Stream the download to disk, do not `io.ReadAll` the binary response body into RAM.
- **Hardcoding `~/.claude/skills/`:** Always resolve `os.UserHomeDir()` at runtime; never hardcode the home path.
- **Version string format mismatch:** Release tags are `v0.3.0`. The binary `version` variable injected by ldflags is typically the same. Strip leading `v` before semver comparison, or compare tag strings directly (tags already include `v`; `rootCmd.Version` likely strips it — verify at implementation time).

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Shell completion scripts | Custom completion output | `cobra`'s built-in `completion` subcommand generators | Already handles all edge cases per shell; already wired in the project |
| Semver comparison | String parsing + comparison logic | Simple string equality or `strings.TrimPrefix(tag, "v")` + `version != latest` | Only need "is there a newer version", not full semver ordering |
| Archive extraction | Custom tar/zip parser | N/A — release assets are plain binaries | No tarballs in this project's release convention |
| Checksum verification | Custom SHA256 | `crypto/sha256` stdlib | Already used in install.sh; simple to add in Go |

**Key insight:** The release assets are plain binaries (not archives), so the "download and replace" logic is truly just HTTP GET + write + rename. No extraction layer needed.

## Common Pitfalls

### Pitfall 1: Version String Format
**What goes wrong:** `rootCmd.Version` is set via `-ldflags "-X main.version=v0.3.0"`. The GitHub API returns `tag_name: "v0.3.0"`. If the binary strips the `v` prefix, comparison `version == latestTag` will always say "update available".
**Why it happens:** Cobra sometimes displays version without `v`; ldflags value is under developer control.
**How to avoid:** In `updateCmd`, compare `"v"+version == latestTag` OR normalize both sides. Check the actual injected value at test time.
**Warning signs:** Update command always reports a newer version even when running the latest.

### Pitfall 2: Replace Fails When Binary is in Read-Only Location
**What goes wrong:** If dtasks was installed to `/usr/local/bin` by root, the current user cannot write the temp file to that directory.
**Why it happens:** `os.CreateTemp(filepath.Dir(exePath), ...)` returns permission denied.
**How to avoid:** Catch the error and print a helpful message: "Run with sudo or reinstall with install.sh".
**Warning signs:** `permission denied` error on temp file creation.

### Pitfall 3: Completion Shell Detection
**What goes wrong:** `$SHELL` env var may point to `/bin/zsh` but the user's interactive shell is bash, or vice versa.
**Why it happens:** `$SHELL` reflects the login shell, not the current interactive shell.
**How to avoid:** Detect `$SHELL` as primary signal, but allow user to confirm or override in the interactive prompt. In non-TTY (CI/pipe) mode, skip entirely.
**Warning signs:** Completions installed for wrong shell.

### Pitfall 4: Claude Config Dir Migration
**What goes wrong:** Claude Code v1.0.30 moved config from `~/.claude/` to `~/.config/claude/`, but `~/.claude/` still exists as a symlink or legacy dir on some installs.
**Why it happens:** Undocumented breaking change in Claude Code.
**How to avoid:** Check both paths AND `exec.LookPath("claude")`. Install skill to `~/.claude/skills/dtasks-cli/` (confirmed by official Claude Code docs as the personal skills path regardless of config dir migration).
**Warning signs:** Skill install succeeds but Claude doesn't discover it.

### Pitfall 5: Self-Update on Windows
**What goes wrong:** `os.Rename` on Windows for a running executable may work or may not depending on Windows version and antivirus. The old file cannot always be deleted while the process runs.
**Why it happens:** Windows file locking semantics differ from POSIX.
**How to avoid:** Attempt rename; if it fails, write to `dtasks.new` and print instructions. Do not exit with error — treat as partial success.
**Warning signs:** `access denied` or sharing violation on Windows.

### Pitfall 6: Embedding vs. Shipping Skill File
**What goes wrong:** If the skill content is embedded via `//go:embed`, a stale build could ship an old skill version. If it's read from disk, the file must be present after install.
**Why it happens:** `//go:embed` bakes the file at compile time.
**How to avoid:** Recommend embedding (`//go:embed`) for the `dtasks update` path (binary is self-contained). For the initial install, `install.sh` copies from the repo. These two sources must be kept in sync (the embedded version in the binary is always current for that release).
**Warning signs:** Skill installed by `install.sh` differs from what `dtasks update` installs.

## Code Examples

### Update Command Skeleton

```go
// Source: stdlib patterns, consistent with existing cmd/ structure
var updateCmd = &cobra.Command{
    Use:   "update",
    Short: "Check for and install updates",
    RunE: func(cmd *cobra.Command, args []string) error {
        current := rootCmd.Version // set via Execute(version)
        latest, err := fetchLatestVersion("danielmrdev/dtasks-cli")
        if err != nil {
            return fmt.Errorf("check update: %w", err)
        }

        result := UpdateResult{
            Current: current,
            Latest:  strings.TrimPrefix(latest, "v"),
        }

        if "v"+current == latest || current == latest {
            result.Message = "already up to date"
            return printUpdateResult(result)
        }

        // Download and replace
        asset, err := assetName()
        if err != nil {
            return err
        }
        url := fmt.Sprintf(
            "https://github.com/danielmrdev/dtasks-cli/releases/download/%s/%s",
            latest, asset,
        )
        if err := downloadAndReplace(url); err != nil {
            return fmt.Errorf("update failed: %w", err)
        }
        result.Updated = true
        result.Message = fmt.Sprintf("updated to %s", latest)
        return printUpdateResult(result)
    },
}
```

### TTY Check in install.sh

```sh
# Source: POSIX sh standard
install_completions() {
    # Skip in non-interactive (pipe/CI) environments
    [ -t 0 ] || return 0

    # Detect shell
    shell_name="$(basename "${SHELL:-}")"
    printf "Install shell completions for %s? [y/N] " "$shell_name"
    read -r answer
    case "$answer" in
        [Yy]*) ;;
        *) return 0 ;;
    esac

    case "$shell_name" in
        bash)
            dir="${HOME}/.local/share/bash-completion/completions"
            mkdir -p "$dir"
            dtasks completion bash > "${dir}/dtasks"
            echo "Completions installed for bash"
            ;;
        zsh)
            dir="${HOME}/.zsh/completions"
            mkdir -p "$dir"
            dtasks completion zsh > "${dir}/_dtasks"
            echo "Completions installed for zsh"
            echo "Ensure your ~/.zshrc has: fpath=(~/.zsh/completions \$fpath)"
            ;;
        fish)
            dir="${HOME}/.config/fish/completions"
            mkdir -p "$dir"
            dtasks completion fish > "${dir}/dtasks.fish"
            echo "Completions installed for fish"
            ;;
        *)
            echo "Unsupported shell: $shell_name. Run 'dtasks completion --help' manually."
            ;;
    esac
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Manual skill copying | `~/.claude/skills/` auto-discovery | Claude Code skills system (2025) | Skills placed in `~/.claude/skills/<name>/SKILL.md` are loaded automatically |
| `~/.claude/` config dir | `~/.config/claude/` (v1.0.30+) | Late 2025 | Both paths work; `~/.claude/` still valid for skills dir |
| External update library | stdlib `net/http` + `os.Rename` | Ongoing best practice for small CLIs | Avoids dependency bloat; sufficient for plain-binary releases |
| System-level completion install | User-level completion dirs | Modern shell tooling | User-level paths (`~/.local/share/bash-completion/`, `~/.zsh/completions/`, `~/.config/fish/`) don't require sudo |

## Open Questions

1. **Skill install trigger point**
   - What we know: SKIL-01/02 say "on first run (or during install)" — ambiguous between the Go binary and install.sh.
   - What's unclear: Should `PersistentPreRunE` detect Claude and prompt? That runs on every command invocation, which would re-prompt after the first time unless a state flag is written.
   - Recommendation: Handle skill install in `install.sh` (consistent with COMP-04 pattern) and in `updateCmd` post-update. Add a `--install-skill` flag to the binary for manual re-run. Write a `~/.dtasks/skill-installed` marker to avoid re-prompting.

2. **PowerShell completion (COMP-03)**
   - What we know: Cobra generates PowerShell completion via `GenPowerShellCompletion`. The canonical install is appending to `$PROFILE`.
   - What's unclear: `install.sh` is a POSIX shell script — it cannot run on Windows directly. PowerShell completion would need to be handled by `install.ps1` (already exists).
   - Recommendation: Add PowerShell completion install to `install.ps1`; document COMP-03 as "bash/zsh/fish in install.sh, PowerShell in install.ps1".

3. **`dtasks update` + completion re-run (COMP-04)**
   - What we know: After updating the binary, completions should be re-installed.
   - What's unclear: The update command is Go code; calling shell completion install logic from Go means either re-exec'ing `install.sh` or duplicating the logic.
   - Recommendation: After successful update, `updateCmd` prints a hint: "Run install.sh to update completions" rather than attempting to do it inline. This avoids coupling Go and shell logic.

## Validation Architecture

### Test Framework

| Property | Value |
|----------|-------|
| Framework | Go testing (stdlib), v1.22 |
| Config file | none (go test ./...) |
| Quick run command | `go test ./internal/... -run TestUpdate -v` |
| Full suite command | `go test ./...` |

### Phase Requirements → Test Map

| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| UPDT-01 | `dtasks update` command exists and runs | smoke | `go build ./... && ./dist/dtasks update --help` | ❌ Wave 0 |
| UPDT-02 | Shows current + latest version | unit | `go test ./internal/updater/... -run TestFetchLatestVersion -v` | ❌ Wave 0 |
| UPDT-03 | Downloads and atomically replaces binary | unit | `go test ./internal/updater/... -run TestAtomicReplace -v` | ❌ Wave 0 |
| UPDT-04 | `--json` outputs structured result | unit | `go test ./cmd/... -run TestUpdateCmd_JSON -v` | ❌ Wave 0 |
| COMP-01 | `install.sh` detects shell via `$SHELL` | manual | `SHELL=/bin/zsh bash install.sh` (inspect output) | manual-only |
| COMP-02 | Skips prompt in non-TTY | manual | `echo "" \| bash install.sh` (should not prompt) | manual-only |
| COMP-03 | Writes completions to canonical location | integration | `bash install.sh` then check target files exist | manual-only |
| COMP-04 | Completions run on upgrade | manual | Run update command, verify completion hint printed | manual-only |
| SKIL-01 | Detects Claude via `~/.claude/` or `claude` in PATH | unit | `go test ./internal/skill/... -run TestDetectClaude -v` | ❌ Wave 0 |
| SKIL-02 | Prompts for consent (skips non-TTY) | unit | `go test ./internal/skill/... -run TestSkillInstall_NonTTY -v` | ❌ Wave 0 |
| SKIL-03 | Copies to `~/.claude/skills/dtasks-cli/SKILL.md` | unit | `go test ./internal/skill/... -run TestInstallSkill_Path -v` | ❌ Wave 0 |
| SKIL-04 | Overwrites silently; skips if not found | unit | `go test ./internal/skill/... -run TestInstallSkill_Overwrite -v` | ❌ Wave 0 |

### Sampling Rate

- **Per task commit:** `go test ./...`
- **Per wave merge:** `go test ./... && go vet ./...`
- **Phase gate:** Full suite green before `/gsd:verify-work`

### Wave 0 Gaps

- [ ] `internal/updater/updater.go` + `internal/updater/updater_test.go` — covers UPDT-02, UPDT-03
- [ ] `internal/skill/skill.go` + `internal/skill/skill_test.go` — covers SKIL-01..04
- [ ] `cmd/update.go` — the update cobra command (UPDT-01, UPDT-04)
- [ ] Test helpers: mock HTTP server for GitHub API responses (for UPDT-02 unit tests without network)

## Sources

### Primary (HIGH confidence)

- stdlib `os`, `net/http`, `encoding/json`, `runtime` — Go 1.22 standard library (confirmed in go.mod)
- `github.com/spf13/cobra` v1.8.0 — already in go.mod; completion generators verified via pkg.go.dev
- GitHub REST API docs (https://docs.github.com/en/rest/releases/releases) — official; anonymous access confirmed for public repos
- Claude Code Skills docs (https://code.claude.com/docs/en/skills) — official; `~/.claude/skills/<name>/SKILL.md` path confirmed
- This project's `release.yml` — definitive source for binary asset naming convention

### Secondary (MEDIUM confidence)

- Cobra shell completion docs (https://cobra.dev/docs/how-to-guides/shell-completion/) — canonical installation paths for bash/zsh/fish
- WebSearch: Claude Code config dir migration to `~/.config/claude/` in v1.0.30 — multiple sources agree; `~/.claude/` still valid for skills

### Tertiary (LOW confidence)

- Windows atomic rename behavior for running executables — multiple sources mention it works for Go binaries (no file lock on the binary itself in recent Go), but not definitively verified; flag as needing Windows UAT.

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — no new dependencies; all stdlib
- Architecture: HIGH — patterns derived from existing project conventions and official docs
- Pitfalls: HIGH — most derived from official sources or directly observable project state
- Windows self-update: LOW — behavior depends on Windows version and AV; flag for manual testing

**Research date:** 2026-03-06
**Valid until:** 2026-06-06 (stable domain; cobra and GitHub API are stable)
