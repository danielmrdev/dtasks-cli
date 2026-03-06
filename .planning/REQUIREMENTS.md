# Requirements: dtasks-cli v0.3.0

**Defined:** 2026-03-06
**Core Value:** Tasks are always reachable from the terminal with a single fast command — no UI, no login, no overhead.

## v1 Requirements

Requirements for v0.3.0 release. Each maps to a roadmap phase.

### Filters

- [x] **FILT-01**: User can list tasks due today or earlier with `ls --today`
- [x] **FILT-02**: User can list tasks past their due date with `ls --overdue`
- [x] **FILT-03**: User can list tasks due tomorrow with `ls --tomorrow`
- [x] **FILT-04**: User can list tasks due within the next 7 days with `ls --week`

### Sorting

- [x] **SORT-01**: User can sort task listing by due date, created, or completed with `ls --sort=<field>` (priority sort covered by PRIO-04 in Phase 2)
- [x] **SORT-02**: User can reverse the sort order with `ls --reverse`

### Search

- [x] **SRCH-01**: User can search tasks by keyword across title and notes with `dtasks find <keyword>` (case-insensitive)
- [x] **SRCH-02**: User can scope search to a specific list with `find --list <id>`
- [x] **SRCH-03**: User can search with a regex pattern with `find --regex`

### Priority

- [x] **PRIO-01**: User can set task priority (high/medium/low) when adding a task with `add --priority <level>`
- [x] **PRIO-02**: User can set task priority when editing a task with `edit --priority <level>`
- [x] **PRIO-03**: Priority is shown as a visual indicator in table output
- [x] **PRIO-04**: Task listing can be sorted by priority

### Maintenance

- [x] **MAINT-01**: User can bulk-delete completed tasks on or before a date with `rm --completed <date>`
- [x] **MAINT-02**: User can preview what would be deleted without committing with `rm --completed <date> --dry-run`
- [x] **MAINT-03**: Bulk delete requires explicit confirmation unless `--yes` is passed
- [x] **MAINT-04**: Bulk delete respects `--json` flag (emits `{"deleted": N}`)
- [x] **MAINT-05**: Bulk delete can be scoped to a specific list with `--list <id>`

### Stats

- [x] **STAT-01**: User can view a task summary with `dtasks stats` (total, pending, done, % by list, upcoming due dates)
- [x] **STAT-02**: Stats command respects `--json` flag

### Self-update

- [x] **UPDT-01**: User can check for and install updates with `dtasks update`
- [x] **UPDT-02**: `dtasks update` shows current version and latest available
- [x] **UPDT-03**: `dtasks update` downloads and atomically replaces the running binary for the correct OS/arch
- [x] **UPDT-04**: `dtasks update` respects `--json` flag

### Install Completions

- [ ] **COMP-01**: `install.sh` detects the user's current shell automatically
- [ ] **COMP-02**: `install.sh` prompts interactively to install shell completions (skips when stdin is not a TTY)
- [ ] **COMP-03**: Completions are written to the canonical location for bash, zsh, fish, and PowerShell
- [ ] **COMP-04**: Completion setup also runs on upgrade (update path)

### Skill Auto-install

- [x] **SKIL-01**: On first run (or during install), the CLI detects whether Claude is installed (`~/.claude/` or `claude` command)
- [x] **SKIL-02**: User is prompted for consent before copying the skill
- [x] **SKIL-03**: Skill is copied to the correct platform path (`~/.claude/skills/dtasks-cli/`)
- [x] **SKIL-04**: If skill already exists it is overwritten silently; if platform not found, install is skipped gracefully

## v2 Requirements

Deferred to future releases.

### Notifications

- System notifications on task due — high complexity, not core value

### Tagging

- Task tags / labels — deferred until priority UX is validated

## Out of Scope

| Feature | Reason |
|---------|--------|
| Sync / cloud backend | Not in scope for v0.3.0 |
| Mobile app | CLI-first |
| OAuth / team features | Single-user tool |
| Real-time collaboration | Architecture mismatch |

## Traceability

Which phases cover which requirements. Updated during roadmap creation.

| Requirement | Phase | Status |
|-------------|-------|--------|
| FILT-01 | Phase 1 | Complete |
| FILT-02 | Phase 1 | Complete |
| FILT-03 | Phase 1 | Complete |
| FILT-04 | Phase 1 | Complete |
| SORT-01 | Phase 1 | Complete |
| SORT-02 | Phase 1 | Complete |
| SRCH-01 | Phase 1 | Complete |
| SRCH-02 | Phase 1 | Complete |
| SRCH-03 | Phase 1 | Complete |
| PRIO-01 | Phase 2 | Complete |
| PRIO-02 | Phase 2 | Complete |
| PRIO-03 | Phase 2 | Complete |
| PRIO-04 | Phase 2 | Complete |
| MAINT-01 | Phase 2 | Complete |
| MAINT-02 | Phase 2 | Complete |
| MAINT-03 | Phase 2 | Complete |
| MAINT-04 | Phase 2 | Complete |
| MAINT-05 | Phase 2 | Complete |
| STAT-01 | Phase 2 | Complete |
| STAT-02 | Phase 2 | Complete |
| UPDT-01 | Phase 3 | Complete |
| UPDT-02 | Phase 3 | Complete |
| UPDT-03 | Phase 3 | Complete |
| UPDT-04 | Phase 3 | Complete |
| COMP-01 | Phase 3 | Pending |
| COMP-02 | Phase 3 | Pending |
| COMP-03 | Phase 3 | Pending |
| COMP-04 | Phase 3 | Pending |
| SKIL-01 | Phase 3 | Complete |
| SKIL-02 | Phase 3 | Complete |
| SKIL-03 | Phase 3 | Complete |
| SKIL-04 | Phase 3 | Complete |

**Coverage:**
- v1 requirements: 32 total
- Mapped to phases: 32
- Unmapped: 0 ✓

---
*Requirements defined: 2026-03-06*
*Last updated: 2026-03-06 after initial definition*
