---
phase: 03-tooling
plan: 05
subsystem: infra
tags: [shell-completion, bash, zsh, fish, powershell, install-script]

# Dependency graph
requires:
  - phase: 03-tooling-04
    provides: dtasks update command with post-update hint for completions (COMP-04)
provides:
  - install.sh install_completions() function: TTY-aware, $SHELL-detected, writes to canonical bash/zsh/fish paths
  - install.ps1 PowerShell completion block: idempotent append to $PROFILE
  - COMP-01..04 requirements fulfilled
affects: []

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "POSIX TTY check via [ -t 0 ] before interactive prompts in shell scripts"
    - "PowerShell interactive guard via [System.Environment]::UserInteractive"
    - "Canonical completion paths: bash=~/.local/share/bash-completion/completions/, zsh=~/.zsh/completions/_<name>, fish=~/.config/fish/completions/<name>.fish"

key-files:
  created: []
  modified:
    - install.sh
    - install.ps1

key-decisions:
  - "install_completions() uses ${install_dir}/${BINARY} (the just-installed binary) to generate completions — not bare dtasks, which may not be in PATH yet"
  - "[ -t 0 ] || return 0 is the POSIX TTY check — silently skips completions in pipe/CI environments (COMP-02)"
  - "PowerShell block guards with [System.Environment]::UserInteractive — equivalent TTY semantics on Windows"
  - "install.ps1 completion append is idempotent: checks for *dtasks completion* before adding (safe for upgrade — COMP-04)"

patterns-established:
  - "TTY-aware install prompts: check stdin/interactive before any user prompt in install scripts"
  - "Idempotent completion install: detect existing entry before appending to shell profile"

requirements-completed: [COMP-01, COMP-02, COMP-03, COMP-04]

# Metrics
duration: ~10min
completed: 2026-03-06
---

# Phase 3 Plan 05: Shell Completion Install Summary

**Shell completion auto-install added to install.sh (bash/zsh/fish) and install.ps1 (PowerShell) with POSIX TTY detection and idempotent profile append**

## Performance

- **Duration:** ~10 min
- **Started:** 2026-03-06T13:20:00Z
- **Completed:** 2026-03-06T13:33:34Z
- **Tasks:** 2 (1 auto + 1 human-verify checkpoint)
- **Files modified:** 2

## Accomplishments

- install.sh: `install_completions()` function with `[ -t 0 ]` TTY guard, `$SHELL` detection, and canonical path install for bash, zsh, and fish
- install.ps1: interactive PowerShell completion block that appends to `$PROFILE` (idempotent, guards against duplicate entries)
- Both scripts skip the completions prompt entirely when stdin is not a TTY (CI/pipe environments)
- Fulfills COMP-01 (shell detection), COMP-02 (TTY skip), COMP-03 (canonical paths), COMP-04 (upgrade hint from Plan 04)

## Task Commits

1. **Task 1: Add install_completions() to install.sh and completion block to install.ps1** - `722bbf4` (feat)
2. **Task 2: Verify shell completion install interactively** - human-verify checkpoint, approved by user

**Plan metadata:** (docs commit — see below)

## Files Created/Modified

- `install.sh` - Added `install_completions()` function (46 lines): TTY check, $SHELL detection, bash/zsh/fish cases with mkdir -p and canonical paths
- `install.ps1` - Added PowerShell completion block (18 lines): interactive guard, $PROFILE creation if missing, idempotent append

## Decisions Made

- Used `${install_dir}/${BINARY}` to invoke the binary for completion generation, not the bare `dtasks` name, because the binary may not be in PATH immediately after install
- `[ -t 0 ] || return 0` chosen as the POSIX-portable TTY check (works with dash, bash, zsh as /bin/sh)
- PowerShell uses `[System.Environment]::UserInteractive` as the equivalent interactive check
- Completion append uses `*dtasks completion*` substring check to prevent duplicate profile entries on upgrade

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None.

## User Setup Required

None - no external service configuration required. Shell completions are opt-in at install time.

## Next Phase Readiness

- Phase 03-tooling is now fully complete (Plans 01-05)
- All tooling requirements delivered: Makefile CI targets (P01), GitHub Actions release workflow (P02), skill package (P03), update command (P04), shell completions (P05)
- Ready for Phase 04 (release) — milestone v0.3.0 artifacts are production-ready

---
*Phase: 03-tooling*
*Completed: 2026-03-06*
