# Project Retrospective

*A living document updated after each milestone. Lessons feed forward into future planning.*

## Milestone: v0.3 — Querying, Richness & Tooling

**Shipped:** 2026-03-07
**Phases:** 7 | **Plans:** 18 | **Files changed:** 31

### What Was Built
- Time-based filters on `task ls`: `--today`, `--overdue`, `--tomorrow`, `--week`
- Flexible sorting (`--sort=due|created|completed|priority`, `--reverse`) and keyword/regex search (`dtasks find`)
- Task priorities (high/medium/low) with visual indicator in table and sort support
- Bulk delete completed tasks and `dtasks stats` summary command
- Self-update binary in-place via GitHub Releases API with atomic OS swap
- Shell completions auto-install (bash/zsh/fish/PowerShell) and Claude skill auto-install
- GitHub release pipeline with 6 platform binaries triggered by tag push

### What Worked
- **TDD red-phase pattern**: Defining failing test scaffolds before implementation caught interface mismatches upfront and kept implementation plans honest. Every phase with a test plan (01, 02, 03) required no rework.
- **Milestone audit before closing**: Catching 4 gaps (sort-help-text, skill-first-install-path, json-update-flow, updt-04-json-contamination) via the audit prevented these from becoming long-term debt. Closing them via short phases (5, 6, 7) worked well.
- **Gap closure as named phases**: Treating each audit gap as an explicit phase (5, 6, 7) with its own PLAN.md kept work traceable and auditable.
- **Feature branch workflow**: PR to main + squash merge + tag produced a clean, linear history. No rework on main.
- **output.JSONMode as SSOT**: Consolidating JSON mode detection in PersistentPreRunE eliminated per-command flag-read duplication.

### What Was Inefficient
- **Accomplishments field in milestone CLI**: The `gsd-tools milestone complete` CLI left accomplishments empty because SUMMARY.md frontmatter doesn't use `one_liner` field — had to fill manually. Consider adding summary extraction fallback to phase goal or `provides` field.
- **Phase count creep via gap closure**: 3 core phases grew to 7 with gap closures. Audit-driven phases are necessary but add overhead — could pre-empt some gaps during integration checks within each phase.
- **STATE.md progress not fully updated by gsd-tools**: After milestone complete, STATE.md still showed old milestone/focus data. Needed manual update.

### Patterns Established
- **Skip list in PersistentPreRunE**: Commands that must work without DB (update, install-skill) are explicitly listed in root.go PersistentPreRunE skip check. Pattern: `cmd.Name() == "update" || cmd.Name() == "install-skill"`.
- **skilldata wrapper package**: Go `//go:embed` prohibits `..` traversal — use a thin wrapper package at the embed source location.
- **DryRun pattern in repo**: Fetch rows first (SELECT), skip DELETE if DryRun=true. Single code path for preview and destructive execute.
- **Shell TTY guard in install.sh**: `[ -t 0 ] || return 0` before any interactive prompt in shell script functions.

### Key Lessons
1. **Audit before close, always**: The 4 gaps caught by the milestone audit justified the process investment. Skipping audit would have shipped a broken `--json update` flow.
2. **TDD red phases pay off even for small plans**: Phase 01-01 and 02-01 each took <5 min but saved debugging time in implementation plans by establishing the exact API contract upfront.
3. **Atomic binary replacement is non-trivial**: `filepath.Dir(exePath)` for temp file location avoids cross-device rename failures. This is a hidden footgun that the pattern now encodes explicitly.
4. **output.JSONMode duplication is a smell**: Any command that re-reads `--json` locally is likely wrong. Audit for this pattern when adding new commands.

### Cost Observations
- Model mix: ~100% Sonnet 4.6 (claude-sonnet-4-6)
- Sessions: estimated 6-8 sessions across 2 days
- Notable: Short gap-closure phases (5, 6, 7) each completed in <15 min — well-scoped atomic plans are highly efficient

---

## Cross-Milestone Trends

### Process Evolution

| Milestone | Phases | Plans | Key Change |
|-----------|--------|-------|------------|
| v0.3 | 7 | 18 | First milestone with full GSD workflow: TDD red phases, audit, gap closure phases |

### Cumulative Quality

| Milestone | Go LOC (prod) | Test coverage (estimated) | Zero-dep additions |
|-----------|---------------|--------------------------|-------------------|
| v0.3 | 2,797 | High (repo/output/updater/skill) | golang.org/x/term |

### Top Lessons (Verified Across Milestones)

1. **Audit-driven gap closure produces higher quality than skipping**: All 4 gaps caught in v0.3 audit were real defects (not theoretical).
2. **Named skip lists in CLI framework boilerplate scale poorly** — watch for PersistentPreRunE skip list growth as commands increase.
