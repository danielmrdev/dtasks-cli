package cmd

import (
	"encoding/json"
	"fmt"
	"io"
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
		w := cmd.OutOrStdout()

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
			return emitUpdateResult(w, result, output.JSONMode)
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
				return json.NewEncoder(w).Encode(result)
			}
			return err
		}

		if !output.JSONMode {
			homeDir, err := os.UserHomeDir()
			if err == nil {
				if skillErr := skill.PromptAndInstall(homeDir, skilldata.Content, os.Stdin, w); skillErr != nil {
					fmt.Fprintln(cmd.ErrOrStderr(), "warning: skill install:", skillErr)
				}
			}
			fmt.Fprintln(w, "Run install.sh to update shell completions")
		}

		result.Updated = true
		result.Message = fmt.Sprintf("updated to v%s", latest)
		return emitUpdateResult(w, result, output.JSONMode)
	},
}

func emitUpdateResult(w io.Writer, r UpdateResult, useJSON bool) error {
	if useJSON {
		return json.NewEncoder(w).Encode(r)
	}
	if r.Updated {
		fmt.Fprintln(w, r.Message)
	} else {
		fmt.Fprintf(w, "Already up to date (%s)\n", r.Current)
	}
	return nil
}
