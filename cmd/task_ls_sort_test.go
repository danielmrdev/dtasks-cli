package cmd

import (
	"bytes"
	"strings"
	"testing"
)

// TestLsCmd_SortFlagIncludesPriority verifies that `dtasks ls --help` output
// contains "priority" in the --sort flag description, protecting SORT-01 from regression.
func TestLsCmd_SortFlagIncludesPriority(t *testing.T) {
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"ls", "--help"})
	t.Cleanup(func() {
		rootCmd.SetOut(nil)
		rootCmd.SetErr(nil)
		rootCmd.SetArgs(nil)
	})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("ls --help failed: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "priority") {
		t.Errorf("--sort flag description must include 'priority', got:\n%s", out)
	}
}
