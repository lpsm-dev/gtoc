package cmd

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/lpsm-dev/gtoc/internal/logger"
	"github.com/spf13/cobra"
)

// defaultGitHubAPIURL is the GitHub API endpoint used to look up the latest
// gtoc release.
const defaultGitHubAPIURL = "https://api.github.com/repos/lpsm-dev/gtoc/releases/latest"

// httpUserAgent identifies gtoc when talking to GitHub.
const httpUserAgent = "gtoc-cli"

// checksumsAssetName is the name goreleaser gives the checksums file
// published alongside each release's archives.
const checksumsAssetName = "checksums.txt"

// errNoReleases is returned when the GitHub API reports no releases exist
// for the repository (HTTP 404).
var errNoReleases = errors.New("no releases found")

// Release represents a GitHub API release response.
type Release struct {
	TagName     string         `json:"tag_name"`
	Name        string         `json:"name"`
	PublishedAt time.Time      `json:"published_at"`
	Assets      []ReleaseAsset `json:"assets"`
}

// ReleaseAsset represents a single downloadable file attached to a release.
type ReleaseAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

var (
	forceUpgrade bool
	apiEndpoint  string
)

// upgradeCmd upgrades the current gtoc installation to the latest release.
var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrade gtoc to the latest version",
	Long: `Upgrade checks GitHub for a newer version of gtoc and upgrades
the current installation if a newer version is available.`,
	RunE: runUpgrade,
}

// runUpgrade checks GitHub for a newer release and, if one is available (or
// --force is set), downloads, verifies, and installs it.
func runUpgrade(cmd *cobra.Command, args []string) error {
	logger.Info("Checking for updates", "current_version", Version)

	endpoint := apiEndpoint
	if endpoint == "" {
		endpoint = defaultGitHubAPIURL
	}

	release, err := getLatestRelease(endpoint)
	if err != nil {
		if errors.Is(err, errNoReleases) {
			logger.Warn("No releases found", "details", "development version or no releases available yet")
			fmt.Println("No official releases found. You are using a development version or no releases are available yet.")
			return nil
		}
		return fmt.Errorf("failed to check for updates: %w", err)
	}

	logger.Info("Latest version information",
		"version", release.TagName,
		"released_at", release.PublishedAt.Format("2006-01-02"))

	if !shouldUpgrade(Version, release.TagName) && !forceUpgrade {
		fmt.Println("You already have the latest version installed!")
		return nil
	}

	fmt.Printf("Upgrading from version %s to %s\n", Version, release.TagName)

	if err := downloadAndInstall(release); err != nil {
		return fmt.Errorf("failed to download and install new version: %w", err)
	}

	fmt.Println("Successfully upgraded to version", release.TagName)
	return nil
}

// getLatestRelease fetches the latest release metadata from the GitHub API.
func getLatestRelease(apiURL string) (*Release, error) {
	logger.Debug("Fetching latest release information", "url", apiURL)

	body, err := httpGetBytes(apiURL, 10*time.Second)
	if err != nil {
		return nil, err
	}

	var release Release
	if err := json.Unmarshal(body, &release); err != nil {
		return nil, fmt.Errorf("failed to parse release information: %w", err)
	}

	return &release, nil
}

// shouldUpgrade reports whether latestVersion is newer than currentVersion,
// using a numeric, segment-by-segment comparison of the dot-separated parts
// of each version (e.g. "1.2.10" > "1.2.9"). A currentVersion that cannot be
// parsed as a numeric version (including the "dev" placeholder used in local
// builds) is always considered upgradable. A latestVersion that cannot be
// parsed is treated conservatively as not upgradable.
func shouldUpgrade(currentVersion, latestVersion string) bool {
	current, ok := parseVersionSegments(currentVersion)
	if !ok {
		return true
	}

	latest, ok := parseVersionSegments(latestVersion)
	if !ok {
		return false
	}

	return compareVersionSegments(latest, current) > 0
}

