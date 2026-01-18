# Managing Servers

This guide covers everyday operations: starting, stopping, and inspecting your Hytale servers using the `hsm` binary.

## Using `hsm`

The main entrypoint for managing servers is the `hsm` TUI:

```bash
sudo hsm       # Launch interactive TUI (install, updates, status, logs)
```

From the TUI you can:

- **Install or repair servers** (wizard).
- **Start / stop / restart all servers**.
- **Check status and logs**.
- **Run game or plugin updates**.
- **View server configuration**.

## TUI Navigation

The TUI uses keyboard navigation:

- **Arrow Keys (↑/↓)** or **j/k**: Navigate menu items
- **Arrow Keys (←/→)** or **h/l**: Switch between tabs
- **Enter**: Select/execute action
- **Esc** or **q**: Exit current view (returns to main menu)
- **Ctrl+C** or **q** (from main menu): Quit application

## TUI Tabs

### Install Tab

- **Run Installation Wizard**: Interactive wizard to configure and install servers
  - Configure number of servers, ports, hostname, max players
  - Set JVM arguments and admin password
  - Automatically installs Java 25 and dependencies
  - Guides through Hytale OAuth authentication
  - Downloads server files via `hytale-downloader`

### Updates Tab

- **Update Game**: Download and install latest Hytale server files
- **Update Plugins**: Update server plugins and addons
- **Enable Auto-Update Monitor**: Automatically check for updates (future feature)

### Servers Tab

- **Start All Servers**: Launch all server instances
- **Stop All Servers**: Gracefully stop all running servers
- **Restart All Servers**: Restart all server instances
- **View Server Logs**: View logs for a specific server
- **Scale Up Servers**: Add more server instances
- **Scale Down Servers**: Remove server instances

### Tools Tab

- **Edit Server Configs**: Edit shared server configuration files
- **View Server Status**: View detailed server status dashboard
  - Shows all servers with their current status (running/stopped)
  - Displays port numbers and tmux session names
  - Auto-updates every 2 seconds
  - Color-coded status indicators

## Console and logs via tmux

Servers run inside tmux sessions for easy console access:

```bash
# List all tmux sessions
tmux ls

# Attach to server 1 console
tmux attach-session -t hytale-server-1

# View server logs
tail -f data/logs/server-*.log
```

When attached to a tmux session:

- Press **Ctrl+B, then D** to detach without stopping the server.
- Type commands directly into the Hytale server console.

## Where servers live

By default, server data is stored in the `data/` directory:

```text
data/
├── Server/HytaleServer.jar   # Server JAR
├── Assets.zip                # Assets file
├── config.json               # Server config
├── universe/                 # World data
├── logs/                     # Server logs
└── ...
```

Each server runs in its own tmux session named `hytale-server-N` (where N is the server number).

## Real-time status

The TUI provides real-time server status in multiple ways:

### Status Bar

The bottom status bar shows:
- Number of servers running vs stopped (e.g., "2 running, 1 stopped")
- Current operation status (e.g., "Starting all servers...")
- Error messages if operations fail

Status is updated automatically every 2 seconds and after server operations (start/stop/restart).

### Server Status Page

Access the detailed status page via **Tools Tab → View Server Status**:

- **Table view** showing all servers:
  - Server ID
  - Status (running/stopped) with color coding
  - Port number
  - Tmux session name
- **Auto-refreshes** every 2 seconds
- **Real-time updates** when servers start/stop

This page is useful for monitoring all servers at a glance and verifying their current state.
