# Milestones

## v0.3 Querying, Richness & Tooling (Shipped: 2026-03-07)

**Phases completed:** 7 phases, 18 plans
**Lines of code:** 2,797 Go (production)
**Files changed:** 31
**Timeline:** 2026-03-06 → 2026-03-07 (2 days)

**Key accomplishments:**
- Time filters added to `task ls`: `--today`, `--overdue`, `--tomorrow`, `--week`
- Flexible sorting (`--sort=due|created|completed|priority`, `--reverse`) and keyword search (`dtasks find <keyword>`, `--list`, `--regex`)
- Task priorities (high/medium/low) with visual indicator in table output and sort support
- Bulk delete completed tasks (`rm --completed`, `--dry-run`, `--yes`) and new `dtasks stats` command
- Self-update command (`dtasks update`) downloads and atomically replaces the running binary for the correct OS/arch
- Shell completions installer and Claude skill auto-install in first-install path via `install.sh`
- v0.3.0 released as GitHub release with 6 platform binaries via feature branch → PR → squash merge → tag pipeline

**Known tech debt:**
- COMP-04: completions post-update are hint-only, not re-installed automatically
- `TaskUpdate` in repo/task.go does not persist `Priority` in SQL UPDATE (no user-facing regression)
- Pre-existing scheduler test failures with hardcoded dates (TestAutocomplete_DueTimePassed) — out of scope

---