// parseVersionSegments splits a semantic-version-like string (with an
// optional leading "v") into its numeric dot-separated segments. It returns
// ok=false if the string is empty or any segment is not a non-negative
// integer.
func parseVersionSegments(version string) ([]int, bool) {
	trimmed := strings.TrimPrefix(strings.TrimSpace(version), "v")
	if trimmed == "" {
		return nil, false
	}

	parts := strings.Split(trimmed, ".")
	segments := make([]int, len(parts))
	for i, part := range parts {
		n, err := strconv.Atoi(part)
		if err != nil || n < 0 {
			return nil, false
		}
		segments[i] = n
	}
	return segments, true
}

// compareVersionSegments compares two version segment slices, returning a
// positive number if a > b, negative if a < b, and 0 if equal. Missing
// trailing segments are treated as 0 (so "1.2" == "1.2.0").
func compareVersionSegments(a, b []int) int {
	length := len(a)
	if len(b) > length {
		length = len(b)
	}

	for i := 0; i < length; i++ {
		if diff := segmentAt(a, i) - segmentAt(b, i); diff != 0 {
			return diff
		}
	}
	return 0
}

// segmentAt returns segments[i], or 0 when i is out of range.
func segmentAt(segments []int, i int) int {
	if i >= len(segments) {
		return 0
	}
	return segments[i]
}

// platformOSNames maps a Go runtime.GOOS value to the name goreleaser uses
// in archive filenames (see .goreleaser.yml's archives.name_template).
var platformOSNames = map[string]string{
	"darwin":  "Darwin",
	"linux":   "Linux",
	"windows": "Windows",
}

// platformArchNames maps a Go runtime.GOARCH value to the name goreleaser
// uses in archive filenames.
var platformArchNames = map[string]string{
	"amd64": "x86_64",
	"386":   "i386",
	"arm64": "arm64",
	"arm":   "arm",
}

// assetExtension returns the archive extension goreleaser publishes for
// goos: ".zip" for Windows, ".tar.gz" for everything else.
func assetExtension(goos string) string {
	if goos == "windows" {
		return "zip"
	}
	return "tar.gz"
}

// expectedAssetName returns the goreleaser archive filename for the given
// platform, e.g. "gtoc_Darwin_arm64.tar.gz". goreleaser's name_template does
// not include the release version in the archive name.
func expectedAssetName(goos, goarch string) (string, error) {
	osName, ok := platformOSNames[goos]
	if !ok {
		return "", fmt.Errorf("unsupported operating system: %s", goos)
	}

	archName, ok := platformArchNames[goarch]
	if !ok {
		return "", fmt.Errorf("unsupported architecture: %s", goarch)
	}

	return fmt.Sprintf("gtoc_%s_%s.%s", osName, archName, assetExtension(goos)), nil
}

// selectAsset picks the release asset matching the given platform. It is a
// pure function of the release metadata, so it can be unit tested without
// any network access.
func selectAsset(release *Release, goos, goarch string) (ReleaseAsset, error) {
	name, err := expectedAssetName(goos, goarch)
	if err != nil {
		return ReleaseAsset{}, err
	}

	for _, asset := range release.Assets {
		if asset.Name == name {
			return asset, nil
		}
	}

	return ReleaseAsset{}, fmt.Errorf("no release asset found matching %q", name)
}

// downloadAndInstall downloads the release archive for the current
// platform, verifies its checksum when available, extracts the gtoc binary,
// and installs it in place of the running executable.
func downloadAndInstall(release *Release) error {
	asset, err := selectAsset(release, runtime.GOOS, runtime.GOARCH)
	if err != nil {
		return err
	}

	tempDir, err := os.MkdirTemp("", "gtoc-upgrade")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	archivePath := filepath.Join(tempDir, asset.Name)
	logger.Debug("Downloading release asset", "name", asset.Name, "url", asset.BrowserDownloadURL)
	if err := downloadFile(asset.BrowserDownloadURL, archivePath); err != nil {
		return fmt.Errorf("failed to download %s: %w", asset.Name, err)
	}

	if err := verifyChecksum(release, asset.Name, archivePath); err != nil {
		return err
	}

	binPath, err := extractBinary(archivePath, asset.Name, tempDir)
	if err != nil {
		return fmt.Errorf("failed to extract archive: %w", err)
	}

	return installBinary(binPath)
}

