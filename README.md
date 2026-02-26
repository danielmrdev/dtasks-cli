# dtasks

[![CI](https://github.com/danielmrdev/dtasks/actions/workflows/ci.yml/badge.svg)](https://github.com/danielmrdev/dtasks/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/danielmrdev/dtasks?sort=semver)](https://github.com/danielmrdev/dtasks/releases/latest)
[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![Platforms](https://img.shields.io/badge/platform-macOS%20%7C%20Linux%20%7C%20Windows-lightgrey)](#install)
[![License](https://img.shields.io/badge/license-MIT-blue)](LICENSE)

A fast, scriptable CLI task manager. Single static binary — no runtime dependencies, no daemon, no sync service.

Tasks and lists live in a local SQLite database. The same file can be shared between a macOS host and a Linux container via a volume mount, making it a good fit for Docker-based workflows.

## Install

Download the latest binary for your platform from the [Releases](https://github.com/danielmrdev/dtasks/releases/latest) page:

| Platform | File |
|----------|------|
| macOS Apple Silicon | `dtasks-macos-arm64` |
| macOS Intel | `dtasks-macos-amd64` |
| Linux x86-64 | `dtasks-linux-amd64` |
| Linux ARM64 | `dtasks-linux-arm64` |
| Windows x86-64 | `dtasks-windows-amd64.exe` |
| Windows ARM64 | `dtasks-windows-arm64.exe` |

Each release includes a `checksums.txt` file with SHA-256 hashes.

**macOS / Linux**

```bash
# Example: Linux amd64
curl -Lo dtasks https://github.com/danielmrdev/dtasks/releases/latest/download/dtasks-linux-amd64
chmod +x dtasks
sudo mv dtasks /usr/local/bin/
```

**Windows** — download the `.exe`, place it somewhere in your `PATH`, and run from PowerShell or Command Prompt.

## Build from source

Requires Go 1.22+.

```bash
git clone https://github.com/danielmrdev/dtasks
cd dtasks
go mod tidy
go build -o dtasks .
```

Build all release targets at once:

```bash
make build-all   # outputs to dist/
```

## First run

On first run, dtasks asks where to store the database and writes a config file:

| Platform | Config file |
|----------|-------------|
| macOS | `~/.dtasks/.env` |
| Linux | `~/.config/dtasks/.env` (respects `$XDG_CONFIG_HOME`) |
| Windows | `%AppData%\dtasks\.env` |

```
$ dtasks ls
Welcome to dtasks! No configuration found.

Database path [~/.local/share/dtasks/tasks.db]:
```

Press Enter to accept the default or type a custom path. The database is created automatically.

To skip the wizard, pass `--db` or set `DB_PATH` in the config file:

```bash
dtasks --db /path/to/tasks.db ls
```

## Usage

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
dtasks add --list 1 "Buy milk" --due 2026-03-01 --due-time 10:00 --notes "organic"
dtasks add --list 1 --parent 5 "Subtask title"

# Read
dtasks ls                    # pending tasks
dtasks ls --list 1           # filter by list
dtasks ls --due-today        # due today or overdue
dtasks ls --all              # include completed
dtasks show 42               # full detail + subtasks

# Update
dtasks edit 42 --title "New title"
dtasks edit 42 --due 2026-04-01 --notes "updated"
dtasks done 42
dtasks undone 42

# Delete
dtasks rm 42
```

### Recurrence

```bash
dtasks recur daily 42 --every 1 --time 09:00
dtasks recur weekly 42 --every 2 --day thu --time 10:00 --ends-after 30
dtasks recur monthly 42 --every 1 --day 25 --ends never
dtasks recur monthly 42 --every 3 --day 1 --ends 2027-01-01
dtasks recur rm 42
```

### JSON output

All commands support `--json` for scripting and integration with other tools:

```bash
dtasks ls --json
dtasks show 42 --json
dtasks list ls --json
```

## Shared database (Docker)

```yaml
# docker-compose.yml
services:
  app:
    volumes:
      - ~/.local/share/dtasks:/data/dtasks
    environment:
      - DB_PATH=/data/dtasks/tasks.db
```

Both the macOS host and the Linux container can point to the same file. The database uses WAL mode and a 5-second busy timeout to handle concurrent access safely.

## Global flags

| Flag | Description |
|------|-------------|
| `--db PATH` | Override the database path from config |
| `--json` | Output as JSON |
| `--version` | Print version and exit |

## License

MIT
