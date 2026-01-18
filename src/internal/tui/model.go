package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/bubbles/viewport"
)

// View modes
type viewMode int

const (
	viewMain viewMode = iota
	viewInstallWizard
	viewViewport
	viewActionResult
	viewEditServerConfigs
)

// Tabs
type tab int

const (
	tabInstall tab = iota
	tabUpdates
	tabServers
	tabTools
)

// Menu item kinds
type itemKind int

const (
	itemInstallWizard itemKind = iota
	itemStartAllGo
	itemStopAllGo
	itemRestartAllGo
	itemViewLogs
	itemEditConfigs
	itemUpdateGame
	itemUpdatePlugins
	itemScaleUp
	itemScaleDown
)

// Wizard cancel message
type wizardCancelMsg struct{}

// Menu item
type menuItem struct {
	title       string
	description string
	kind        itemKind
}

// Styling
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205")).
			PaddingTop(1).
			PaddingBottom(1)

	normalText = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")).
			Bold(true)

	dimmedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	tabStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Padding(0, 1)

	tabActiveStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")).
			Bold(true).
			Padding(0, 1)

	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Italic(true)
)

// Model holds all TUI state
type model struct {
	view   viewMode
	tab    tab
	items  []menuItem
	cursor int
	width  int
	height int

	status  string
	running bool

	// Server status (would be populated from backend)
	serverStatuses []serverStatus

	// Install wizard
	wizard installWizard

	// Viewport for logs/status
	viewport viewport.Model
	viewportContent string
}

type serverStatus struct {
	ID     int
	Status string // "running", "stopped"
	Port   int
}

func initialModel() model {
	vp := viewport.New(80, 20)
	
	m := model{
		view:    viewMain,
		tab:     tabServers,
		cursor:  0,
		status:  "Ready",
		items:   getTabItems(tabServers),
		wizard:  newInstallWizard(),
		viewport: vp,
	}

	return m
}

func getTabItems(t tab) []menuItem {
	switch t {
	case tabInstall:
		return []menuItem{
			{title: "Run Installation Wizard", description: "Install and configure Hytale servers", kind: itemInstallWizard},
		}

	case tabUpdates:
		return []menuItem{
			{title: "Update Game", description: "Download and install latest Hytale server", kind: itemUpdateGame},
			{title: "Update Plugins", description: "Update server plugins and addons", kind: itemUpdatePlugins},
			{title: "Enable Auto-Update Monitor", description: "Automatically check for updates", kind: itemUpdatePlugins},
		}

	case tabServers:
		return []menuItem{
			{title: "Start All Servers", description: "Start all Hytale server instances", kind: itemStartAllGo},
			{title: "Stop All Servers", description: "Stop all running server instances", kind: itemStopAllGo},
			{title: "Restart All Servers", description: "Restart all server instances", kind: itemRestartAllGo},
			{title: "View Server Logs", description: "View logs for a specific server", kind: itemViewLogs},
			{title: "Scale Up Servers", description: "Add more server instances", kind: itemScaleUp},
			{title: "Scale Down Servers", description: "Remove server instances", kind: itemScaleDown},
		}

	case tabTools:
		return []menuItem{
			{title: "Edit Server Configs", description: "Edit shared server configuration", kind: itemEditConfigs},
			{title: "View Server Status", description: "View detailed server status", kind: itemViewLogs},
		}

	default:
		return []menuItem{}
	}
}

func (m model) Init() tea.Cmd {
	// Start polling server status on init
	return getServerStatus()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.viewport.Width = msg.Width - 4
		m.viewport.Height = msg.Height - 10
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			// Cancel install if running
			if m.view == viewInstallWizard {
				CancelInstall()
			}
			return m, tea.Quit

		case "left", "h":
			// Switch to previous tab
			if m.tab > tabInstall {
				m.tab--
				m.items = getTabItems(m.tab)
				m.cursor = 0
			}
			return m, nil

		case "right", "l":
			// Switch to next tab
			if m.tab < tabTools {
				m.tab++
				m.items = getTabItems(m.tab)
				m.cursor = 0
			}
			return m, nil

		case "up", "k":
			if m.view == viewViewport {
				m.viewport.LineUp(1)
				return m, nil
			}
			if m.cursor > 0 {
				m.cursor--
			}
			return m, nil

		case "down", "j":
			if m.view == viewViewport {
				m.viewport.LineDown(1)
				return m, nil
			}
			if m.cursor < len(m.items)-1 {
				m.cursor++
			}
			return m, nil

		case "enter", " ":
			// Execute selected action
			if len(m.items) > 0 {
				return m, m.executeAction(m.items[m.cursor].kind)
			}
			return m, nil

		case "esc":
			// Back to main menu
			if m.view != viewMain {
				m.view = viewMain
			}
			return m, nil
		}

	case commandFinishedMsg:
		// Handle command completion
		m.running = false
		if msg.err != nil {
			m.status = fmt.Sprintf("Error: %v", msg.err)
		} else {
			m.status = msg.output
		}
		// Refresh server status after command completion
		return m, getServerStatus()

	case serverStatusMsg:
		// Update server statuses
		m.serverStatuses = make([]serverStatus, len(msg.statuses))
		for i, st := range msg.statuses {
			m.serverStatuses[i] = serverStatus{
				ID:     st.Server,
				Status: st.Status,
				Port:   st.Port,
			}
		}
		
		// Update status bar with server count
		if len(msg.statuses) > 0 {
			m.status = formatServerStatus(msg.statuses)
		} else {
			m.status = "Ready"
		}
		
		// Continue polling
		return m, pollServerStatus()
	}

	return m, nil
}

