package hytale

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// downloadHytaleDownloaderWithHTTP downloads a file using http.Get as fallback
func downloadHytaleDownloaderWithHTTP(ctx context.Context, url, destPath string, progressCallback ProgressCallback) error {
	if progressCallback != nil {
		progressCallback(0.0, "Downloading via HTTP...")
	}

	// Create HTTP request with context
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	client := &http.Client{
		Timeout: 30 * time.Minute, // Long timeout for large downloads
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status: %d", resp.StatusCode)
	}

	// Create destination file
	destFile, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer destFile.Close()

	// Get content length for progress tracking
	contentLength := resp.ContentLength
	var totalBytes int64

	// Copy with progress tracking
	if contentLength > 0 && progressCallback != nil {
		buffer := make([]byte, 32*1024) // 32KB buffer
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			n, err := resp.Body.Read(buffer)
			if n > 0 {
				written, writeErr := destFile.Write(buffer[:n])
				if writeErr != nil {
					return fmt.Errorf("failed to write file: %w", writeErr)
				}
				totalBytes += int64(written)

				// Update progress
				progress := float64(totalBytes) / float64(contentLength)
				progressCallback(progress, "Downloading hytale-downloader...")
			}

			if err == io.EOF {
				break
			}
			if err != nil {
				return fmt.Errorf("failed to read response: %w", err)
			}
		}
	} else {
		// No content length or progress callback - simple copy
		_, err = io.Copy(destFile, resp.Body)
		if err != nil {
			return fmt.Errorf("failed to copy file: %w", err)
		}
	}

	if progressCallback != nil {
		progressCallback(1.0, "Download complete")
	}

	return nil
}
