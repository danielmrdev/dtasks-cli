# dtasks-cli

## What This Is

CLI task manager written in Go. Single static binary, no runtime dependencies. SQLite as the database backend, designed to run on macOS, Linux, and Windows. The database path can point to a synced folder (Dropbox, iCloud Drive, Syncthing…) to share tasks across machines.

## Core Value

Tasks are always reachable from the terminal with a single fast command — no UI, no login, no overhead.

## Requirements

### Validated

- ✓ Task CRUD (add, edit, done, undone, rm) — existing
- ✓ Task lists (create, ls, edit, rm) — existing
- ✓ Recurring tasks (daily/weekly/monthly, autocomplete, scheduler) — existing
- ✓ Subtasks — existing
- ✓ Due dates and times — existing
- ✓ Table and JSON output — existing
- ✓ Cross-platform config and DB paths — existing
- ✓ Shell completions (cobra-generated) — existing
- ✓ `--due-today` filter — existing

### Active

- [ ] Essential filters: --overdue, --today, --tomorrow, --week (issue #13)
- [ ] Sorting options for task listing: --sort=due|priority|created|completed, --reverse (issue #14)
- [ ] Search tasks by keyword: `dtasks find <keyword>`, --list, --regex (issue #15)
- [ ] Task priorities: high/medium/low field, visual indicator, sort by priority (issue #17)
- [ ] Bulk delete completed tasks: `dtasks rm --completed <date>`, --dry-run, --yes (issue #9)
- [ ] Stats command: `dtasks stats` with totals, pending, done, % by list (issue #16)
- [ ] Self-update command: `dtasks update` fetches latest GitHub release and replaces binary (issue #7)
- [ ] Shell completions setup during install/update: interactive prompt, shell auto-detection (issue #12)
- [ ] Auto-install dtasks skill to Claude/Codex/OpenCode on first run (issue #18)

### Out of Scope

- System notifications — high complexity, not core value
- Sync / cloud backend — out of v0.3.0 scope
- Tags, priorities (beyond #17 scope) — deferred
- Mobile app — web-first or CLI-first

## Context

Brownfield project at v0.2.0. Codebase fully mapped in `.planning/codebase/`. All 9 open GitHub issues tracked and scoped for v0.3.0. Release flow: feature branch → PR to main → tag v0.3.0 → GitHub Actions CI builds and publishes release assets.

## Constraints

- **Tech stack**: Go, modernc.org/sqlite (CGO_ENABLED=0), Cobra — no new runtime dependencies unless strictly necessary
- **Binary**: Must remain a single static binary for all platforms (macOS arm64/amd64, Linux amd64/arm64, Windows amd64/arm64)
- **CI**: Release triggered by git tag push (GH Actions already configured)
- **Branch strategy**: Work on feature branch, PR to main, tag v0.3.0 after merge

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| One milestone (v0.3.0) for all 9 issues | Issues are cohesive and form a natural feature layer | — Pending |
| 3 phases: querying → richness → tooling | Dependencies flow cleanly; querying needed before sort/filter UX | — Pending |
| Feature branch workflow | PR to main + tag for CI release automation | — Pending |

---
*Last updated: 2026-03-06 after initialization*
