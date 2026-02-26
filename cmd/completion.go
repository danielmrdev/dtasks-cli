package cmd

import (
	"fmt"

	"github.com/danielmrdev/dtasks-cli/internal/repo"
	"github.com/spf13/cobra"
)

// completePendingTasks returns pending task IDs for shell completion.
func completePendingTasks(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	if DB == nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	f := false
	tasks, err := repo.TaskList(DB, repo.TaskListOptions{Completed: &f})
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	comps := make([]string, 0, len(tasks))
	for _, t := range tasks {
		comps = append(comps, fmt.Sprintf("%d\t%s", t.ID, t.Title))
	}
	return comps, cobra.ShellCompDirectiveNoFileComp
}

// completeCompletedTasks returns completed task IDs for shell completion.
func completeCompletedTasks(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	if DB == nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	tr := true
	tasks, err := repo.TaskList(DB, repo.TaskListOptions{Completed: &tr})
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	comps := make([]string, 0, len(tasks))
	for _, t := range tasks {
		comps = append(comps, fmt.Sprintf("%d\t%s", t.ID, t.Title))
	}
	return comps, cobra.ShellCompDirectiveNoFileComp
}

// completeAllTasks returns all task IDs (pending + completed) for shell completion.
func completeAllTasks(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	if DB == nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	tasks, err := repo.TaskList(DB, repo.TaskListOptions{})
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	comps := make([]string, 0, len(tasks))
	for _, t := range tasks {
		comps = append(comps, fmt.Sprintf("%d\t%s", t.ID, t.Title))
	}
	return comps, cobra.ShellCompDirectiveNoFileComp
}

// completeLists returns list IDs for shell completion.
func completeLists(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	if DB == nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	lists, err := repo.ListAll(DB)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	comps := make([]string, 0, len(lists))
	for _, l := range lists {
		comps = append(comps, fmt.Sprintf("%d\t%s", l.ID, l.Name))
	}
	return comps, cobra.ShellCompDirectiveNoFileComp
}

// isCompletionScript reports whether cmd is part of the auto-generated
// "completion" subtree (e.g. "completion bash", "completion zsh", …).
func isCompletionScript(cmd *cobra.Command) bool {
	for c := cmd; c != nil; c = c.Parent() {
		if c.Name() == "completion" {
			return true
		}
	}
	return false
}
