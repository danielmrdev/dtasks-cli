# dtasks-cli

## What This Is

CLI task manager written in Go. Single static binary, no runtime dependencies. SQLite as the database backend, designed to run on macOS, Linux, and Windows. The database path can point to a synced folder (Dropbox, iCloud Drive, Syncthing…) to share tasks across machines. Ships with self-update, shell completions, and Claude skill auto-install.

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
- ✓ Essential filters: `--today`, `--overdue`, `--tomorrow`, `--week` — v0.3
- ✓ Sorting: `--sort=due|created|completed|priority`, `--reverse` — v0.3
- ✓ Keyword search: `dtasks find <keyword>`, `--list`, `--regex` — v0.3
- ✓ Task priorities: high/medium/low, visual indicator, sort by priority — v0.3
- ✓ Bulk delete completed tasks: `dtasks task rm --completed`, `--dry-run`, `--yes` — v0.3
- ✓ Stats command: `dtasks stats` with totals, pending, done, % by list — v0.3
- ✓ Self-update: `dtasks update` fetches latest GitHub release and replaces binary atomically — v0.3
- ✓ Shell completions setup during install/update: interactive prompt, shell auto-detection — v0.3
- ✓ Claude skill auto-install on first run and via `install-skill` command — v0.3

### Active

(None — all scoped requirements shipped in v0.3)

### Out of Scope

- System notifications — high complexity, not core value
- Sync / cloud backend — out of current scope
- Tags / labels — deferred until priority UX is validated in production
- Mobile app — CLI-first

## Context

Shipped v0.3.0 with 2,797 LOC Go (production). Tech stack: Go, modernc.org/sqlite (CGO_ENABLED=0), Cobra, golang.org/x/term. Release pipeline: GitHub Actions triggered by tag push, produces 6 platform binaries.

Known tech debt:
- COMP-04: shell completions are hint-only after `update` (not re-installed automatically)
- `TaskUpdate` in repo/task.go does not persist `Priority` in SQL UPDATE — no regression, but confusing
- Pre-existing scheduler tests with hardcoded dates fail intermittently (out of scope)

## Constraints

- **Tech stack**: Go, modernc.org/sqlite (CGO_ENABLED=0), Cobra — no new runtime dependencies unless strictly necessary
- **Binary**: Must remain a single static binary for all platforms (macOS arm64/amd64, Linux amd64/arm64, Windows amd64/arm64)
- **CI**: Release triggered by git tag push (GH Actions already configured)

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| One milestone (v0.3.0) for all 9 issues | Issues are cohesive and form a natural feature layer | ✓ Good — clean shipping unit |
| 3 core phases: querying → richness → tooling | Dependencies flow cleanly; querying needed before sort/filter UX | ✓ Good — no rework needed |
| Feature branch workflow | PR to main + tag for CI release automation | ✓ Good — clean linear history on main |
| TDD red phase for every implementation phase | Forces contract definition before implementation, catches interface mismatches early | ✓ Good — caught several interface issues upfront |
| Dynamic ORDER BY builder in TaskList | Append WHERE/ORDER BY at runtime; base const has no ORDER BY clause | ✓ Good — flexible, testable |
| Dual-mode search (SQL LIKE + Go regexp) | SQL LIKE for keyword mode, post-fetch Go regexp for regex mode | ✓ Good — avoids SQLite REGEXP extension dependency |
| DryRun pattern: fetch rows first, skip DELETE if DryRun=true | Enables preview without separate query path | ✓ Good — single code path for preview and execute |
| skilldata wrapper package to embed SKILL.md | Go `//go:embed` prohibits `..` path traversal | ✓ Good — clean workaround |
| output.JSONMode as SSOT for updateCmd | Avoids double flag-read, single source of truth for JSON mode | ✓ Good — fixed contamination bug |
| Squash merge strategy for PRs | Clean linear history on main | ✓ Good — easy bisect |
| PersistentPreRunE skip list for update/install-skill | Commands must work without DB on fresh installs | ✓ Good — no false errors on first install |

---
*Last updated: 2026-03-07 after v0.3 milestone*
