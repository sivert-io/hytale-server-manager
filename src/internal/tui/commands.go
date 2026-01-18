package tui

import (
	"context"
	"fmt"
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

		out, err := hytale.BootstrapWithContext(ctx, cfg)
		return commandFinishedMsg{
			output: out,
			err:    err,
		}
	}
}

func runStartAllGo() tea.Cmd {
	return func() tea.Msg {
		// TODO: Get num servers from config/state
		numServers := 1 // Placeholder
		tm := hytale.NewTmuxManager(hytale.DefaultBasePort)
		
		// TODO: Get actual paths from config
		dataDirBase := hytale.DataDirBase
		jarPath := fmt.Sprintf("%s/Server/HytaleServer.jar", dataDirBase)
		
		err := tm.StartAll(numServers, dataDirBase, jarPath, hytale.DefaultJVMArgs)
		if err != nil {
			return commandFinishedMsg{
				output: "",
				err:    err,
			}
		}

		return commandFinishedMsg{
			output: "All servers started successfully",
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
		
		// Start all
		err := tm.StartAll(numServers, dataDirBase, jarPath, hytale.DefaultJVMArgs)
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
		return commandFinishedMsg{
			output: out,
			err:    err,
		}
	}
}
