package tui

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/sivert-io/hytale-server-manager/src/internal/hytale"
)

// Status message for server status updates
type serverStatusMsg struct {
	statuses []hytale.ServerStatus
}

// Poll server status periodically
func pollServerStatus() tea.Cmd {
	return tea.Tick(2*time.Second, func(time.Time) tea.Msg {
		numServers := hytale.DetectNumServers()
		if numServers == 0 {
			return serverStatusMsg{statuses: []hytale.ServerStatus{}}
		}

		tm := hytale.NewTmuxManager(hytale.DefaultBasePort)
		statuses := tm.Status(numServers)

		return serverStatusMsg{statuses: statuses}
	})
}

// Get server status once
func getServerStatus() tea.Cmd {
	return func() tea.Msg {
		numServers := hytale.DetectNumServers()
		if numServers == 0 {
			return serverStatusMsg{statuses: []hytale.ServerStatus{}}
		}

		tm := hytale.NewTmuxManager(hytale.DefaultBasePort)
		statuses := tm.Status(numServers)

		return serverStatusMsg{statuses: statuses}
	}
}

// Format server status for display
func formatServerStatus(statuses []hytale.ServerStatus) string {
	if len(statuses) == 0 {
		return "No servers installed"
	}

	var s string
	running := 0
	stopped := 0

	for _, st := range statuses {
		if st.Status == "running" {
			running++
		} else {
			stopped++
		}
	}

	if running > 0 && stopped > 0 {
		s = fmt.Sprintf("%d running, %d stopped", running, stopped)
	} else if running > 0 {
		s = fmt.Sprintf("All %d running", running)
	} else {
		s = fmt.Sprintf("All %d stopped", stopped)
	}

	return s
}
