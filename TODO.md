# TODO List

This document tracks remaining implementation tasks and improvements for Hytale Server Manager.

**Reference:** All implementations must follow the [Hytale Server Manual](https://support.hytale.com/hc/en-us/articles/45326769420827-Hytale-Server-Manual) for official requirements and best practices.

## High Priority

### 1. hytale-downloader Integration

**Status:** ✅ Implementation Complete  
**Files:** `src/internal/hytale/downloader.go`, `src/internal/hytale/downloader_install.go`, `src/internal/hytale/bootstrap.go`, `src/internal/hytale/update.go`

Integrate the official `hytale-downloader` CLI tool for downloading and updating Hytale server files.

**Requirements:**
- Execute `hytale-downloader` command during bootstrap and updates
- Use OAuth credentials from `BootstrapConfig` (Client ID, Client Secret, or Access Token)
- Support authentication methods:
  - OAuth Client ID + Client Secret
  - OAuth Access Token (alternative)
- Handle `hytale-downloader` command output and progress
- Verify downloaded files (HytaleServer.jar, Assets.zip)
- Extract files to `master-install/` directory

**Implementation Notes:**
- OAuth fields are already collected in the wizard and passed to `BootstrapConfig`
- Reference: [Hytale Server Manual](https://support.hytale.com/hc/en-us/articles/45326769420827-Hytale-Server-Manual)
- Command format: `hytale-downloader -patchline production -output-dir <dir>`
- Authentication via `-credentials-path` or environment variables

**Tasks:**
- [x] Research hytale-downloader CLI options and authentication methods
- [x] Implement OAuth credential handling (Client ID/Secret or Access Token)
- [x] Add hytale-downloader execution in `BootstrapWithContextAndProgress()`
- [x] Add hytale-downloader execution in `UpdateGame()`
- [x] Basic progress tracking (generic feedback during download)
- [x] Handle authentication failures gracefully (falls back to existing files)
- [x] Add error handling for missing hytale-downloader binary
- [x] Auto-download hytale-downloader from official URL during installation
- [x] Device Code Flow support (hytale-downloader handles authentication interactively)
- [ ] Parse hytale-downloader output for detailed progress tracking (future enhancement)

---

### 1b. Game Session Token Creation

**Status:** ⚠️ Partial - Infrastructure Complete, API Integration Pending  
**Files:** `src/internal/hytale/session_tokens.go`, `src/internal/hytale/tmux.go`, `src/internal/tui/commands.go`

Implement game session creation API call to obtain session and identity tokens for server authentication. This allows servers to start authenticated automatically without manual `/auth login device` on each server.

**Requirements:**
- Call Hytale's `/game-session/new` endpoint with OAuth access token and profile UUID
- Store `sessionToken` and `identityToken` in `shared/.session-tokens.json`
- Integrate with bootstrap/wizard to create game session after OAuth authentication
- Handle token expiry (1 hour TTL) and refresh logic

**Current State:**
- ✅ `SessionTokens` struct and storage/loading functions implemented
- ✅ `Start()` and `StartAll()` accept and pass session tokens via `--session-token` and `--identity-token` flags
- ✅ TUI loads session tokens before starting servers (servers start authenticated if tokens exist)
- ⚠️ `CreateGameSession()` is a placeholder - needs actual API integration

**Implementation Notes:**
- Per Server Provider Authentication Guide: https://support.hytale.com/hc/en-us/articles/45328341414043
- Requires OAuth access token (from device code flow or Client ID/Secret)
- Requires profile UUID (obtained from `/my-account/get-profiles` endpoint)
- API endpoint: `POST /game-session/new` with `{"uuid": "<profileUUID>"}` in body
- Response includes `sessionToken` and `identityToken` JWTs
- Tokens expire after 1 hour - need refresh logic before expiry (5 minutes before)

**Tasks:**
- [ ] Research Hytale API endpoints and base URL
- [ ] Implement OAuth access token extraction from credentials file
- [ ] Implement profile list fetching (`GET /my-account/get-profiles`)
- [ ] Implement game session creation (`POST /game-session/new`)
- [ ] Add profile UUID selection to wizard (if multiple profiles)
- [ ] Integrate game session creation into bootstrap process
- [ ] Add token refresh logic (refresh 5 minutes before expiry)
- [ ] Handle token expiry gracefully (re-create session automatically)

---

### 2. Plugin Update Logic

**Status:** ✅ Implementation Complete  
**Files:** `src/internal/hytale/update.go`

Implement automatic plugin update functionality.

**Requirements:**
- Download plugin updates from source (e.g., GitHub releases)
- Update plugin files in `shared/mods/` directory
- Propagate updates to all server instances
- Optionally restart servers after plugin update

**Tasks:**
- [x] Implement plugin download and update logic (Performance Saver plugin)
- [x] Update `UpdatePlugins()` function
- [x] Propagate updates to all server instances via `CopySharedToServer()`
- [ ] Define plugin source configuration for multiple plugins (URL, version check)
- [ ] Add plugin version checking (check for newer versions before downloading)
- [ ] Add option to restart servers after plugin update (future enhancement)

---

## Medium Priority

### 3. Real-time Wizard Progress Display

**Status:** Partially Complete  
**Files:** `src/internal/tui/install_wizard.go`, `src/internal/tui/commands.go`

Currently, bootstrap progress is collected and shown in the receipt view after completion. Improve this to show real-time progress during installation.

**Current State:**
- Progress callback collects steps in `runBootstrapGo()`
- Steps are displayed in receipt view after completion
- Activity logs show during execution (last 4 lines)

**Desired State:**
- Display progress steps in wizard view during installation
- Show percentage and current step in wizard
- Update progress in real-time without blocking UI

**Challenges:**
- Bootstrap runs in blocking `func() tea.Msg` which can't send intermediate messages
- Would require refactoring to use channels/goroutines or progress message stream

**Tasks:**
- [ ] Research Bubble Tea patterns for real-time progress updates
- [ ] Refactor bootstrap to use progress message stream (channels)
- [ ] Update wizard view to display progress steps
- [ ] Add progress percentage display
- [ ] Ensure UI remains responsive during installation

---

## Low Priority / Future Enhancements

### 4. Config Editor UI

**Status:** Placeholder  
**Files:** `src/internal/tui/model.go` (viewEditServerConfigs)

Currently shows a placeholder message. Implement a full config editor in the TUI.

**Tasks:**
- [ ] Design config editor UI (JSON editor or form-based)
- [ ] Implement config file reading/writing with validation
- [ ] Add support for editing shared configs
- [ ] Add support for editing server-specific configs
- [ ] Add config validation and error handling
- [ ] Add config backup before editing

---

### 5. Advanced Server Configuration

**Status:** Not Started

Add more server configuration options to the wizard or config editor.

**Potential Options:**
- Custom world generation settings
- Advanced JVM tuning parameters
- Network configuration (bind address, firewall rules)
- Performance tuning (chunk loading, entity limits)
- Mod/plugin management UI

---

### 6. Backup Management UI

**Status:** Not Started

Add UI for managing backups (list, restore, delete).

**Tasks:**
- [ ] Add backup listing view
- [ ] Add backup restore functionality
- [ ] Add backup deletion functionality
- [ ] Add backup scheduling UI
- [ ] Show backup sizes and timestamps

---

## Code Quality / Maintenance

### 7. CLI Commands

**Status:** Placeholder  
**Files:** `src/cmd/hytale-tui/main.go`

Currently, CLI mode shows a placeholder message. Implement non-interactive CLI commands.

**Requirements:**
- Commands: `start`, `stop`, `restart`, `status`, `logs`
- Support for specifying server number or `all`
- JSON output option for scripting
- Error codes for automation

**Tasks:**
- [ ] Design CLI command structure
- [ ] Implement `start` command (server number or all)
- [ ] Implement `stop` command
- [ ] Implement `restart` command
- [ ] Implement `status` command (JSON output option)
- [ ] Implement `logs` command (follow option)
- [ ] Add CLI help/usage documentation

---

### 8. Remove Legacy Scale Up/Down Handlers

**Status:** Partially Complete  
**Files:** `src/internal/tui/model.go`

The `itemScaleUp` and `itemScaleDown` menu items are still in the menu, but they redirect to `viewServerSelection` which handles quantity selection. The direct handlers in `executeAction()` now have fallback implementations, but these could be removed if we verify the redirect always works.

**Tasks:**
- [ ] Verify `itemScaleUp`/`itemScaleDown` always go through `viewServerSelection`
- [ ] Remove fallback handlers if not needed
- [ ] Consider removing menu items if `itemAddServers`/`itemRemoveServers` are sufficient

---

### 8. Error Handling Improvements

**Status:** Ongoing

Improve error handling throughout the codebase.

**Areas:**
- More specific error messages
- Better error recovery
- User-friendly error display in TUI
- Logging for debugging

---

### 9. Testing

**Status:** Not Started

Add unit tests and integration tests.

**Tasks:**
- [ ] Set up testing framework
- [ ] Add unit tests for backend functions
- [ ] Add integration tests for bootstrap/update flows
- [ ] Add TUI component tests

---

## Documentation

### 10. Update Documentation for New Features

**Status:** Pending

Update documentation to reflect newly implemented features.

**Tasks:**
- [ ] Document "Add/Remove Servers" feature
- [ ] Document "Wipe Everything" feature
- [ ] Document activity logs feature
- [ ] Document OAuth authentication fields in wizard
- [ ] Document session token authentication (automatic server auth via --session-token/--identity-token)
- [ ] Document hytale-downloader auto-download feature
- [ ] Update troubleshooting guide

---

## Notes

- Most TODOs are documented in code comments where applicable
- This list is prioritized by user impact and implementation complexity
- Some items may be addressed as part of other work (e.g., hytale-downloader integration will naturally improve progress display)
