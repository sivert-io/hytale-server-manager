# Instructions for Creating a Hytale Server Manager TUI

This document provides comprehensive instructions for creating a Terminal User Interface (TUI) for Hytale server management that matches the architecture and design patterns of the CS2 Server Manager.

## Overview

The CS2 Server Manager uses a **Bubble Tea**-based TUI with a clean separation between:
- **TUI Layer** (`src/internal/tui/`) - User interface, views, input handling
- **Backend Layer** (`src/internal/csm/`) - Game-specific operations, server management, file I/O

Create a similar structure for Hytale: `src/internal/tui/` and `src/internal/hytale/` (or `hsm/`).

---

## Technology Stack

### Core Dependencies (from `go.mod`)
- **Bubble Tea** (`github.com/charmbracelet/bubbletea`) - TUI framework
- **Bubbles** (`github.com/charmbracelet/bubbles`) - Reusable TUI components:
  - `viewport` - Scrollable content display
  - `textinput` - Text input fields
  - `spinner` - Loading animations
  - `progress` - Progress bars
- Standard library for file I/O, process management, etc.

### Project Structure
```
hytale-server-manager/
├── src/
│   ├── cmd/
│   │   └── hytale-tui/
│   │       └── main.go           # Entry point, CLI parsing
│   └── internal/
│       ├── tui/                  # TUI layer (UI logic only)
│       │   ├── model.go          # Main state machine, Update/View
│       │   ├── install_wizard.go # Installation wizard
│       │   ├── config_editor.go  # Config editing interface
│       │   ├── viewport.go       # Viewport-based views
│       │   └── commands.go       # TUI command wrappers
│       └── hytale/               # Backend layer (game operations)
│           ├── bootstrap.go      # Server installation/setup
│           ├── servers.go        # Server lifecycle management
│           ├── tmux.go           # Tmux session management
│           ├── update_configs.go # Config update operations
│           ├── update.go         # Game/plugin updates
│           └── consts.go         # Defaults and constants
├── go.mod
└── README.md
```

---

## Architecture Patterns

### 1. State Machine Pattern (`model.go`)

The TUI uses a `model` struct that holds all state and implements Bubble Tea's `Model` interface:

```go
type model struct {
    view   viewMode           // Current view (main menu, wizard, etc.)
    tab    tab                // Current tab (Install, Updates, Servers, Tools)
    items  []menuItem         // Menu items for current tab
    cursor int                // Selected menu item
    status string             // Status bar message
    running bool              // Is a command running?
    spin   spinner.Model      // Loading spinner
    wizard installWizard      // Install wizard state
    // ... other state
}
```

**Key Methods:**
- `Init() tea.Cmd` - Initialize and return initial commands
- `Update(msg tea.Msg) (tea.Model, tea.Cmd)` - Handle messages, return updated model + commands
- `View() string` - Render the current view as a string

### 2. View Modes (Views)

Use an enum for different views:

```go
type viewMode int
const (
    viewMain viewMode = iota
    viewInstallWizard
    viewViewport
    viewActionResult
    viewEditServerConfigs
    // ... other views
)
```

### 3. Tabbed Menu System

Organize menu items into tabs:
- **Install Tab**: Installation wizard, dependencies
- **Updates Tab**: Game updates, plugin updates, auto-update monitor
- **Servers Tab**: Status dashboard, logs, start/stop/restart, scale up/down
- **Tools Tab**: Config editing, cleanup, utilities

Each tab has its own `[]menuItem` list.

### 4. Menu Item System

```go
type menuItem struct {
    title       string    // Menu item text
    description string    // Help text
    kind        itemKind  // Action type enum
}

type itemKind int
const (
    itemInstallWizard itemKind = iota
    itemStartAllGo
    itemStopAllGo
    // ... other actions
)
```

### 5. Command Pattern

Long-running operations return commands via `tea.Cmd`:

```go
func runUpdateGameGo() tea.Cmd {
    return func() tea.Msg {
        // Run operation
        out, err := hytale.UpdateGame(...)
        return commandFinishedMsg{
            output: out,
            err: err,
        }
    }
}
```

The `Update()` method handles these messages and updates the UI.

---

## Core Components to Implement

### 1. Main Menu (`View()` in `model.go`)

- Header with title and version
- Tab bar (Install | Updates | Servers | Tools)
- Menu items list (arrows to navigate, Enter to select)
- Status bar at bottom
- Update banner (if newer version available)

**Navigation:**
- `↑/↓` or `j/k` - Navigate menu items
- `←/→` or `h/l` - Switch tabs
- `Enter` - Execute selected action
- `q` or `Ctrl+C` - Quit (double-press confirmation)
- `Esc` - Back to main menu (from sub-views)

