package cmd

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/danielmrdev/dtasks-cli/internal/updater"
)

// TestUpdateCmd_JSON verifies that `dtasks update --json` emits a JSON object
// with at least the "current" and "latest" keys and updated=false when versions match.
func TestUpdateCmd_JSON(t *testing.T) {
	// Mock the GitHub API to return a fixed latest version.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"tag_name":"v0.3.0"}`))
	}))
	t.Cleanup(srv.Close)

	origBase := updater.GHAPIBase
	updater.GHAPIBase = srv.URL
	t.Cleanup(func() { updater.GHAPIBase = origBase })

	origVersion := rootCmd.Version
	rootCmd.Version = "0.3.0"
	t.Cleanup(func() { rootCmd.Version = origVersion })

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"update", "--json"})
	t.Cleanup(func() {
		rootCmd.SetOut(nil)
		rootCmd.SetErr(nil)
		rootCmd.SetArgs(nil)
	})

	_ = updateCmd // compile check — requires cmd/update.go to exist

	_ = rootCmd.Execute()

	out := buf.String()
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("output is not valid JSON: %v\noutput: %s", err, out)
	}
	if _, ok := result["current"]; !ok {
		t.Errorf("JSON output missing 'current' key: %s", out)
	}
	if _, ok := result["latest"]; !ok {
		t.Errorf("JSON output missing 'latest' key: %s", out)
	}
	if updated, _ := result["updated"].(bool); updated {
		t.Errorf("expected updated=false when versions match, got true")
	}
}

// TestUpdateCmd_Help verifies that `dtasks update --help` succeeds and mentions "Check for".
func TestUpdateCmd_Help(t *testing.T) {
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"update", "--help"})
	t.Cleanup(func() {
		rootCmd.SetOut(nil)
		rootCmd.SetErr(nil)
		rootCmd.SetArgs(nil)
	})

	_ = updateCmd // compile check — requires cmd/update.go to exist

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("update --help failed: %v", err)
	}
	if !strings.Contains(buf.String(), "Check for") {
		t.Errorf("--help output should contain 'Check for', got:\n%s", buf.String())
	}
}

// TestUpdateCmd_JSON_NoContamination verifies that `dtasks --json update` emits a single
// valid JSON object with no plain-text prefix or suffix (no stdout contamination).
func TestUpdateCmd_JSON_NoContamination(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"tag_name":"v0.3.0"}`))
	}))
	t.Cleanup(srv.Close)

	origBase := updater.GHAPIBase
	updater.GHAPIBase = srv.URL
	t.Cleanup(func() { updater.GHAPIBase = origBase })

	origVersion := rootCmd.Version
	rootCmd.Version = "0.3.0"
	t.Cleanup(func() { rootCmd.Version = origVersion })

	// Reset any lingering --help flag state from a previous test.
	if f := updateCmd.Flags().Lookup("help"); f != nil {
		_ = f.Value.Set("false")
	}

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"update", "--json"})
	t.Cleanup(func() {
		rootCmd.SetOut(nil)
		rootCmd.SetErr(nil)
		rootCmd.SetArgs(nil)
	})

	_ = rootCmd.Execute()

	out := buf.String()
	trimmed := strings.TrimSpace(out)
	if len(trimmed) == 0 || trimmed[0] != '{' {
		t.Fatalf("output does not start with '{' (no plain-text contamination expected):\n%s", out)
	}
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("output is not valid JSON: %v\noutput: %s", err, out)
	}
}

// TestUpdateCmd_AlreadyUpToDate verifies that when current == latest, updated is false.
func TestUpdateCmd_AlreadyUpToDate(t *testing.T) {
	cases := []struct {
		name    string
		current string
		latest  string
	}{
		{"same without v prefix", "0.3.0", "v0.3.0"},
		{"same with v prefix current", "v0.3.0", "v0.3.0"},
		{"same both without v", "0.3.0", "0.3.0"},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"tag_name":"v0.3.0"}`))
	}))
	t.Cleanup(srv.Close)

	origBase := updater.GHAPIBase
	updater.GHAPIBase = srv.URL
	t.Cleanup(func() { updater.GHAPIBase = origBase })

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			origVersion := rootCmd.Version
			rootCmd.Version = tc.current
			t.Cleanup(func() { rootCmd.Version = origVersion })

			// Reset any lingering --help flag state from a previous test.
			if f := updateCmd.Flags().Lookup("help"); f != nil {
				_ = f.Value.Set("false")
			}

			var buf bytes.Buffer
			rootCmd.SetOut(&buf)
			rootCmd.SetErr(&buf)
			rootCmd.SetArgs([]string{"update", "--json"})
			t.Cleanup(func() {
				rootCmd.SetOut(nil)
				rootCmd.SetErr(nil)
				rootCmd.SetArgs(nil)
			})

			_ = rootCmd.Execute()

			out := buf.String()
			var result map[string]interface{}
			if err := json.Unmarshal([]byte(out), &result); err != nil {
				t.Fatalf("output is not valid JSON: %v\noutput: %s", err, out)
			}
			if updated, _ := result["updated"].(bool); updated {
				t.Errorf("expected updated=false for current=%q latest=v0.3.0, got true", tc.current)
			}
		})
	}
}
