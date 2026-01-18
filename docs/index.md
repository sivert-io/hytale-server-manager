---
title: Hytale Server Manager
hide:
  - navigation
  - toc
---

# Hytale Server Manager

Terminal UI (TUI) for managing Hytale dedicated servers natively. Deploy multiple server instances in minutes with auto-updates, OAuth authentication, and native Java performance.

## What it does

- **Multi-server deployment**: Spin up multiple Hytale servers with a single command.
- **Native Java**: Best performance with lower CPU and memory usage (no Docker overhead).
- **Interactive TUI**: Terminal-based interface with tabbed navigation for all operations.
- **Installation wizard**: Guided setup with automatic dependency installation (Java 25, tmux, etc.).
- **Real-time status**: Live server status dashboard showing running/stopped servers, ports, and sessions.
- **Safe updates**: Handles game and plugin updates automatically while preserving your configs.
- **OAuth authentication**: Integrated Hytale authentication with token management.
- **Persistent data**: Worlds, tokens, and logs survive restarts.
- **Observability & control**: Go-based TUI and tmux integration for logs and debugging.

## Quick Start

For most users, installing `hsm` globally and running the TUI is all you need:

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

Read the **Getting Started** section for a full walkthrough.

## Project layout

```
hytale-server-manager/
├── src/                    # Go source code
│   ├── cmd/hytale-tui/    # TUI entry point
│   └── internal/
│       ├── tui/           # TUI layer (user interface)
│       └── hytale/        # Backend layer (server management)
├── tools/                  # Helper scripts
│   ├── release.sh         # GitHub release script
│   └── start.sh           # Development build script
├── data/                   # Server data (worlds, configs, logs)
├── install.sh             # Global installation script (build from source)
└── docs/                   # Documentation
```

Server data is stored in `data/` (worlds, configs, logs, tokens).

## TUI Features

The HSM TUI provides a complete interface for managing your servers:

- **Install Tab**: Interactive wizard for initial setup and configuration
- **Updates Tab**: Update game files and plugins
- **Servers Tab**: Start/stop/restart servers, view logs, scale instances
- **Tools Tab**: Edit configs, view detailed server status dashboard

All operations are accessible via keyboard navigation (arrow keys, Enter, Esc).

See:

- **Getting Started → Quick Start** – first-time setup and TUI navigation.
- **Guides → Managing Servers** – everyday operations and TUI features.
- **Guides → Configuration** – customizing your servers.
- **Guides → Auto Updates** – how updates are handled behind the scenes.
- **Guides → Troubleshooting** – common problems and fixes.

---

## Support

- [GitHub Issues](https://github.com/sivert-io/hytale-server-manager/issues) – report bugs or request features.
- [Discussions](https://github.com/sivert-io/hytale-server-manager/discussions) – ask questions and share ideas.

---

## License & credits

<div align="center" markdown>

MIT License • Made with :material-heart: for the Hytale community

</div>
