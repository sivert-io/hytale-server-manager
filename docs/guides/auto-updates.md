# Auto Updates

Hytale Server Manager includes automated update capabilities for Hytale server files.

## What gets automated

The update system:

- Downloads latest Hytale server files using `hytale-downloader`.
- Extracts and verifies server files automatically.
- Preserves your `data/` directory (worlds, configs, tokens).
- Can be triggered manually via TUI or automatically in the future.

## How updates work

Hytale server files are managed via `hytale-downloader`:

1. **OAuth authentication** – Uses your Hytale account tokens (stored in `data/.cache/.hytale-downloader-credentials.json`).
2. **Download** – Fetches latest server files from Hytale's patchline servers.
3. **Extraction** – Extracts server JAR and assets to `data/Server/` and `data/Assets.zip`.
4. **Verification** – Ensures required files (HytaleServer.jar, Assets.zip) are present.

## Updating via TUI

The easiest way to update:

```bash
sudo hsm
```

Then:

1. Navigate to the **Updates** tab.
2. Select **"Update Game"** to download latest server files.
3. Restart servers after update completes (via **Servers** tab).

## Manual updates

You can also update manually using `hytale-downloader`:

```bash
# Ensure you're authenticated
hytale-downloader login

# Download latest files (uses stored credentials)
hytale-downloader -patchline production -credentials-path data/.cache/.hytale-downloader-credentials.json

# Files will be downloaded to data/.cache/, then extract them
unzip -o data/.cache/*.zip -d data/
```

## Update safety

- **World data preserved** – Your `data/universe/` directory is never touched during updates.
- **Configs preserved** – `config.json`, `permissions.json`, etc. remain intact.
- **Tokens preserved** – OAuth tokens in `auth.enc` and `.cache/.hytale-downloader-credentials.json` are preserved.
- **Backup recommended** – Before major updates, consider backing up your `data/` directory.

## Version tracking

HSM tracks the downloaded server version in:

- `data/.version_info` – Contains patchline and version information.

You can check the current version:

```bash
cat data/.version_info
```

## Future: Automatic update monitoring

Future versions may include automatic update checking and monitoring, similar to auto-update monitors in other server managers.
