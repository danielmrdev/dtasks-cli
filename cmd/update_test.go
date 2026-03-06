//go:build ignore
// +build ignore

// update_test.go contains the TDD red-phase tests for the update subcommand.
// This file uses //go:build ignore so it compiles without errors while updateCmd
// does not yet exist. Plan 03-04 creates cmd/update.go and removes the build tag,
// activating these tests.
//
// Tests covered:
//   - TestUpdateCmd_JSON: --json output contains "current" and "latest" keys
//   - TestUpdateCmd_Help: --help exits 0 and mentions "Check for"

package cmd

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestUpdateCmd_JSON verifies that `dtasks update --json` emits a JSON object
// with at least the "current" and "latest" keys.
func TestUpdateCmd_JSON(t *testing.T) {
	// Mock the GitHub API to return a fixed latest version.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"tag_name":"v0.3.0"}`))
	}))
	t.Cleanup(srv.Close)

	// NOTE: The updater package exposes ghAPIBase for test overrides.
	// Uncomment and adjust the import when update.go is created.
	// updater.SetGHAPIBase(srv.URL)

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"update", "--json"})

	// This reference ensures the test won't compile until updateCmd exists.
	// Plan 03-04 adds updateCmd to rootCmd in cmd/update.go.
	_ = updateCmd

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
}

// TestUpdateCmd_Help verifies that `dtasks update --help` succeeds and mentions "Check for".
func TestUpdateCmd_Help(t *testing.T) {
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"update", "--help"})

	_ = updateCmd // compile check — requires cmd/update.go to exist

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("update --help failed: %v", err)
	}
	if !strings.Contains(buf.String(), "Check for") {
		t.Errorf("--help output should contain 'Check for', got:\n%s", buf.String())
	}
}
