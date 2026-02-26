package cmd

import (
	"fmt"
	"strconv"

	"github.com/danielmrdev/dtasks-cli/internal/output"
	"github.com/danielmrdev/dtasks-cli/internal/repo"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Manage lists",
}

var createColor string

var listCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new list",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var color *string
		if cmd.Flags().Changed("color") {
			color = &createColor
		}
		l, err := repo.ListCreate(DB, args[0], color)
		if err != nil {
			return err
		}
		output.PrintList(l)
		return nil
	},
}

var listLsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List all lists",
	RunE: func(cmd *cobra.Command, args []string) error {
		lists, err := repo.ListAll(DB)
		if err != nil {
			return err
		}
		output.PrintLists(lists)
		return nil
	},
}

var listRenameCmd = &cobra.Command{
	Use:               "rename <id> <new-name>",
	Short:             "Rename a list",
	Args:              cobra.ExactArgs(2),
	ValidArgsFunction: completeLists,
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid id: %s", args[0])
		}
		if err := repo.ListRename(DB, id, args[1]); err != nil {
			return err
		}
		output.PrintSuccess(fmt.Sprintf("List #%d renamed to %q", id, args[1]))
		return nil
	},
}

var listRmCmd = &cobra.Command{
	Use:               "rm <id>",
	Short:             "Delete a list (and all its tasks)",
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: completeLists,
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid id: %s", args[0])
		}
		if err := repo.ListDelete(DB, id); err != nil {
			return err
		}
		output.PrintSuccess(fmt.Sprintf("List #%d deleted", id))
		return nil
	},
}

func init() {
	listCreateCmd.Flags().StringVar(&createColor, "color", "", "hex color for the list (e.g. #22ff33)")

	listCmd.AddCommand(listCreateCmd)
	listCmd.AddCommand(listLsCmd)
	listCmd.AddCommand(listRenameCmd)
	listCmd.AddCommand(listRmCmd)
}
