# Skill: dtasks CLI

Use this skill whenever the user asks you to manage their tasks or lists with
dtasks. It covers every command, flag, and constraint you need to operate the
tool correctly.

---

## What dtasks is

A CLI task manager. Data lives in a local SQLite file. No daemon, no server.
All state is read and written through the `dtasks` binary.

---

## Global flags (prepend to any command)

| Flag | Effect |
|------|--------|
| `--db <path>` | Use a specific database file instead of the default |
| `--json` | Output as JSON (use this when you need to parse results) |

---

## Lists

Tasks must belong to a list. Always create a list first if none exist.

```bash
dtasks list create "<name>"        # create — prints the new list with its ID
dtasks list ls                     # show all lists (ID + name)
dtasks list rename <id> "<name>"   # rename
dtasks list rm <id>                # delete list AND all its tasks (irreversible)
```

**Getting a list ID:** run `dtasks list ls --json` and read `.id`.

---

## Tasks

### Create

```bash
dtasks add --list <list-id> "<title>"
dtasks add --list <list-id> "<title>" \
  --due 2026-03-01 \
  --due-time 10:00 \
  --notes "any free text" \
  --date 2026-02-28 \       # scheduled date (different from due date)
  --time 09:00              # scheduled time
```

- `--list` / `-l` is **required**.
- `--parent <task-id>` creates a subtask under the given task.
- Dates: `YYYY-MM-DD`. Times: `HH:MM`. Both are optional.

### Read

```bash
dtasks ls                    # pending root tasks (no subtasks)
dtasks ls --list <id>        # filter by list
dtasks ls --due-today        # tasks due today or overdue
dtasks ls --all              # include completed tasks
dtasks show <id>             # full detail + subtasks
```

### Update

```bash
dtasks edit <id> --title "<new title>"
dtasks edit <id> --due 2026-04-01
dtasks edit <id> --notes "updated notes"
dtasks edit <id> --list <new-list-id>   # move to another list
dtasks done <id>
dtasks undone <id>
```

Only pass the flags you want to change — omitted flags are left untouched.

### Delete

```bash
dtasks rm <id>   # deletes task and its subtasks
```

---

## Recurrence

Recurrence stores metadata on a task; it does **not** auto-create future
occurrences (out of scope for v1).

```bash
# Daily — every N days
dtasks recur daily <id> --every 1 --time 09:00

# Weekly — requires --day (mon tue wed thu fri sat sun)
dtasks recur weekly <id> --every 2 --day thu --time 10:00

# Monthly — requires --day (1-31)
dtasks recur monthly <id> --every 1 --day 25
dtasks recur monthly <id> --every 3 --day 1 --ends 2027-01-01
dtasks recur monthly <id> --every 1 --day 1 --ends never
dtasks recur monthly <id> --every 1 --day 1 --ends-after 12

# Remove recurrence
dtasks recur rm <id>
```

End conditions (mutually exclusive):
- `--ends never` (default)
- `--ends YYYY-MM-DD`
- `--ends-after <n>` — stop after N repetitions

---

## JSON output

Add `--json` to any command to get machine-readable output. Use this whenever
you need to extract an ID or pass data to another command.

```bash
# Get list ID by name
dtasks list ls --json | jq '.[] | select(.name=="Work") | .id'

# Get all overdue task IDs
dtasks ls --due-today --json | jq '.[].id'

# Create a task and capture its ID
id=$(dtasks add --list 1 "Fix bug" --json | jq '.id')
dtasks edit "$id" --due 2026-03-01
```

JSON shapes:
- List: `{"id": 1, "name": "Work", "created_at": "..."}`
- Task: `{"id": 42, "list_id": 1, "title": "...", "notes": "...", "date": "...", "time": "...", "due_date": "...", "due_time": "...", "completed": false, "parent_task_id": null, "created_at": "...", "updated_at": "..."}`

---

## Common workflows

### Add several tasks to a list

```bash
list_id=$(dtasks list ls --json | jq '.[] | select(.name=="Personal") | .id')
dtasks add --list "$list_id" "Buy groceries"
dtasks add --list "$list_id" "Call dentist" --due 2026-03-05
```

### Bulk-complete tasks due today

```bash
dtasks ls --due-today --json | jq '.[].id' | xargs -I{} dtasks done {}
```

### Move all tasks from one list to another

```bash
from=1; to=2
dtasks ls --list "$from" --all --json | jq '.[].id' | xargs -I{} dtasks edit {} --list "$to"
```

---

## Constraints and edge cases

- `list rm` cascades: deleting a list deletes all its tasks and their subtasks.
- `rm` on a parent task also deletes all its subtasks.
- `--due-today` includes tasks where `due_date <= today` (overdue tasks show up).
- `ls` without `--all` hides completed tasks.
- `ls` only returns root tasks (no subtasks); use `show <id>` to see subtasks.
- Recurrence is stored but does **not** trigger automatic task creation.
- Dates must be `YYYY-MM-DD`; any other format will be rejected by the DB layer.
- Times must be `HH:MM` (24-hour).
