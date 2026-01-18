package hytale

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

// BootstrapWithContext performs the initial server installation and setup
func BootstrapWithContext(ctx context.Context, cfg BootstrapConfig) (string, error) {
	return BootstrapWithContextAndProgress(ctx, cfg, nil)
}

// BootstrapWithContextAndProgress performs the initial server installation and setup with progress tracking
// If progressCallback is provided, it will be called with progress updates (0.0 to 1.0)
func BootstrapWithContextAndProgress(ctx context.Context, cfg BootstrapConfig, progressCallback ProgressCallback) (string, error) {
	// 1. Create base directory structure
	if progressCallback != nil {
		progressCallback(0.0, "Creating base directories...")
	}
	if err := os.MkdirAll(DataDirBase, 0755); err != nil {
		return "", fmt.Errorf("failed to create base directory: %w", err)
	}

	masterDir := filepath.Join(DataDirBase, "master-install")
	sharedDir := GetSharedConfigDir()

	// 2. Create master-install and shared directories
	if progressCallback != nil {
		progressCallback(0.02, "Creating master-install and shared directories...")
	}
	if err := os.MkdirAll(masterDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create master directory: %w", err)
	}
	if err := os.MkdirAll(sharedDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create shared directory: %w", err)
	}
	if err := os.MkdirAll(filepath.Join(sharedDir, "mods"), 0755); err != nil {
		return "", fmt.Errorf("failed to create shared mods directory: %w", err)
	}

	// 3. Validate all required dependencies
	if progressCallback != nil {
		progressCallback(0.05, "Validating system dependencies...")
	}
	if err := ValidateDependencies(); err != nil {
		return "", fmt.Errorf("dependency check failed: %w", err)
	}

	// 4. Install Performance Saver plugin (default mod)
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}
	if progressCallback != nil {
		progressCallback(0.1, "Installing Performance Saver plugin...")
	}
	if err := InstallPerformanceSaverPlugin(ctx, progressCallback); err != nil {
		// Log warning but don't fail bootstrap - plugin is optional
		// In production, you might want to make this non-fatal
		fmt.Printf("Warning: Failed to install Performance Saver plugin: %v\n", err)
	}
	
	if progressCallback != nil {
		progressCallback(0.3, "Creating server instances...")
	}

	// 5. Ensure hytale-downloader is installed (downloads from official URL if needed)
	if progressCallback != nil {
		progressCallback(0.25, "Ensuring hytale-downloader is installed...")
	}
	downloaderBinaryPath, err := EnsureHytaleDownloaderInstalled(ctx, progressCallback)
	if err != nil {
		// Failed to install hytale-downloader - check if files already exist
		jarPath := filepath.Join(masterDir, "Server", "HytaleServer.jar")
		if _, statErr := os.Stat(jarPath); os.IsNotExist(statErr) {
			// Files don't exist and hytale-downloader not available
			return "", fmt.Errorf("failed to install hytale-downloader and server files missing. %v. Please install hytale-downloader manually or copy server files to %s", err, masterDir)
		}
		// Files exist, continue with existing files
		if progressCallback != nil {
			progressCallback(0.3, "Using existing server files (hytale-downloader installation failed)")
		}
	} else {
		// hytale-downloader is available - download server files
		if progressCallback != nil {
			progressCallback(0.27, "Downloading server files via hytale-downloader...")
		}

		// Create downloader instance with the installed binary path
		downloader, err := NewHytaleDownloaderWithPath(cfg, downloaderBinaryPath)
		if err != nil {
			return "", fmt.Errorf("failed to create hytale-downloader instance: %w", err)
		}

		if err := downloader.Download(ctx, masterDir, progressCallback); err != nil {
			return "", fmt.Errorf("failed to download server files: %w", err)
		}
		if progressCallback != nil {
			progressCallback(0.3, "Server files downloaded successfully")
		}
	}

	// 6. Create server instances
	for i := 1; i <= cfg.NumServers; i++ {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
		}
		
		if progressCallback != nil {
			progress := 0.3 + (float64(i-1) / float64(cfg.NumServers)) * 0.6
			progressCallback(progress, fmt.Sprintf("Creating server %d/%d directory...", i, cfg.NumServers))
		}

		serverDir := GetServerDir(i)
		
		// Create server directory
		if err := os.MkdirAll(serverDir, 0755); err != nil {
			return "", fmt.Errorf("failed to create server %d directory: %w", i, err)
		}

		// Create server-specific directories
		os.MkdirAll(filepath.Join(serverDir, "universe"), 0755)
		os.MkdirAll(filepath.Join(serverDir, "logs"), 0755)

		// Copy from master-install (if it has content)
		if _, err := os.Stat(filepath.Join(masterDir, "Server")); err == nil {
			if progressCallback != nil {
				progress := 0.3 + (float64(i-1) / float64(cfg.NumServers)) * 0.6 + 0.05
				progressCallback(progress, fmt.Sprintf("Copying master files to server %d/%d...", i, cfg.NumServers))
			}
			if err := CopyMasterToServer(ctx, i); err != nil {
				return "", fmt.Errorf("failed to copy master files to server %d: %w", i, err)
			}
		}

		// Copy shared configs
		if progressCallback != nil {
			progress := 0.3 + (float64(i-1) / float64(cfg.NumServers)) * 0.6 + 0.08
			progressCallback(progress, fmt.Sprintf("Copying shared configs to server %d/%d...", i, cfg.NumServers))
		}
		if err := CopySharedToServer(ctx, i); err != nil {
			return "", fmt.Errorf("failed to copy shared configs to server %d: %w", i, err)
		}

		// 7. Create server-specific config.json
		if progressCallback != nil {
			progress := 0.3 + (float64(i-1) / float64(cfg.NumServers)) * 0.6 + 0.09
			progressCallback(progress, fmt.Sprintf("Creating config.json for server %d/%d...", i, cfg.NumServers))
		}
		port := cfg.BasePort + (i - 1)
		hostname := fmt.Sprintf("%s-%d", cfg.HostnamePrefix, i)
		
		if err := UpdateServerConfig(i, port, hostname, cfg.MaxPlayers, cfg.MaxViewRadius, cfg.GameMode, cfg.ServerPassword); err != nil {
			return "", fmt.Errorf("failed to create config for server %d: %w", i, err)
		}
	}

	// 8. Save backup configuration to shared config
	if progressCallback != nil {
		progressCallback(0.95, "Saving backup configuration...")
	}
	backupConfig := &BackupConfig{
		Enabled:  cfg.BackupEnabled,
		Frequency: cfg.BackupFrequency,
	}
	if err := WriteBackupConfig(backupConfig); err != nil {
		return "", fmt.Errorf("failed to save backup config: %w", err)
	}

	return fmt.Sprintf("Bootstrap completed: %d server(s) created", cfg.NumServers), nil
}
