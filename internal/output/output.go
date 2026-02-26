package output

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/danielmrdev/dtasks-cli/internal/models"
)

var JSONMode bool

// --- Lists ---

func PrintLists(lists []models.List) {
	if JSONMode {
		printJSON(map[string]any{"lists": lists})
		return
	}
	if len(lists) == 0 {
		fmt.Println("No lists found.")
		return
	}
	w := newTabWriter()
	fmt.Fprintln(w, "ID\tNAME\tCREATED")
	for _, l := range lists {
		fmt.Fprintf(w, "%d\t%s\t%s\n", l.ID, l.Name, l.CreatedAt.Format("2006-01-02"))
	}
	w.Flush()
}

func PrintList(l *models.List) {
	if JSONMode {
		printJSON(l)
		return
	}
	fmt.Printf("List #%d: %s (created %s)\n", l.ID, l.Name, l.CreatedAt.Format("2006-01-02"))
}

// --- Tasks ---

func PrintTasks(tasks []models.Task) {
	if JSONMode {
		printJSON(map[string]any{"tasks": tasks})
		return
	}
	if len(tasks) == 0 {
		fmt.Println("No tasks found.")
		return
	}
	w := newTabWriter()
	fmt.Fprintln(w, "ID\tLIST\tTITLE\tDATE\tDUE\tDONE")
	for _, t := range tasks {
		done := " "
		if t.Completed {
			done = "✓"
		}
		fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\t%s\n",
			t.ID, t.ListName, t.Title,
			formatDate(t.Date, t.Time),
			formatDate(t.DueDate, t.DueTime),
			done,
		)
	}
	w.Flush()
}

func PrintTask(t *models.Task) {
	if JSONMode {
		printJSON(t)
		return
	}
	fmt.Printf("Task #%d\n", t.ID)
	fmt.Printf("  Title    : %s\n", t.Title)
	fmt.Printf("  List     : %s (#%d)\n", t.ListName, t.ListID)
	if t.ParentTaskID != nil {
		fmt.Printf("  Parent   : #%d\n", *t.ParentTaskID)
	}
	if t.Notes != nil && *t.Notes != "" {
		fmt.Printf("  Notes    : %s\n", *t.Notes)
	}
	if t.Date != nil {
		fmt.Printf("  Date     : %s\n", formatDate(t.Date, t.Time))
	}
	if t.DueDate != nil {
		fmt.Printf("  Due      : %s\n", formatDate(t.DueDate, t.DueTime))
	}
	if t.Completed {
		comp := ""
		if t.CompletedAt != nil {
			comp = " on " + t.CompletedAt.Format("2006-01-02 15:04")
		}
		fmt.Printf("  Status   : ✓ completed%s\n", comp)
	} else {
		fmt.Printf("  Status   : pending\n")
	}
	if t.Recurring {
		fmt.Printf("  Recur    : %s\n", formatRecur(t))
	}
	fmt.Printf("  Created  : %s\n", t.CreatedAt.Format("2006-01-02"))
}

func PrintSuccess(msg string) {
	if JSONMode {
		printJSON(map[string]string{"status": "ok", "message": msg})
		return
	}
	fmt.Println(msg)
}

func PrintError(msg string) {
	if JSONMode {
		printJSON(map[string]string{"status": "error", "message": msg})
		return
	}
	fmt.Fprintln(os.Stderr, "Error: "+msg)
}

// --- Helpers ---

func newTabWriter() *tabwriter.Writer {
	return tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
}

func printJSON(v any) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(v)
}

func formatDate(date, t *string) string {
	if date == nil {
		return "-"
	}
	if t == nil {
		return *date
	}
	return *date + " " + *t
}

func formatRecur(t *models.Task) string {
	if t.RecurType == nil {
		return ""
	}
	var sb strings.Builder
	interval := t.RecurInterval
	if interval <= 0 {
		interval = 1
	}

	switch *t.RecurType {
	case "daily":
		if interval == 1 {
			sb.WriteString("every day")
		} else {
			fmt.Fprintf(&sb, "every %d days", interval)
		}
	case "weekly":
		day := ""
		if t.RecurDayOfWeek != nil {
			day = weekdayName(*t.RecurDayOfWeek)
		}
		if interval == 1 {
			fmt.Fprintf(&sb, "every week on %s", day)
		} else {
			fmt.Fprintf(&sb, "every %d weeks on %s", interval, day)
		}
	case "monthly":
		dom := ""
		if t.RecurDayOfMonth != nil {
			dom = fmt.Sprintf("day %d", *t.RecurDayOfMonth)
		}
		if interval == 1 {
			fmt.Fprintf(&sb, "every month on %s", dom)
		} else {
			fmt.Fprintf(&sb, "every %d months on %s", interval, dom)
		}
	}

	if t.RecurTime != nil {
		fmt.Fprintf(&sb, " at %s", *t.RecurTime)
	}

	if t.RecurEndsType != nil {
		switch *t.RecurEndsType {
		case "on_date":
			if t.RecurEndsDate != nil {
				fmt.Fprintf(&sb, " until %s", *t.RecurEndsDate)
			}
		case "after_n":
			if t.RecurEndsAfter != nil {
				fmt.Fprintf(&sb, " (%d/%d times)", t.RecurCount, *t.RecurEndsAfter)
			}
		case "never":
			sb.WriteString(" (no end)")
		}
	}

	if t.RecurStarts != nil {
		fmt.Fprintf(&sb, " starting %s", *t.RecurStarts)
	}

	return sb.String()
}

func weekdayName(d int) string {
	days := []string{"Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"}
	if d >= 0 && d < len(days) {
		return days[d]
	}
	return fmt.Sprintf("%d", d)
}
