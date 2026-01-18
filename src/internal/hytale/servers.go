package hytale

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

// AddServerInstanceWithContext adds a new server instance
// Enforces MaxServersPerLicense limit per Hytale Server Manual
func AddServerInstanceWithContext(ctx context.Context, numServers int, basePort int) error {
	newServerNum := numServers + 1
	
	// Enforce server limit per Hytale Server Manual
	// Default limit: 100 servers per game license
	if newServerNum > MaxServersPerLicense {
		return fmt.Errorf("maximum %d servers allowed per game license (Hytale Server Manual). Additional licenses or Server Provider account required for more", MaxServersPerLicense)
	}
	
	serverDir := GetServerDir(newServerNum)

	// Create server directory
	if err := os.MkdirAll(serverDir, 0755); err != nil {
		return fmt.Errorf("failed to create server directory: %w", err)
	}

	// Create server-specific directories
	os.MkdirAll(filepath.Join(serverDir, "universe"), 0755)
	os.MkdirAll(filepath.Join(serverDir, "logs"), 0755)

	// Copy from master-install
	if err := CopyMasterToServer(ctx, newServerNum); err != nil {
		return fmt.Errorf("failed to copy master files: %w", err)
	}

	// Copy shared configs
	if err := CopySharedToServer(ctx, newServerNum); err != nil {
		return fmt.Errorf("failed to copy shared configs: %w", err)
	}

	// Create server-specific config.json
	// Read config from server-1 to get defaults, or use defaults
	port := basePort + (newServerNum - 1)
	hostname := fmt.Sprintf("%s-%d", DefaultHostnamePrefix, newServerNum)
	
	// Try to get defaults from server-1
	maxPlayers := DefaultMaxPlayers
	maxViewRadius := DefaultMaxViewRadius
	gameMode := DefaultGameMode
	serverPassword := ""
	if config, err := ReadConfig(GetServerConfigPath(1)); err == nil {
		maxPlayers = config.MaxPlayers
		maxViewRadius = config.MaxViewRadius
		if config.Defaults.GameMode != "" {
			gameMode = config.Defaults.GameMode
		}
		serverPassword = config.Password
	}

	if err := UpdateServerConfig(newServerNum, port, hostname, maxPlayers, maxViewRadius, gameMode, serverPassword); err != nil {
		return fmt.Errorf("failed to create config: %w", err)
	}

	return nil
}

// RemoveLastServerInstance removes the last server instance
func RemoveLastServerInstance(numServers int) error {
	if numServers <= 1 {
		return fmt.Errorf("cannot remove last server")
	}

	lastServer := numServers
	dataDir := fmt.Sprintf("%s/server-%d", DataDirBase, lastServer)

	// Stop server if running
	tm := NewTmuxManager(DefaultBasePort)
	if tm.HasSession(lastServer) {
		_ = tm.Stop(lastServer) // Continue even if stop fails
	}

	// Remove directory
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
