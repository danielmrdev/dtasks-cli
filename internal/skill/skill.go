package skill

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"golang.org/x/term"
)

// DetectClaude reports whether Claude Code appears to be installed.
// homeDir is the user's home directory (injected for testability).
// It checks ~/.claude/, ~/.config/claude/, and the claude binary in PATH.
func DetectClaude(homeDir string) bool {
	if _, err := os.Stat(filepath.Join(homeDir, ".claude")); err == nil {
		return true
	}
	if _, err := os.Stat(filepath.Join(homeDir, ".config", "claude")); err == nil {
		return true
	}
	if _, err := exec.LookPath("claude"); err == nil {
		return true
	}
	return false
}

// InstallSkill writes content to <homeDir>/.claude/skills/dtasks-cli/SKILL.md.
// If the file already exists it is overwritten silently.
func InstallSkill(homeDir string, content []byte) error {
	dir := filepath.Join(homeDir, ".claude", "skills", "dtasks-cli")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("skill: create dir: %w", err)
	}
	dest := filepath.Join(dir, "SKILL.md")
	if err := os.WriteFile(dest, content, 0644); err != nil {
		return fmt.Errorf("skill: write file: %w", err)
	}
	return nil
}

// PromptAndInstall checks whether in is a real TTY.
// If it is not a TTY, it calls InstallSkill directly without prompting.
// If it is a TTY, it prompts "Install dtasks skill for Claude Code? [y/N] " and installs on y/Y.
// If DetectClaude returns false, it returns nil immediately (graceful skip).
func PromptAndInstall(homeDir string, content []byte, in io.Reader, out io.Writer) error {
	if !DetectClaude(homeDir) {
		return nil
	}

	type fder interface {
		Fd() uintptr
	}

	if f, ok := in.(fder); ok && term.IsTerminal(int(f.Fd())) {
		fmt.Fprint(out, "Install dtasks skill for Claude Code? [y/N] ")
		reader := bufio.NewReader(in)
		line, err := reader.ReadString('\n')
		if err != nil {
			return nil
		}
		answer := strings.TrimSpace(line)
		if answer != "y" && answer != "Y" {
			return nil
		}
	}

	return InstallSkill(homeDir, content)
}
