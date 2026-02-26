# CLAUDE.md — dtasks

## Qué es esto

CLI task manager escrito en Go. Binario estático, sin runtime deps. SQLite como base de datos. Pensado para correr en macOS y Linux (incluyendo Docker), compartiendo el mismo fichero `.db` vía volumen o carpeta sincronizada.

## Build y test

```bash
# Dependencias (primera vez — requiere red; usa goproxy.io si proxy.golang.org falla)
GOPROXY=https://goproxy.io,direct go mod tidy

# Compilar para el sistema actual
go build ./...

# Compilar todos los targets
make build-all          # darwin-arm64, linux-amd64, linux-arm64 → dist/

# Tests
go test ./...
go test ./internal/... -v   # verbose

# Lint
go vet ./...
```

> **Nota sobre dependencias:** `proxy.golang.org` redirige las descargas a `storage.googleapis.com`, que puede estar bloqueado en este entorno. Usar `GOPROXY=https://goproxy.io,direct` como workaround.

## Estructura

```
dtasks/
├── main.go                   # entrypoint, llama a cmd.Execute()
├── cmd/
│   ├── root.go               # cobra root, flags globales (--json, --db), inicializa DB
│   ├── list.go               # subcomando list (create/ls/rename/rm)
│   ├── task.go               # add, ls, show, edit, done, undone, rm
│   └── recur.go              # recur daily/weekly/monthly/rm
├── internal/
│   ├── config/config.go      # carga .env, first-run wizard
│   ├── db/db.go              # abre SQLite, aplica PRAGMAs, ejecuta migración
│   ├── models/models.go      # structs List y Task
│   ├── repo/
│   │   ├── list.go           # CRUD de listas
│   │   └── task.go           # CRUD de tareas + recurrencia
│   └── output/output.go      # imprime en tabla o JSON (controlado por output.JSONMode)
└── Makefile
```

## Arquitectura

- **Entrada:** `cmd/root.go` usa `PersistentPreRunE` para abrir la DB antes de cualquier subcomando. La variable global `cmd.DB *sql.DB` se pasa directamente a las funciones de `repo`.
- **Config:** `internal/config` busca `DB_PATH` en el fichero `.env` específico de plataforma. Si no existe, lanza un wizard interactivo que pregunta la ruta y crea el fichero.
- **DB:** `internal/db` abre SQLite con WAL + busy_timeout y ejecuta el `CREATE TABLE IF NOT EXISTS` en cada arranque (migración idempotente).
- **Repo:** funciones puras que reciben `*sql.DB` y devuelven modelos o error. Sin estado global en este paquete.
- **Output:** `output.JSONMode` es un bool global que se activa con `--json`. Todas las funciones de impresión lo comprueban.

## Convenciones de código

- Fechas: `YYYY-MM-DD` como `string` (puntero `*string` cuando es nullable).
- Horas: `HH:MM` como `string`.
- IDs de DB: `int64`.
- Flags opcionales: se comprueban con `cmd.Flags().Changed("flag")` antes de asignar al struct de input, para distinguir "no dado" de "dado con valor vacío".
- `TaskPatch` para edición parcial (solo actualiza los campos no-nil).
- Driver SQLite: `modernc.org/sqlite` (pure Go, CGO_ENABLED=0). El driver se registra con el nombre `"sqlite"`.

## Config paths

| Plataforma | Config             | DB por defecto                               |
|------------|--------------------|----------------------------------------------|
| macOS      | `~/.dtasks/.env`   | `~/Library/Application Support/dtasks/tasks.db` |
| Linux      | `~/.config/dtasks/.env` (respeta `$XDG_CONFIG_HOME`) | `~/.local/share/dtasks/tasks.db` (respeta `$XDG_DATA_HOME`) |

## Tests existentes

Los tests están en `internal/` junto al paquete que prueban:

| Fichero | Qué cubre |
|---|---|
| `internal/db/db_test.go` | `Open`, creación de dirs, migración de esquema |
| `internal/config/config_test.go` | `DefaultDBPath`, `EnvFilePath`, `Load` desde `.env` |
| `internal/output/output_test.go` | Salida tabla y JSON para lists/tasks/success/error |
| `internal/repo/repo_test.go` | CRUD completo listas y tareas, filtros, done/undone, subtareas, recurrencia, cascade delete |

Los tests de `repo` y `db` crean una DB SQLite temporal con `os.CreateTemp` y la limpian con `t.Cleanup`.

## Dependencias clave

| Módulo | Uso |
|---|---|
| `modernc.org/sqlite v1.29.0` | Driver SQLite pure-Go |
| `github.com/spf13/cobra v1.8.0` | Framework CLI |
| `github.com/joho/godotenv v1.5.1` | Lectura de ficheros `.env` |

## Qué falta (v1 out of scope)

- Notificaciones del sistema
- Sync / backend cloud
- Tags, prioridades, adjuntos
- La lógica de "crear siguiente ocurrencia" para tareas recurrentes (los campos de recurrencia se guardan pero no hay scheduler)
