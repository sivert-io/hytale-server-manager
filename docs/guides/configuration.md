# Configuration

Hytale Server Manager stores server data in the `data/` directory. This page explains how configuration is structured and how to safely customize servers.

## Installation methods

You can install with one of two common flows:

### 1. Install from GitHub releases (recommended)

```bash
# One-line installation from GitHub releases
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
sudo hsm
```

This installs the `hsm` binary globally. Server data is stored in a `data/` directory relative to where you run `hsm`.

### 2. Build from source

```bash
git clone https://github.com/sivert-io/hytale-server-manager.git
cd hytale-server-manager

# Build and install globally (requires sudo)
sudo ./install.sh

# Run the TUI
sudo hsm
```

## Data directory structure

The `data/` directory contains all server data:

```text
data/
├── Server/                  # Server JAR and assets (master-install)
├── Assets.zip              # Hytale assets file
├── config.json             # Server configuration
├── permissions.json        # Player permissions
├── auth.enc                # OAuth authentication tokens
├── whitelist.json          # Server whitelist
├── bans.json               # Server bans
├── logs/                   # Server logs
├── universe/               # World data
├── mods/                   # Server mods/plugins
└── .cache/                 # Cache and temporary files
```

Files in `data/` are **never deleted** during updates, so your worlds, configs, and settings persist.

## Ports and networking

Default ports (incrementing from base port):

| Server | Port  |
| ------ | ----- |
| 1      | 5520  |
| 2      | 5521  |
| 3      | 5522  |

- **Default base port**: 5520
- You can adjust ports in the installation wizard or server configuration.
- Each server runs in its own tmux session (e.g., `hytale-server-1`).

## JVM arguments

Default JVM memory settings:

- **Initial heap**: `-Xms4G`
- **Maximum heap**: `-Xmx8G`

You can customize JVM arguments in the installation wizard during setup.

## Best practices

- Keep all long-term customizations inside `data/`.
- Use a git repo for your `data/` directory so you can version changes (but exclude sensitive files like `auth.enc`).
- Avoid editing server files directly unless testing something temporarily.
- After changing configs, restart the relevant server(s) via the TUI or manually.
- Backup `data/` regularly, especially `universe/` (world data) and `auth.enc` (tokens).
