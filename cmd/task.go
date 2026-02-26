package cmd

import (
	"fmt"
	"strconv"

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
	_ = addCmd.RegisterFlagCompletionFunc("list", completeLists)
	_ = addCmd.RegisterFlagCompletionFunc("parent", completePendingTasks)
}

// --- ls ---

var (
	lsListID   int64
	lsAll      bool
	lsDueToday bool
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
		if lsDueToday {
			opts.DueToday = true
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
	lsCmd.Flags().BoolVar(&lsDueToday, "due-today", false, "Only tasks due today or overdue")
	_ = lsCmd.RegisterFlagCompletionFunc("list", completeLists)
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
	_ = editCmd.RegisterFlagCompletionFunc("list", completeLists)
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

var rmCmd = &cobra.Command{
	Use:               "rm <id>",
	Short:             "Delete a task",
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: completeAllTasks,
	RunE: func(cmd *cobra.Command, args []string) error {
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

// --- helpers ---

func parseID(s string) (int64, error) {
	id, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid id: %q", s)
	}
	return id, nil
}
