# External Integrations

**Analysis Date:** 2026-03-06

## APIs & External Services

**None detected.**

This is a standalone CLI with no external API integrations. All functionality is local-only.

## Data Storage

**Databases:**
- SQLite (local file-based)
  - Connection: Direct file path via `modernc.org/sqlite` driver
  - Client: Go standard library `database/sql`
  - Location: Platform-specific (default paths in `internal/config/config.go`)
    - macOS: `~/Library/Application Support/dtasks/tasks.db`
    - Linux: `~/.local/share/dtasks/tasks.db`
    - Windows: `%LOCALAPPDATA%\dtasks\tasks.db`
  - Configuration: `PRAGMA journal_mode=WAL`, `PRAGMA busy_timeout=5000`, `PRAGMA foreign_keys=ON` in `internal/db/db.go`

**File Storage:**
- Local filesystem only
- No cloud/remote storage integration (database syncing delegated to user tools: Dropbox, Google Drive, iCloud, Syncthing)

**Caching:**
- None - SQLite handles all data access with built-in caching

## Authentication & Identity

**None required.**

dtasks is a single-user CLI with no user authentication or identity system. All tasks belong to the local database owner.

## Monitoring & Observability

**Error Tracking:**
- None - Errors logged directly to stderr via standard Go logging

**Logs:**
- stdout/stderr only
- JSON mode available via `--json` flag for structured output (controlled by `output.JSONMode` global in `internal/output/output.go`)
- No persistent logging or aggregation

## CI/CD & Deployment

**Hosting:**
- GitHub (source repository only)
- No deployed backend or cloud infrastructure

**CI Pipeline:**
- GitHub Actions (inferred from `make release TAG=v1.2.3` in `Makefile`)
- Release workflow triggered by git tags
- Builds: `make build-all` generates binaries for:
  - macOS: arm64, amd64
  - Linux: amd64, arm64
  - Windows: amd64.exe, arm64.exe

**Local Installation:**
- `make install` - Copies binary to `~/.local/bin/dtasks` or `/usr/local/bin/dtasks`

## Environment Configuration

**Required env vars:**
- `DB_PATH` - Path to SQLite database file (string, required in `.env`)

**Optional env vars:**
- `XDG_CONFIG_HOME` - Linux only, overrides default `~/.config` for `.env` location
- `XDG_DATA_HOME` - Linux only, overrides default `~/.local/share` for database
- `LOCALAPPDATA` - Windows only, overrides default AppData for database

**Secrets location:**
- No secrets stored - dtasks has no authentication or API keys
- `.env` file contains only `DB_PATH`

## Webhooks & Callbacks

**Incoming:**
- None - No server or API endpoints

**Outgoing:**
- None - All operations are synchronous, local-only

---

*Integration audit: 2026-03-06*
