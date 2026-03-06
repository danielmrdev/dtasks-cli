package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/danielmrdev/dtasks-cli/internal/output"
	"github.com/danielmrdev/dtasks-cli/internal/skill"
	"github.com/danielmrdev/dtasks-cli/internal/updater"
	skilldata "github.com/danielmrdev/dtasks-cli/skills/dtasks-cli"
	"github.com/spf13/cobra"
)

// UpdateResult holds the result of an update check or download operation.
type UpdateResult struct {
	Current string `json:"current"`
	Latest  string `json:"latest"`
	Updated bool   `json:"updated"`
	Message string `json:"message,omitempty"`
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Check for and install updates from GitHub releases",
	RunE: func(cmd *cobra.Command, args []string) error {
		current := strings.TrimPrefix(rootCmd.Version, "v")

		latestTag, err := updater.FetchLatestVersion("danielmrdev/dtasks-cli")
		if err != nil {
			return fmt.Errorf("fetch latest version: %w", err)
		}
		latest := strings.TrimPrefix(latestTag, "v")

		result := UpdateResult{
			Current: current,
			Latest:  latest,
		}

		if current == latest {
			result.Updated = false
			result.Message = "already up to date"
			return emitUpdateResult(result)
		}

		asset, err := updater.AssetName()
		if err != nil {
			return fmt.Errorf("resolve asset name: %w", err)
		}

		exePath, err := os.Executable()
		if err != nil {
			return fmt.Errorf("resolve executable path: %w", err)
		}

		assetURL := fmt.Sprintf("https://github.com/danielmrdev/dtasks-cli/releases/download/v%s/%s", latest, asset)
		if err := updater.DownloadAndReplace(assetURL, exePath); err != nil {
			result.Updated = false
			result.Message = fmt.Sprintf("update failed: %s", err)
			if output.JSONMode {
				return json.NewEncoder(os.Stdout).Encode(result)
			}
			return err
		}

		homeDir, err := os.UserHomeDir()
		if err == nil {
			if skillErr := skill.PromptAndInstall(homeDir, skilldata.Content, os.Stdin, os.Stdout); skillErr != nil {
				fmt.Fprintln(os.Stderr, "warning: skill install:", skillErr)
			}
		}

		fmt.Fprintln(os.Stdout, "Run install.sh to update shell completions")

		result.Updated = true
		result.Message = fmt.Sprintf("updated to v%s", latest)
		return emitUpdateResult(result)
	},
}

func emitUpdateResult(r UpdateResult) error {
	if output.JSONMode {
		return json.NewEncoder(os.Stdout).Encode(r)
	}
	if r.Updated {
		fmt.Fprintln(os.Stdout, r.Message)
	} else {
		fmt.Fprintf(os.Stdout, "Already up to date (%s)\n", r.Current)
	}
	return nil
}
