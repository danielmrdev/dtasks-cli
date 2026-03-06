package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/danielmrdev/dtasks-cli/internal/output"
	"github.com/danielmrdev/dtasks-cli/internal/repo"
	"github.com/spf13/cobra"
)

// --- add ---

var (
	addListID       int64
	addParent       int64
	addNotes        string
	addDueDate      string
	addDueTime      string
	addAutocomplete bool
	addPriority     string
)

var addCmd = &cobra.Command{
	Use:   "add <title>",
	Short: "Add a new task",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if cmd.Flags().Changed("due-time") && !cmd.Flags().Changed("due") {
			return fmt.Errorf("--due-time requires --due")
		}
		in := repo.TaskInput{
			ListID: addListID,
			Title:  args[0],
		}
		if cmd.Flags().Changed("parent") {
			in.ParentTaskID = &addParent
		}
		if cmd.Flags().Changed("notes") {
			in.Notes = &addNotes
		}
		if cmd.Flags().Changed("due") {
			in.DueDate = &addDueDate
		}
		if cmd.Flags().Changed("due-time") {
			in.DueTime = &addDueTime
		}
		if cmd.Flags().Changed("autocomplete") {
			in.Autocomplete = addAutocomplete
		}
		if cmd.Flags().Changed("priority") {
			validPriorities := map[string]bool{"high": true, "medium": true, "low": true}
			if !validPriorities[addPriority] {
				return fmt.Errorf("invalid priority %q: must be high, medium, or low", addPriority)
			}
			in.Priority = &addPriority
		}

		t, err := repo.TaskCreate(DB, in)
		if err != nil {
			return err
		}
		output.PrintTask(t)
		return nil
	},
}

func init() {
	addCmd.Flags().Int64VarP(&addListID, "list", "l", 0, "List ID (required)")
	addCmd.MarkFlagRequired("list")
	addCmd.Flags().Int64Var(&addParent, "parent", 0, "Parent task ID (for subtasks)")
	addCmd.Flags().StringVarP(&addNotes, "notes", "n", "", "Notes")
	addCmd.Flags().StringVar(&addDueDate, "due", "", "Due date (YYYY-MM-DD)")
	addCmd.Flags().StringVar(&addDueTime, "due-time", "", "Due time (HH:MM, requires --due)")
	addCmd.Flags().BoolVar(&addAutocomplete, "autocomplete", false, "Auto-complete when due date passes")
	addCmd.Flags().StringVar(&addPriority, "priority", "", "Priority: high, medium, or low")
	_ = addCmd.RegisterFlagCompletionFunc("list", completeLists)
	_ = addCmd.RegisterFlagCompletionFunc("parent", completePendingTasks)
	_ = addCmd.RegisterFlagCompletionFunc("priority", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"high", "medium", "low"}, cobra.ShellCompDirectiveNoFileComp
	})
}

// --- ls ---

var (
	lsListID   int64
	lsAll      bool
	lsToday    bool
	lsOverdue  bool
	lsTomorrow bool
	lsWeek     bool
	lsSort     string
	lsReverse  bool
)

var lsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List tasks",
	Long:  "List tasks.\n\nColumn symbols:\n  DONE  ✓ = completed\n  AC    ✓ = auto-complete on due date",
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := repo.TaskListOptions{OnlyRoot: true}

		if cmd.Flags().Changed("list") {
			opts.ListID = &lsListID
		}
		if lsToday {
			opts.DueToday = true
		}
		if lsOverdue {
			opts.Overdue = true
		}
		if lsTomorrow {
			opts.DueTomorrow = true
		}
		if lsWeek {
			opts.DueWeek = true
		}
		if cmd.Flags().Changed("sort") {
			opts.SortBy = lsSort
		}
		if lsReverse {
			opts.Reverse = true
		}
		if !lsAll {
			f := false
			opts.Completed = &f
		}

		tasks, err := repo.TaskList(DB, opts)
		if err != nil {
			return err
		}
		output.PrintTasks(tasks)
		return nil
	},
}

