package updater

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
)

// ghAPIBase is the base URL for GitHub API requests.
// Tests override this to point at a mock server.
var ghAPIBase = "https://api.github.com"

// FetchLatestVersion queries the GitHub Releases API for the latest tag of repo.
// repo is in the form "owner/repository".
func FetchLatestVersion(repo string) (string, error) {
	url := fmt.Sprintf("%s/repos/%s/releases/latest", ghAPIBase, repo)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	req.Header.Set("User-Agent", "dtasks-cli")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("github API request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("github API: status %d", resp.StatusCode)
	}

	var payload struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}

	return payload.TagName, nil
}

// AssetName returns the release asset filename for the current OS and architecture.
// Example: "dtasks-macos-arm64", "dtasks-linux-amd64", "dtasks-windows-amd64.exe".
func AssetName() (string, error) {
	var platform string
	switch runtime.GOOS {
	case "darwin":
		platform = "macos"
	case "linux":
		platform = "linux"
	case "windows":
		platform = "windows"
	default:
		return "", fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}

	var arch string
	switch runtime.GOARCH {
	case "amd64":
		arch = "amd64"
	case "arm64":
		arch = "arm64"
	default:
		return "", fmt.Errorf("unsupported arch: %s", runtime.GOARCH)
	}

	name := fmt.Sprintf("dtasks-%s-%s", platform, arch)
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	return name, nil
}

// DownloadAndReplace downloads the binary at url and atomically replaces exePath.
// The download is streamed to a temp file in the same directory as exePath.
// exePath must be writable by the current user.
func DownloadAndReplace(url, exePath string) error {
	resp, err := http.Get(url) //nolint:noctx
	if err != nil {
		return fmt.Errorf("download: %w", err)
	}
	defer resp.Body.Close()

	dir := filepath.Dir(exePath)
	tmp, err := os.CreateTemp(dir, ".dtasks-update-*")
	if err != nil {
		return fmt.Errorf("create temp: %w", err)
	}
	tmpName := tmp.Name()
	defer func() { os.Remove(tmpName) }()

	if _, err := io.Copy(tmp, resp.Body); err != nil {
		tmp.Close()
		return fmt.Errorf("write temp: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close temp: %w", err)
	}

	if err := os.Chmod(tmpName, 0755); err != nil {
		return fmt.Errorf("chmod: %w", err)
	}

	if err := os.Rename(tmpName, exePath); err != nil {
		return fmt.Errorf("rename: %w", err)
	}
	return nil
}
