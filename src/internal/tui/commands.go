package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/sivert-io/hytale-server-manager/src/internal/hytale"
)

// Command finished message
type commandFinishedMsg struct {
	output string
	err    error
}

// Global cancellation function for long-running operations
var installCancel context.CancelFunc

func SetInstallCancel(cancel context.CancelFunc) {
	installCancel = cancel
}

func CancelInstall() {
	if installCancel != nil {
		installCancel()
		installCancel = nil
	}
}

// Wrapper commands for backend operations

func runBootstrapGo(cfg hytale.BootstrapConfig) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithCancel(context.Background())
		SetInstallCancel(cancel)
		defer CancelInstall()

		// Collect progress messages
		var progressSteps []string

		// Progress callback that collects steps
		progressCallback := func(percent float64, label string) {
			progressSteps = append(progressSteps, label)
		}

		// Run bootstrap with progress callback
		out, err := hytale.BootstrapWithContextAndProgress(ctx, cfg, progressCallback)

		// Combine all progress messages into output
		var output strings.Builder
		if err == nil {
			output.WriteString("Bootstrap completed successfully!\n\n")
			output.WriteString("Steps completed:\n")
			for _, step := range progressSteps {
				output.WriteString(fmt.Sprintf("  • %s\n", step))
			}
			if out != "" {
				output.WriteString(fmt.Sprintf("\n%s", out))
			}
		} else {
			output.WriteString("Bootstrap failed:\n")
			for _, step := range progressSteps {
				output.WriteString(fmt.Sprintf("  • %s\n", step))
			}
		}

		return commandFinishedMsg{
			output: output.String(),
			err:    err,
		}
	}
}

func runStartAllGo() tea.Cmd {
	return func() tea.Msg {
		numServers := hytale.DetectNumServers()
		if numServers == 0 {
			return commandFinishedMsg{
				output: "",
				err:    fmt.Errorf("no servers installed - run installation wizard first"),
			}
		}

		tm := hytale.NewTmuxManager(hytale.DefaultBasePort)
		
		// Get server JAR path (servers share the same JAR)
		jarPath := hytale.GetServerJarPath(1)
		dataDirBase := hytale.DataDirBase
		
		// Get backup settings from shared config file
		backupConfig, backupErr := hytale.ReadBackupConfig()
		if backupErr != nil {
			// Fall back to defaults if config can't be read
			backupConfig = &hytale.BackupConfig{
				Enabled:  hytale.DefaultBackupEnabled,
				Frequency: hytale.DefaultBackupFrequency,
			}
		}
		backupEnabled := backupConfig.Enabled
		backupFrequency := backupConfig.Frequency
		
		// Try to load session tokens (for automatic authentication)
		// Per Server Provider Authentication Guide: https://support.hytale.com/hc/en-us/articles/45328341414043
		// If tokens exist and are valid, servers start authenticated without manual /auth
		sessionTokens, _ := hytale.LoadSessionTokens()
		
		err := tm.StartAll(numServers, dataDirBase, jarPath, hytale.DefaultJVMArgs, backupEnabled, backupFrequency, sessionTokens)
		if err != nil {
			return commandFinishedMsg{
				output: "",
				err:    err,
			}
		}

		return commandFinishedMsg{
			output: fmt.Sprintf("All %d servers started successfully", numServers),
			err:    nil,
		}
	}
}

func runStopAllGo() tea.Cmd {
	return func() tea.Msg {
		numServers := hytale.DetectNumServers()
		if numServers == 0 {
			return commandFinishedMsg{
				output: "",
				err:    fmt.Errorf("no servers installed"),
			}
		}

		tm := hytale.NewTmuxManager(hytale.DefaultBasePort)
		
		err := tm.StopAll(numServers)
		if err != nil {
			return commandFinishedMsg{
				output: "",
				err:    err,
			}
		}

		return commandFinishedMsg{
			output: fmt.Sprintf("All %d servers stopped successfully", numServers),
			err:    nil,
		}
	}
}

func runRestartAllGo() tea.Cmd {
	return func() tea.Msg {
		numServers := hytale.DetectNumServers()
		if numServers == 0 {
			return commandFinishedMsg{
				output: "",
				err:    fmt.Errorf("no servers installed"),
			}
		}

		tm := hytale.NewTmuxManager(hytale.DefaultBasePort)
		
		// Stop all
		_ = tm.StopAll(numServers)
		
		// Small delay before starting
		time.Sleep(1 * time.Second)
		
		// Get server JAR path
		jarPath := hytale.GetServerJarPath(1)
		dataDirBase := hytale.DataDirBase
		
		// Get backup settings from shared config file
		backupConfig, backupErr := hytale.ReadBackupConfig()
		if backupErr != nil {
			// Fall back to defaults if config can't be read
			backupConfig = &hytale.BackupConfig{
				Enabled:  hytale.DefaultBackupEnabled,
				Frequency: hytale.DefaultBackupFrequency,
			}
		}
		backupEnabled := backupConfig.Enabled
		backupFrequency := backupConfig.Frequency
		
		// Try to load session tokens (for automatic authentication)
		sessionTokens, _ := hytale.LoadSessionTokens()
		
		// Start all
		err := tm.StartAll(numServers, dataDirBase, jarPath, hytale.DefaultJVMArgs, backupEnabled, backupFrequency, sessionTokens)
		if err != nil {
			return commandFinishedMsg{
				output: "",
				err:    err,
			}
		}

		return commandFinishedMsg{
			output: fmt.Sprintf("All %d servers restarted successfully", numServers),
			err:    nil,
		}
	}
}

