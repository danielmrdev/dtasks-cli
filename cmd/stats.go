package cmd

import (
	"github.com/danielmrdev/dtasks-cli/internal/output"
	"github.com/danielmrdev/dtasks-cli/internal/repo"
	"github.com/spf13/cobra"
)

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show task statistics per list",
	Long:  "Show total, pending, and done task counts per list.\n\nCounts include only root tasks (subtasks are excluded).",
	RunE: func(cmd *cobra.Command, args []string) error {
		s, err := repo.TaskStats(DB)
		if err != nil {
			return err
		}
		output.PrintStats(s)
		return nil
	},
}
