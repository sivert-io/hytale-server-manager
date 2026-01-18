# Troubleshooting

This page collects common issues and how to diagnose them.

## Server won't start

Check server status and logs via the TUI:

```bash
sudo hsm
```

Then navigate to:
- **Tools Tab → View Server Status**: See detailed status of all servers (running/stopped, ports, sessions)
- **Servers Tab → View Server Logs**: View logs for specific servers

Or attach directly to tmux session:

```bash
tmux attach-session -t hytale-server-1
```

Check for:

- Missing or invalid server files in `data/Server/HytaleServer.jar` or `data/Assets.zip`.
- Java version issues (requires Java 25+).
- Port conflicts if you changed defaults.
- Authentication errors (missing or expired OAuth tokens).

## Java version issues

HSM requires Java 25+. Check your version:

```bash
java --version
```

If Java 25+ is not installed:

1. The installation wizard will auto-install Java 25 (Eclipse Temurin via Adoptium).
2. Or install manually: Visit https://adoptium.net/

## Authentication errors

If you see authentication errors:

1. **Check tokens exist**:
   ```bash
   ls -la data/auth.enc
   ls -la data/.cache/.hytale-downloader-credentials.json
   ```

2. **Re-authenticate via TUI**:
   - Run `sudo hsm`
   - Go to **Install** tab
   - Re-run the installation wizard (or update authentication)

3. **Manual authentication**: Use `hytale-downloader` directly or the TUI installation wizard to re-authenticate.

## hytale-downloader not found

If the installer reports `hytale-downloader` is missing:

1. **Check if installed**:
   ```bash
   which hytale-downloader
   ```

2. **Install manually**:
   - Visit: https://support.hytale.com/hc/en-us/articles/45326769420827-Hytale-Server-Manual
   - Download the latest `hytale-downloader` binary
   - Install to `/usr/local/bin/hytale-downloader`

3. **Verify**:
   ```bash
   hytale-downloader -version
   ```

## Can't connect to server

Check:

- **Firewall rules**: Ensure UDP port 5520 (and other server ports) are open and forwarded on your router.
- **Server running**: Check status via TUI (`sudo hsm`) or `tmux ls`.
- **Network**: Hytale uses QUIC over UDP (not TCP) – ensure UDP traffic is allowed.
- **Server logs**: Check `data/logs/` for errors.

## Port conflicts

If you see port binding errors:

1. **Check what's using the port**:
   ```bash
   sudo netstat -tulpn | grep 5520
   ```

2. **Change base port**: Configure a different base port in the installation wizard.

3. **Kill conflicting processes**:
   ```bash
   sudo killall java  # Be careful - kills all Java processes
   ```

## Permission errors

If you encounter permission errors:

- Ensure you're running with `sudo hsm`.
- Check that `data/` directory is writable.
- Verify tmux is installed and accessible: `tmux -V`

## When in doubt

- **Re-run the installer**: Use the TUI installation wizard to repair a broken installation.
- **Check logs**: Review `data/logs/` for detailed error messages.
- **Verify files**: Ensure `data/Server/HytaleServer.jar` and `data/Assets.zip` exist.
- **Check tmux logs**: Attach to tmux session with `tmux attach-session -t hytale-server-1` to see full output.
