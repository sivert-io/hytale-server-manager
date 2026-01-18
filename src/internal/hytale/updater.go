package hytale

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

const (
	// GitHubRepo is the GitHub repository for HSM
	GitHubRepo = "sivert-io/hytale-server-manager"
	// GitHubAPIBase is the base URL for GitHub API
	GitHubAPIBase = "https://api.github.com"
)

// ReleaseInfo represents GitHub release information
type ReleaseInfo struct {
	TagName    string `json:"tag_name"`
	Name       string `json:"name"`
	Body       string `json:"body"`
	PublishedAt string `json:"published_at"`
	Assets     []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
		Size               int64  `json:"size"`
	} `json:"assets"`
}

// CheckForUpdates checks if a newer version is available on GitHub
func CheckForUpdates(ctx context.Context) (*ReleaseInfo, bool, error) {
	// Get latest release from GitHub API
	apiURL := fmt.Sprintf("%s/repos/%s/releases/latest", GitHubAPIBase, GitHubRepo)
	
	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, false, fmt.Errorf("failed to create request: %w", err)
	}
	
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, false, fmt.Errorf("failed to fetch release info: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, false, fmt.Errorf("failed to fetch release info: status %d", resp.StatusCode)
	}
	
	var release ReleaseInfo
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, false, fmt.Errorf("failed to parse release info: %w", err)
	}
	
	// Compare versions
	currentVersion := GetVersion()
	latestVersion := strings.TrimPrefix(release.TagName, "v")
	
	// If current version is "dev", always consider update available
	if currentVersion == "dev" {
		return &release, true, nil
	}
	
	// Simple version comparison (assumes semantic versioning)
	isNewer := compareVersions(latestVersion, currentVersion) > 0
	
	return &release, isNewer, nil
}

// compareVersions compares two version strings
// Returns: 1 if v1 > v2, -1 if v1 < v2, 0 if equal
func compareVersions(v1, v2 string) int {
	// Remove 'v' prefix if present
	v1 = strings.TrimPrefix(v1, "v")
	v2 = strings.TrimPrefix(v2, "v")
	
	// Simple string comparison for semantic versions
	// This works for versions like "1.0.0", "1.0.1", etc.
	if v1 > v2 {
		return 1
	}
	if v1 < v2 {
		return -1
	}
	return 0
}

// DownloadUpdate downloads the latest release binary
func DownloadUpdate(ctx context.Context, release *ReleaseInfo, progressCallback ProgressCallback) (string, error) {
	// Determine architecture
	arch := runtime.GOARCH
	var assetName string
	
	switch arch {
	case "amd64":
		assetName = "hsm-linux-amd64"
	case "arm64":
		assetName = "hsm-linux-arm64"
	default:
		return "", fmt.Errorf("unsupported architecture: %s", arch)
	}
	
	// Find matching asset
	var downloadURL string
	for _, asset := range release.Assets {
		if asset.Name == assetName {
			downloadURL = asset.BrowserDownloadURL
			break
		}
	}
	
	if downloadURL == "" {
		return "", fmt.Errorf("no matching binary found for architecture %s", arch)
	}
	
	// Create temp file
	tmpFile, err := os.CreateTemp("", "hsm-update-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	
	// Download using wget or curl
	if wgetPath, err := exec.LookPath("wget"); err == nil {
		if err := downloadWithWget(ctx, wgetPath, downloadURL, tmpPath, progressCallback); err != nil {
			os.Remove(tmpPath)
			return "", err
		}
	} else if curlPath, err := exec.LookPath("curl"); err == nil {
		if err := downloadWithCurl(ctx, curlPath, downloadURL, tmpPath, progressCallback); err != nil {
			os.Remove(tmpPath)
			return "", err
		}
	} else {
		// Fallback to http.Get
		if err := downloadWithHTTP(ctx, downloadURL, tmpPath, progressCallback); err != nil {
			os.Remove(tmpPath)
			return "", err
		}
	}
	
	// Make executable
	if err := os.Chmod(tmpPath, 0755); err != nil {
		os.Remove(tmpPath)
		return "", fmt.Errorf("failed to make binary executable: %w", err)
	}
	
	return tmpPath, nil
}

// InstallUpdate installs the downloaded binary and returns a command to restart
func InstallUpdate(ctx context.Context, binaryPath string) (string, error) {
	// Check if we need sudo
	needsSudo := os.Geteuid() != 0
	
	// Install to /usr/local/bin/hsm
	installPath := "/usr/local/bin/hsm"
	
	var installCmd *exec.Cmd
	if needsSudo {
		installCmd = exec.CommandContext(ctx, "sudo", "install", "-m", "0755", binaryPath, installPath)
	} else {
		installCmd = exec.CommandContext(ctx, "install", "-m", "0755", binaryPath, installPath)
	}
	
	output, err := installCmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to install update: %w\nOutput: %s", err, string(output))
	}
	
	// Clean up temp file
	os.Remove(binaryPath)
	
	// Return command to restart
	if needsSudo {
		return "sudo hsm", nil
	}
	return "hsm", nil
}

