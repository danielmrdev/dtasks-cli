package skill

import (
	"fmt"
	"io"
)

// DetectClaude reports whether Claude Code appears to be installed.
// homeDir is the user's home directory (injected for testability).
// It checks ~/.claude/, ~/.config/claude/, and the claude binary in PATH.
func DetectClaude(homeDir string) bool {
	return false
}

// InstallSkill writes content to <homeDir>/.claude/skills/dtasks-cli/SKILL.md.
// If the file already exists it is overwritten silently.
func InstallSkill(homeDir string, content []byte) error {
	return fmt.Errorf("not implemented")
}

// PromptAndInstall checks whether in is a real TTY.
// If it is not a TTY, it calls InstallSkill directly without prompting.
// If it is a TTY, it prompts "Install dtasks skill for Claude Code? [y/N] " and installs on y/Y.
func PromptAndInstall(homeDir string, content []byte, in io.Reader, out io.Writer) error {
	return fmt.Errorf("not implemented")
}
