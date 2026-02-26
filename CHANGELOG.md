# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.0] - 2026-02-26

### Added
- `dtasks add` — create tasks with title, notes, scheduled date/time, due date/time, and parent task
- `dtasks ls` — list pending tasks; filter by list (`--list`), due today (`--due-today`), or include completed (`--all`)
- `dtasks show` — full task detail including subtasks
- `dtasks edit` — partial update of any task field
- `dtasks done` / `dtasks undone` — toggle completion state
- `dtasks rm` — delete a task
- `dtasks list create/ls/rename/rm` — manage lists
- `dtasks recur daily/weekly/monthly/rm` — set or remove recurrence rules (interval, day, time, start/end)
- `--json` global flag for machine-readable output on all commands
- `--db PATH` global flag to override the database path
- `--version` flag, injected at build time via `-ldflags`
- First-run wizard: interactive prompt for database path; writes platform-specific `.env` config
- SQLite database with WAL mode and 5-second busy timeout for safe concurrent access
- Cross-platform config and data paths: macOS, Linux (XDG), Windows (`%AppData%` / `%LocalAppData%`)
- Static binaries for macOS arm64/amd64, Linux amd64/arm64, Windows amd64/arm64 (`CGO_ENABLED=0`)
- GitHub Actions CI (vet + test + build on every push/PR)
- GitHub Actions release workflow (triggered by `v*.*.*` tags, builds all 6 targets, publishes release with checksums)

[Unreleased]: https://github.com/danielmrdev/dtasks-cli/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/danielmrdev/dtasks-cli/releases/tag/v0.1.0
