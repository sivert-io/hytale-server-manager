package hytale

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

// UpdateGame downloads and updates the Hytale server files
func UpdateGame(ctx context.Context) (string, error) {
	masterDir := filepath.Join(DataDirBase, "master-install")

	// 1. Download latest server files to master-install using hytale-downloader
	// Create a minimal bootstrap config for downloader (with empty OAuth fields for now)
	// In the future, we could read saved credentials from config
	cfg := BootstrapConfig{} // Empty config - downloader will use existing credentials file or environment
	
	// Try to download using hytale-downloader
	downloader, err := NewHytaleDownloader(cfg)
	if err != nil {
		// hytale-downloader not available - check if files already exist
		jarPath := filepath.Join(masterDir, "Server", "HytaleServer.jar")
		if _, statErr := os.Stat(jarPath); os.IsNotExist(statErr) {
			return "", fmt.Errorf("hytale-downloader not found and server files missing. %v. Please install hytale-downloader or copy server files manually to %s", err, masterDir)
		}
		// Files exist, use existing files
	} else {
		// Download latest files
		if err := downloader.Download(ctx, masterDir, nil); err != nil {
			return "", fmt.Errorf("failed to download server files: %w", err)
		}
	}

	// Verify master-install has server files
	jarPath := filepath.Join(masterDir, "Server", "HytaleServer.jar")
	if _, err := os.Stat(jarPath); os.IsNotExist(err) {
		return "", fmt.Errorf("server files not found in master-install after download. Please check hytale-downloader output")
	}

	// 2. Update all server instances from master-install
	numServers := DetectNumServers()
	if numServers == 0 {
		return "", fmt.Errorf("no servers installed")
	}

	updated := 0
	for i := 1; i <= numServers; i++ {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
		}

		// Copy master-install to server (preserves universe/ and logs/)
		if err := CopyMasterToServer(ctx, i); err != nil {
			return "", fmt.Errorf("failed to update server %d: %w", i, err)
		}

		// Copy shared configs (in case they were updated)
		if err := CopySharedToServer(ctx, i); err != nil {
			return "", fmt.Errorf("failed to copy shared configs to server %d: %w", i, err)
		}

		// Note: config.json is NOT overwritten (server-specific settings preserved)
		updated++
	}

	return fmt.Sprintf("Updated %d server(s) from master-install", updated), nil
}

// UpdatePlugins updates server plugins and addons
// This updates plugins in the shared/mods directory and propagates them to all servers
func UpdatePlugins(ctx context.Context) (string, error) {
	sharedModsDir := filepath.Join(GetSharedConfigDir(), "mods")
	
	// Ensure shared mods directory exists
	if err := os.MkdirAll(sharedModsDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create shared mods directory: %w", err)
	}

	// Update Performance Saver plugin (default plugin)
	// In the future, we could check for updates and only download if newer version available
	var progressCallback ProgressCallback // Can be nil for now
	if err := InstallPerformanceSaverPlugin(ctx, progressCallback); err != nil {
		// Log warning but don't fail - plugin updates are optional
		return fmt.Sprintf("Plugins updated (warning: failed to update Performance Saver: %v)", err), nil
	}

	// Propagate updated plugins to all server instances
	numServers := DetectNumServers()
	if numServers == 0 {
		return "Plugins updated in shared directory (no servers to update)", nil
	}

	updatedCount := 0
	for i := 1; i <= numServers; i++ {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
		}

		// Copy shared configs (includes mods) to server
		if err := CopySharedToServer(ctx, i); err != nil {
			// Continue on error - log but don't fail entire operation
			fmt.Printf("Warning: Failed to copy plugins to server %d: %v\n", i, err)
			continue
		}

		updatedCount++
	}

	return fmt.Sprintf("Plugins updated: %d/%d servers updated", updatedCount, numServers), nil
}
