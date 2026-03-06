# Technology Stack

**Analysis Date:** 2026-03-06

## Languages

**Primary:**
- Go 1.22 - Entire codebase (`main.go`, `cmd/`, `internal/`)

## Runtime

**Environment:**
- Go 1.22+
- Cross-platform (macOS, Linux, Windows)

**Build Output:**
- Static binary (`dtasks` / `dtasks.exe`)
- CGO disabled (`CGO_ENABLED=0`) for pure Go compilation and no runtime dependencies

## Frameworks

**CLI:**
- Cobra v1.8.0 - Command structure and subcommand routing (`cmd/root.go`, `cmd/list.go`, `cmd/task.go`, `cmd/recur.go`)

**Testing:**
- Go testing (built-in) - Test files: `internal/db/db_test.go`, `internal/config/config_test.go`, `internal/output/output_test.go`, `internal/repo/repo_test.go`

**Build/Dev:**
- Makefile - Cross-platform builds with environment overrides
- Go modules - Dependency management

## Key Dependencies

**Critical:**
- `modernc.org/sqlite v1.29.0` - Pure-Go SQLite driver (no CGO required), registered as driver `"sqlite"`
- `github.com/spf13/cobra v1.8.0` - CLI framework for commands and flags
- `github.com/joho/godotenv v1.5.1` - Environment file loading (`.env` parsing in `internal/config/config.go`)

**Transitive (modernc.org/sqlite dependencies):**
- `modernc.org/libc v1.41.0` - C library bindings
- `modernc.org/memory v1.7.2` - Memory management
- `golang.org/x/sys v0.16.0` - Platform-specific system calls
- `github.com/google/uuid v1.3.0` - UUID generation (used by modernc.org/sqlite)
- `github.com/mattn/go-runewidth v0.0.20` - Unicode rune width calculation

## Configuration

**Environment:**
- Loaded via `godotenv` from platform-specific `.env` file
- Single env var required: `DB_PATH` (path to SQLite database)
- Config file locations:
  - macOS: `~/.dtasks/.env`
  - Linux: `~/.config/dtasks/.env` (respects `$XDG_CONFIG_HOME`)
  - Windows: `%AppData%\dtasks\.env`

**First-run:** Interactive wizard in `internal/config/config.go` (`runWizard()`) creates `.env` and database directory if not found

**Build:**
- `Makefile` - Defines build targets for all platforms
- Version injection via `-ldflags "-X main.version=<tag>"` at build time (`main.go`)
- Optimization flags: `-s -w` (strip symbols and dwarf info for smaller binaries)

## Platform Requirements

**Development:**
- Go 1.22+
- Make (for build automation)
- Git (for version tagging and release workflow)

**Runtime:**
- macOS 10.14+ (tested), Linux (glibc/musl), Windows 10+
- No external dependencies (static binary)
- SQLite database file write permissions in `%LOCALAPPDATA%` (Windows), `~/Library/Application Support/` (macOS), or `~/.local/share/` (Linux)

**Deployment:**
- GitHub Actions (inferred from `make release` publishing git tags) - triggers release workflow
- Binary distribution via GitHub Releases (per `make release` documentation)

---

*Stack analysis: 2026-03-06*
