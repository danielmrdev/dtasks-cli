# Roadmap: dtasks-cli v0.3.0

## Overview

Brownfield project at v0.2.0. This milestone adds task querying (filters, sorting, search), task richness (priorities, bulk delete, stats), and tooling (self-update, completions, skill auto-install), then ships via a feature branch → PR → tag v0.3.0 workflow.

## Phases

- [x] **Phase 1: Querying** - Filters, sorting, and keyword search for task listing (completed 2026-03-06)
- [x] **Phase 2: Richness** - Task priorities, bulk maintenance, and usage stats (completed 2026-03-06)
- [ ] **Phase 3: Tooling** - Self-update, shell completions setup, and skill auto-install
- [ ] **Phase 4: Release** - Feature branch, CI validation, tag v0.3.0, and publish release assets

## Phase Details

### Phase 1: Querying
**Goal**: Users can find and list tasks by time window, sort order, or keyword
**Depends on**: Nothing (first phase)
**Requirements**: FILT-01, FILT-02, FILT-03, FILT-04, SORT-01, SORT-02, SRCH-01, SRCH-02, SRCH-03
**Success Criteria** (what must be TRUE):
  1. User can run `dtasks task ls --today`, `--overdue`, `--tomorrow`, `--week` and receive only tasks matching the time filter
  2. User can run `dtasks task ls --sort=due` (and `created`, `completed`) and see tasks ordered accordingly; `--reverse` inverts the order (`--sort=priority` in Phase 2)
  3. User can run `dtasks find <keyword>` and receive all tasks whose title or notes contain the keyword (case-insensitive)
  4. User can scope `find` to a specific list with `--list <id>` and search with a regex pattern with `--regex`
**Plans**: 3 plans

Plans:
- [ ] 01-01-PLAN.md — Test scaffold: failing tests for all 9 filter/sort/search requirements (TDD red phase)
- [ ] 01-02-PLAN.md — Repo layer: extend TaskListOptions, refactor ORDER BY, add TaskSearch
- [ ] 01-03-PLAN.md — CLI layer: lsCmd filter/sort flags, new findCmd top-level command

### Phase 2: Richness
**Goal**: Users can assign priorities to tasks, bulk-clean completed tasks, and view task statistics
**Depends on**: Phase 1
**Requirements**: PRIO-01, PRIO-02, PRIO-03, PRIO-04, MAINT-01, MAINT-02, MAINT-03, MAINT-04, MAINT-05, STAT-01, STAT-02
**Success Criteria** (what must be TRUE):
  1. User can set `--priority high|medium|low` when adding or editing a task, and the priority is visible as a visual indicator in table output
  2. User can sort tasks by priority with `ls --sort=priority`
  3. User can run `dtasks rm --completed <date>` to bulk-delete completed tasks; `--dry-run` previews without deleting; `--yes` skips confirmation; `--list <id>` scopes to one list
  4. User can run `dtasks stats` and see total, pending, done, and percentage per list; `--json` outputs structured JSON
**Plans**: 4 plans

Plans:
- [ ] 02-01-PLAN.md — TDD red phase: failing tests for all 11 Phase 2 requirements
- [ ] 02-02-PLAN.md — Repo layer: schema migration, model extension, TaskDeleteCompleted, TaskStats
- [ ] 02-03-PLAN.md — Output layer: PRIO column in PrintTasks, new PrintStats function
- [ ] 02-04-PLAN.md — CLI layer: priority flags on add/edit, extended rmCmd, new statsCmd

### Phase 3: Tooling
**Goal**: Users can update the binary in-place, install shell completions, and have the dtasks skill auto-installed for Claude
**Depends on**: Phase 2
**Requirements**: UPDT-01, UPDT-02, UPDT-03, UPDT-04, COMP-01, COMP-02, COMP-03, COMP-04, SKIL-01, SKIL-02, SKIL-03, SKIL-04
**Success Criteria** (what must be TRUE):
  1. User can run `dtasks update` to see current and latest version, and the command downloads and atomically replaces the binary for the correct OS/arch; `--json` outputs structured result
  2. `install.sh` detects the user's shell, prompts for completion install (skips if non-TTY), and writes completions to the canonical location for bash, zsh, fish, and PowerShell; this also runs on upgrade
  3. On first run (or install), the CLI detects Claude and prompts for consent before copying the skill to `~/.claude/skills/dtasks-cli/`; overwrites silently if already present; skips gracefully if platform not found
**Plans**: TBD

### Phase 4: Release
**Goal**: v0.3.0 ships as a tagged GitHub release with compiled assets for all platforms
**Depends on**: Phase 3
**Requirements**: (no REQUIREMENTS.md entries — release infrastructure)
**Success Criteria** (what must be TRUE):
  1. All Phase 1-3 work lands in a dedicated feature branch with a PR to main
  2. CI passes on the PR (tests, lint, build for all platforms)
  3. Merging to main and pushing tag v0.3.0 triggers GH Actions to publish release assets for all 6 platform targets
**Plans**: TBD

## Progress

| Phase | Plans Complete | Status | Completed |
|-------|----------------|--------|-----------|
| 1. Querying | 3/3 | Complete   | 2026-03-06 |
| 2. Richness | 4/4 | Complete   | 2026-03-06 |
| 3. Tooling | 0/? | Not started | - |
| 4. Release | 0/? | Not started | - |
