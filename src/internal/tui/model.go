package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/sivert-io/hytale-server-manager/src/internal/hytale"
)

// View modes
type viewMode int

const (
	viewMain viewMode = iota
	viewInstallWizard
	viewViewport
	viewActionResult
	viewEditServerConfigs
	viewServerStatus
	viewServerSelection
	viewConfirmWipe
)

// Tabs
type tab int

const (
	tabInstall tab = iota
	tabUpdates
	tabServers
	tabTools
	tabAdvanced
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
	itemAddServers
	itemRemoveServers
	itemViewServerStatus
	itemInstallDependencies
	itemCheckUpdates
	itemWipeEverything
)

// Wizard cancel message
type wizardCancelMsg struct{}

// Viewport content message
type viewportContentMsg struct {
	content string
	title   string
}

// Activity log message for verbose output
type activityLogMsg struct {
	message string
}

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
	
	// Progress tracking
	progress progressModel
	showProgress bool
	
	// Server selection for logs/scale operations
	selectedServer int
	serverList     []int
	// Track which action we're doing in server selection
	serverSelectionAction itemKind
	
	// Update status
	updateAvailable bool
	latestVersion    string
	currentVersion   string
	
	// Action result (receipt)
	actionResult string
	actionError   error
	actionTitle   string
	
	// Activity logs (last 4 lines for verbose output)
	activityLogs []string
	maxActivityLogs int
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
		tab:     tabInstall,
		cursor:  0,
		status:  "Ready",
		items:   getTabItems(tabInstall),
		wizard:  newInstallWizard(),
		viewport: vp,
		progress: newProgressModel(""),
		showProgress: false,
		currentVersion: hytale.GetShortVersion(),
		actionResult: "",
		actionError: nil,
		actionTitle: "",
		activityLogs: make([]string, 0),
		maxActivityLogs: 4,
	}

	return m
}

func getTabItems(t tab) []menuItem {
	return getTabItemsWithUpdate(t, false, "")
}

func getTabItemsWithUpdate(t tab, updateAvailable bool, latestVersion string) []menuItem {
	switch t {
	case tabInstall:
		items := []menuItem{
			{title: "Install Dependencies", description: "Install required system dependencies (Java, tmux, etc.)", kind: itemInstallDependencies},
			{title: "Run Installation Wizard", description: "Install and configure Hytale servers", kind: itemInstallWizard},
		}
		// Add update option at the end if available
		if updateAvailable {
			items = append(items, menuItem{
				title: fmt.Sprintf("Update HSM (v%s available)", latestVersion),
				description: fmt.Sprintf("Current: v%s → Latest: v%s", hytale.GetShortVersion(), latestVersion),
				kind: itemCheckUpdates,
			})
		}
		return items

	case tabUpdates:
		return []menuItem{
			{title: "Check for Updates", description: "Check for HSM updates and install if available", kind: itemCheckUpdates},
			{title: "Update Game", description: "Download and install latest Hytale server", kind: itemUpdateGame},
			{title: "Update Plugins", description: "Update server plugins and addons", kind: itemUpdatePlugins},
		}

	case tabServers:
		items := []menuItem{
			{title: "Start All Servers", description: "Start all Hytale server instances", kind: itemStartAllGo},
			{title: "Stop All Servers", description: "Stop all running server instances", kind: itemStopAllGo},
			{title: "Restart All Servers", description: "Restart all server instances", kind: itemRestartAllGo},
			{title: "View Server Logs", description: "View logs for a specific server", kind: itemViewLogs},
			{title: "Add Servers", description: "Add multiple server instances", kind: itemAddServers},
			{title: "Remove Servers", description: "Remove multiple server instances", kind: itemRemoveServers},
			{title: "Scale Up Servers", description: "Add one server instance", kind: itemScaleUp},
			{title: "Scale Down Servers", description: "Remove one server instance", kind: itemScaleDown},
		}
		// Add update option at the end if available
		if updateAvailable {
			items = append(items, menuItem{
				title: fmt.Sprintf("Update HSM (v%s available)", latestVersion),
				description: fmt.Sprintf("Current: v%s → Latest: v%s", hytale.GetShortVersion(), latestVersion),
				kind: itemCheckUpdates,
			})
		}
		return items

	case tabTools:
		items := []menuItem{
			{title: "Edit Server Configs", description: "Edit shared server configuration", kind: itemEditConfigs},
			{title: "View Server Status", description: "View detailed server status", kind: itemViewServerStatus},
		}
		// Add update option at the end if available
		if updateAvailable {
			items = append(items, menuItem{
				title: fmt.Sprintf("Update HSM (v%s available)", latestVersion),
				description: fmt.Sprintf("Current: v%s → Latest: v%s", hytale.GetShortVersion(), latestVersion),
				kind: itemCheckUpdates,
			})
		}
		return items

	case tabAdvanced:
		items := []menuItem{
			{title: "⚠️  Wipe Everything", description: "Delete all servers, configs, and data (DANGER: irreversible)", kind: itemWipeEverything},
		}
		return items

	default:
		return []menuItem{}
	}
}

