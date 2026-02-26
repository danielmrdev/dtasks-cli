# dtasks — Product Requirements Document

**Version:** 1.0  
**Date:** 2026-02-26  
**Status:** Draft

-----

## 1. Overview

`dtasks` is a CLI task/reminder manager that compiles to a single static binary targeting macOS Apple Silicon and Linux (amd64/arm64). A companion macOS status bar app consumes the same SQLite database.

-----

## 2. Goals

- Single binary, zero runtime dependencies
- Cross-platform: macOS Silicon (`aarch64-apple-darwin`) + Linux (`aarch64/x86_64-unknown-linux`)
- Shared SQLite DB between macOS host and Linux Docker container via volume mount
- Fast, scriptable, composable (human output + `--json` flag)

-----

## 3. Tech Stack

|Component    |Choice              |Reason                                  |
|-------------|--------------------|----------------------------------------|
|Language     |Go                  |Simple cross-compile, good CLI ecosystem|
|SQLite driver|`modernc.org/sqlite`|Pure Go, no CGO, static binary          |
|CLI framework|`cobra`             |Standard, subcommand support            |
|Date parsing |`natural` / manual  |User-friendly date input                |
|Config       |`.env` file         |Shared between platforms                |

-----

## 4. Configuration

### First Run Wizard

If no config file is found, the CLI interactively asks for the DB path and creates:

- The `.env` file at the platform-specific location
- An empty SQLite database at the specified path

### Config File Locations

|Platform|Config Path            |
|--------|-----------------------|
|macOS   |`~/.dtasks/.env`       |
|Linux   |`~/.config/dtasks/.env`|

### Default DB Paths (if user accepts defaults)

|Platform|Default DB                                     |
|--------|-----------------------------------------------|
|macOS   |`~/Library/Application Support/dtasks/tasks.db`|
|Linux   |`~/.local/share/dtasks/tasks.db`               |

### `.env` format

```
DB_PATH=/path/to/tasks.db
```

Both config files can point to the same physical file (e.g., a Docker volume or synced folder).

### SQLite settings

```sql
PRAGMA journal_mode=WAL;
PRAGMA busy_timeout=5000;
```

Required for concurrent access between macOS host and Docker Linux container.

-----

## 5. Data Model

### lists

|Column    |Type      |Notes |
|----------|----------|------|
|id        |INTEGER PK|auto  |
|name      |TEXT      |unique|
|created_at|DATETIME  |      |

### tasks

|Column            |Type      |Notes                                     |
|------------------|----------|------------------------------------------|
|id                |INTEGER PK|auto                                      |
|list_id           |INTEGER FK|→ lists.id                                |
|parent_task_id    |INTEGER FK|→ tasks.id, nullable (subtask)            |
|title             |TEXT      |required                                  |
|notes             |TEXT      |nullable                                  |
|date              |DATE      |nullable, scheduled date                  |
|time              |TEXT      |`HH:MM`, nullable — if null, all-day      |
|due_date          |DATE      |nullable                                  |
|due_time          |TEXT      |`HH:MM`, nullable — if null, all-day      |
|completed         |BOOLEAN   |default false                             |
|completed_at      |DATETIME  |nullable                                  |
|recurring         |BOOLEAN   |default false                             |
|recur_type        |TEXT      |nullable — `daily` | `weekly` | `monthly` |
|recur_interval    |INTEGER   |default 1 — every N days/weeks/months     |
|recur_time        |TEXT      |`HH:MM`, nullable                         |
|recur_day_of_week |INTEGER   |0–6, for weekly                           |
|recur_day_of_month|INTEGER   |1–31, for monthly                         |
|recur_starts      |DATE      |nullable                                  |
|recur_ends_type   |TEXT      |nullable — `never` | `on_date` | `after_n`|
|recur_ends_date   |DATE      |nullable                                  |
|recur_ends_after  |INTEGER   |nullable, max repetitions                 |
|recur_count       |INTEGER   |completed repetitions, default 0          |
|created_at        |DATETIME  |                                          |

-----

## 6. CLI Interface

### Global flags

```
--json       Output as JSON
--db PATH    Override DB_PATH from env
```

### Lists

```bash
dtasks list create "Personal"
dtasks list ls
dtasks list rename 1 "Home"
dtasks list rm 1
```

### Tasks

```bash
# Create
dtasks add --list 1 "Buy milk"
dtasks add --list 1 "Buy milk" --due "2026-03-01 10:00" --notes "organic"
dtasks add --list 1 --parent 5 "Subtask title"

# Read
dtasks ls                        # all pending tasks
dtasks ls --list 1               # tasks in list
dtasks ls --due-today            # due today
dtasks ls --all                  # including completed
dtasks show 42                   # full detail

# Update
dtasks edit 42 --title "New title"
dtasks edit 42 --due "2026-04-01" --notes "updated"
dtasks done 42
dtasks undone 42

# Delete
dtasks rm 42
```

### Recurrence

```bash
dtasks recur 42 daily --time "09:00" --starts "2026-03-01"
dtasks recur 42 weekly --day thu --time "10:00" --ends-after 30
dtasks recur 42 monthly --day 25 --ends "2027-01-01"
dtasks recur 42 monthly --day 25 --ends never
dtasks recur rm 42
```

-----

## 7. Output

### Human (default)

Formatted tables with color support (TTY detection). Example:

```
 ID  LIST      TITLE           DUE          
 42  Personal  Buy milk        Mar 01 10:00 
 43  Work      Review PR #12   Today 15:00  ⚠
```

### JSON (`--json`)

```json
{
  "tasks": [
    {
      "id": 42,
      "list": "Personal",
      "title": "Buy milk",
      "due_datetime": "2026-03-01T10:00:00Z",
      "completed": false,
      ...
    }
  ]
}
```

-----

## 8. Build Targets

```
GOOS=darwin  GOARCH=arm64  CGO_ENABLED=0  → dtasks-macos-arm64
GOOS=linux   GOARCH=amd64  CGO_ENABLED=0  → dtasks-linux-amd64
GOOS=linux   GOARCH=arm64  CGO_ENABLED=0  → dtasks-linux-arm64
```

Makefile target: `make build-all`

-----

## 9. Project Structure

```
dtasks/
├── cmd/
│   ├── root.go
│   ├── list.go
│   ├── task.go
│   └── recur.go
├── internal/
│   ├── config/     # env loading, first-run wizard
│   ├── db/         # SQLite init, migrations, WAL setup
│   ├── models/     # List, Task, Recurrence structs
│   ├── repo/       # CRUD operations
│   └── output/     # table printer + JSON serializer
├── main.go
├── Makefile
└── go.mod
```

-----

## 10. Out of Scope (v1)

- Notifications / system alerts
- Sync / cloud backend
- Import from other task managers
- Tags / labels
- Priority levels
- Attachments

-----

## 11. Open Questions

- Formato de fecha en input: ISO estricto — `2026-03-01`, hora `09:00`