func init() {
	lsCmd.Flags().Int64VarP(&lsListID, "list", "l", 0, "Filter by list ID")
	lsCmd.Flags().BoolVar(&lsAll, "all", false, "Include completed tasks")
	lsCmd.Flags().BoolVar(&lsToday, "today", false, "Tasks due today or earlier")
	lsCmd.Flags().BoolVar(&lsOverdue, "overdue", false, "Tasks past their due date")
	lsCmd.Flags().BoolVar(&lsTomorrow, "tomorrow", false, "Tasks due tomorrow")
	lsCmd.Flags().BoolVar(&lsWeek, "week", false, "Tasks due within the next 7 days")
	lsCmd.Flags().StringVar(&lsSort, "sort", "", "Sort by: due, created, completed")
	lsCmd.Flags().BoolVar(&lsReverse, "reverse", false, "Reverse sort order")
	_ = lsCmd.RegisterFlagCompletionFunc("list", completeLists)
	_ = lsCmd.RegisterFlagCompletionFunc("sort", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"due", "created", "completed", "priority"}, cobra.ShellCompDirectiveNoFileComp
	})
}

// --- show ---

var showCmd = &cobra.Command{
	Use:               "show <id>",
	Short:             "Show full task detail",
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: completeAllTasks,
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := parseID(args[0])
		if err != nil {
			return err
		}
		t, err := repo.TaskGet(DB, id)
		if err != nil {
			return err
		}
		output.PrintTask(t)

		// Show subtasks if any
		subtasks, err := repo.TaskList(DB, repo.TaskListOptions{ParentID: &id})
		if err != nil {
			return err
		}
		if len(subtasks) > 0 {
			fmt.Println("\nSubtasks:")
			output.PrintTasks(subtasks)
		}
		return nil
	},
}

// --- edit ---

var (
	editTitle        string
	editNotes        string
	editDueDate      string
	editDueTime      string
	editListID       int64
	editAutocomplete bool
	editPriority     string
)

var editCmd = &cobra.Command{
	Use:               "edit <id>",
	Short:             "Edit a task",
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: completePendingTasks,
	RunE: func(cmd *cobra.Command, args []string) error {
		if cmd.Flags().Changed("due-time") && !cmd.Flags().Changed("due") {
			return fmt.Errorf("--due-time requires --due")
		}
		id, err := parseID(args[0])
		if err != nil {
			return err
		}

		p := repo.TaskPatch{}
		if cmd.Flags().Changed("title") {
			p.Title = &editTitle
		}
		if cmd.Flags().Changed("notes") {
			p.Notes = &editNotes
		}
		if cmd.Flags().Changed("due") {
			p.DueDate = &editDueDate
		}
		if cmd.Flags().Changed("due-time") {
			p.DueTime = &editDueTime
		}
		if cmd.Flags().Changed("list") {
			p.ListID = &editListID
		}
		if cmd.Flags().Changed("autocomplete") {
			p.Autocomplete = &editAutocomplete
		}
		if cmd.Flags().Changed("priority") {
			if editPriority != "" {
				validPriorities := map[string]bool{"high": true, "medium": true, "low": true}
				if !validPriorities[editPriority] {
					return fmt.Errorf("invalid priority %q: must be high, medium, or low", editPriority)
				}
			}
			p.Priority = &editPriority
		}

		t, err := repo.TaskPatchFields(DB, id, p)
		if err != nil {
			return err
		}
		output.PrintTask(t)
		return nil
	},
}

func init() {
	editCmd.Flags().StringVar(&editTitle, "title", "", "New title")
	editCmd.Flags().StringVarP(&editNotes, "notes", "n", "", "New notes")
	editCmd.Flags().StringVar(&editDueDate, "due", "", "Due date (YYYY-MM-DD)")
	editCmd.Flags().StringVar(&editDueTime, "due-time", "", "Due time (HH:MM, requires --due)")
	editCmd.Flags().Int64VarP(&editListID, "list", "l", 0, "Move to list ID")
	editCmd.Flags().BoolVar(&editAutocomplete, "autocomplete", false, "Enable/disable autocomplete")
	editCmd.Flags().StringVar(&editPriority, "priority", "", `Priority: high, medium, low, or "" to clear`)
	_ = editCmd.RegisterFlagCompletionFunc("list", completeLists)
	_ = editCmd.RegisterFlagCompletionFunc("priority", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"high", "medium", "low"}, cobra.ShellCompDirectiveNoFileComp
	})
}

// --- done / undone ---