func (m model) Init() tea.Cmd {
	// Start polling server status and check for updates on init
	return tea.Batch(
		getServerStatus(),
		checkForUpdates(),
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle special keys in wizard view before forwarding
	if m.view == viewInstallWizard {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "q", "esc":
				// Return to main menu from wizard
				CancelInstall()
				m.view = viewMain
				m.items = getTabItemsWithUpdate(m.tab, m.updateAvailable, m.latestVersion)
				return m, nil
			}
		case wizardCancelMsg:
			// Return to main menu from wizard
			m.view = viewMain
			m.items = getTabItemsWithUpdate(m.tab, m.updateAvailable, m.latestVersion)
			return m, nil
		case commandFinishedMsg:
			// Handle command completion from wizard (e.g., bootstrap)
			// Let it fall through to the main commandFinishedMsg handler below
			// by NOT returning early here - we need to update view state
		}
		// Only forward non-commandFinishedMsg messages to wizard
		// (commandFinishedMsg will be handled by the main switch below)
		if _, ok := msg.(commandFinishedMsg); !ok {
			var cmd tea.Cmd
			m.wizard, cmd = m.wizard.Update(msg)
			return m, cmd
		}
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.viewport.Width = msg.Width - 4
		m.viewport.Height = msg.Height - 10
		return m, nil

	case wizardCancelMsg:
		// Return to main menu from wizard
		m.view = viewMain
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			// Quit application
			if m.view == viewInstallWizard {
				CancelInstall()
			}
			return m, tea.Quit
		case "q":
			// Quit only from main menu, otherwise handled above
			if m.view == viewMain {
				return m, tea.Quit
			}

		case "left", "h":
			// Switch to previous tab
			if m.tab > tabInstall {
				m.tab--
				m.items = getTabItemsWithUpdate(m.tab, m.updateAvailable, m.latestVersion)
				m.cursor = 0
			}
			return m, nil

		case "right", "l":
			// Switch to next tab
			if m.tab < tabAdvanced {
				m.tab++
				m.items = getTabItemsWithUpdate(m.tab, m.updateAvailable, m.latestVersion)
				m.cursor = 0
			}
			return m, nil

		case "up", "k":
			if m.view == viewViewport {
				m.viewport.LineUp(1)
				return m, nil
			}
			if m.view == viewServerSelection {
				if m.selectedServer > 0 {
					m.selectedServer--
				}
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
			if m.view == viewServerSelection {
				if m.selectedServer < len(m.serverList)-1 {
					m.selectedServer++
				}
				return m, nil
			}
			if m.cursor < len(m.items)-1 {
				m.cursor++
			}
			return m, nil

		case "enter", " ":
			// Handle action result view - return to main menu
			if m.view == viewActionResult {
				m.view = viewMain
				return m, nil
			}
			
			// Handle wipe confirmation - require typing "WIPE" (or just Enter twice for safety)
			if m.view == viewConfirmWipe {
				// Execute wipe (double Enter confirmation)
				m.running = true
				m.actionTitle = "⚠️  Wipe Everything"
				return m, tea.Batch(
					sendActivityLog("Wiping all data... (this may take a moment)"),
					runWipeEverythingGo(),
				)
			}
			
			// Handle server selection view
			if m.view == viewServerSelection {
				// Use serverSelectionAction to determine what to do
				switch m.serverSelectionAction {
				case itemViewLogs:
					serverNum := m.serverList[m.selectedServer]
					return m, runViewLogsGo(serverNum)
				case itemScaleUp:
					numToAdd := m.serverList[m.selectedServer]
					m.running = true
					m.actionTitle = fmt.Sprintf("⬆️  Add %d Server(s)", numToAdd)
					return m, tea.Batch(
						sendActivityLog(fmt.Sprintf("Adding %d server instance(s)...", numToAdd)),
						runScaleUpGo(numToAdd),
					)
				case itemScaleDown:
					m.running = true
					m.actionTitle = "⬇️  Remove Server"
					return m, tea.Batch(
						sendActivityLog("Removing last server instance..."),
						runScaleDownGo(),
					)
				case itemAddServers:
					numToAdd := m.serverList[m.selectedServer]
					m.running = true
					m.actionTitle = fmt.Sprintf("➕ Add %d Server(s)", numToAdd)
					return m, tea.Batch(
						sendActivityLog(fmt.Sprintf("Adding %d server instance(s)...", numToAdd)),
						runAddServersGo(numToAdd),
					)
				case itemRemoveServers:
					numToRemove := -m.serverList[m.selectedServer] // Negative values represent count
					m.running = true
					m.actionTitle = fmt.Sprintf("➖ Remove %d Server(s)", numToRemove)
					return m, tea.Batch(
						sendActivityLog(fmt.Sprintf("Removing %d server instance(s)...", numToRemove)),
						runRemoveServersGo(numToRemove),
					)
				default:
					// Fallback: try to determine from list structure (legacy behavior)
					if len(m.serverList) > 0 && m.serverList[0] <= 5 && len(m.serverList) == 5 {
						numToAdd := m.serverList[m.selectedServer]
						return m, runScaleUpGo(numToAdd)
					} else if len(m.serverList) == 1 && m.serverList[0] == 1 {
						return m, runScaleDownGo()
					} else if len(m.serverList) > 0 && m.serverList[0] > 0 {
						serverNum := m.serverList[m.selectedServer]
						return m, runViewLogsGo(serverNum)
					}
				}
			}
			
			// Only handle Enter in main view (or wizard/sub-views handled above)
			if m.view != viewMain {
				return m, nil
			}
			
			// Execute selected action
			if len(m.items) == 0 {
				// No items available - refresh items list
				m.items = getTabItemsWithUpdate(m.tab, m.updateAvailable, m.latestVersion)
				return m, nil
			}
			
			if m.cursor >= len(m.items) {
				// Cursor out of bounds - reset it
				m.cursor = 0
				return m, nil
			}
			
			kind := m.items[m.cursor].kind
			// Update model state for actions that change view
			switch kind {
			case itemInstallWizard:
				m.view = viewInstallWizard
				m.wizard = newInstallWizard()
				return m, nil
			case itemViewLogs:
				// Show server selection
				numServers := hytale.DetectNumServers()
				if numServers == 0 {
					m.status = "No servers installed"
					return m, nil
				}
				m.serverList = make([]int, numServers)
				for i := 0; i < numServers; i++ {
					m.serverList[i] = i + 1
				}
				m.selectedServer = 0
				m.serverSelectionAction = itemViewLogs
				m.view = viewServerSelection
				return m, nil
			case itemEditConfigs:
				m.view = viewEditServerConfigs
				return m, nil
			case itemViewServerStatus:
				m.view = viewServerStatus
				return m, getServerStatus()
			case itemScaleUp:
				// Show scale up selection (1-5 servers)
				m.serverList = []int{1, 2, 3, 4, 5}
				m.selectedServer = 0
				m.serverSelectionAction = itemScaleUp
				m.view = viewServerSelection
				return m, nil
			case itemScaleDown:
				// Show scale down confirmation (use viewServerSelection with special list)
				numServers := hytale.DetectNumServers()
				if numServers <= 1 {
					m.status = "Cannot remove - only one server remaining"
					return m, nil
				}
				m.serverList = []int{1} // Special marker for scale down
				m.selectedServer = 0
				m.serverSelectionAction = itemScaleDown
				m.view = viewServerSelection
				return m, nil
			case itemAddServers:
				// Show add servers selection (1-10 servers)
				// Enforce MaxServersPerLicense limit per Hytale Server Manual
				maxPossible := hytale.MaxServersPerLicense
				numServers := hytale.DetectNumServers()
				maxToAdd := maxPossible - numServers
				if maxToAdd > 10 {
					maxToAdd = 10
				}
				if maxToAdd < 1 {
					m.status = "Cannot add more servers - max capacity reached"
					return m, nil
				}
				m.serverList = make([]int, maxToAdd)
				for i := 0; i < maxToAdd; i++ {
					m.serverList[i] = i + 1
				}
				m.selectedServer = 0
				m.serverSelectionAction = itemAddServers
				m.view = viewServerSelection
				return m, nil
			case itemRemoveServers:
				// Show remove servers selection (1 to all but last)
				numServers := hytale.DetectNumServers()
				if numServers <= 1 {
					m.status = "Cannot remove - only one server remaining"
					return m, nil
				}
				maxToRemove := numServers - 1
				if maxToRemove > 10 {
					maxToRemove = 10
				}
				// Use negative numbers to mark "remove" operation
				m.serverList = make([]int, maxToRemove)
				for i := 0; i < maxToRemove; i++ {
					m.serverList[i] = -(i + 1)
				}
				m.selectedServer = 0
				m.serverSelectionAction = itemRemoveServers
				m.view = viewServerSelection
				return m, nil
			case itemWipeEverything:
				// Show wipe confirmation view
				m.view = viewConfirmWipe
				return m, nil
			case itemInstallDependencies, itemCheckUpdates, itemStartAllGo, itemStopAllGo, itemRestartAllGo, itemUpdateGame, itemUpdatePlugins:
				// All command actions - run via executeAction
				cmd := (&m).executeAction(kind)
				return m, cmd
			default:
				// Fallback - should not reach here for known items
				cmd := (&m).executeAction(kind)
				return m, cmd
			}

		case "esc":
			// Back to main menu
			if m.view != viewMain {
				m.view = viewMain
				m.items = getTabItemsWithUpdate(m.tab, m.updateAvailable, m.latestVersion)
				m.cursor = 0
			}
			return m, nil
		}

	case progressMsg:
		// Update progress bar
		m.showProgress = true
		var cmd tea.Cmd
		m.progress, cmd = m.progress.Update(msg)
		return m, cmd

	case progressCompleteMsg:
		// Hide progress bar
		m.showProgress = false
		var cmd tea.Cmd
		m.progress, cmd = m.progress.Update(msg)
		return m, cmd

	case commandFinishedMsg:
		// Handle command completion - show receipt view
		m.running = false
		m.showProgress = false
		m.actionResult = msg.output
		m.actionError = msg.err
		// Clear activity logs after command completes
		m.activityLogs = make([]string, 0)
		// Update action title if error (keep the one set in executeAction if success)
		if m.actionError != nil {
			// Prepend error indicator to existing title
			if m.actionTitle != "" {
				m.actionTitle = "❌ " + m.actionTitle
			} else {
				m.actionTitle = "❌ Action Failed"
			}
		} else if m.actionTitle == "" {
			m.actionTitle = "✅ Action Completed"
		} else {
			// Prepend success indicator
			m.actionTitle = "✅ " + m.actionTitle
		}
		// Switch to receipt view
		m.view = viewActionResult
		// Refresh server status after command completion (but don't overwrite receipt)
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
		
		// Continue polling only if in main view or server status view
		if m.view == viewMain || m.view == viewServerStatus {
			return m, pollServerStatus()
		}
		return m, nil

	case viewportContentMsg:
		// Set viewport content and switch to viewport view
		m.view = viewViewport
		m.viewportContent = msg.content
		m.viewport.SetContent(msg.content)
		return m, nil
	}

	return m, nil
}

