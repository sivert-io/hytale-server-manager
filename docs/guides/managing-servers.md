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

## Server operations via TUI

Navigate the TUI with arrow keys:

- **Install Tab** – Run the installation wizard or reconfigure servers.
- **Updates Tab** – Update game files or plugins.
- **Servers Tab** – Start/stop/restart all servers, view logs, scale up/down.
- **Tools Tab** – Edit configs, view detailed status.

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

The TUI polls server status every 2 seconds and displays:

- Number of servers running vs stopped
- Live status in the status bar
- Server port information

Status is updated automatically after server operations (start/stop/restart).
