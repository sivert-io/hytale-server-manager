# Quick Start

This guide gets you from zero to running Hytale servers in a few minutes.

## Prerequisites

- **OS**: Ubuntu 22.04+ (or similar modern Linux distro).
- **Root / sudo access** on the machine that will host the servers.
- **Go 1.19+** installed (for building the TUI binary).
- **Enough resources** for multiple Hytale servers (CPU, RAM, and disk).

## 1. Install HSM globally and run the installer wizard

From your target server, install from GitHub releases:

```bash
# One-line installation from GitHub releases (recommended)
arch=$(uname -m); \
case "$arch" in \
  x86_64)  asset="hsm-linux-amd64" ;; \
  aarch64|arm64) asset="hsm-linux-arm64" ;; \
  *) echo "Unsupported architecture: $arch" && exit 1 ;; \
esac; \
tmp=$(mktemp); \
curl -L "https://github.com/sivert-io/hytale-server-manager/releases/latest/download/$asset" -o "$tmp" && \
sudo install -m 0755 "$tmp" /usr/local/bin/hsm && \
rm "$tmp" && \
sudo hsm            # launches the interactive TUI installer
```

**Alternative: Build from source**

```bash
# Clone the repository
git clone https://github.com/sivert-io/hytale-server-manager.git
cd hytale-server-manager

# Build and install globally (requires sudo)
sudo ./install.sh

# Run the TUI
sudo hsm
```

The installer wizard will:

- Install required system dependencies (Java 25, tmux, etc.).
- Guide you through Hytale OAuth authentication.
- Download Hytale server files using `hytale-downloader`.
- Configure multiple Hytale instances with sane defaults.

## 2. Use the HSM TUI

Once installation completes, you can re-open the TUI at any time with:

```bash
sudo hsm            # run all TUI actions (install, updates, status, logs)
```

From the TUI you can:

- **Install Tab**: Run the installation wizard to set up servers
- **Updates Tab**: Update game files or plugins
- **Servers Tab**: Start/stop/restart all servers, view logs, scale servers
- **Tools Tab**: Edit server configs, view detailed server status

### TUI Navigation

- **Arrow Keys (↑/↓)** or **j/k**: Navigate menu items
- **Arrow Keys (←/→)** or **h/l**: Switch between tabs
- **Enter**: Select/execute action
- **Esc** or **q**: Exit current view (returns to main menu)
- **Ctrl+C** or **q** (from main menu): Quit application

## 3. Next steps

- See **Guides → Managing Servers** for day-to-day operations.
- See **Guides → Configuration** to customize configs before or after installation.
- See **Guides → Auto Updates** to understand how updates are handled.
