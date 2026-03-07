package cmd

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

// compile check — requires cmd/install_skill.go to define installSkillCmd.
var _ = installSkillCmd

// TestInstallSkillCmd_NonTTY verifies that `dtasks install-skill` exits 0
// in a non-TTY environment (bytes.Buffer does not implement Fd()).
// HOME is redirected to a temp dir to avoid touching the real ~/.claude directory.
func TestInstallSkillCmd_NonTTY(t *testing.T) {
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	if err := os.Setenv("HOME", tmpDir); err != nil {
		t.Fatalf("setenv HOME: %v", err)
	}
	t.Cleanup(func() { os.Setenv("HOME", origHome) })

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetIn(bytes.NewBufferString(""))
	rootCmd.SetArgs([]string{"install-skill"})
	t.Cleanup(func() {
		rootCmd.SetOut(nil)
		rootCmd.SetErr(nil)
		rootCmd.SetIn(nil)
		rootCmd.SetArgs(nil)
	})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("install-skill non-TTY failed: %v", err)
	}
}

// TestInstallSkillCmd_Help verifies that `dtasks install-skill --help` exits 0
// and output contains "Install".
func TestInstallSkillCmd_Help(t *testing.T) {
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"install-skill", "--help"})
	t.Cleanup(func() {
		rootCmd.SetOut(nil)
		rootCmd.SetErr(nil)
		rootCmd.SetArgs(nil)
	})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("install-skill --help failed: %v", err)
	}
	if !strings.Contains(buf.String(), "Install") {
		t.Errorf("--help output should contain 'Install', got:\n%s", buf.String())
	}
}
