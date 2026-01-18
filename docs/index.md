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
- **Auto-installation**: Automatic Java 25 setup and hytale-downloader installation.
- **Safe updates**: Handles game and plugin updates automatically while preserving your configs.
- **OAuth authentication**: Integrated Hytale authentication with token management.
- **Persistent data**: Worlds, tokens, and logs survive restarts.
- **Observability & control**: Go-based TUI and tmux integration for logs and debugging.

## Quick Start

For most users, installing `hsm` globally and running the TUI is all you need:

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

- `data/` – server data (worlds, configs, logs, tokens).

See:

- **Getting Started → Quick Start** – first-time setup.
- **Guides → Managing Servers** – everyday operations.
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
