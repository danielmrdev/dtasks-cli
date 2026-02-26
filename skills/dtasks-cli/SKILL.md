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
dtasks list create "<name>"                    # create — prints the new list with its ID
dtasks list create "<name>" --color "#rrggbb"  # create with hex color
dtasks list ls                                 # show all lists (ID + name + color dot)
dtasks list edit <id> --name "<new name>"      # rename
dtasks list edit <id> --color "#rrggbb"        # set or change color
dtasks list edit <id> --no-color               # remove color
dtasks list rm <id>                            # delete list AND all its tasks (irreversible)
```

`--name` and `--color`/`--no-color` can be combined in a single call. `--color` and `--no-color` are mutually exclusive.

**Getting a list ID:** run `dtasks list ls --json` and read `.id`.

---

## Tasks

### Create

```bash
dtasks add --list <list-id> "<title>"
dtasks add --list <list-id> "<title>" \
  --due 2026-03-01 --due-time 17:00 \
  --notes "any free text" \
  --autocomplete
```

- `--list` / `-l` is **required**.
- `--parent <task-id>` creates a subtask under the given task.
- `--due` — due date (`YYYY-MM-DD`). Drives `--due-today` filtering, autocomplete, recurrence base date, and sort order.
- `--due-time` — due time (`HH:MM`). **Requires `--due`** — error if used alone.
- `--autocomplete` — the task is automatically marked done the next time any dtasks command runs after its `due_date` has passed. If recurring, the next occurrence is created at that point.

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
dtasks edit <id> --due 2026-04-01 --due-time 17:00
dtasks edit <id> --notes "updated notes"
dtasks edit <id> --list <new-list-id>   # move to another list
dtasks edit <id> --autocomplete         # enable/disable autocomplete
dtasks done <id>    # marks done; if recurring, prints the next scheduled occurrence
dtasks undone <id>
```

Only pass the flags you want to change — omitted flags are left untouched. `--due-time` requires `--due`.

### Delete

```bash
dtasks rm <id>   # deletes task and its subtasks
```

---

## Recurrence

Recurrence stores metadata on a task. When `dtasks done <id>` is run on a recurring task, the next occurrence is **created automatically** (title, notes, `due_time`, recurrence settings, and `autocomplete` are inherited; `due_date` is advanced by the configured interval).

```bash
# Daily — every N days
dtasks recur daily <id> --every 1

# Weekly — requires --day (mon tue wed thu fri sat sun)
dtasks recur weekly <id> --every 2 --day thu

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
- `--ends YYYY-MM-DD` — no more occurrences after this date
- `--ends-after <n>` — stop after N repetitions

**Month overflow clamping:** if `due_date` is Jan 31 and recurrence is monthly, the next date is Feb 28 (last valid day of the month).

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
- List: `{"id": 1, "name": "Work", "color": "#0077ff", "created_at": "..."}` (`color` is `null` if not set)
- Task: `{"id": 42, "list_id": 1, "title": "...", "notes": "...", "due_date": "...", "due_time": "...", "completed": false, "autocomplete": false, "parent_task_id": null, "created_at": "..."}`

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
- `done` on a recurring task creates the next occurrence automatically; prints it inline.
- `autocomplete` tasks are silently completed on the next command invocation after their due date passes.
- `--due-time` without `--due` is an error.
- Dates must be `YYYY-MM-DD`; times must be `HH:MM` (24-hour).
