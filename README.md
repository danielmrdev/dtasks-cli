# dtasks

CLI task and reminder manager. Single static binary, no runtime dependencies.

## Setup

```bash
go mod tidy
make build-all
# binaries in dist/
```

Or install to `/usr/local/bin`:

```bash
make install
```

## First run

On first run, dtasks will ask for a database path and create the config:

- **macOS**: `~/.dtasks/.env`
- **Linux**: `~/.config/dtasks/.env`

Both can point to the same `.db` file (shared volume, sync folder).

## Usage

```bash
# Lists
dtasks list create "Personal"
dtasks list ls
dtasks list rename 1 "Home"
dtasks list rm 1

# Tasks
dtasks add --list 1 "Buy milk"
dtasks add --list 1 "Buy milk" --due 2026-03-01 --due-time 10:00 --notes "organic"
dtasks add --list 1 --parent 5 "Subtask"

dtasks ls                    # pending tasks
dtasks ls --list 1           # filter by list
dtasks ls --due-today        # due today or overdue
dtasks ls --all              # include completed

dtasks show 42
dtasks edit 42 --title "New title" --due 2026-04-01
dtasks done 42
dtasks undone 42
dtasks rm 42

# Recurrence
dtasks recur daily 42 --every 3 --time 12:00
dtasks recur weekly 42 --every 2 --day thu --time 10:00
dtasks recur monthly 42 --every 1 --day 25 --ends never
dtasks recur monthly 42 --every 3 --day 1 --ends-after 30
dtasks recur monthly 42 --every 1 --day 1 --ends 2027-01-01
dtasks recur rm 42

# JSON output (for scripting / companion app)
dtasks ls --json
dtasks show 42 --json
```

## Docker / shared DB

```yaml
# docker-compose.yml
volumes:
  - ~/.local/share/dtasks:/data/dtasks

# in container .env
DB_PATH=/data/dtasks/tasks.db
```