func (m model) executeAction(kind itemKind) tea.Cmd {
	switch kind {
	case itemInstallWizard:
		m.view = viewInstallWizard
		m.wizard = newInstallWizard()
		return nil

	case itemStartAllGo:
		m.status = "Starting all servers..."
		m.running = true
		return runStartAllGo()

	case itemStopAllGo:
		m.status = "Stopping all servers..."
		m.running = true
		return runStopAllGo()

	case itemRestartAllGo:
		m.status = "Restarting all servers..."
		m.running = true
		return runRestartAllGo()

	case itemViewLogs:
		// Show logs viewport
		m.view = viewViewport
		m.viewportContent = "Server logs will appear here...\n\n(TODO: Fetch actual logs from tmux)"
		m.viewport.SetContent(m.viewportContent)
		return nil

	case itemEditConfigs:
		m.view = viewEditServerConfigs
		return nil // TODO: Show config editor

	case itemUpdateGame:
		m.status = "Updating game..."
		m.running = true
		return runUpdateGameGo()

	case itemUpdatePlugins:
		m.status = "Updating plugins..."
		m.running = true
		return runUpdatePluginsGo()

	case itemScaleUp:
		// TODO: Prompt for number
		return nil

	case itemScaleDown:
		// TODO: Prompt for number
		return nil

	default:
		return nil
	}
}

func (m model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	var s string

	// Header
	s += titleStyle.Render(" 🎮 Hytale Server Manager (HSM)") + "\n"

	// Tab bar
	tabs := []string{"Install", "Updates", "Servers", "Tools"}
	tabBar := ""
	for i, tabName := range tabs {
		if i == int(m.tab) {
			tabBar += tabActiveStyle.Render(tabName)
		} else {
			tabBar += tabStyle.Render(tabName)
		}
		if i < len(tabs)-1 {
			tabBar += " │ "
		}
	}
	s += "\n" + tabBar + "\n\n"

	// Menu items or other views
	if m.view == viewMain {
		for i, item := range m.items {
			cursor := "  "
			if i == m.cursor {
				cursor = selectedStyle.Render("▶ ")
			} else {
				cursor = "  "
			}

			itemText := fmt.Sprintf("%s%s", cursor, item.title)
			if i == m.cursor {
				itemText = selectedStyle.Render(itemText)
			} else {
				itemText = normalText.Render(itemText)
			}

			s += itemText
			if item.description != "" {
				s += " " + dimmedStyle.Render("- "+item.description)
			}
			s += "\n"
		}
	} else if m.view == viewInstallWizard {
		// Install wizard view
		s += m.wizard.View()
	} else if m.view == viewViewport {
		// Viewport view (logs, status, etc.)
		s += titleStyle.Render(" 📋 Server Logs") + "\n\n"
		s += m.viewport.View()
		s += "\n" + dimmedStyle.Render("↑/↓: Scroll  |  Esc: Back")
	} else {
		// Other views
		s += "View not yet implemented\n"
		s += dimmedStyle.Render("Press Esc to return")
	}

	// Status bar
	s += "\n\n"
	if m.running {
		s += statusStyle.Render("⏳ " + m.status)
	} else {
		s += statusStyle.Render("✓ " + m.status)
	}

	// Help text
	helpText := "↑/↓: Navigate  |  ←/→: Tabs  |  Enter: Select  |  Esc: Back  |  q: Quit"
	s += "\n" + dimmedStyle.Render(helpText)

	return lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Padding(1, 2).
		Render(s)
}

// Command functions are implemented in commands.go

func Run() error {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	_, err := p.Run()
	return err
}