var doneCmd = &cobra.Command{
	Use:               "done <id>",
	Short:             "Mark a task as completed",
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: completePendingTasks,
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := parseID(args[0])
		if err != nil {
			return err
		}
		if err := repo.TaskDone(DB, id, true); err != nil {
			return err
		}
		output.PrintSuccess(fmt.Sprintf("Task #%d marked as done ✓", id))

		next, err := repo.TaskScheduleNext(DB, id)
		if err != nil {
			return err
		}
		if next != nil {
			fmt.Println("\nNext occurrence scheduled:")
			output.PrintTask(next)
		}
		return nil
	},
}

var undoneCmd = &cobra.Command{
	Use:               "undone <id>",
	Short:             "Mark a task as pending",
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: completeCompletedTasks,
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := parseID(args[0])
		if err != nil {
			return err
		}
		if err := repo.TaskDone(DB, id, false); err != nil {
			return err
		}
		output.PrintSuccess(fmt.Sprintf("Task #%d marked as pending", id))
		return nil
	},
}

// --- rm ---

var (
	rmCompleted bool
	rmDryRun    bool
	rmYes       bool
	rmListID    int64
)

var rmCmd = &cobra.Command{
	Use:               "rm <id>",
	Short:             "Delete a task",
	Args:              cobra.RangeArgs(0, 1),
	ValidArgsFunction: completeAllTasks,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Bulk delete path
		if cmd.Flags().Changed("completed") {
			if len(args) > 0 {
				return fmt.Errorf("--completed cannot be used with a task ID argument")
			}
			// Always use DryRun:true for the first call to get the affected task list.
			// This is the only safe way to display a count before asking for confirmation.
			// Never set DryRun:rmDryRun here — that would execute the DELETE before the
			// confirmation prompt runs, bypassing MAINT-03.
			opts := repo.DeleteCompletedOptions{
				DryRun: true,
			}
			if cmd.Flags().Changed("list") {
				opts.ListID = &rmListID
			}
			result, err := repo.TaskDeleteCompleted(DB, opts)
			if err != nil {
				return err
			}
			// --dry-run path: show preview and return without deleting
			if rmDryRun {
				if len(result.Tasks) == 0 {
					fmt.Println("No tasks would be deleted.")
					return nil
				}
				fmt.Printf("Would delete %d task(s):\n", len(result.Tasks))
				output.PrintTasks(result.Tasks)
				return nil
			}
			// Non-dry-run: require confirmation (MAINT-03)
			if !rmYes && !output.JSONMode {
				if !isTerminal(os.Stdin) {
					return fmt.Errorf("bulk delete requires --yes in non-interactive mode")
				}
				fmt.Printf("This will permanently delete %d task(s). Confirm? [y/N]: ", len(result.Tasks))
				scanner := bufio.NewScanner(os.Stdin)
				if scanner.Scan() {
					answer := strings.TrimSpace(strings.ToLower(scanner.Text()))
					if answer != "y" && answer != "yes" {
						fmt.Println("Aborted.")
						return nil
					}
				}
			}
			// Execute actual delete: second call with DryRun:false
			opts.DryRun = false
			result, err = repo.TaskDeleteCompleted(DB, opts)
			if err != nil {
				return err
			}
			output.PrintDeletedCount(result.Deleted)
			return nil
		}
		// Single task delete path (existing behavior)
		if len(args) != 1 {
			return fmt.Errorf("accepts 1 arg(s), received %d", len(args))
		}
		id, err := parseID(args[0])
		if err != nil {
			return err
		}
		if err := repo.TaskDelete(DB, id); err != nil {
			return err
		}
		output.PrintSuccess(fmt.Sprintf("Task #%d deleted", id))
		return nil
	},
}

func init() {
	rmCmd.Flags().BoolVar(&rmCompleted, "completed", false, "Bulk-delete all completed tasks")
	rmCmd.Flags().BoolVar(&rmDryRun, "dry-run", false, "Preview without deleting")
	rmCmd.Flags().BoolVar(&rmYes, "yes", false, "Skip confirmation prompt")
	rmCmd.Flags().Int64VarP(&rmListID, "list", "l", 0, "Scope bulk delete to list ID")
	_ = rmCmd.RegisterFlagCompletionFunc("list", completeLists)
}

// --- helpers ---

func parseID(s string) (int64, error) {
	id, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid id: %q", s)
	}
	return id, nil
}

func isTerminal(f *os.File) bool {
	info, err := f.Stat()
	if err != nil {
		return false
	}
	return (info.Mode() & os.ModeCharDevice) != 0
}
