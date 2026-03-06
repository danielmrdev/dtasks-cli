---
phase: 03-tooling
plan: "03"
subsystem: tooling
tags: [skill, claude-detection, tty, term, go]

requires:
  - phase: 03-tooling plan 01
    provides: "Skill package scaffold (skill.go stub, skill_test.go)"

provides:
  - "DetectClaude: checks ~/.claude, ~/.config/claude, exec.LookPath for Claude Code"
  - "InstallSkill: writes SKILL.md to <homeDir>/.claude/skills/dtasks-cli/ with MkdirAll + WriteFile"
  - "PromptAndInstall: TTY-gated consent flow using golang.org/x/term, non-TTY installs directly"

affects: [03-04, install.sh]

tech-stack:
  added: []
  patterns:
    - "TTY detection via interface{ Fd() uintptr } type assertion on io.Reader"
    - "homeDir injection for testability — no direct os.UserHomeDir() calls"
    - "Graceful skip pattern: DetectClaude=false returns nil without error"

key-files:
  created: []
  modified:
    - internal/skill/skill.go

key-decisions:
  - "Non-TTY path installs directly without prompting (programmatic path for updateCmd)"
  - "TTY detection uses golang.org/x/term.IsTerminal via Fd() uintptr interface assertion on io.Reader"
  - "PromptAndInstall returns nil (no error) when Claude not detected or user declines — graceful skip"

patterns-established:
  - "Dependency injection for testability: homeDir, in io.Reader, out io.Writer all injected"
  - "Error wrapping convention: 'skill: create dir: %w' and 'skill: write file: %w'"

requirements-completed: [SKIL-01, SKIL-02, SKIL-03, SKIL-04]

duration: 2min
completed: 2026-03-06
---

# Phase 3 Plan 03: Skill Package Summary

**Claude detection + skill auto-install with TTY-gated consent using golang.org/x/term and io.Reader interface assertion**

## Performance

- **Duration:** ~2 min
- **Started:** 2026-03-06T12:21:12Z
- **Completed:** 2026-03-06T12:22:19Z
- **Tasks:** 2
- **Files modified:** 1

## Accomplishments

- `DetectClaude` checks three paths in order: `~/.claude` dir, `~/.config/claude` dir, `claude` binary in PATH
- `InstallSkill` creates `<homeDir>/.claude/skills/dtasks-cli/SKILL.md` via `os.MkdirAll` + `os.WriteFile`; idempotent and overwrites silently
- `PromptAndInstall` detects TTY via `interface{ Fd() uintptr }` type assertion; non-TTY installs without prompting; TTY prompts "Install dtasks skill for Claude Code? [y/N]"
- All 6 skill tests pass green; `go vet` and `go build ./...` clean

## Task Commits

Each task was committed atomically:

1. **Task 1: Implement DetectClaude and InstallSkill** - `e6185f5` (feat)
2. **Task 2: Implement PromptAndInstall** - `e6185f5` (feat, included in Task 1 commit — single coherent implementation)

## Files Created/Modified

- `internal/skill/skill.go` — Full implementation of DetectClaude, InstallSkill, PromptAndInstall

## Decisions Made

- Non-TTY path installs directly without prompting. This is the correct behavior for `updateCmd` (programmatic invocation) per RESEARCH.md.
- TTY detection: `in io.Reader` is checked for `interface{ Fd() uintptr }` — if it implements Fd, use `term.IsTerminal`; if not (e.g. `bytes.Buffer`), treat as non-TTY.
- `PromptAndInstall` returns nil (not an error) when Claude is not detected or user declines — graceful skip matches SKIL-01 contract.

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Skill package fully implemented and tested; ready for Plan 03-04 (updateCmd integration)
- `PromptAndInstall` can be called directly from `updateCmd` with `os.Stdin` / `os.Stdout`
- `InstallSkill` can be called from `install.sh` via `dtasks install-skill` subcommand (if added)

---
*Phase: 03-tooling*
*Completed: 2026-03-06*
