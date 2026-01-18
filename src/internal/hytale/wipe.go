package hytale

import (
	"context"
	"fmt"
	"os"
	"os/exec"
)

// WipeEverything permanently deletes all Hytale server data, configs, and system user
// WARNING: This is irreversible!
func WipeEverything(ctx context.Context) (string, error) {
	// 1. Stop all running servers and kill all tmux sessions
	numServers := DetectNumServers()
	if numServers > 0 {
		tm := NewTmuxManager(DefaultBasePort)
		// Stop all servers (sends /stop, then kills sessions)
		_ = tm.StopAll(numServers)
		
		// Kill any remaining tmux sessions matching our pattern
		for i := 1; i <= numServers; i++ {
			sessionName := tm.SessionName(i)
			cmd := exec.Command("tmux", "kill-session", "-t", sessionName)
			_ = cmd.Run() // Ignore errors (session might not exist)
		}
	}

	// 2. Delete all server directories
	if err := os.RemoveAll(DataDirBase); err != nil {
		return "", fmt.Errorf("failed to remove data directory: %w", err)
	}

	// 3. Delete config directory
	if err := os.RemoveAll(ConfigDir); err != nil {
		return "", fmt.Errorf("failed to remove config directory: %w", err)
	}

	// 4. Delete system user (if it exists)
	// Read default user from config if possible, or use default
	// For now, try to remove the default user
	userToRemove := DefaultHytaleUser
	cmd := exec.Command("id", "-u", userToRemove)
	if err := cmd.Run(); err == nil {
		// User exists, remove it
		// First, kill any processes owned by the user
		exec.Command("pkill", "-u", userToRemove).Run()
		
		// Remove user and home directory
		removeUserCmd := exec.Command("userdel", "-r", userToRemove)
		if err := removeUserCmd.Run(); err != nil {
			// Log warning but continue (user might not be removable, or doesn't exist)
			fmt.Printf("Warning: Failed to remove user %s: %v\n", userToRemove, err)
		}
	}

	return "All Hytale server data, configurations, and system user have been permanently deleted", nil
}
