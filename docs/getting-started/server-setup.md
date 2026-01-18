# Server Setup

This page explains how to prepare a machine for Hytale Server Manager and what the installation script expects.

## System requirements

- **Linux server** (Ubuntu 22.04+ recommended).
- **Root / sudo access**.
- **64-bit CPU** with sufficient cores for multiple game servers.
- **At least 16 GB RAM** recommended (8 GB minimum).
- **4 CPU cores** recommended (2 cores minimum).
- **Stable network** with UDP port 5520 open and forwarded.

## Required packages

The installer will attempt to install most dependencies for you, but in locked-down environments you may need to do it manually:

```bash
sudo apt-get update
sudo apt-get install -y \
  tmux curl unzip tar git
```

You also need:

- **Java 25+** – Auto-installed by the installation wizard if missing (Eclipse Temurin via Adoptium).
- **hytale-downloader** – Required for downloading server files. The installer will guide you through setup if missing.
- **Go 1.19+** – Required for building the TUI binary (only needed if building from source).

## Network and ports

By default, servers use incrementing ports. A typical layout is:

| Server | Port  |
|--------|-------|
| 1      | 5520  |
| 2      | 5521  |
| 3      | 5522  |

> **Note:** Hytale uses **QUIC over UDP** (not TCP). Make sure your firewall allows UDP traffic on these ports and forward them on your router.

## Filesystem layout

After installation, your key locations are:

- The `hsm` binary installed to `/usr/local/bin/hsm` (globally available).
- The `data/` directory (default: `./data/` from where you run `hsm`) containing server data (worlds, configs, logs, tokens).

Server instances are managed via tmux sessions, with each server running in its own session (e.g., `hytale-server-1`, `hytale-server-2`).

### Project structure

```
hytale-server-manager/
├── src/                    # Go source code (if cloned)
│   ├── cmd/hytale-tui/    # TUI entry point
│   └── internal/
│       ├── tui/           # TUI layer (user interface)
│       └── hytale/        # Backend layer (server management)
├── tools/                  # Helper scripts (if cloned)
│   ├── release.sh         # GitHub release script
│   └── start.sh           # Development build script
├── data/                   # Server data (worlds, configs, logs)
├── install.sh             # Global installation script (build from source)
└── docs/                   # Documentation
```

## Hytale authentication

HSM uses OAuth2 authentication for Hytale servers:

1. **Device code flow** – The installation wizard guides you through authenticating with your Hytale account.
2. **Token storage** – OAuth tokens are stored securely in `data/` and refreshed automatically.
3. **Profile selection** – Choose which Hytale profile to use for server management.
4. **Game sessions** – Each server creates a game session for authentication.

See the installation wizard in the TUI for step-by-step authentication setup.

## Running the installer

Once prerequisites are in place, follow **Getting Started → Quick Start** to run the installer and bring up your first servers.
