package cmd

import (
	"fmt"

	"github.com/danielmrdev/dtasks-cli/internal/output"
	"github.com/danielmrdev/dtasks-cli/internal/repo"
	"github.com/spf13/cobra"
)

var (
	findListID int64
	findRegex  bool
)

var findCmd = &cobra.Command{
	Use:   "find <keyword>",
	Short: "Search tasks by keyword",
	Long:  "Search tasks by keyword across title and notes (case-insensitive).\n\nUse --regex to treat the keyword as a regular expression.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := repo.TaskSearchOptions{
			Keyword: args[0],
			Regex:   findRegex,
		}
		if cmd.Flags().Changed("list") {
			opts.ListID = &findListID
		}

		tasks, err := repo.TaskSearch(DB, opts)
		if err != nil {
			return fmt.Errorf("search: %w", err)
		}
		output.PrintTasks(tasks)
		return nil
	},
}

func init() {
	findCmd.Flags().Int64VarP(&findListID, "list", "l", 0, "Scope to list ID")
	findCmd.Flags().BoolVar(&findRegex, "regex", false, "Treat keyword as a regex pattern")
	_ = findCmd.RegisterFlagCompletionFunc("list", completeLists)
}
