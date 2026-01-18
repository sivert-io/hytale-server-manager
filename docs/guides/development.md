# Development Guide

This guide covers the architecture and development patterns used in HSM for contributors.

## Architecture Overview

HSM uses a **Bubble Tea**-based TUI with a clean separation between:

- **TUI Layer** (`src/internal/tui/`) - User interface, views, input handling
- **Backend Layer** (`src/internal/hytale/`) - Game-specific operations, server management, file I/O

## Project Structure

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
│       │   ├── status.go         # Server status polling
│       │   ├── viewport.go       # Viewport-based views
│       │   └── commands.go       # TUI command wrappers
│       └── hytale/               # Backend layer (game operations)
│           ├── bootstrap.go      # Server installation/setup
│           ├── servers.go        # Server lifecycle management
│           ├── tmux.go           # Tmux session management
│           ├── update.go         # Game/plugin updates
│           └── consts.go         # Defaults and constants
├── scripts/                      # Server scripts (hytale-auth, etc.)
├── tools/                        # Helper scripts
└── docs/                        # Documentation
```

## Technology Stack

### Core Dependencies

- **Bubble Tea** (`github.com/charmbracelet/bubbletea`) - TUI framework
- **Bubbles** (`github.com/charmbracelet/bubbles`) - Reusable TUI components:
  - `viewport` - Scrollable content display
  - `textinput` - Text input fields
  - `spinner` - Loading animations
  - `progress` - Progress bars
- **Lip Gloss** (`github.com/charmbracelet/lipgloss`) - Styling and colors

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
    wizard installWizard      // Install wizard state
    viewport viewport.Model   // Viewport for logs/status
    serverStatuses []serverStatus  // Server status tracking
    // ... other state
}
```

**Key Methods:**
- `Init() tea.Cmd` - Initialize and return initial commands
- `Update(msg tea.Msg) (tea.Model, tea.Cmd)` - Handle messages, return updated model + commands
- `View() string` - Render the current view as a string

### 2. View Modes

Use an enum for different views:

```go
type viewMode int
const (
    viewMain viewMode = iota
    viewInstallWizard
    viewViewport
    // ... other views
)
```

### 3. Tabbed Menu System

Organize menu items into tabs:

- **Install Tab**: Installation wizard, dependencies
- **Updates Tab**: Game updates, plugin updates
- **Servers Tab**: Status dashboard, logs, start/stop/restart, scale up/down
- **Tools Tab**: Config editing, cleanup, utilities

### 4. Command Pattern

Long-running operations return commands via `tea.Cmd`:

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

The `Update()` method handles these messages and updates the UI.

## Backend Layer (`src/internal/hytale/`)

### Bootstrap (`bootstrap.go`)

Main installation function that:
- Creates directory structure (`data/master-install`, `data/server-1`, etc.)
- Downloads/installs Hytale server files via `hytale-downloader`
- Configures server instances
- Writes configuration files

**Key Function:**
```go
func BootstrapWithContext(ctx context.Context, cfg BootstrapConfig) (string, error)
```

### Server Management (`servers.go`)

Functions for:
- `AddServerInstanceWithContext()` - Scale up servers
- `RemoveLastServerInstance()` - Scale down servers
- `DetectNumServers()` - Detection from existing configs
- `GetServerJarPath()` - Get server JAR path

### Process Management (`tmux.go`)

Manages server processes via tmux:
- `TmuxManager` struct with server count detection
- `StartAll()`, `StopAll()`, `RestartAll()` - Bulk operations
- `Status()` - Human-readable status for all servers
- `Logs()` - Read server logs
- Session naming: `hytale-server-1`, `hytale-server-2`, etc.

### Updates (`update.go`)

Handle game and plugin updates:
- Download latest Hytale server files via `hytale-downloader`
- Update plugins/addons
- Deploy to all servers

### Constants (`consts.go`)

Defaults and constants:
```go
const (
    DataDirBase = "/home/hytale"
    DefaultHytaleUser = "hytale"
    DefaultNumServers = 1
    DefaultBasePort = 5520
    DefaultJVMArgs = "-Xms4G -Xmx8G"
    // ...
)
```

## TUI-Backend Integration

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
- Call `Cancel()` on user interrupt (Ctrl+C in wizard)
- Backend operations check `ctx.Done()`

