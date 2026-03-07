package cmd

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/danielmrdev/dtasks-cli/internal/config"
	"github.com/danielmrdev/dtasks-cli/internal/db"
	"github.com/danielmrdev/dtasks-cli/internal/output"
	"github.com/danielmrdev/dtasks-cli/internal/repo"
	"github.com/spf13/cobra"
)

var (
	DB         *sql.DB
	jsonFlag   bool
	dbPathFlag string
)

var rootCmd = &cobra.Command{
	Use:   "dtasks",
	Short: "dtasks — CLI task and reminder manager",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		output.JSONMode = jsonFlag

		// Skip DB init for the built-in completion script generator, help, update, and install-skill.
		if isCompletionScript(cmd) || cmd.Name() == "help" || cmd.Name() == "update" || cmd.Name() == "install-skill" {
			return nil
		}

		// For shell completion queries (__complete/__completeNoDesc): open the
		// DB silently so ValidArgsFunction can return dynamic suggestions.
		// Suppress any errors — completions simply return empty on failure.
		if cmd.Name() == "__complete" || cmd.Name() == "__completeNoDesc" {
			dbPath := dbPathFlag
			if dbPath == "" {
				if _, statErr := os.Stat(config.EnvFilePath()); statErr == nil {
					if cfg, loadErr := config.Load(); loadErr == nil {
						dbPath = cfg.DBPath
					}
				}
			}
			if dbPath != "" {
				DB, _ = db.Open(dbPath)
			}
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

		if err := repo.ProcessAutocompleteTasks(DB); err != nil {
			fmt.Fprintf(os.Stderr, "warning: autocomplete processing failed: %v\n", err)
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
	rootCmd.AddCommand(findCmd)
	rootCmd.AddCommand(statsCmd)
	rootCmd.AddCommand(updateCmd)
	rootCmd.AddCommand(installSkillCmd)
}

func Execute(version string) {
	rootCmd.Version = version
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
