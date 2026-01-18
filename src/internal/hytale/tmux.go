package hytale

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// TmuxManager manages Hytale server processes via tmux sessions
type TmuxManager struct {
	basePort int
}

// NewTmuxManager creates a new TmuxManager instance
func NewTmuxManager(basePort int) *TmuxManager {
	return &TmuxManager{
		basePort: basePort,
	}
}

// SessionName returns the tmux session name for a given server number
func (tm *TmuxManager) SessionName(server int) string {
	return fmt.Sprintf("%s-%d", TmuxSessionPrefix, server)
}

// HasSession checks if a tmux session exists
func (tm *TmuxManager) HasSession(server int) bool {
	sessionName := tm.SessionName(server)
	cmd := exec.Command("tmux", "has-session", "-t", sessionName)
	return cmd.Run() == nil
}

// Start launches a Hytale server in a tmux session
// If sessionTokens is provided, servers start authenticated (no need for /auth login device)
// Per Server Provider Authentication Guide: https://support.hytale.com/hc/en-us/articles/45328341414043
func (tm *TmuxManager) Start(server int, dataDir, jarPath string, jvmArgs string, backupEnabled bool, backupFrequency int, sessionTokens *SessionTokens) error {
	sessionName := tm.SessionName(server)

	// Check if session already exists
	if tm.HasSession(server) {
		return fmt.Errorf("session %s already exists", sessionName)
	}

	// Calculate port from basePort (Hytale doesn't store port in config.json, it's passed via --bind)
	port := tm.basePort + (server - 1)
	
	// Parse JVM args into array
	args := strings.Fields(jvmArgs)
	args = append(args, "-jar", jarPath)
	
	// Hytale server arguments (Host Havoc optimization guide)
	// --bind: Bind to specific port
	args = append(args, "--bind", fmt.Sprintf("0.0.0.0:%d", port))
	
	// --assets: Specify assets file location (relative to server directory)
	assetsPath := filepath.Join(dataDir, "Assets.zip")
	// Also check master-install location
	if _, err := os.Stat(assetsPath); os.IsNotExist(err) {
		masterAssets := filepath.Join(DataDirBase, "master-install", "Assets.zip")
		if _, err := os.Stat(masterAssets); err == nil {
			assetsPath = masterAssets
		}
	}
	if _, err := os.Stat(assetsPath); err == nil {
		args = append(args, "--assets", assetsPath)
	}
	
	// --backup: Enable automatic backups (if enabled)
	if backupEnabled {
		args = append(args, "--backup")
		// --backup-frequency: Backup frequency in minutes
		args = append(args, "--backup-frequency", fmt.Sprintf("%d", backupFrequency))
	}

	// Add session and identity tokens if provided (per Server Provider Authentication Guide)
	// This allows servers to start authenticated without manual /auth login device
	if sessionTokens != nil && sessionTokens.SessionToken != "" && sessionTokens.IdentityToken != "" {
		args = append(args, "--session-token", sessionTokens.SessionToken)
		args = append(args, "--identity-token", sessionTokens.IdentityToken)
		if sessionTokens.OwnerUUID != "" {
			args = append(args, "--owner-uuid", sessionTokens.OwnerUUID)
		}
	}

	// Create tmux session and run server
	cmd := exec.Command("tmux", "new-session", "-d", "-s", sessionName,
		"-c", dataDir,
		"java", strings.Join(args, " "))
	
	return cmd.Run()
}

// Stop gracefully stops a server by sending /stop command, then kills the tmux session
func (tm *TmuxManager) Stop(server int) error {
	sessionName := tm.SessionName(server)

	if !tm.HasSession(server) {
		return fmt.Errorf("session %s does not exist", sessionName)
	}

	// Send /stop command to server
	cmd := exec.Command("tmux", "send-keys", "-t", sessionName, "/stop", "C-m")
	_ = cmd.Run()

	// Wait a moment, then kill session
	cmd = exec.Command("sleep", "2")
	_ = cmd.Run()

	cmd = exec.Command("tmux", "kill-session", "-t", sessionName)
	return cmd.Run()
}

// StartAll starts all servers with optional session tokens
func (tm *TmuxManager) StartAll(numServers int, dataDirBase, jarPath string, jvmArgs string, backupEnabled bool, backupFrequency int, sessionTokens *SessionTokens) error {
	for i := 1; i <= numServers; i++ {
		serverDir := GetServerDir(i)
		if err := tm.Start(i, serverDir, jarPath, jvmArgs, backupEnabled, backupFrequency, sessionTokens); err != nil {
			return fmt.Errorf("failed to start server %d: %w", i, err)
		}
	}
	return nil
}

// StopAll stops all running servers
func (tm *TmuxManager) StopAll(numServers int) error {
	for i := 1; i <= numServers; i++ {
		if tm.HasSession(i) {
			_ = tm.Stop(i) // Continue on error
		}
	}
	return nil
}

// Status returns human-readable status for all servers
func (tm *TmuxManager) Status(numServers int) []ServerStatus {
	statuses := make([]ServerStatus, numServers)
	
	for i := 1; i <= numServers; i++ {
		sessionName := tm.SessionName(i)
		
		// Calculate port from basePort (Hytale doesn't store port in config.json, it's passed via --bind)
		port := tm.basePort + (i - 1)
		
		running := tm.HasSession(i)
		status := "stopped"
		if running {
			status = "running"
		}

		statuses[i-1] = ServerStatus{
			Server: i,
			Status: status,
			Port:   port,
			Session: sessionName,
		}
	}

	return statuses
}

// Logs reads the last N lines from a server's tmux session
func (tm *TmuxManager) Logs(server int, lines int) (string, error) {
	sessionName := tm.SessionName(server)

	if !tm.HasSession(server) {
		return "", fmt.Errorf("session %s does not exist", sessionName)
	}

	// Use tmux capture-pane to get output
	cmd := exec.Command("tmux", "capture-pane", "-t", sessionName, "-p", "-S", strconv.Itoa(-lines))
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return string(output), nil
}

// ServerStatus represents the status of a single server
type ServerStatus struct {
	Server  int
	Status  string // "running", "stopped"
	Port    int
	Session string
}
