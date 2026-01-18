# Quick Start

This guide gets you from zero to running Hytale servers in a few minutes.

## Prerequisites

- **OS**: Ubuntu 22.04+ (or similar modern Linux distro).
- **Root / sudo access** on the machine that will host the servers.
- **Go 1.19+** installed (for building the TUI binary).
- **Enough resources** for multiple Hytale servers (CPU, RAM, and disk).

## 1. Install HSM globally and run the installer wizard

From your target server, clone the repository and install:

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

- Install or repair servers (wizard).
- Start/stop/restart all servers.
- Check status and logs.
- Run game/plugin updates.

## 3. Next steps

- See **Guides → Managing Servers** for day-to-day operations.
- See **Guides → Configuration** to customize configs before or after installation.
- See **Guides → Auto Updates** to understand how updates are handled.
