package hytale

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// CopyFile copies a file from src to dst
func CopyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer sourceFile.Close()

	// Create destination directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	destFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, sourceFile); err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	// Copy file permissions
	sourceInfo, err := sourceFile.Stat()
	if err == nil {
		os.Chmod(dst, sourceInfo.Mode())
	}

	return nil
}

// CopyDir copies a directory recursively, excluding certain patterns
func CopyDir(ctx context.Context, src, dst string, excludeDirs []string) error {
	excludeMap := make(map[string]bool)
	for _, dir := range excludeDirs {
		excludeMap[dir] = true
	}

	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Check for cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Get relative path
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		// Skip excluded directories
		for excludeDir := range excludeMap {
			if relPath == excludeDir || filepath.HasPrefix(relPath, excludeDir+string(filepath.Separator)) {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}

		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		return CopyFile(path, dstPath)
	})
}

// CopyMasterToServer copies files from master-install to a server instance
// Excludes universe/, logs/, and server-specific configs
func CopyMasterToServer(ctx context.Context, serverNum int) error {
	masterDir := filepath.Join(DataDirBase, "master-install")
	serverDir := GetServerDir(serverNum)

	// Exclude directories that should be unique per server
	excludeDirs := []string{
		"universe",
		"logs",
		"config.json", // We'll create this separately
	}

	return CopyDir(ctx, masterDir, serverDir, excludeDirs)
}

// CopySharedToServer copies shared configs and mods to a server instance
func CopySharedToServer(ctx context.Context, serverNum int) error {
	sharedDir := GetSharedConfigDir()
	serverDir := GetServerDir(serverNum)

	// Check if shared directory exists
	if _, err := os.Stat(sharedDir); os.IsNotExist(err) {
		// Shared directory doesn't exist yet, create it with defaults
		if err := os.MkdirAll(sharedDir, 0755); err != nil {
			return fmt.Errorf("failed to create shared directory: %w", err)
		}
		// Create default directories
		os.MkdirAll(filepath.Join(sharedDir, "mods"), 0755)
		return nil
	}

	// Copy shared configs (but not config.json - that's server-specific)
	excludeDirs := []string{
		"config.json", // Server-specific, handled separately
	}

	return CopyDir(ctx, sharedDir, serverDir, excludeDirs)
}

// UpdateAllServersFromMaster copies master-install to all servers
func UpdateAllServersFromMaster(ctx context.Context) error {
	numServers := DetectNumServers()
	if numServers == 0 {
		return fmt.Errorf("no servers installed")
	}

	for i := 1; i <= numServers; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err := CopyMasterToServer(ctx, i); err != nil {
			return fmt.Errorf("failed to update server %d: %w", i, err)
		}

		// Also copy shared configs
		if err := CopySharedToServer(ctx, i); err != nil {
			return fmt.Errorf("failed to copy shared configs to server %d: %w", i, err)
		}
	}

	return nil
}
