package output

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/danielmrdev/dtasks-cli/internal/models"
	"github.com/danielmrdev/dtasks-cli/internal/repo"
	"github.com/mattn/go-runewidth"
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

	headers := []string{"", "ID", "NAME", "CREATED"}
	var plain, styled [][]string
	for _, l := range lists {
		dot := " "
		styledDot := " "
		if l.Color != nil && *l.Color != "" {
			dot = "●"
			styledDot = colorDot(*l.Color)
		}
		id := strconv.FormatInt(l.ID, 10)
		created := l.CreatedAt.Format("2006-01-02")
		plain = append(plain, []string{dot, id, l.Name, created})
		styled = append(styled, []string{styledDot, id, l.Name, created})
	}
	printBorderedTable(headers, plain, styled)
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

	headers := []string{"ID", "LIST", "TITLE", "DUE", "PRIO", "✔", "AC"}
	var plain, styled [][]string
	for _, t := range tasks {
		done := " "
		if t.Completed {
			done = "✓"
		}
		ac := " "
		if t.Autocomplete {
			ac = "✓"
		}
		title := t.Title
		if t.Recurring {
			title += " ↻"
		}
		id := strconv.FormatInt(t.ID, 10)
		due := formatDate(t.DueDate, t.DueTime)
		plainList := t.ListName
		styledList := t.ListName
		if t.ListColor != nil && *t.ListColor != "" {
			plainList = "● " + t.ListName
			styledList = colorDot(*t.ListColor) + " " + t.ListName
		}
		prio := " "
		switch {
		case t.Priority != nil && *t.Priority == "high":
			prio = "!"
		case t.Priority != nil && *t.Priority == "medium":
			prio = "~"
		case t.Priority != nil && *t.Priority == "low":
			prio = "-"
		}
		plain = append(plain, []string{id, plainList, title, due, prio, done, ac})
		styled = append(styled, []string{id, styledList, title, due, prio, done, ac})
	}
	printBorderedTable(headers, plain, styled)
}

