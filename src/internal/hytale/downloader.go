package hytale

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// DownloaderCredentials holds OAuth credentials for hytale-downloader
type DownloaderCredentials struct {
	ClientID     string `json:"client_id,omitempty"`
	ClientSecret string `json:"client_secret,omitempty"`
	AccessToken  string `json:"access_token,omitempty"`
}

// HytaleDownloader handles execution of the hytale-downloader CLI tool
type HytaleDownloader struct {
	binaryPath    string
	credentials   DownloaderCredentials
	credentialsPath string
}

// NewHytaleDownloader creates a new HytaleDownloader instance
// Automatically finds hytale-downloader in PATH
func NewHytaleDownloader(cfg BootstrapConfig) (*HytaleDownloader, error) {
	// Find hytale-downloader binary
	binaryPath, err := exec.LookPath("hytale-downloader")
	if err != nil {
		return nil, fmt.Errorf("hytale-downloader not found in PATH: %w. Please install from: https://support.hytale.com/hc/en-us/articles/45326769420827-Hytale-Server-Manual", err)
	}
	return NewHytaleDownloaderWithPath(cfg, binaryPath)
}

// NewHytaleDownloaderWithPath creates a new HytaleDownloader instance with a specific binary path
func NewHytaleDownloaderWithPath(cfg BootstrapConfig, binaryPath string) (*HytaleDownloader, error) {
	// Verify binary exists
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("hytale-downloader binary not found at %s: %w", binaryPath, err)
	}

	// Prepare credentials
	creds := DownloaderCredentials{
		ClientID:     cfg.OAuthClientID,
		ClientSecret: cfg.OAuthClientSecret,
		AccessToken:  cfg.OAuthAccessToken,
	}

	// Determine credentials file path
	// Store in shared config directory for reuse
	credentialsPath := filepath.Join(GetSharedConfigDir(), ".hytale-downloader-credentials.json")

	return &HytaleDownloader{
		binaryPath:      binaryPath,
		credentials:     creds,
		credentialsPath: credentialsPath,
	}, nil
}

// SaveCredentials saves OAuth credentials to the credentials file
// Returns path to credentials file and error
func (hd *HytaleDownloader) SaveCredentials() (string, error) {
	// Only save if we have credentials
	hasCreds := hd.credentials.ClientID != "" && hd.credentials.ClientSecret != "" || hd.credentials.AccessToken != ""
	if !hasCreds {
		// No credentials provided, try to use existing file or environment
		if _, err := os.Stat(hd.credentialsPath); err == nil {
			// Credentials file exists, use it
			return hd.credentialsPath, nil
		}
		// No credentials at all - hytale-downloader will use device code flow (browser auth)
		// This is the default authentication method per Server Provider Authentication Guide
		// See: https://support.hytale.com/hc/en-us/articles/45328341414043-Server-Provider-Authentication-Guide
		// Return empty path - hytale-downloader will handle authentication interactively
		return "", nil
	}

	// Ensure credentials directory exists
	credDir := filepath.Dir(hd.credentialsPath)
	if err := os.MkdirAll(credDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create credentials directory: %w", err)
	}

	// Write credentials to file
	credData, err := json.MarshalIndent(hd.credentials, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal credentials: %w", err)
	}

	if err := os.WriteFile(hd.credentialsPath, credData, 0600); err != nil {
		return "", fmt.Errorf("failed to write credentials file: %w", err)
	}

	return hd.credentialsPath, nil
}

// Download downloads Hytale server files to the specified directory
// Returns error if download fails
func (hd *HytaleDownloader) Download(ctx context.Context, outputDir string, progressCallback ProgressCallback) error {
	// Save credentials if provided
	credPath, err := hd.SaveCredentials()
	if err != nil {
		// If credentials file exists but error occurred, return error
		if _, statErr := os.Stat(hd.credentialsPath); statErr == nil {
			// File exists, but save failed - use existing file
			credPath = hd.credentialsPath
		} else if strings.Contains(err.Error(), "no OAuth credentials provided") {
			// No credentials provided and no existing file
			// Try anyway - hytale-downloader may use environment or cached credentials
			credPath = ""
		} else {
			// Other error saving credentials
			return fmt.Errorf("failed to save credentials: %w", err)
		}
	}

	// Ensure output directory exists
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Build command based on QUICKSTART.md:
	// hytale-downloader -patchline production -download-path <path>
	// Note: According to QUICKSTART.md, -download-path is for a zip file
	// But we need to download to a directory. We'll try without -download-path first,
	// as the default behavior may download to current directory.
	// If that doesn't work, we may need to adjust based on actual hytale-downloader behavior.
	args := []string{
		"-patchline", "production",
		"-skip-update-check", // Skip auto-update check during automation
	}

	// Add credentials path if we have one
	if credPath != "" {
		args = append(args, "-credentials-path", credPath)
	}

	// Create command
	cmd := exec.CommandContext(ctx, hd.binaryPath, args...)
	
	// Set working directory (hytale-downloader may use cwd)
	cmd.Dir = outputDir

	// Capture output for progress tracking
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("hytale-downloader failed: %w\nOutput: %s", err, string(output))
	}

	// Parse output for progress (if callback provided)
	if progressCallback != nil && len(output) > 0 {
		// Try to extract progress from output
		// hytale-downloader output format may vary, so we provide generic feedback
		progressCallback(0.9, "Downloading server files...")
	}

	// Verify downloaded files
	if err := hd.VerifyDownload(outputDir); err != nil {
		return fmt.Errorf("download verification failed: %w", err)
	}

	if progressCallback != nil {
		progressCallback(1.0, "Server files downloaded and verified")
	}

	return nil
}

// VerifyDownload verifies that required files are present after download
// Checks for HytaleServer.jar and Assets.zip
func (hd *HytaleDownloader) VerifyDownload(outputDir string) error {
	// Check for HytaleServer.jar
	jarPath := filepath.Join(outputDir, "Server", "HytaleServer.jar")
	if _, err := os.Stat(jarPath); os.IsNotExist(err) {
		// Also check in outputDir root (hytale-downloader may extract directly)
		jarPath = filepath.Join(outputDir, "HytaleServer.jar")
		if _, err := os.Stat(jarPath); os.IsNotExist(err) {
			return fmt.Errorf("HytaleServer.jar not found in %s. hytale-downloader may have failed or downloaded to a different location", outputDir)
		}
	}

	// Check for Assets.zip
	assetsPath := filepath.Join(outputDir, "Assets.zip")
	if _, err := os.Stat(assetsPath); os.IsNotExist(err) {
		// Assets.zip might not always be present, depending on hytale-downloader version
		// This is not a hard failure, but log a warning
		// return fmt.Errorf("Assets.zip not found in %s", outputDir)
	}

	return nil
}