// verifyChecksum downloads checksums.txt (if present among the release
// assets) and verifies that archivePath's SHA-256 matches the entry for
// assetName. If no checksums.txt asset is published, it logs a warning and
// skips verification.
func verifyChecksum(release *Release, assetName, archivePath string) error {
	checksumsURL := findAssetURL(release, checksumsAssetName)
	if checksumsURL == "" {
		logger.Warn("Release has no checksums.txt asset; skipping checksum verification")
		return nil
	}

	checksumsBody, err := httpGetBytes(checksumsURL, 30*time.Second)
	if err != nil {
		return fmt.Errorf("failed to download checksums.txt: %w", err)
	}

	expected, err := findChecksum(string(checksumsBody), assetName)
	if err != nil {
		return err
	}

	actual, err := sha256File(archivePath)
	if err != nil {
		return fmt.Errorf("failed to hash downloaded archive: %w", err)
	}

	if actual != expected {
		return fmt.Errorf("checksum mismatch for %s: expected %s, got %s", assetName, expected, actual)
	}

	return nil
}

// findAssetURL returns the BrowserDownloadURL of the release asset with the
// given name, or "" if none matches.
func findAssetURL(release *Release, name string) string {
	for _, asset := range release.Assets {
		if asset.Name == name {
			return asset.BrowserDownloadURL
		}
	}
	return ""
}

// findChecksum locates the SHA-256 digest for assetName inside the contents
// of a checksums.txt file, whose lines look like "<hex digest>  <filename>".
func findChecksum(checksumsText, assetName string) (string, error) {
	for _, line := range strings.Split(checksumsText, "\n") {
		fields := strings.Fields(line)
		if len(fields) == 2 && fields[1] == assetName {
			return fields[0], nil
		}
	}
	return "", fmt.Errorf("no checksum entry found for %s", assetName)
}

// sha256File returns the lowercase hex-encoded SHA-256 digest of the file at path.
func sha256File(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

// isBinaryMember reports whether an archive entry name (by basename) is the
// gtoc executable.
func isBinaryMember(name string) bool {
	base := filepath.Base(name)
	return base == "gtoc" || base == "gtoc.exe"
}

// extractBinary extracts the gtoc executable from a downloaded archive
// (tar.gz or zip, detected from assetName's extension) into destDir, and
// returns the path to the extracted file.
func extractBinary(archivePath, assetName, destDir string) (string, error) {
	if strings.HasSuffix(assetName, ".zip") {
		return extractBinaryFromZip(archivePath, destDir)
	}
	return extractBinaryFromTarGz(archivePath, destDir)
}

// extractBinaryFromTarGz extracts the gtoc binary from a .tar.gz archive.
// Only the archive member matching isBinaryMember is written to disk; every
// other member is ignored, which also prevents zip-slip style path
// traversal since no entry path is ever joined into the destination.
func extractBinaryFromTarGz(archivePath, destDir string) (string, error) {
	f, err := os.Open(archivePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return "", fmt.Errorf("failed to open gzip stream: %w", err)
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("failed to read tar entry: %w", err)
		}

		if hdr.Typeflag != tar.TypeReg || !isBinaryMember(hdr.Name) {
			continue
		}
		return writeBinaryMember(destDir, filepath.Base(hdr.Name), tr)
	}

	return "", fmt.Errorf("no gtoc binary found in archive")
}

// extractBinaryFromZip extracts the gtoc binary from a .zip archive. As
// with extractBinaryFromTarGz, only the matching member is ever written,
// which also prevents zip-slip path traversal.
func extractBinaryFromZip(archivePath, destDir string) (string, error) {
	r, err := zip.OpenReader(archivePath)
	if err != nil {
		return "", fmt.Errorf("failed to open zip archive: %w", err)
	}
	defer r.Close()

	for _, entry := range r.File {
		if entry.FileInfo().IsDir() || !isBinaryMember(entry.Name) {
			continue
		}

		rc, err := entry.Open()
		if err != nil {
			return "", fmt.Errorf("failed to open zip entry: %w", err)
		}
		path, err := writeBinaryMember(destDir, filepath.Base(entry.Name), rc)
		rc.Close()
		if err != nil {
			return "", err
		}
		return path, nil
	}

	return "", fmt.Errorf("no gtoc binary found in archive")
}

