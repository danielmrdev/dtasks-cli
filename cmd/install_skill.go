package cmd

import (
	"fmt"
	"os"

	"github.com/danielmrdev/dtasks-cli/internal/skill"
	skilldata "github.com/danielmrdev/dtasks-cli/skills/dtasks-cli"
	"github.com/spf13/cobra"
)

var installSkillCmd = &cobra.Command{
	Use:   "install-skill",
	Short: "Install the dtasks skill for Claude Code",
	RunE: func(cmd *cobra.Command, args []string) error {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("resolve home dir: %w", err)
		}
		if err := skill.PromptAndInstall(homeDir, skilldata.Content, os.Stdin, cmd.OutOrStdout()); err != nil {
			return fmt.Errorf("skill install: %w", err)
		}
		return nil
	},
}
