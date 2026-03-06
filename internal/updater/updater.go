package updater

import "fmt"

// ghAPIBase is the base URL for GitHub API requests.
// Tests override this to point at a mock server.
var ghAPIBase = "https://api.github.com"

// FetchLatestVersion queries the GitHub Releases API for the latest tag of repo.
// repo is in the form "owner/repository".
func FetchLatestVersion(repo string) (string, error) {
	return "", fmt.Errorf("not implemented")
}

// AssetName returns the release asset filename for the current OS and architecture.
// Example: "dtasks-macos-arm64", "dtasks-linux-amd64", "dtasks-windows-amd64.exe".
func AssetName() (string, error) {
	return "", fmt.Errorf("not implemented")
}

// DownloadAndReplace downloads the binary at url and atomically replaces exePath.
// The download is streamed to a temp file in the same directory as exePath.
// exePath must be writable by the current user.
func DownloadAndReplace(url, exePath string) error {
	return fmt.Errorf("not implemented")
}
