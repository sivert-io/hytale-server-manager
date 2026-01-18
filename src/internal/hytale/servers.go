package hytale

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

// AddServerInstanceWithContext adds a new server instance
func AddServerInstanceWithContext(ctx context.Context, numServers int, basePort int) error {
	newServerNum := numServers + 1
	dataDir := fmt.Sprintf("%s/server-%d", DataDirBase, newServerNum)

	// Create server directory
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return fmt.Errorf("failed to create server directory: %w", err)
	}

	// TODO: Copy config files, create server config, etc.
	// This is a placeholder - actual implementation would:
	// - Copy base configs from master-install or server-1
	// - Update port numbers
	// - Update server.cfg/properties.json
	// - Create symlinks if needed

	return nil
}

// RemoveLastServerInstance removes the last server instance
func RemoveLastServerInstance(numServers int) error {
	if numServers <= 1 {
		return fmt.Errorf("cannot remove last server")
	}

	lastServer := numServers
	dataDir := fmt.Sprintf("%s/server-%d", DataDirBase, lastServer)

	// TODO: Stop server if running
	// TODO: Remove directory
	if err := os.RemoveAll(dataDir); err != nil {
		return fmt.Errorf("failed to remove server directory: %w", err)
	}

	return nil
}

// DetectNumServers detects how many server instances exist
func DetectNumServers() int {
	count := 0
	for i := 1; ; i++ {
		dataDir := fmt.Sprintf("%s/server-%d", DataDirBase, i)
		if _, err := os.Stat(dataDir); os.IsNotExist(err) {
			break
		}
		count++
	}
	return count
}

// ServerExists checks if a server instance exists
func ServerExists(serverNum int) bool {
	dataDir := fmt.Sprintf("%s/server-%d", DataDirBase, serverNum)
	_, err := os.Stat(dataDir)
	return !os.IsNotExist(err)
}

// GetServerJarPath returns the path to the server JAR for a server instance
func GetServerJarPath(serverNum int) string {
	// Servers share the same JAR from master-install or server-1
	return filepath.Join(DataDirBase, "master-install", "Server", "HytaleServer.jar")
}