### 2. Install Wizard (`install_wizard.go`)

Multi-step form with:
- DB mode selection (Docker-managed or external)
- Number of servers
- Base ports (game, query)
- Hostname prefix
- RCON/admin password
- Max players
- Feature toggles (plugins, auto-updates)
- External DB settings (if applicable)

**Pattern:**
- One-page scrolling wizard with cursor-based navigation
- Text input fields for editing values
- Boolean toggles (checkboxes)
- Validation before submission
- Progress steps during installation

### 3. Viewport Views (`viewport.go`)

For displaying long content (logs, status, configs):
- Scrollable viewport component
- Header with title and instructions
- `Enter/q/Esc` to return

**Use for:**
- Server status dashboard
- Server logs
- Config file viewing
- Action result display

### 4. Config Editor (`config_editor.go`)

Form for editing server configs:
- Fields: RCON password, max players, etc.
- Read current values from shared config
- Apply changes to all servers
- Validation (e.g., RCON password required)

### 5. Prompt Views

Simple prompts for single input:
- "Add servers" - Enter count
- "Remove servers" - Enter count  
- "View logs" - Enter server number
- "View config" - Enter server number

**Pattern:**
- Text input field
- Enter to submit, Esc to cancel
- Error messages if validation fails

---

## Backend Layer Structure (`src/internal/hytale/`)

### 1. Bootstrap (`bootstrap.go`)

Main installation function that:
- Creates dedicated system user
- Downloads/installs Hytale server files
- Sets up directory structure (`master-install`, `server-1`, `server-2`, etc.)
- Configures shared config directory (`hytale-config/`)
- Creates server instances
- Writes config files (server.cfg, properties.json, etc.)

**Key Function:**
```go
func BootstrapWithContext(ctx context.Context, cfg BootstrapConfig) (string, error)
```

**BootstrapConfig Struct:**
```go
type BootstrapConfig struct {
    HytaleUser      string
    NumServers      int
    BasePort        int
    QueryPort       int
    HostnamePrefix  string
    AdminPassword   string
    MaxPlayers      int
    // ... Hytale-specific settings
}
```

### 2. Server Management (`servers.go`)

Functions for:
- `AddServerInstanceWithContext()` - Scale up servers
- `RemoveLastServerInstance()` - Scale down servers
- Detection functions (read existing configs)
- Server directory/file management

### 3. Process Management (`tmux.go`)

Manages server processes via tmux:
- `TmuxManager` struct with server count detection
- `Start(server int)`, `Stop(server int)`, `StartAll()`, `StopAll()`
- `Status()` - Human-readable status for all servers
- `Logs(server int, lines int)` - Read server logs
- Session naming convention: `hytale-server-1`, `hytale-server-2`, etc.

### 4. Config Management (`update_configs.go`)

Update shared configs:
- Read from shared config location (`hytale-config/`)
- Write to shared config
- Apply to all servers (restart required)
- Validate inputs

### 5. Updates (`update.go`)

Handle game and plugin updates:
- Download latest Hytale server files
- Update plugins/addons
- Deploy to all servers

### 6. Constants (`consts.go`)

Defaults and constants:
```go
const (
    DefaultHytaleUser = "hytaleservermanager"
    DefaultNumServers = 3
    DefaultBasePort = 27015
    DefaultQueryPort = 27016
    DefaultAdminPassword = ""
    // ...
)
```

---

## TUI-Backend Integration Pattern

### Commands in `tui/commands.go`

Wrap backend operations as TUI commands:

```go
func runBootstrapGo(cfg hytale.BootstrapConfig) tea.Cmd {
    return func() tea.Msg {
        ctx, cancel := context.WithCancel(context.Background())
        SetInstallCancel(cancel)
        defer CancelInstall()
        
        out, err := hytale.BootstrapWithContext(ctx, cfg)
        return commandFinishedMsg{
            output: out,
            err: err,
        }
    }
}
```

### Context Cancellation

Support cancellation for long operations:
- Store `context.CancelFunc` globally
- Call `Cancel()` on user interrupt
- Backend operations check `ctx.Done()`

---

## Hytale-Specific Adaptations

### Config Files

**Hytale uses:**
- `server.properties` - Main server configuration
- `whitelist.json` - Whitelist
- `banned-players.json` - Banned players
- Config directory structure differs from CS2

**Adaptation:**
- Update file paths in detection/writing functions
- Match Hytale config format (properties files, JSON)
- Store shared configs in `hytale-config/` directory

