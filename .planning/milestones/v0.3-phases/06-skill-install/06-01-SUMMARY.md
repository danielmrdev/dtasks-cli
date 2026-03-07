---
phase: 06-skill-install
plan: 01
subsystem: cli
tags: [cobra, skill, install.sh, shell]

# Dependency graph
requires:
  - phase: 03-tooling
    provides: internal/skill package (PromptAndInstall, InstallSkill, DetectClaude) and skills/dtasks-cli skilldata embed wrapper
provides:
  - cmd/install_skill.go: Cobra install-skill subcommand delegating to skill.PromptAndInstall
  - cmd/root.go: install-skill registered in init() and skipped in PersistentPreRunE
  - install.sh: install_skill() POSIX function + call after install_completions
affects: []

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "PersistentPreRunE skip list extended with cmd.Name() string match for commands that must work without DB (install-skill, update)"
    - "Shell TTY guard pattern: [ -t 0 ] || return 0 before calling binary interactively"
    - "|| true in install.sh binary calls to prevent non-zero exit from aborting installer"

key-files:
  created:
    - cmd/install_skill.go
    - cmd/install_skill_test.go
  modified:
    - cmd/root.go
    - install.sh

key-decisions:
  - "install-skill uses os.Stdin (not cmd.InOrStdin()) so PromptAndInstall receives real TTY — consistent with updateCmd pattern"
  - "Shell-level [ -t 0 ] guard in install_skill() is the primary non-TTY gate; binary-level guard is secondary"
  - "No success message from RunE: PromptAndInstall handles all output via the passed io.Writer"

patterns-established:
  - "Commands that must run without config/DB are explicitly added to PersistentPreRunE skip list"
  - "TDD RED: test compiled but import undefined; GREEN: implementation makes tests pass"

requirements-completed: [SKIL-01, SKIL-02, SKIL-03, SKIL-04]

# Metrics
duration: 1min
completed: 2026-03-07
---

# Phase 6 Plan 01: Skill Install Command Summary

**`dtasks install-skill` Cobra command wiring internal/skill into CLI and install.sh TTY-guarded shell function closing the first-install skill consent gap**

## Performance

- **Duration:** 1 min
- **Started:** 2026-03-07T14:52:12Z
- **Completed:** 2026-03-07T14:53:30Z
- **Tasks:** 3
- **Files modified:** 4

## Accomplishments

- Created `cmd/install_skill.go`: Cobra command that calls `skill.PromptAndInstall` with same pattern as `updateCmd`
- Extended `cmd/root.go` to register `installSkillCmd` and skip DB init for `install-skill` in `PersistentPreRunE`
- Added POSIX-compliant `install_skill()` function to `install.sh` with TTY guard and safe `|| true` exit handling
- All tests green (cmd, internal/skill, repo, and all other packages pass)

## Task Commits

Each task was committed atomically:

1. **Task 1: Write install-skill test scaffold** - `8c29591` (test - RED phase)
2. **Task 2: Create install-skill command and wire to root** - `d89818b` (feat - GREEN phase)
3. **Task 3: Add install_skill to install.sh** - `afc6cb8` (feat)

## Files Created/Modified

- `cmd/install_skill.go` - Cobra subcommand delegating to skill.PromptAndInstall
- `cmd/install_skill_test.go` - TestInstallSkillCmd_NonTTY and TestInstallSkillCmd_Help
- `cmd/root.go` - install-skill added to PersistentPreRunE skip list and init() AddCommand
- `install.sh` - install_skill() function + call at end of script

## Decisions Made

- `install-skill` uses `os.Stdin` (not `cmd.InOrStdin()`) so `PromptAndInstall` receives the real TTY file descriptor — consistent with the `updateCmd` reference pattern from the plan
- Shell-level `[ -t 0 ]` guard in `install_skill()` is the primary non-TTY gate; the binary-level check via `skill.PromptAndInstall` is the secondary guard
- No success message from `RunE`: `PromptAndInstall` handles all output internally via the passed `io.Writer`

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- All SKIL-01..04 requirements satisfied
- `dtasks install-skill` available for users who want to manually trigger skill consent after initial install
- Phase 6 complete: skill auto-install flow fully wired from install.sh through CLI to internal/skill

## Self-Check: PASSED

All files verified present. All commits verified in git history.

---
*Phase: 06-skill-install*
*Completed: 2026-03-07*
