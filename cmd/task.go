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
	addListID  int64
	addParent  int64
	addNotes   string
	addDate    string
	addTime    string
	addDueDate string
	addDueTime string
)

var addCmd = &cobra.Command{
	Use:   "add <title>",
	Short: "Add a new task",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
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
		if cmd.Flags().Changed("date") {
			in.Date = &addDate
		}
		if cmd.Flags().Changed("time") {
			in.Time = &addTime
		}
		if cmd.Flags().Changed("due") {
			in.DueDate = &addDueDate
		}
		if cmd.Flags().Changed("due-time") {
			in.DueTime = &addDueTime
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
	addCmd.Flags().StringVar(&addDate, "date", "", "Scheduled date (YYYY-MM-DD)")
	addCmd.Flags().StringVar(&addTime, "time", "", "Scheduled time (HH:MM)")
	addCmd.Flags().StringVar(&addDueDate, "due", "", "Due date (YYYY-MM-DD)")
	addCmd.Flags().StringVar(&addDueTime, "due-time", "", "Due time (HH:MM)")
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
}

// --- show ---

var showCmd = &cobra.Command{
	Use:   "show <id>",
	Short: "Show full task detail",
	Args:  cobra.ExactArgs(1),
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
	editTitle   string
	editNotes   string
	editDate    string
	editTime    string
	editDueDate string
	editDueTime string
	editListID  int64
)

var editCmd = &cobra.Command{
	Use:   "edit <id>",
	Short: "Edit a task",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
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
		if cmd.Flags().Changed("date") {
			p.Date = &editDate
		}
		if cmd.Flags().Changed("time") {
			p.Time = &editTime
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
	editCmd.Flags().StringVar(&editDate, "date", "", "Scheduled date (YYYY-MM-DD)")
	editCmd.Flags().StringVar(&editTime, "time", "", "Scheduled time (HH:MM)")
	editCmd.Flags().StringVar(&editDueDate, "due", "", "Due date (YYYY-MM-DD)")
	editCmd.Flags().StringVar(&editDueTime, "due-time", "", "Due time (HH:MM)")
	editCmd.Flags().Int64VarP(&editListID, "list", "l", 0, "Move to list ID")
}

// --- done / undone ---

var doneCmd = &cobra.Command{
	Use:   "done <id>",
	Short: "Mark a task as completed",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := parseID(args[0])
		if err != nil {
			return err
		}
		if err := repo.TaskDone(DB, id, true); err != nil {
			return err
		}
		output.PrintSuccess(fmt.Sprintf("Task #%d marked as done ✓", id))
		return nil
	},
}

var undoneCmd = &cobra.Command{
	Use:   "undone <id>",
	Short: "Mark a task as pending",
	Args:  cobra.ExactArgs(1),
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
	Use:   "rm <id>",
	Short: "Delete a task",
	Args:  cobra.ExactArgs(1),
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