func PrintTask(t *models.Task) {
	if JSONMode {
		printJSON(t)
		return
	}
	fmt.Printf("Task #%d\n", t.ID)
	fmt.Printf("  Title       : %s\n", t.Title)
	listLabel := t.ListName
	if t.ListColor != nil && *t.ListColor != "" {
		listLabel = colorDot(*t.ListColor) + " " + t.ListName
	}
	fmt.Printf("  List        : %s (#%d)\n", listLabel, t.ListID)
	if t.ParentTaskID != nil {
		fmt.Printf("  Parent      : #%d\n", *t.ParentTaskID)
	}
	if t.Notes != nil && *t.Notes != "" {
		fmt.Printf("  Notes       : %s\n", *t.Notes)
	}
	if t.DueDate != nil {
		fmt.Printf("  Due         : %s\n", formatDate(t.DueDate, t.DueTime))
	}
	if t.Completed {
		comp := ""
		if t.CompletedAt != nil {
			comp = " on " + t.CompletedAt.Format("2006-01-02 15:04")
		}
		fmt.Printf("  Status      : ✓ completed%s\n", comp)
	} else {
		fmt.Printf("  Status      : pending\n")
	}
	if t.Autocomplete {
		fmt.Printf("  Autocomplete: ✓\n")
	}
	if t.Priority != nil && *t.Priority != "" {
		prioSymbol := map[string]string{"high": "!", "medium": "~", "low": "-"}[*t.Priority]
		fmt.Printf("  Priority    : %s (%s)\n", prioSymbol, *t.Priority)
	}
	if t.Recurring {
		fmt.Printf("  Recurring   : %s\n", formatRecur(t))
	}
	fmt.Printf("  Created     : %s\n", t.CreatedAt.Format("2006-01-02"))
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

func PrintDeletedCount(n int) {
	if JSONMode {
		printJSON(map[string]any{"deleted": n})
		return
	}
	fmt.Printf("Deleted %d task(s).\n", n)
}

func PrintStats(s *repo.StatsSummary) {
	if JSONMode {
		printJSON(s)
		return
	}
	fmt.Printf("Total: %d  Pending: %d  Done: %d  (%.1f%% complete)\n",
		s.Total, s.Pending, s.Done, s.PctDone)
	if len(s.ByList) == 0 {
		fmt.Println("No lists found.")
		return
	}
	headers := []string{"LIST", "Total", "Pending", "Done", "Done%"}
	var rows [][]string
	for _, ls := range s.ByList {
		pct := fmt.Sprintf("%.1f%%", ls.PctDone)
		rows = append(rows, []string{
			ls.ListName,
			strconv.Itoa(ls.Total),
			strconv.Itoa(ls.Pending),
			strconv.Itoa(ls.Done),
			pct,
		})
	}
	printBorderedTable(headers, rows, rows)
}

// --- Table renderer ---

// printBorderedTable renders a bordered table with bold headers.
// plain holds visible-only text (used for column width calculation).
// styled holds the same rows with optional ANSI codes (same visible width as plain).
func printBorderedTable(headers []string, plain [][]string, styled [][]string) {
	n := len(headers)
	widths := make([]int, n)
	for i, h := range headers {
		widths[i] = runeWidth(h)
	}
	for _, row := range plain {
		for i := 0; i < n && i < len(row); i++ {
			if w := runeWidth(row[i]); w > widths[i] {
				widths[i] = w
			}
		}
	}

	border := func(left, mid, right, fill string) string {
		var sb strings.Builder
		sb.WriteString(left)
		for i, w := range widths {
			sb.WriteString(strings.Repeat(fill, w+2))
			if i < n-1 {
				sb.WriteString(mid)
			}
		}
		sb.WriteString(right)
		return sb.String()
	}

	fmt.Println(border("┌", "┬", "┐", "─"))

	var hdr strings.Builder
	hdr.WriteString("│")
	for i, h := range headers {
		hdr.WriteString(" ")
		hdr.WriteString(bold(h))
		hdr.WriteString(strings.Repeat(" ", widths[i]-runeWidth(h)))
		hdr.WriteString(" │")
	}
	fmt.Println(hdr.String())

	fmt.Println(border("├", "┼", "┤", "─"))

	for ri, row := range styled {
		var line strings.Builder
		line.WriteString("│")
		for i := 0; i < n; i++ {
			styledCell := ""
			plainWidth := 0
			if i < len(row) {
				styledCell = row[i]
			}
			if ri < len(plain) && i < len(plain[ri]) {
				plainWidth = runeWidth(plain[ri][i])
			}
			line.WriteString(" ")
			line.WriteString(styledCell)
			line.WriteString(strings.Repeat(" ", widths[i]-plainWidth))
			line.WriteString(" │")
		}
		fmt.Println(line.String())
	}

	fmt.Println(border("└", "┴", "┘", "─"))
}

func runeWidth(s string) int {
	return runewidth.StringWidth(s)
}

func bold(s string) string {
	if s == "" {
		return ""
	}
	return "\033[1m" + s + "\033[0m"
}

func colorDot(hex string) string {
	r, g, b, err := hexToRGB(hex)
	if err != nil {
		return "●"
	}
	return fmt.Sprintf("\033[38;2;%d;%d;%dm●\033[0m", r, g, b)
}

func hexToRGB(hex string) (r, g, b uint8, err error) {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) != 6 {
		return 0, 0, 0, fmt.Errorf("invalid hex color: %q", hex)
	}
	v, err := strconv.ParseUint(hex, 16, 32)
	if err != nil {
		return 0, 0, 0, err
	}
	return uint8(v >> 16), uint8((v >> 8) & 0xff), uint8(v & 0xff), nil
}

// --- Helpers ---

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

	if t.DueTime != nil {
		fmt.Fprintf(&sb, " at %s", *t.DueTime)
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
