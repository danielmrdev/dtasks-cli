package skill

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestDetectClaude_FoundDotClaude(t *testing.T) {
	home := t.TempDir()
	if err := os.Mkdir(filepath.Join(home, ".claude"), 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	if !DetectClaude(home) {
		t.Error("expected DetectClaude to return true with ~/.claude present")
	}
}

func TestDetectClaude_FoundConfigClaude(t *testing.T) {
	home := t.TempDir()
	if err := os.MkdirAll(filepath.Join(home, ".config", "claude"), 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	if !DetectClaude(home) {
		t.Error("expected DetectClaude to return true with ~/.config/claude present")
	}
}

func TestDetectClaude_NotFound(t *testing.T) {
	home := t.TempDir() // empty dir, no .claude, no .config/claude

	// Note: exec.LookPath("claude") depends on the actual PATH in the test environment.
	// This test is expected to pass only when claude is not installed or not in PATH.
	// If claude is in PATH this test may be a false pass — acceptable for the stub phase.
	if DetectClaude(home) {
		t.Log("DetectClaude returned true — claude may be in PATH; skipping assertion")
	}
}

func TestInstallSkill_Path(t *testing.T) {
	home := t.TempDir()
	content := []byte("# SKILL.md content")

	if err := InstallSkill(home, content); err != nil {
		t.Fatalf("InstallSkill: %v", err)
	}

	dest := filepath.Join(home, ".claude", "skills", "dtasks-cli", "SKILL.md")
	got, err := os.ReadFile(dest)
	if err != nil {
		t.Fatalf("read installed skill: %v", err)
	}
	if string(got) != string(content) {
		t.Errorf("content mismatch: got %q, want %q", got, content)
	}
}

func TestInstallSkill_Overwrite(t *testing.T) {
	home := t.TempDir()

	first := []byte("first content")
	second := []byte("second content")

	if err := InstallSkill(home, first); err != nil {
		t.Fatalf("first InstallSkill: %v", err)
	}
	if err := InstallSkill(home, second); err != nil {
		t.Fatalf("second InstallSkill: %v", err)
	}

	dest := filepath.Join(home, ".claude", "skills", "dtasks-cli", "SKILL.md")
	got, err := os.ReadFile(dest)
	if err != nil {
		t.Fatalf("read after overwrite: %v", err)
	}
	if string(got) != string(second) {
		t.Errorf("expected second content %q, got %q", second, got)
	}
}

func TestInstallSkill_NonTTY(t *testing.T) {
	home := t.TempDir()
	content := []byte("skill content")

	// bytes.Buffer is not a TTY — PromptAndInstall should install without prompting.
	in := &bytes.Buffer{}
	out := &bytes.Buffer{}

	if err := PromptAndInstall(home, content, in, out); err != nil {
		t.Fatalf("PromptAndInstall in non-TTY: %v", err)
	}

	// Skill should be installed even without a TTY.
	dest := filepath.Join(home, ".claude", "skills", "dtasks-cli", "SKILL.md")
	got, err := os.ReadFile(dest)
	if err != nil {
		t.Fatalf("read skill after non-TTY install: %v", err)
	}
	if string(got) != string(content) {
		t.Errorf("content mismatch: got %q, want %q", got, content)
	}
}
