package hytale

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// EnsureHytaleDownloaderInstalled downloads and installs hytale-downloader if not already present
// Returns the path to the hytale-downloader binary
func EnsureHytaleDownloaderInstalled(ctx context.Context, progressCallback ProgressCallback) (string, error) {
	// First, check if hytale-downloader is already in PATH
	if binaryPath, err := exec.LookPath("hytale-downloader"); err == nil {
		if progressCallback != nil {
			progressCallback(1.0, "hytale-downloader already installed")
		}
		return binaryPath, nil
	}

	// Check if we've installed it to the system path
	if _, err := os.Stat(HytaleDownloaderBinPath); err == nil {
		if progressCallback != nil {
			progressCallback(1.0, "hytale-downloader found in system path")
		}
		return HytaleDownloaderBinPath, nil
	}

	// Need to download and install
	if progressCallback != nil {
		progressCallback(0.0, "Downloading hytale-downloader...")
	}

	// Create temporary directory for download
	tempDir, err := os.MkdirTemp("", "hytale-downloader-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	zipPath := filepath.Join(tempDir, "hytale-downloader.zip")

	// Download the zip file
	if progressCallback != nil {
		progressCallback(0.1, "Downloading hytale-downloader.zip...")
	}
	if err := downloadHytaleDownloaderZip(ctx, HytaleDownloaderURL, zipPath, progressCallback); err != nil {
		return "", fmt.Errorf("failed to download hytale-downloader: %w", err)
	}

	// Extract the binary from the zip
	if progressCallback != nil {
		progressCallback(0.7, "Extracting hytale-downloader...")
	}
	binaryName, err := extractHytaleDownloaderBinary(ctx, zipPath, tempDir)
	if err != nil {
		return "", fmt.Errorf("failed to extract hytale-downloader: %w", err)
	}

	extractedPath := filepath.Join(tempDir, binaryName)

	// Install to /usr/local/bin (requires root/sudo)
	if progressCallback != nil {
		progressCallback(0.9, "Installing hytale-downloader to /usr/local/bin...")
	}
	if err := installHytaleDownloaderBinary(extractedPath, HytaleDownloaderBinPath); err != nil {
		return "", fmt.Errorf("failed to install hytale-downloader: %w", err)
	}

	if progressCallback != nil {
		progressCallback(1.0, "hytale-downloader installed successfully")
	}

	return HytaleDownloaderBinPath, nil
}

// downloadHytaleDownloaderZip downloads the hytale-downloader.zip file
func downloadHytaleDownloaderZip(ctx context.Context, url, destPath string, progressCallback ProgressCallback) error {
	// Try wget first, then curl, then fallback to http.Get
	if wgetPath, err := exec.LookPath("wget"); err == nil {
		return downloadWithWget(ctx, wgetPath, url, destPath, progressCallback)
	}

	if curlPath, err := exec.LookPath("curl"); err == nil {
		return downloadWithCurl(ctx, curlPath, url, destPath, progressCallback)
	}

	// Fallback to http.Get (similar to plugins.go)
	return downloadWithHTTP(ctx, url, destPath, progressCallback)
}

// extractHytaleDownloaderBinary extracts the appropriate binary from the zip file
// Returns the extracted binary filename
func extractHytaleDownloaderBinary(ctx context.Context, zipPath, extractDir string) (string, error) {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return "", fmt.Errorf("failed to open zip: %w", err)
	}
	defer r.Close()

	// Determine which binary we need based on OS/architecture
	var targetBinary string
	if runtime.GOOS == "linux" && runtime.GOARCH == "amd64" {
		targetBinary = "hytale-downloader-linux-amd64"
	} else if runtime.GOOS == "windows" && runtime.GOARCH == "amd64" {
		targetBinary = "hytale-downloader-windows-amd64.exe"
	} else {
		return "", fmt.Errorf("unsupported platform: %s/%s (only linux/amd64 and windows/amd64 are supported)", runtime.GOOS+"/"+runtime.GOARCH)
	}

	// Find and extract the target binary
	var found bool
	for _, f := range r.File {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
		}

		if f.Name == targetBinary {
			// Extract the file
			rc, err := f.Open()
			if err != nil {
				return "", fmt.Errorf("failed to open file in zip: %w", err)
			}

			destPath := filepath.Join(extractDir, targetBinary)
			destFile, err := os.Create(destPath)
			if err != nil {
				rc.Close()
				return "", fmt.Errorf("failed to create destination file: %w", err)
			}

			_, err = io.Copy(destFile, rc)
			rc.Close()
			destFile.Close()

			if err != nil {
				return "", fmt.Errorf("failed to extract file: %w", err)
			}

			// Make executable (Linux/Mac)
			if runtime.GOOS != "windows" {
				if err := os.Chmod(destPath, 0755); err != nil {
					return "", fmt.Errorf("failed to make binary executable: %w", err)
				}
			}

			found = true
			break
		}
	}

	if !found {
		return "", fmt.Errorf("binary %s not found in zip file", targetBinary)
	}

	return targetBinary, nil
}

// installHytaleDownloaderBinary copies the binary to /usr/local/bin
// Requires root/sudo permissions
func installHytaleDownloaderBinary(srcPath, destPath string) error {
	// Ensure /usr/local/bin exists
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return fmt.Errorf("failed to create bin directory: %w", err)
	}

	// Copy file (requires root)
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer srcFile.Close()

	destFile, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create destination file (may need sudo): %w", err)
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, srcFile); err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	// Make executable (Linux/Mac)
	if runtime.GOOS != "windows" {
		if err := os.Chmod(destPath, 0755); err != nil {
			return fmt.Errorf("failed to make binary executable: %w", err)
		}
	}

	return nil
}