// Helper to send activity log message
func sendActivityLog(msg string) tea.Cmd {
	return func() tea.Msg {
		return activityLogMsg{message: msg}
	}
}

func (m *model) executeAction(kind itemKind) tea.Cmd {
	switch kind {
	case itemStartAllGo:
		m.status = "Starting all servers..."
		m.running = true
		m.actionTitle = "🚀 Start All Servers"
		return tea.Batch(
			sendActivityLog("Starting all servers..."),
			runStartAllGo(),
		)

	case itemStopAllGo:
		m.status = "Stopping all servers..."
		m.running = true
		m.actionTitle = "🛑 Stop All Servers"
		return tea.Batch(
			sendActivityLog("Stopping all servers..."),
			runStopAllGo(),
		)

	case itemRestartAllGo:
		m.status = "Restarting all servers..."
		m.running = true
		m.actionTitle = "🔄 Restart All Servers"
		return tea.Batch(
			sendActivityLog("Restarting all servers..."),
			runRestartAllGo(),
		)

	case itemUpdateGame:
		m.status = "Updating game..."
		m.running = true
		m.actionTitle = "🎮 Update Game"
		return tea.Batch(
			sendActivityLog("Updating game files..."),
			runUpdateGameGo(),
		)

	case itemUpdatePlugins:
		m.status = "Updating plugins..."
		m.running = true
		m.actionTitle = "🔌 Update Plugins"
		return tea.Batch(
			sendActivityLog("Updating plugins..."),
			runUpdatePluginsGo(),
		)

	case itemInstallDependencies:
		m.status = "Installing dependencies..."
		m.running = true
		m.showProgress = true
		m.actionTitle = "📦 Install Dependencies"
		return tea.Batch(
			sendActivityLog("Checking and installing system dependencies..."),
			runInstallDependenciesGo(),
		)

	case itemScaleUp:
		// Scale up - add 1 server (handled via viewServerSelection)
		// This case shouldn't be reached as itemScaleUp goes to viewServerSelection
		// But if it is, default to adding 1 server
		m.running = true
		m.actionTitle = "⬆️  Add Server"
		return tea.Batch(
			sendActivityLog("Adding 1 server instance..."),
			runScaleUpGo(1),
		)

	case itemScaleDown:
		// Scale down - remove last server (handled via viewServerSelection)
		// This case shouldn't be reached as itemScaleDown goes to viewServerSelection
		// But if it is, default to removing last server
		m.running = true
		m.actionTitle = "⬇️  Remove Server"
		return tea.Batch(
			sendActivityLog("Removing last server instance..."),
			runScaleDownGo(),
		)

	case itemCheckUpdates:
		m.status = "Checking for updates..."
		m.running = true
		m.actionTitle = "🔍 Check for Updates"
		return tea.Batch(
			sendActivityLog("Checking for HSM updates..."),
			runCheckUpdatesGo(),
		)

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

	// Tab bar - only show in main view
	if m.view == viewMain {
		tabs := []string{"Install", "Updates", "Servers", "Tools", "Advanced"}
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
	} else {
		s += "\n"
	}

	// Show update notification if available
	if m.updateAvailable && m.view == viewMain {
		updateStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("46")). // Green
			Bold(true).
			Padding(0, 1).
			MarginBottom(1)
		s += "\n" + updateStyle.Render(fmt.Sprintf("🆕 New version available: v%s (current: v%s)", m.latestVersion, m.currentVersion)) + "\n"
	}

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
	} else if m.view == viewServerStatus {
		// Server status view
		s += titleStyle.Render(" 📊 Server Status") + "\n\n"
		if len(m.serverStatuses) == 0 {
			s += dimmedStyle.Render("No servers installed or detected.\n")
			s += dimmedStyle.Render("Run the installation wizard to set up servers.")
		} else {
			// Table header
			header := fmt.Sprintf("%-8s %-12s %-8s %-10s", "Server", "Status", "Port", "Session")
			s += selectedStyle.Render(header) + "\n"
			s += dimmedStyle.Render("────────────────────────────────────────") + "\n"
			
			// Server rows
			for _, st := range m.serverStatuses {
				statusColor := "241" // dimmed (stopped)
				statusText := "stopped"
				if st.Status == "running" {
					statusColor = "46" // green
					statusText = "running"
				}
				
				statusStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(statusColor))
				row := fmt.Sprintf("%-8d %-12s %-8d %-10s",
					st.ID,
					statusStyle.Render(statusText),
					st.Port,
					fmt.Sprintf("hytale-server-%d", st.ID),
				)
				s += row + "\n"
			}
		}
		s += "\n" + dimmedStyle.Render("Esc: Back")
	} else if m.view == viewServerSelection {
		// Server selection view (for logs, scale up/down)
		if len(m.serverList) > 0 && m.serverList[0] <= 5 && len(m.serverList) == 5 {
			// Scale up selection
			s += titleStyle.Render(" ⬆️  Scale Up Servers") + "\n\n"
			s += dimmedStyle.Render("Select number of servers to add:") + "\n\n"
			for i, num := range m.serverList {
				cursor := "  "
				if i == m.selectedServer {
					cursor = selectedStyle.Render("▶ ")
				}
				text := fmt.Sprintf("Add %d server(s)", num)
				if i == m.selectedServer {
					text = selectedStyle.Render(text)
				}
				s += fmt.Sprintf("%s%s\n", cursor, text)
			}
		} else if len(m.serverList) == 1 && m.serverList[0] == 1 {
			// Scale down confirmation
			s += titleStyle.Render(" ⬇️  Scale Down Servers") + "\n\n"
			s += dimmedStyle.Render("This will remove the last server instance.") + "\n\n"
			s += selectedStyle.Render("▶ Confirm Removal") + "\n"
			s += "\n" + dimmedStyle.Render("Enter: Confirm  |  Esc: Cancel")
		} else {
			// Server selection for logs
			s += titleStyle.Render(" 📋 Select Server for Logs") + "\n\n"
			for i, serverNum := range m.serverList {
				cursor := "  "
				if i == m.selectedServer {
					cursor = selectedStyle.Render("▶ ")
				}
				text := fmt.Sprintf("Server %d", serverNum)
				if i == m.selectedServer {
					text = selectedStyle.Render(text)
				}
				s += fmt.Sprintf("%s%s\n", cursor, text)
			}
			s += "\n" + dimmedStyle.Render("Enter: View Logs  |  Esc: Back")
		}
	} else if m.view == viewEditServerConfigs {
		// Config editor view
		s += titleStyle.Render(" ⚙️  Edit Server Configs") + "\n\n"
		s += dimmedStyle.Render("Config editor coming soon...") + "\n"
		s += dimmedStyle.Render("For now, edit config files directly in:") + "\n"
		s += fmt.Sprintf("  %s\n", hytale.DataDirBase)
		s += "\n" + dimmedStyle.Render("Esc: Back")
	} else if m.view == viewActionResult {
		// Action result receipt view
		s += titleStyle.Render(" " + m.actionTitle) + "\n\n"
		
		// Box style for receipt
		boxStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("205")).
			Padding(1, 2).
			Width(m.width - 8)
		
		if m.actionError != nil {
			// Error display
			errorStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("196")). // Red
				Bold(true)
			content := errorStyle.Render("Error:") + "\n\n"
			content += normalText.Render(m.actionError.Error())
			s += boxStyle.Render(content)
		} else {
			// Success display
			successStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("46")). // Green
				Bold(true)
			content := successStyle.Render("Success!") + "\n\n"
			if m.actionResult != "" {
				content += normalText.Render(m.actionResult)
			} else {
				content += normalText.Render("Operation completed successfully.")
			}
			s += boxStyle.Render(content)
		}
		
		s += "\n\n" + dimmedStyle.Render("Press Esc or Enter to return to main menu")
	} else if m.view == viewConfirmWipe {
		// Wipe Everything confirmation view
		warningStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")). // Red
			Bold(true).
			Padding(1, 0)
		
		s += warningStyle.Render("⚠️  WIPE EVERYTHING - CONFIRMATION REQUIRED") + "\n\n"
		
		boxStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("196")). // Red border
			Padding(1, 2).
			Width(m.width - 8)
		
		content := normalText.Render("This will PERMANENTLY DELETE:") + "\n\n"
		content += "  • All server instances and data\n"
		content += "  • All configuration files\n"
		content += "  • All world/universe data\n"
		content += "  • All backups\n"
		content += "  • Master server files\n"
		content += "  • System user (if created)\n"
		content += "  • All authentication tokens\n"
		content += "  • Everything under /var/lib/hytale and /etc/hytale\n\n"
		content += warningStyle.Render("THIS ACTION CANNOT BE UNDONE!") + "\n\n"
		content += normalText.Render("Press Enter to confirm this action (or Esc to cancel)")
		
		s += boxStyle.Render(content)
		s += "\n\n" + dimmedStyle.Render("Enter: Confirm Wipe  |  Esc: Cancel")
	} else {
		// Other views
		s += "View not yet implemented\n"
		s += dimmedStyle.Render("Press Esc to return")
	}

	// Show progress bar if active (above status bar)
	if m.showProgress {
		s += "\n" + m.progress.View()
	}
	
	// Status bar
	s += "\n\n"
	if m.running {
		s += statusStyle.Render("⏳ " + m.status)
	} else {
		// Show helpful hints based on current view and status
		statusText := m.status
		if statusText == "Ready" {
			// Add context-specific hints
			if m.view == viewMain {
				switch m.tab {
				case tabInstall:
					statusText = "Ready - Run installation wizard to set up servers"
				case tabUpdates:
					statusText = "Ready - Update game files or plugins"
				case tabServers:
					if len(m.serverStatuses) > 0 {
						running := 0
						for _, st := range m.serverStatuses {
							if st.Status == "running" {
								running++
							}
						}
						statusText = fmt.Sprintf("Ready - %d/%d servers running", running, len(m.serverStatuses))
					} else {
						statusText = "Ready - No servers installed (run installation wizard)"
					}
				case tabTools:
					statusText = "Ready - Edit configs or view server status"
				case tabAdvanced:
					statusText = "Ready - Advanced/dangerous operations"
				}
			}
		}
		s += statusStyle.Render("✓ " + statusText)
	}

	// Activity logs (last 4 lines) - show above status bar when running
	if m.running && len(m.activityLogs) > 0 {
		s += "\n"
		logStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Italic(true).
			MarginTop(1)
		
		// Show last maxActivityLogs lines
		start := len(m.activityLogs) - m.maxActivityLogs
		if start < 0 {
			start = 0
		}
		for i := start; i < len(m.activityLogs); i++ {
			s += logStyle.Render("  " + m.activityLogs[i]) + "\n"
		}
	}

	// Version info at bottom
	versionText := fmt.Sprintf("HSM v%s", hytale.GetShortVersion())
	s += "\n" + dimmedStyle.Render(versionText)
	
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
