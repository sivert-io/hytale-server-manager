<div align="center">
  <img src="assets/kgs-icon.svg" alt="Hytale Server Manager" width="140" height="140">
  
  # Hytale Server Manager (HSM)
  
  ⚡ **Terminal UI (TUI) for managing Hytale dedicated servers natively**
  
  <p>Beautiful TUI built with Bubble Tea, native Java performance, tmux-based server management. Install once, manage forever.</p>

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://github.com/sivert-io/hytale-server-manager/blob/main/LICENSE)
[![Java](https://img.shields.io/badge/Java-25-ED8B00?logo=openjdk&logoColor=white)](https://adoptium.net)
[![Go](https://img.shields.io/badge/Go-1.19+-00ADD8?logo=go&logoColor=white)](https://golang.org)

**📚 <a href="#-quick-start" target="_blank">Quick Start</a>** • <a href="#-installation" target="_blank">Installation</a> • <a href="#-features" target="_blank">Features</a> • <a href="https://github.com/sivert-io/hytale-server-manager/issues" target="_blank">💬 Issues & Support</a>

</div>

---

## ✨ Features

🎮 **Beautiful TUI** — Tabbed interface built with Bubble Tea for intuitive server management  
⚡ **Native Java** — Best performance with lower CPU and memory usage (no Docker overhead)  
🔧 **Auto-Installation** — Automatic Java 25 setup and hytale-downloader installation  
📦 **Auto-Updates** — Downloads and updates server files automatically  
🔐 **OAuth2 Authentication** — Integrated Hytale authentication with token management  
🔄 **Auto-Refresh Tokens** — Background daemon keeps tokens valid (30-day refresh tokens)  
🖥️ **Tmux Integration** — Each server runs in its own tmux session for easy log access  
⚡ **Fast Boot** — AOT cache support for quicker server startup  
💾 **Persistent Data** — Worlds, tokens, and logs survive restarts  
📊 **Real-Time Status** — Live server status polling and monitoring  
🎯 **Multi-Server Support** — Manage multiple server instances from one interface  

---

## ⚙️ Requirements

- **Go 1.19+** (for building the TUI binary)
- **Java 25+** (auto-installed by installation wizard if missing)
- **Tmux** (for server process management)
- **hytale-downloader** (auto-installed during setup)
- **Hytale account** for server authentication
- **16GB RAM** recommended (8GB minimum)
- **4 CPU cores** recommended (2 cores minimum)
- **UDP port 5520** open and forwarded on your firewall/router

---

## 🚀 Quick Start

### Installation

Install the `hsm` binary globally:

```bash
# Clone the repository
git clone https://github.com/sivert-io/hytale-server-manager.git
cd hytale-server-manager

# Build and install globally (requires sudo)
sudo ./install.sh

# Run the TUI
sudo hsm
```

> **Note:** The TUI needs elevated privileges to manage server files and tmux sessions.

### First Run

1. **Open the TUI**: Run `sudo hsm` to launch the Terminal User Interface
2. **Installation Wizard**: Navigate to the **Install** tab and run the installation wizard
3. **Configure Servers**: Set number of servers, ports, and settings
4. **Authenticate**: The wizard will guide you through Hytale OAuth authentication
5. **Download Files**: Server files will be downloaded automatically using `hytale-downloader`
6. **Start Servers**: Use the **Servers** tab to start/stop/restart your servers

### TUI Navigation

- **Tabs**: `←/→` or `h/l` - Switch between Install | Updates | Servers | Tools
- **Menu**: `↑/↓` or `j/k` - Navigate menu items
- **Select**: `Enter` - Execute selected action
- **Back**: `Esc` - Return to main menu
- **Quit**: `q` or `Ctrl+C` - Exit TUI

### Connect to Your Server

Once servers are running, connect at `your-ip:5520` using the Hytale client.

> **Note:** Hytale uses **QUIC over UDP** (not TCP). Forward UDP port 5520 on your firewall.

---

## 📖 TUI Overview

The Hytale Server Manager provides a tabbed interface:

### Install Tab
- **Installation Wizard** - Multi-step form to configure and install Hytale servers
  - Number of servers
  - Base port configuration
  - JVM arguments
  - Authentication setup

### Updates Tab
- **Update Game** - Download latest Hytale server files
- **Update Plugins** - Update server plugins and addons
- **Auto-Update Monitor** - Enable automatic update checking

### Servers Tab
- **Start All Servers** - Launch all server instances
- **Stop All Servers** - Gracefully stop all servers
- **Restart All Servers** - Restart all server instances
- **View Server Logs** - View logs for specific servers
- **Scale Up/Down** - Add or remove server instances

### Tools Tab
- **Edit Server Configs** - Edit shared server configuration
- **View Server Status** - Detailed server status dashboard

---

---

## 📁 Project Structure

```
hytale-server-manager/
├── src/                    # Go source code
│   ├── cmd/hytale-tui/    # TUI entry point
│   └── internal/
│       ├── tui/           # TUI layer (user interface)
│       └── hytale/        # Backend layer (server management)
├── scripts/                # Server scripts (hytale-auth, etc.)
├── tools/                  # Helper scripts
│   ├── release.sh         # GitHub release script
│   ├── start.sh           # Development build script
│   └── ...
├── data/                   # Server data (worlds, configs, logs)
├── install.sh             # Global installation script
└── README.md
```

---

## 🔧 Hytale CLI Tools

The manager uses official Hytale CLI tools:

### hytale-downloader

Downloads and updates Hytale server files. Installed automatically during setup.

**Manual Installation:**
```bash
# Check the official Hytale documentation for latest download link
# https://support.hytale.com/hc/en-us/articles/45326769420827-Hytale-Server-Manual

# Typically installed to /usr/local/bin/hytale-downloader
```

### hytale-auth

OAuth authentication tool for Hytale servers. Available in `scripts/hytale-auth.sh`.

**Usage:**
```bash
# Login to Hytale account
./scripts/hytale-auth.sh login

# List profiles
./scripts/hytale-auth.sh profile list

# Select profile
./scripts/hytale-auth.sh profile select 1

# Create game session
./scripts/hytale-auth.sh session
```

---

## 🛑 Stopping Servers

### Via TUI

Use the **Servers** tab in the TUI to stop all servers gracefully.

### Manually

```bash
# Stop all servers (sends /stop command, then kills tmux sessions)
pkill -f "HytaleServer.jar"

# Or kill specific tmux session
tmux kill-session -t hytale-server-1
```

---

## 🔄 Restarting Servers

### Via TUI

Use the **Servers** tab → **Restart All Servers** option.

### Manually

```bash
# Restart via TUI (recommended)
sudo hsm

# Or manually attach to tmux session and restart
tmux attach-session -t hytale-server-1
```

---

## 📖 Documentation

### Official Hytale Documentation

- **[Hytale Server Manual](https://support.hytale.com/hc/en-us/articles/45326769420827-Hytale-Server-Manual)** — Official server setup guide
- **[Server Provider Authentication Guide](https://support.hytale.com/hc/en-us/articles/45328341414043-Server-Provider-Authentication-Guide)** — Authentication setup

### Project Documentation

- **Installation Wizard** - Guides you through initial server setup
- **Server Management** - Start/stop/restart servers via TUI
- **Update System** - Automatic server file updates

---

## 🤝 Contributing

Contributions are welcome! Whether you're fixing bugs, adding features, improving docs, or sharing ideas.

Feel free to open an [issue](https://github.com/sivert-io/hytale-server-manager/issues) or submit a pull request.

---

## 📜 License

MIT License - see [LICENSE](LICENSE) for details

---

<div align="center">
  <strong>Made with ❤️ for the Hytale community</strong>
</div>