// downloadWithWgetForUpdate downloads using wget
func downloadWithWgetForUpdate(ctx context.Context, wgetPath, url, destPath string, progressCallback ProgressCallback) error {
	if progressCallback != nil {
		progressCallback(0.0, "Downloading update...")
	}
	
	cmd := exec.CommandContext(ctx, wgetPath,
		"--progress=bar:force",
		"--output-document="+destPath,
		url,
	)
	
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}
	
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start wget: %w", err)
	}
	
	scanner := bufio.NewScanner(stderr)
	go func() {
		for scanner.Scan() {
			line := scanner.Text()
			if progressCallback != nil {
				if idx := strings.Index(line, "%"); idx > 0 {
					var percent float64
					if _, err := fmt.Sscanf(line[:idx+1], "%f%%", &percent); err == nil {
						if percent >= 0 && percent <= 100 {
							progressCallback(percent/100.0, "Downloading update...")
						}
					}
				}
			}
		}
	}()
	
	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("wget failed: %w", err)
	}
	
	if progressCallback != nil {
		progressCallback(1.0, "Download complete")
	}
	
	return nil
}

// downloadWithCurlForUpdate downloads using curl
func downloadWithCurlForUpdate(ctx context.Context, curlPath, url, destPath string, progressCallback ProgressCallback) error {
	if progressCallback != nil {
		progressCallback(0.0, "Downloading update...")
	}
	
	cmd := exec.CommandContext(ctx, curlPath,
		"--progress-bar",
		"--output", destPath,
		url,
	)
	
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}
	
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start curl: %w", err)
	}
	
	scanner := bufio.NewScanner(stderr)
	go func() {
		for scanner.Scan() {
			line := scanner.Text()
			if progressCallback != nil {
				if idx := strings.LastIndex(line, "%"); idx > 0 {
					var percent float64
					if _, err := fmt.Sscanf(line[idx-5:idx+1], "%f%%", &percent); err == nil {
						if percent >= 0 && percent <= 100 {
							progressCallback(percent/100.0, "Downloading update...")
						}
					}
				}
			}
		}
	}()
	
	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("curl failed: %w", err)
	}
	
	if progressCallback != nil {
		progressCallback(1.0, "Download complete")
	}
	
	return nil
}

// downloadWithHTTPForUpdate downloads using http.Get
func downloadWithHTTPForUpdate(ctx context.Context, url, destPath string, progressCallback ProgressCallback) error {
	if progressCallback != nil {
		progressCallback(0.0, "Downloading update...")
	}
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	
	client := &http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download: status %d", resp.StatusCode)
	}
	
	file, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()
	
	total := resp.ContentLength
	var downloaded int64
	
	buf := make([]byte, 32*1024)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		
		n, err := resp.Body.Read(buf)
		if n > 0 {
			if _, writeErr := file.Write(buf[:n]); writeErr != nil {
				return fmt.Errorf("failed to write file: %w", writeErr)
			}
			downloaded += int64(n)
			
			if progressCallback != nil && total > 0 {
				percent := float64(downloaded) / float64(total)
				progressCallback(percent, "Downloading update...")
			}
		}
		
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read response: %w", err)
		}
	}
	
	if progressCallback != nil {
		progressCallback(1.0, "Download complete")
	}
	
	return nil
}