### Ports

**Hytale typically uses:**
- Game port (default 27015)
- Query port (default 27016)
- RCON port (may be same as game port or separate)

**Adaptation:**
- Adjust port increment logic (Hytale may use different spacing)
- Update port detection in configs

### Process Management

**Hytale server:**
- Launched via Java: `java -jar HytaleServer.jar`
- Requires Java runtime
- May use different process management (systemd, screen, tmux)

**Adaptation:**
- Update `tmux.go` Start() method for Hytale launch command
- Adjust JVM arguments, memory settings
- Handle Hytale-specific server files

### Update Process

**Hytale:**
- Download server JAR from official source
- May have different update mechanism than CS2/Steam

**Adaptation:**
- Replace SteamCMD logic with Hytale download mechanism
- Update file paths and extraction logic

---

## Styling and UX

### Colors and Styles

Use Charm Lip Gloss styles:

```go
import "github.com/charmbracelet/lipgloss"

var (
    titleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
    menuSelectedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
    menuItemStyle = lipgloss.NewStyle()
    // ...
)
```

### Key UX Features

1. **Progress Indicators**
   - Spinner for "loading..." states
   - Progress bars for file transfers/updates
   - Status messages in footer

2. **Error Handling**
   - Red error messages in status bar
   - Validation feedback on forms
   - Graceful error messages (not raw stack traces)

3. **Responsive Design**
   - Adjust viewport height based on terminal size
   - Scrollable wizard if fields exceed terminal height
   - Wrapping text in viewports

4. **Keyboard Shortcuts**
   - Consistent navigation (arrows, vim keys)
   - Double-press to quit (prevents accidental exits)
   - Escape always goes back

---

## Installation Flow

1. **Dependencies Check**
   - Java runtime
   - Tmux (for process management)
   - Required system packages

2. **Bootstrap Process**
   - Create system user
   - Download Hytale server
   - Set up directory structure
   - Create shared config
   - Create server instances
   - Write configs

3. **Post-Install**
   - Install auto-update monitor (cron)
   - Start all servers
   - Show completion summary

---

## Testing Checklist

- [ ] Install wizard completes successfully
- [ ] Multiple servers can be created
- [ ] Servers can be started/stopped/restarted
- [ ] Config editor reads/writes correctly
- [ ] Logs viewport displays correctly
- [ ] Status dashboard shows server states
- [ ] Server scaling (add/remove) works
- [ ] Config updates apply to all servers
- [ ] Game/plugin updates work
- [ ] Error handling is graceful
- [ ] Navigation (tabs, menus) is smooth
- [ ] Cancellation works for long operations

---

## Key Files to Create (Priority Order)

1. **Backend Core:**
   - `src/internal/hytale/consts.go` - Constants
   - `src/internal/hytale/tmux.go` - Process management
   - `src/internal/hytale/servers.go` - Server lifecycle
   - `src/internal/hytale/bootstrap.go` - Installation

2. **TUI Core:**
   - `src/internal/tui/model.go` - Main state machine
   - `src/cmd/hytale-tui/main.go` - Entry point

3. **TUI Views:**
   - `src/internal/tui/viewport.go` - Viewport views
   - `src/internal/tui/install_wizard.go` - Install wizard
   - `src/internal/tui/commands.go` - Command wrappers

4. **Extended Features:**
   - `src/internal/tui/config_editor.go` - Config editing
   - `src/internal/hytale/update_configs.go` - Config updates
   - `src/internal/hytale/update.go` - Update operations

---

## Reference Implementation

Study the CS2 Server Manager codebase:
- **Architecture**: `src/internal/tui/model.go` - See how state is managed
- **Wizard**: `src/internal/tui/install_wizard.go` - Form handling pattern
- **Backend**: `src/internal/csm/bootstrap.go` - Installation logic
- **Tmux**: `src/internal/csm/tmux.go` - Process management pattern

---

## Notes

- Keep TUI layer **completely** separate from game-specific logic
- Backend should be testable without TUI
- Use context for cancellation
- Log all operations for debugging
- Follow existing naming conventions
- Handle edge cases (no servers installed, permission errors, etc.)
- Support both interactive (TUI) and non-interactive (CLI) modes

---

## Getting Started

1. Set up Go project structure
2. Add Bubble Tea dependencies (`go get`)
3. Create basic `model.go` with `Init/Update/View`
4. Implement main menu first
5. Add tab navigation
6. Implement install wizard
7. Add backend bootstrap function
8. Wire up commands
9. Add remaining views/features incrementally

Start simple, iterate. The CS2 TUI evolved over time - begin with core functionality and expand.
