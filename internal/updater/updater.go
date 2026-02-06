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
	"strings"
)

const releaseURL = "https://api.github.com/repos/dulait/grit/releases/latest"

// Release represents a GitHub release.
type Release struct {
	TagName string  `json:"tag_name"`
	Assets  []Asset `json:"assets"`
}

// Asset represents a downloadable file attached to a release.
type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

// Updater checks for and applies updates from GitHub releases.
type Updater struct {
	httpClient *http.Client
	currentVer string
}

// New creates an Updater for the given current version string.
func New(currentVersion string) *Updater {
	return &Updater{
		httpClient: &http.Client{},
		currentVer: currentVersion,
	}
}

// IsDev reports whether the current build is a development build.
func (u *Updater) IsDev() bool {
	return u.currentVer == "dev"
}

// FetchLatestRelease retrieves the latest release metadata from GitHub.
func (u *Updater) FetchLatestRelease(ctx context.Context) (*Release, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, releaseURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := u.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching latest release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d from GitHub API", resp.StatusCode)
	}

	var release Release
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("decoding release: %w", err)
	}
	return &release, nil
}

// IsUpToDate reports whether the current version matches the release tag.
func (u *Updater) IsUpToDate(release *Release) bool {
	latest := strings.TrimPrefix(release.TagName, "v")
	current := strings.TrimPrefix(u.currentVer, "v")
	return current == latest
}

// FindAsset locates the correct archive asset for the current OS and architecture.
func (u *Updater) FindAsset(release *Release) (*Asset, error) {
	version := strings.TrimPrefix(release.TagName, "v")
	ext := "tar.gz"
	if runtime.GOOS == "windows" {
		ext = "zip"
	}
	want := fmt.Sprintf("grit_%s_%s_%s.%s", version, runtime.GOOS, runtime.GOARCH, ext)

	for i := range release.Assets {
		if release.Assets[i].Name == want {
			return &release.Assets[i], nil
		}
	}
	return nil, fmt.Errorf("no asset found for %s/%s (expected %s)", runtime.GOOS, runtime.GOARCH, want)
}

// DownloadAsset downloads the asset to a temporary file and returns its path.
func (u *Updater) DownloadAsset(ctx context.Context, asset *Asset) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, asset.BrowserDownloadURL, nil)
	if err != nil {
		return "", fmt.Errorf("creating download request: %w", err)
	}

	resp, err := u.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("downloading asset: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status %d downloading asset", resp.StatusCode)
	}

	tmp, err := os.CreateTemp("", "grit-update-*"+filepath.Ext(asset.Name))
	if err != nil {
		return "", fmt.Errorf("creating temp file: %w", err)
	}
	defer tmp.Close()

	if _, err := io.Copy(tmp, resp.Body); err != nil {
		os.Remove(tmp.Name())
		return "", fmt.Errorf("writing temp file: %w", err)
	}
	return tmp.Name(), nil
}

// Apply extracts the binary from the archive and replaces the running executable.
func (u *Updater) Apply(archivePath string) error {
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("locating current executable: %w", err)
	}
	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return fmt.Errorf("resolving symlinks: %w", err)
	}

	binaryName := "grit"
	if runtime.GOOS == "windows" {
		binaryName = "grit.exe"
	}

	var data []byte
	if strings.HasSuffix(archivePath, ".zip") {
		data, err = extractFromZip(archivePath, binaryName)
	} else {
		data, err = extractFromTarGz(archivePath, binaryName)
	}
	if err != nil {
		return fmt.Errorf("extracting binary: %w", err)
	}

	return replaceBinary(execPath, data)
}
