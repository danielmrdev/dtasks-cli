package cmd

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/danielmrdev/dtasks-cli/internal/config"
	"github.com/danielmrdev/dtasks-cli/internal/db"
	"github.com/danielmrdev/dtasks-cli/internal/output"
	"github.com/spf13/cobra"
)

var (
	DB      *sql.DB
	jsonFlag bool
	dbPathFlag string
)

var rootCmd = &cobra.Command{
	Use:   "dtasks",
	Short: "dtasks — CLI task and reminder manager",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		output.JSONMode = jsonFlag

		// Skip DB init for completion commands
		if cmd.Name() == "__complete" || cmd.Name() == "help" {
			return nil
		}

		dbPath := dbPathFlag
		if dbPath == "" {
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("config: %w", err)
			}
			dbPath = cfg.DBPath
		}

		var err error
		DB, err = db.Open(dbPath)
		if err != nil {
			return fmt.Errorf("database: %w", err)
		}
		return nil
	},
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&jsonFlag, "json", false, "Output as JSON")
	rootCmd.PersistentFlags().StringVar(&dbPathFlag, "db", "", "Override database path")

	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(lsCmd)
	rootCmd.AddCommand(showCmd)
	rootCmd.AddCommand(editCmd)
	rootCmd.AddCommand(doneCmd)
	rootCmd.AddCommand(undoneCmd)
	rootCmd.AddCommand(rmCmd)
	rootCmd.AddCommand(recurCmd)
}

func Execute(version string) {
	rootCmd.Version = version
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
