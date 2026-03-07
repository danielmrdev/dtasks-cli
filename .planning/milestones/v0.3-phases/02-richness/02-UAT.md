---
status: complete
phase: 02-richness
source: 02-01-SUMMARY.md, 02-02-SUMMARY.md, 02-03-SUMMARY.md, 02-04-SUMMARY.md, 02-05-SUMMARY.md
started: 2026-03-06T10:35:00Z
updated: 2026-03-06T10:45:00Z
---

## Current Test

[testing complete]

## Tests

### 1. Cold Start Smoke Test
expected: Run `make build` y después `dist/dtasks task ls` (o cualquier lista). El binario compila sin errores. El comando devuelve output (aunque esté vacío) sin crashes. La migración de la columna `priority` se aplica silenciosamente.
result: pass

### 2. Add task with priority
expected: Run `dist/dtasks task add "Test priority" --priority high` (en cualquier lista). Comando exitoso. Running `dist/dtasks task ls` muestra `!` en la columna PRIO para esa tarea. Probar también `--priority medium` (muestra `~`) y `--priority low` (muestra `-`).
result: pass

### 3. Edit task priority
expected: Run `dist/dtasks task edit <id> --priority low`. La prioridad se actualiza a low. Running `dist/dtasks task ls` muestra `-` en PRIO. Luego run `dist/dtasks task edit <id> --priority ""` — la prioridad se borra, PRIO queda en blanco.
result: pass

### 4. Task list shows PRIO column
expected: Run `dist/dtasks task ls`. La tabla tiene columna PRIO. Tareas con `high` muestran `!`, `medium` muestra `~`, `low` muestra `-`, sin prioridad muestra espacio en blanco.
result: pass

### 5. Sort tasks by priority
expected: Con tareas de distintas prioridades, run `dist/dtasks task ls --sort priority`. Orden: high primero, luego medium, luego low, luego sin prioridad.
result: pass

### 6. Bulk delete completed — dry-run preview
expected: Marcar varias tareas como done. Run `dist/dtasks task rm --completed --dry-run`. El comando muestra cuántas tareas completadas se borrarían (un número > 0 si hay tareas completadas) SIN pedir confirmación y SIN borrar nada. Running `dist/dtasks task ls` confirma que las tareas siguen existiendo.
result: pass

### 7. Bulk delete completed — with confirmation
expected: Run `dist/dtasks task rm --completed`. El comando muestra el recuento de tareas a borrar y pide confirmación. Escribir `y` las borra. Running `dist/dtasks task ls` confirma que desaparecieron. Con `--yes` salta el prompt y borra directamente sin preguntar.
result: pass

### 8. Bulk delete scoped to a list
expected: Con tareas completadas en múltiples listas, run `dist/dtasks task rm --completed --list <nombre>`. Solo se borran las tareas completadas de esa lista específica. Las tareas completadas de otras listas siguen existiendo.
result: pass

### 9. Stats command — table output
expected: Run `dist/dtasks stats`. Output muestra tabla con columnas: LIST, Total, Pending, Done, Done%. Cada lista aparece como una fila con conteos correctos. Aparece una fila de totales al final.
result: pass

### 10. Stats command — JSON output
expected: Run `dist/dtasks stats --json`. Output es JSON válido con datos de stats (total, pending, done por lista). Sin formato de tabla, JSON puro.
result: pass

## Summary

total: 10
passed: 10
issues: 0
pending: 0
skipped: 0

## Gaps

[none yet]
