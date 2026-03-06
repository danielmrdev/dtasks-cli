package updater

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"runtime"
	"strings"
	"testing"
)

// newMockGHAPI creates a test HTTP server that responds with a GitHub releases JSON payload.
func newMockGHAPI(t *testing.T, tagName string) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"tag_name": tagName})
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestFetchLatestVersion(t *testing.T) {
	srv := newMockGHAPI(t, "v0.3.0")
	// Override the API base URL so the function hits the mock server.
	orig := GHAPIBase
	GHAPIBase = srv.URL
	t.Cleanup(func() { GHAPIBase = orig })

	got, err := FetchLatestVersion("owner/repo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "v0.3.0" {
		t.Errorf("got %q, want %q", got, "v0.3.0")
	}
}

func TestFetchLatestVersion_NetworkError(t *testing.T) {
	orig := GHAPIBase
	GHAPIBase = "http://127.0.0.1:0" // nothing listening
	t.Cleanup(func() { GHAPIBase = orig })

	_, err := FetchLatestVersion("owner/repo")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestAssetName(t *testing.T) {
	name, err := AssetName()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasPrefix(name, "dtasks-") {
		t.Errorf("asset name %q should start with 'dtasks-'", name)
	}

	// Verify OS part
	switch runtime.GOOS {
	case "darwin":
		if !strings.Contains(name, "macos") {
			t.Errorf("darwin: expected 'macos' in %q", name)
		}
	case "linux":
		if !strings.Contains(name, "linux") {
			t.Errorf("linux: expected 'linux' in %q", name)
		}
	case "windows":
		if !strings.Contains(name, "windows") {
			t.Errorf("windows: expected 'windows' in %q", name)
		}
		if !strings.HasSuffix(name, ".exe") {
			t.Errorf("windows: expected .exe suffix in %q", name)
		}
	}

	// Verify arch part
	switch runtime.GOARCH {
	case "amd64":
		if !strings.Contains(name, "amd64") {
			t.Errorf("amd64: expected 'amd64' in %q", name)
		}
	case "arm64":
		if !strings.Contains(name, "arm64") {
			t.Errorf("arm64: expected 'arm64' in %q", name)
		}
	}
}

func TestAtomicReplace(t *testing.T) {
	dir := t.TempDir()
	exePath := dir + "/dtasks"

	// Write initial binary content.
	if err := os.WriteFile(exePath, []byte("old content"), 0755); err != nil {
		t.Fatalf("setup: %v", err)
	}

	newContent := []byte("new binary content")

	// Mock HTTP server serving the new binary.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(newContent)
	}))
	t.Cleanup(srv.Close)

	if err := DownloadAndReplace(srv.URL, exePath); err != nil {
		t.Fatalf("DownloadAndReplace: %v", err)
	}

	got, err := os.ReadFile(exePath)
	if err != nil {
		t.Fatalf("read after replace: %v", err)
	}
	if string(got) != string(newContent) {
		t.Errorf("content after replace: got %q, want %q", got, newContent)
	}

	// Verify the file is executable.
	info, err := os.Stat(exePath)
	if err != nil {
		t.Fatalf("stat after replace: %v", err)
	}
	if info.Mode()&0111 == 0 {
		t.Errorf("file not executable after replace, mode: %v", info.Mode())
	}
}

func TestAtomicReplace_PermissionDenied(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("running as root — permission test not meaningful")
	}

	// Create a read-only directory so CreateTemp fails.
	dir := t.TempDir()
	if err := os.Chmod(dir, 0555); err != nil {
		t.Fatalf("chmod: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(dir, 0755) })

	exePath := dir + "/dtasks"

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, "content")
	}))
	t.Cleanup(srv.Close)

	err := DownloadAndReplace(srv.URL, exePath)
	if err == nil {
		t.Fatal("expected error for read-only dir, got nil")
	}
	if !strings.Contains(err.Error(), "create temp") {
		t.Errorf("error %q should contain 'create temp'", err.Error())
	}
}
