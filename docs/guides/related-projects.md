---
title: Related Projects
---

# Related Projects

Hytale Server Manager is a standalone tool for managing Hytale dedicated servers. This page lists related projects and tools that complement HSM.

## Official Hytale Tools

### hytale-downloader

**URL:** [Hytale Server Manual](https://support.hytale.com/hc/en-us/articles/45326769420827-Hytale-Server-Manual)

**What it is:**  
The official Hytale CLI tool for downloading and updating Hytale server files.

**Key capabilities:**

- **OAuth authentication** – Integrates with Hytale's authentication system.
- **Server file downloads** – Downloads server JAR and assets from Hytale's patchline servers.
- **Version management** – Tracks downloaded versions and handles updates.

**How it works with HSM:**

- HSM uses `hytale-downloader` under the hood for all server file downloads.
- Authentication tokens are shared between HSM and `hytale-downloader`.
- The installation wizard guides you through setting up `hytale-downloader` if needed.

### hytale-auth

**What it is:**  
A helper tool for managing Hytale OAuth authentication.

**How it works with HSM:**

- HSM integrates with Hytale's OAuth authentication during installation.
- The installation wizard guides you through authentication setup.
- Authentication tokens are stored securely in `data/auth.enc` and `data/.cache/.hytale-downloader-credentials.json`.
- You can re-authenticate via the TUI installation wizard if needed.

## Community Resources

### Official Hytale Documentation

- **[Hytale Server Manual](https://support.hytale.com/hc/en-us/articles/45326769420827-Hytale-Server-Manual)** – Official server setup and administration guide.
- **[Server Provider Authentication Guide](https://support.hytale.com/hc/en-us/articles/45328341414043-Server-Provider-Authentication-Guide)** – Detailed authentication setup instructions.

### Hytale Community

- **[Hytale Forums](https://hytale.com/)** – Official Hytale community forums.
- **[Hytale Discord](https://discord.gg/hytale)** – Official Hytale Discord server.

---

If you're building tools that integrate with HSM, or want to list your Hytale-related project here, please open an issue or pull request!
