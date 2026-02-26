# CLAUDE.md — dtasks

## What this is

CLI task manager written in Go. Single static binary, no runtime dependencies. SQLite as the database. Designed to run on macOS, Linux, and Windows. The database path can point to a synced folder (Dropbox, Google Drive, iCloud Drive, Syncthing…) to share tasks across machines.

## Build and test

```bash
# Dependencies (first time — requires network; use goproxy.io if proxy.golang.org fails)
GOPROXY=https://goproxy.io,direct go mod tidy

# Build for the current system
go build ./...

# Build for the current platform
make build              # → dist/dtasks (native OS/arch)

# Build all release targets
make build-all          # macos-arm64, macos-amd64, linux-amd64, linux-arm64, windows-amd64.exe, windows-arm64.exe → dist/

# Publish a release (creates git tag + pushes → triggers GH Actions)
make release TAG=v1.2.3

# Tests
go test ./...
go test ./internal/... -v   # verbose

# Lint
go vet ./...
```

> **Dependency note:** `proxy.golang.org` redirects downloads to `storage.googleapis.com`, which may be blocked in this environment. Use `GOPROXY=https://goproxy.io,direct` as a workaround.

## Structure

```
dtasks/
├── main.go                   # entrypoint, calls cmd.Execute()
├── cmd/
│   ├── root.go               # cobra root, global flags (--json, --db), initialises DB
│   ├── list.go               # list subcommand (create/ls/rename/rm)
│   ├── task.go               # add, ls, show, edit, done, undone, rm
│   └── recur.go              # recur daily/weekly/monthly/rm
├── internal/
│   ├── config/config.go      # loads .env, first-run wizard
│   ├── db/db.go              # opens SQLite, applies PRAGMAs, runs migration
│   ├── models/models.go      # List and Task structs
│   ├── repo/
│   │   ├── list.go              # list CRUD
│   │   ├── task.go              # task CRUD + recurrence
│   │   └── recur_scheduler.go   # scheduler: TaskScheduleNext, ProcessAutocompleteTasks
│   └── output/output.go      # prints as table or JSON (controlled by output.JSONMode)
└── Makefile
```

## Architecture

- **Entry:** `cmd/root.go` uses `PersistentPreRunE` to open the DB and then calls `repo.ProcessAutocompleteTasks` on every command invocation. The global `cmd.DB *sql.DB` is passed directly to `repo` functions.
- **Config:** `internal/config` looks for `DB_PATH` in the platform-specific `.env` file. If not found, it launches an interactive wizard that asks for the path and creates the file.
- **DB:** `internal/db` opens SQLite with WAL + busy_timeout and runs `CREATE TABLE IF NOT EXISTS` on every startup (idempotent migration). Existing DBs without the `autocomplete` column are migrated via `ALTER TABLE` guarded by `pragma_table_info`.
- **Repo:** pure functions that take `*sql.DB` and return models or an error. No global state in this package.
- **Scheduler:** `repo.TaskScheduleNext` is called from `doneCmd` after marking a task done; creates the next occurrence inside a transaction. `ProcessAutocompleteTasks` runs on every command startup and auto-completes overdue tasks flagged with `autocomplete=1`, then spawns their next occurrence.
- **Output:** `output.JSONMode` is a global bool activated by `--json`. All print functions check it.

## Code conventions

- Dates: `YYYY-MM-DD` as `string` (`*string` pointer when nullable).
- Times: `HH:MM` as `string`. `due_time` requires `due_date` — validated at the CLI layer.
- `due_date`/`due_time` are the only date fields. They drive `--due-today`, autocomplete, recurrence scheduling, and ORDER BY.
- DB IDs: `int64`.
- Optional flags: checked with `cmd.Flags().Changed("flag")` before assigning to the input struct, to distinguish "not provided" from "provided with empty value".
- `TaskPatch` for partial edits (only updates non-nil fields).
- SQLite driver: `modernc.org/sqlite` (pure Go, `CGO_ENABLED=0`). Registered under the driver name `"sqlite"`.

## Config paths

| Platform | Config | Default DB |
|----------|--------|------------|
| macOS | `~/.dtasks/.env` | `~/Library/Application Support/dtasks/tasks.db` |
| Linux | `~/.config/dtasks/.env` (respects `$XDG_CONFIG_HOME`) | `~/.local/share/dtasks/tasks.db` (respects `$XDG_DATA_HOME`) |
| Windows | `%AppData%\dtasks\.env` (`os.UserConfigDir()`) | `%LocalAppData%\dtasks\tasks.db` (`$LOCALAPPDATA`) |

The Windows paths use `os.UserConfigDir()` for the config file and `%LOCALAPPDATA%` for the database (non-roaming, machine-local data).

## Documentation

When any command, flag, behaviour, or architecture changes, update all three docs together:

| File | Audience |
|---|---|
| `README.md` | End users — usage examples, flags, workflows |
| `CLAUDE.md` | AI agents working on the codebase — architecture, conventions, structure |
| `skills/dtasks-cli/SKILL.md` | AI agents operating dtasks — commands, flags, constraints, JSON shapes |

## Existing tests

Tests live in `internal/` next to the package they cover:

| File | Covers |
|---|---|
| `internal/db/db_test.go` | `Open`, directory creation, schema migration |
| `internal/config/config_test.go` | `DefaultDBPath`, `EnvFilePath`, `Load` from `.env` |
| `internal/output/output_test.go` | Table and JSON output for lists/tasks/success/error |
| `internal/repo/repo_test.go` | Full CRUD for lists and tasks, filters, done/undone, subtasks, recurrence, cascade delete, scheduler (daily/weekly/monthly/clamp/ends), autocomplete |

`repo` and `db` tests create a temporary SQLite DB with `os.CreateTemp` and clean up with `t.Cleanup`.

## Key dependencies

| Module | Purpose |
|---|---|
| `modernc.org/sqlite v1.29.0` | Pure-Go SQLite driver |
| `github.com/spf13/cobra v1.8.0` | CLI framework |
| `github.com/joho/godotenv v1.5.1` | `.env` file loading |

## Not implemented (v1 out of scope)

- System notifications
- Sync / cloud backend
- Tags, priorities, attachments