// writeBinaryMember copies r into a new executable file named name inside
// destDir and returns its path.
func writeBinaryMember(destDir, name string, r io.Reader) (string, error) {
	path := filepath.Join(destDir, name)

	out, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return "", fmt.Errorf("failed to create extracted binary: %w", err)
	}
	defer out.Close()

	if _, err := io.Copy(out, r); err != nil {
		return "", fmt.Errorf("failed to write extracted binary: %w", err)
	}

	return path, nil
}

// installBinary atomically replaces the running executable with the binary
// at newBinaryPath. It stages the replacement into a temp file in the same
// directory as the current executable (so the rename is atomic on the same
// filesystem), then falls back to a backup-rename-restore sequence for
// platforms (like Windows) where renaming over a running executable can
// fail.
func installBinary(newBinaryPath string) error {
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to determine current executable path: %w", err)
	}

	staged, err := stageBinary(execPath, newBinaryPath)
	if err != nil {
		return err
	}
	defer os.Remove(staged)

	if err := os.Rename(staged, execPath); err == nil {
		return nil
	}

	logger.Warn("Direct rename failed, falling back to backup-and-replace")
	return installViaBackup(execPath, staged)
}

// stageBinary copies newBinaryPath into a temp file in the same directory as
// execPath (required for os.Rename to be atomic) and makes it executable.
func stageBinary(execPath, newBinaryPath string) (string, error) {
	dir := filepath.Dir(execPath)

	tmp, err := os.CreateTemp(dir, ".gtoc-upgrade-*")
	if err != nil {
		return "", fmt.Errorf("failed to create staging file: %w", err)
	}
	defer tmp.Close()

	src, err := os.Open(newBinaryPath)
	if err != nil {
		os.Remove(tmp.Name())
		return "", fmt.Errorf("failed to open new binary: %w", err)
	}
	defer src.Close()

	if _, err := io.Copy(tmp, src); err != nil {
		os.Remove(tmp.Name())
		return "", fmt.Errorf("failed to stage new binary: %w", err)
	}

	if err := os.Chmod(tmp.Name(), 0755); err != nil {
		os.Remove(tmp.Name())
		return "", fmt.Errorf("failed to make staged binary executable: %w", err)
	}

	return tmp.Name(), nil
}

// installViaBackup handles platforms where renaming a new binary over a
// running executable fails (notably Windows): it moves the current
// executable aside, renames the staged binary into place, and removes the
// backup, restoring it if any step fails.
func installViaBackup(execPath, staged string) error {
	backupPath := execPath + ".bak"

	if err := os.Rename(execPath, backupPath); err != nil {
		return fmt.Errorf("failed to back up current binary: %w", err)
	}

	if err := os.Rename(staged, execPath); err != nil {
		if restoreErr := os.Rename(backupPath, execPath); restoreErr != nil {
			return fmt.Errorf("failed to install new binary (%v) and failed to restore backup (%v)", err, restoreErr)
		}
		return fmt.Errorf("failed to install new binary: %w", err)
	}

	if err := os.Remove(backupPath); err != nil {
		logger.Warn("Failed to remove backup binary", "path", backupPath, "error", err)
	}

	return nil
}

// httpDo performs a GET request with a timeout and the gtoc user agent. The
// caller must close the returned response body. It returns errNoReleases
// when the server responds with HTTP 404.
func httpDo(url string, timeout time.Duration) (*http.Response, error) {
	client := &http.Client{Timeout: timeout}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("User-Agent", httpUserAgent)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch %s: %w", url, err)
	}

	if resp.StatusCode == http.StatusNotFound {
		resp.Body.Close()
		return nil, errNoReleases
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("unexpected status %d fetching %s", resp.StatusCode, url)
	}

	return resp, nil
}

// httpGetBytes performs a GET request and returns the full response body.
func httpGetBytes(url string, timeout time.Duration) ([]byte, error) {
	resp, err := httpDo(url, timeout)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

// downloadFile downloads url and streams the response body to destPath.
func downloadFile(url, destPath string) error {
	resp, err := httpDo(url, 5*time.Minute)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func init() {
	upgradeCmd.Flags().BoolVar(&forceUpgrade, "force", false, "Force upgrade even if the current version is the latest")
	upgradeCmd.Flags().StringVar(&apiEndpoint, "endpoint", "", "Specify a custom GitHub API endpoint")
}