func runUpdateGameGo() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		out, err := hytale.UpdateGame(ctx)
		return commandFinishedMsg{
			output: out,
			err:    err,
		}
	}
}

func runUpdatePluginsGo() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		out, err := hytale.UpdatePlugins(ctx)
		if err != nil {
			return commandFinishedMsg{
				output: "",
				err:    err,
			}
		}
		return commandFinishedMsg{
			output: out,
			err:    nil,
		}
	}
}

func runViewLogsGo(serverNum int) tea.Cmd {
	return func() tea.Msg {
		tm := hytale.NewTmuxManager(hytale.DefaultBasePort)
		logs, err := tm.Logs(serverNum, 100) // Get last 100 lines
		if err != nil {
			return commandFinishedMsg{
				output: "",
				err:    fmt.Errorf("failed to get logs for server %d: %w", serverNum, err),
			}
		}
		
		// Return logs as viewport content message
		return viewportContentMsg{
			content: logs,
			title:   fmt.Sprintf("Server %d Logs", serverNum),
		}
	}
}

func runScaleUpGo(numToAdd int) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		numServers := hytale.DetectNumServers()
		basePort := hytale.DefaultBasePort
		
		err := hytale.AddServerInstanceWithContext(ctx, numServers, basePort)
		if err != nil {
			return commandFinishedMsg{
				output: "",
				err:    fmt.Errorf("failed to add server: %w", err),
			}
		}

		return commandFinishedMsg{
			output: fmt.Sprintf("Added %d server instance(s) successfully", numToAdd),
			err:    nil,
		}
	}
}

func runScaleDownGo() tea.Cmd {
	return func() tea.Msg {
		numServers := hytale.DetectNumServers()
		if numServers <= 1 {
			return commandFinishedMsg{
				output: "",
				err:    fmt.Errorf("cannot remove last server"),
			}
		}

		err := hytale.RemoveLastServerInstance(numServers)
		if err != nil {
			return commandFinishedMsg{
				output: "",
				err:    err,
			}
		}

		return commandFinishedMsg{
			output: "Removed last server instance successfully",
			err:    nil,
		}
	}
}

func runAddServersGo(numToAdd int) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		numServers := hytale.DetectNumServers()
		basePort := hytale.DefaultBasePort
		
		var added int
		var lastErr error
		for i := 0; i < numToAdd; i++ {
			err := hytale.AddServerInstanceWithContext(ctx, numServers+i, basePort)
			if err != nil {
				lastErr = err
				break
			}
			added++
		}

		if lastErr != nil {
			return commandFinishedMsg{
				output: fmt.Sprintf("Added %d/%d servers (failed: %v)", added, numToAdd, lastErr),
				err:    lastErr,
			}
		}

		return commandFinishedMsg{
			output: fmt.Sprintf("Added %d server instance(s) successfully", added),
			err:    nil,
		}
	}
}

func runRemoveServersGo(numToRemove int) tea.Cmd {
	return func() tea.Msg {
		numServers := hytale.DetectNumServers()
		if numServers <= 1 {
			return commandFinishedMsg{
				output: "",
				err:    fmt.Errorf("cannot remove - only one server remaining"),
			}
		}

		if numToRemove >= numServers {
			numToRemove = numServers - 1 // Keep at least 1
		}

		var removed int
		var lastErr error
		for i := 0; i < numToRemove; i++ {
			currentNumServers := hytale.DetectNumServers()
			if currentNumServers <= 1 {
				break // Can't remove more
			}
			err := hytale.RemoveLastServerInstance(currentNumServers)
			if err != nil {
				lastErr = err
				break
			}
			removed++
		}

		if lastErr != nil {
			return commandFinishedMsg{
				output: fmt.Sprintf("Removed %d/%d servers (failed: %v)", removed, numToRemove, lastErr),
				err:    lastErr,
			}
		}

		return commandFinishedMsg{
			output: fmt.Sprintf("Removed %d server instance(s) successfully", removed),
			err:    nil,
		}
	}
}

func runInstallDependenciesGo() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		err := hytale.InstallAllDependencies(ctx, nil)
		if err != nil {
			return commandFinishedMsg{
				output: "",
				err:    err,
			}
		}

		return commandFinishedMsg{
			output: "All dependencies installed successfully",
			err:    nil,
		}
	}
}

func runWipeEverythingGo() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		out, err := hytale.WipeEverything(ctx)
		return commandFinishedMsg{
			output: out,
			err:    err,
		}
	}
}

func runCheckUpdatesGo() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		release, isNewer, err := hytale.CheckForUpdates(ctx)
		if err != nil {
			return commandFinishedMsg{
				output: "",
				err:    fmt.Errorf("failed to check for updates: %w", err),
			}
		}

		if !isNewer {
			return commandFinishedMsg{
				output: fmt.Sprintf("You are running the latest version (%s)", hytale.GetVersion()),
				err:    nil,
			}
		}

		// New version available - download and install
		progressCallback := func(percent float64, label string) {
			// Progress updates would be sent via channel in full implementation
		}

		binaryPath, err := hytale.DownloadUpdate(ctx, release, progressCallback)
		if err != nil {
			return commandFinishedMsg{
				output: "",
				err:    fmt.Errorf("failed to download update: %w", err),
			}
		}

		restartCmd, err := hytale.InstallUpdate(ctx, binaryPath)
		if err != nil {
			return commandFinishedMsg{
				output: "",
				err:    fmt.Errorf("failed to install update: %w", err),
			}
		}

		return commandFinishedMsg{
			output: fmt.Sprintf("Update installed successfully! Please restart HSM by running: %s", restartCmd),
			err:    nil,
		}
	}
}