### Status Polling

Server status is polled every 2 seconds:

```go
func pollServerStatus() tea.Cmd {
    return tea.Tick(2*time.Second, func(time.Time) tea.Msg {
        numServers := hytale.DetectNumServers()
        tm := hytale.NewTmuxManager(hytale.DefaultBasePort)
        statuses := tm.Status(numServers)
        return serverStatusMsg{statuses: statuses}
    })
}
```

## Hytale-Specific Adaptations

### Config Files

**Hytale uses:**
- `config.json` - Main server configuration
- `whitelist.json` - Whitelist
- `bans.json` - Banned players
- `permissions.json` - Player permissions

All configs are stored in `data/` directory.

### Ports

**Hytale typically uses:**
- Game port (default 5520 UDP)
- Incrementing ports for multiple servers (5520, 5521, 5522, ...)

**Note:** Hytale uses **QUIC over UDP** (not TCP).

### Process Management

**Hytale server:**
- Launched via Java: `java -jar HytaleServer.jar`
- Requires Java 25+ runtime
- Managed via tmux sessions
- JVM arguments configurable (`-Xms4G -Xmx8G`)

### Update Process

**Hytale:**
- Download server files using `hytale-downloader`
- Uses OAuth authentication for downloads
- Files extracted to `data/Server/` directory
- Assets stored in `data/Assets.zip`

## Styling and UX

### Colors and Styles

Use Charm Lip Gloss styles:

```go
import "github.com/charmbracelet/lipgloss"

var (
    titleStyle = lipgloss.NewStyle().
        Bold(true).
        Foreground(lipgloss.Color("205"))
    selectedStyle = lipgloss.NewStyle().
        Foreground(lipgloss.Color("205"))
    // ...
)
```

### Key UX Features

1. **Progress Indicators**
   - Status messages in footer
   - Real-time status polling

2. **Error Handling**
   - Red error messages in status bar
   - Validation feedback on forms
   - Graceful error messages (not raw stack traces)

3. **Responsive Design**
   - Adjust viewport height based on terminal size
   - Scrollable wizard if fields exceed terminal height
   - Wrapping text in viewports

4. **Keyboard Shortcuts**
   - `↑/↓` or `j/k` - Navigate menu items
   - `←/→` or `h/l` - Switch tabs
   - `Enter` - Execute selected action
   - `q` or `Ctrl+C` - Quit
   - `Esc` - Back to main menu

## Development Workflow

1. **Make Changes**
   ```bash
   # Make your changes in src/
   ```

2. **Build**
   ```bash
   go build -ldflags="-s -w" -o ./hsm ./src/cmd/hytale-tui
   ```

3. **Test**
   ```bash
   sudo ./hsm
   ```

4. **Verify**
   - TUI builds without errors
   - Navigation works smoothly
   - Commands execute correctly
   - Status updates in real-time

## Best Practices

- Keep TUI layer **completely** separate from game-specific logic
- Backend should be testable without TUI
- Use context for cancellation
- Log all operations for debugging
- Follow existing naming conventions
- Handle edge cases (no servers installed, permission errors, etc.)
- Support both interactive (TUI) and non-interactive (CLI) modes (future)

## Testing Checklist

- [ ] Install wizard completes successfully
- [ ] Multiple servers can be created
- [ ] Servers can be started/stopped/restarted
- [ ] Status polling updates correctly
- [ ] Logs viewport displays correctly
- [ ] Status dashboard shows server states
- [ ] Server scaling (add/remove) works
- [ ] Config updates apply to all servers
- [ ] Game/plugin updates work
- [ ] Error handling is graceful
- [ ] Navigation (tabs, menus) is smooth
- [ ] Cancellation works for long operations

## Key Files Reference

- `src/cmd/hytale-tui/main.go` - Entry point
- `src/internal/tui/model.go` - Main state machine
- `src/internal/tui/install_wizard.go` - Installation wizard
- `src/internal/tui/commands.go` - Command wrappers
- `src/internal/tui/status.go` - Status polling
- `src/internal/hytale/bootstrap.go` - Installation logic
- `src/internal/hytale/tmux.go` - Process management
- `src/internal/hytale/servers.go` - Server lifecycle
