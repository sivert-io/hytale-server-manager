package tui

import (
	"fmt"
	"strconv"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/sivert-io/hytale-server-manager/src/internal/hytale"
)

type installWizard struct {
	cursor    int
	fields    []wizardField
	submitted bool
}

type wizardField struct {
	label       string
	value       string
	input       textinput.Model
	kind        fieldKind
	required    bool
	description string // Help text shown below field
}

type fieldKind int

const (
	fieldText fieldKind = iota
	fieldNumber
	fieldPassword
	fieldToggle
)

func newInstallWizard() installWizard {
	w := installWizard{
		fields: []wizardField{
			{
				label:       "System User",
				value:       hytale.DefaultHytaleUser,
				kind:        fieldText,
				required:    true,
				description: "Linux system user to run servers as (will be created if it doesn't exist)",
			},
			{
				label:       "Number of Servers",
				value:       fmt.Sprintf("%d", hytale.DefaultNumServers),
				kind:        fieldNumber,
				required:    true,
				description: "How many server instances to create (ports will increment from base port)",
			},
			{
				label:       "Base Port",
				value:       fmt.Sprintf("%d", hytale.DefaultBasePort),
				kind:        fieldNumber,
				required:    true,
				description: fmt.Sprintf("Starting port (Server 1 = %d, Server 2 = %d, Server 3 = %d, ...)", hytale.DefaultBasePort, hytale.DefaultBasePort+1, hytale.DefaultBasePort+2),
			},
			{
				label:       "Hostname Prefix",
				value:       hytale.DefaultHostnamePrefix,
				kind:        fieldText,
				required:    true,
				description: "Server display name prefix (e.g., 'hytale' → servers show as 'hytale-1', 'hytale-2', etc.)",
			},
			{
				label:       "Max Players",
				value:       fmt.Sprintf("%d", hytale.DefaultMaxPlayers),
				kind:        fieldNumber,
				required:    true,
				description: "Maximum players per server instance",
			},
			{
				label:       "Max View Radius (chunks)",
				value:       fmt.Sprintf("%d", hytale.DefaultMaxViewRadius),
				kind:        fieldNumber,
				required:    true,
				description: "View distance in chunks (12 recommended for performance, default is 32). Lower = better performance",
			},
			{
				label:       "Game Mode",
				value:       hytale.DefaultGameMode,
				kind:        fieldToggle,
				required:    false,
				description: "Default game mode: Adventure, Survival, or Creative (Press Enter to toggle)",
			},
			{
				label:       "Server Password (optional)",
				value:       "",
				kind:        fieldPassword,
				required:    false,
				description: "Server join password (leave empty for public server)",
			},
			{
				label:       "JVM Arguments",
				value:       hytale.DefaultJVMArgs,
				kind:        fieldText,
				required:    false,
				description: "Java VM arguments (memory, GC, AOT cache). Recommended: -Xms6G -Xmx6G -XX:+UseG1GC -XX:AOTCache=HytaleServer.aot",
			},
			{
				label:       "Enable Backups",
				value:       "Yes",
				kind:        fieldToggle,
				required:    false,
				description: "Enable automatic server backups (Press Enter to toggle)",
			},
			{
				label:       "Backup Frequency (minutes)",
				value:       fmt.Sprintf("%d", hytale.DefaultBackupFrequency),
				kind:        fieldNumber,
				required:    false,
				description: "How often to create backups (60 minutes recommended)",
			},
			{
				label:       "OAuth Client ID (optional)",
				value:       "",
				kind:        fieldText,
				required:    false,
				description: "OAuth client ID for advanced authentication. hytale-downloader uses device code flow (browser auth) by default. See: https://support.hytale.com/hc/en-us/articles/45328341414043",
			},
			{
				label:       "OAuth Client Secret (optional)",
				value:       "",
				kind:        fieldPassword,
				required:    false,
				description: "OAuth client secret (used with Client ID). Leave empty to use device code flow instead.",
			},
			{
				label:       "OAuth Access Token (optional)",
				value:       "",
				kind:        fieldPassword,
				required:    false,
				description: "Pre-obtained OAuth access token (alternative to Client ID/Secret). If empty, hytale-downloader will prompt for device code auth.",
			},
		},
	}

	// Initialize text inputs
	for i := range w.fields {
		w.fields[i].input = textinput.New()
		w.fields[i].input.Placeholder = w.fields[i].value
		w.fields[i].input.SetValue(w.fields[i].value)
		if w.fields[i].kind == fieldPassword {
			w.fields[i].input.EchoMode = textinput.EchoPassword
		}
		// Focus the first input
		if i == 0 {
			w.fields[i].input.Focus()
		} else {
			w.fields[i].input.Blur()
		}
	}

	return w
}

func (w installWizard) Update(msg tea.Msg) (installWizard, tea.Cmd) {
	var cmds []tea.Cmd

	// For number fields, filter out non-digit characters from key input
	if w.cursor < len(w.fields) && w.fields[w.cursor].kind == fieldNumber {
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			if keyMsg.Type == tea.KeyRunes {
				// Filter out non-digit characters
				digitsOnly := []rune{}
				for _, r := range keyMsg.Runes {
					if r >= '0' && r <= '9' {
						digitsOnly = append(digitsOnly, r)
					}
				}
				// If we filtered out some characters, create a new key message with only digits
				if len(digitsOnly) < len(keyMsg.Runes) {
					if len(digitsOnly) == 0 {
						// No valid digits, ignore input
						return w, nil
					}
					// Replace msg with filtered version
					keyMsg.Runes = digitsOnly
					msg = keyMsg
				}
			}
		}
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle Enter key BEFORE passing to text input (for toggle fields)
		if msg.String() == "enter" {
			if w.cursor == len(w.fields) {
				// Submit button
				return w, w.submit()
			} else if w.cursor < len(w.fields) && w.fields[w.cursor].kind == fieldToggle {
				// Toggle field - cycle through options
				currentValue := w.fields[w.cursor].value
				
				// Determine which toggle field this is
				if w.fields[w.cursor].label == "Game Mode" {
					gameModes := []string{"Adventure", "Survival", "Creative"}
					// Find current index
					currentIndex := 0
					for i, mode := range gameModes {
						if mode == currentValue {
							currentIndex = i
							break
						}
					}
					// Cycle to next mode
					nextIndex := (currentIndex + 1) % len(gameModes)
					w.fields[w.cursor].value = gameModes[nextIndex]
					w.fields[w.cursor].input.SetValue(gameModes[nextIndex])
				} else if w.fields[w.cursor].label == "Enable Backups" {
					// Toggle Yes/No
					if currentValue == "Yes" {
						w.fields[w.cursor].value = "No"
						w.fields[w.cursor].input.SetValue("No")
					} else {
						w.fields[w.cursor].value = "Yes"
						w.fields[w.cursor].input.SetValue("Yes")
					}
				}
				
				return w, nil
			} else {
				// Move to next field and focus it
				if w.cursor < len(w.fields)-1 {
					// Blur current field
					w.fields[w.cursor].input.Blur()
					w.cursor++
					// Focus next field
					w.fields[w.cursor].input.Focus()
				}
			}
		}
		
		switch msg.String() {
		case "up", "k":
			if w.cursor > 0 {
				// Blur current field
				if w.cursor < len(w.fields) {
					w.fields[w.cursor].input.Blur()
				}
				w.cursor--
				// Focus new field
				if w.cursor < len(w.fields) {
					w.fields[w.cursor].input.Focus()
				}
			}

		case "down", "j":
			// Allow navigation to submit button (len(w.fields))
			if w.cursor < len(w.fields) {
				// Blur current field if it's an input field
				if w.cursor < len(w.fields) {
					w.fields[w.cursor].input.Blur()
				}
				w.cursor++
				// Focus new field if it's an input field (not submit button)
				if w.cursor < len(w.fields) {
					w.fields[w.cursor].input.Focus()
				}
				// If cursor is now at submit button (len(w.fields)), don't try to focus anything
			}

		case "esc":
			// Cancel - signal to return to main menu
			// We'll handle this in the model.Update() by checking for wizardCancelMsg
			return w, func() tea.Msg {
				return wizardCancelMsg{}
			}

		default:
			// Update current field (validation already done above)
			// Skip input updates for toggle fields (they're handled by Enter key)
			if w.cursor < len(w.fields) && w.fields[w.cursor].kind != fieldToggle {
				var cmd tea.Cmd
				w.fields[w.cursor].input, cmd = w.fields[w.cursor].input.Update(msg)
				w.fields[w.cursor].value = w.fields[w.cursor].input.Value()
				cmds = append(cmds, cmd)
			}
		}
	}

	return w, tea.Batch(cmds...)
}

func (w installWizard) View() string {
	var s string

	s += titleStyle.Render(" 🚀 Installation Wizard") + "\n\n"
	s += dimmedStyle.Render("Configure your Hytale server installation") + "\n\n"

	// Form fields
	for i, field := range w.fields {
		cursor := "  "
		if i == w.cursor {
			cursor = selectedStyle.Render("▶ ")
		}

		label := field.label
		if field.required {
			label += " *"
		}

		var value string
		if field.kind == fieldToggle {
			// For toggle fields, show the value directly (not from input)
			value = field.value
			if i == w.cursor {
				value = selectedStyle.Render(value)
			}
		} else {
			value = field.input.View()
			if i == w.cursor {
				value = selectedStyle.Render(value)
			}
		}

		s += fmt.Sprintf("%s%s: %s\n", cursor, label, value)
		
		// Show description for focused field
		if i == w.cursor && field.description != "" {
			s += "   " + dimmedStyle.Render(field.description) + "\n"
		}
	}

	// Submit button
	cursor := "  "
	if w.cursor == len(w.fields) {
		cursor = selectedStyle.Render("▶ ")
	}
	submitText := "Install"
	if w.cursor == len(w.fields) {
		submitText = selectedStyle.Render(submitText)
	}
	s += fmt.Sprintf("\n%s%s\n", cursor, submitText)

	s += "\n" + dimmedStyle.Render("↑/↓: Navigate  |  Enter: Select/Next  |  Esc: Cancel")

	return lipgloss.NewStyle().Padding(1, 2).Render(s)
}

func (w installWizard) submit() tea.Cmd {
	// Validate fields
	for _, field := range w.fields {
		if field.required && field.value == "" {
			return func() tea.Msg {
				return commandFinishedMsg{
					output: "",
					err:    fmt.Errorf("field '%s' is required", field.label),
				}
			}
		}
	}

	// Parse values
	hytaleUser := w.fields[0].value
	numServers, _ := strconv.Atoi(w.fields[1].value)
	basePort, _ := strconv.Atoi(w.fields[2].value)
	hostnamePrefix := w.fields[3].value
	maxPlayers, _ := strconv.Atoi(w.fields[4].value)
	maxViewRadius, _ := strconv.Atoi(w.fields[5].value)
	gameMode := w.fields[6].value
	serverPassword := w.fields[7].value
	jvmArgs := w.fields[8].value
	backupEnabled := w.fields[9].value == "Yes"
	backupFrequency, _ := strconv.Atoi(w.fields[10].value)
	oauthClientID := w.fields[11].value
	oauthClientSecret := w.fields[12].value
	oauthAccessToken := w.fields[13].value
	
	// Validate numServers
	if numServers < 1 {
		return func() tea.Msg {
			return commandFinishedMsg{
				output: "",
				err:    fmt.Errorf("number of servers must be at least 1"),
			}
		}
	}
	
	// Enforce server limit per Hytale Server Manual
	// Default limit: 100 servers per game license
	if numServers > hytale.MaxServersPerLicense {
		return func() tea.Msg {
			return commandFinishedMsg{
				output: "",
				err:    fmt.Errorf("maximum %d servers allowed per game license (Hytale Server Manual). Additional licenses or Server Provider account required for more", hytale.MaxServersPerLicense),
			}
		}
	}
	
	// Validate basePort
	if basePort < 1024 || basePort > 65535 {
		return func() tea.Msg {
			return commandFinishedMsg{
				output: "",
				err:    fmt.Errorf("base port must be between 1024 and 65535"),
			}
		}
	}
	
	// Validate maxViewRadius
	if maxViewRadius < 1 || maxViewRadius > 32 {
		return func() tea.Msg {
			return commandFinishedMsg{
				output: "",
				err:    fmt.Errorf("max view radius must be between 1 and 32 chunks"),
			}
		}
	}
	
	// Validate gameMode
	if gameMode == "" {
		gameMode = hytale.DefaultGameMode
	}
	validGameModes := map[string]bool{"Adventure": true, "Survival": true, "Creative": true}
	if !validGameModes[gameMode] {
		return func() tea.Msg {
			return commandFinishedMsg{
				output: "",
				err:    fmt.Errorf("game mode must be Adventure, Survival, or Creative"),
			}
		}
	}

	// Validate backup frequency
	if backupFrequency < 1 || backupFrequency > 1440 {
		return func() tea.Msg {
			return commandFinishedMsg{
				output: "",
				err:    fmt.Errorf("backup frequency must be between 1 and 1440 minutes (24 hours)"),
			}
		}
	}

	// Validate system user (basic validation - alphanumeric and underscore)
	if hytaleUser == "" {
		hytaleUser = hytale.DefaultHytaleUser
	}

	// Create bootstrap config
	cfg := hytale.BootstrapConfig{
		HytaleUser:        hytaleUser,
		NumServers:        numServers,
		BasePort:          basePort,
		QueryPort:         basePort + 1,
		HostnamePrefix:    hostnamePrefix,
		ServerPassword:    serverPassword,
		MaxPlayers:        maxPlayers,
		MaxViewRadius:     maxViewRadius,
		GameMode:          gameMode,
		JVMArgs:           jvmArgs,
		BackupEnabled:     backupEnabled,
		BackupFrequency:   backupFrequency,
		OAuthClientID:     oauthClientID,
		OAuthClientSecret: oauthClientSecret,
		OAuthAccessToken:  oauthAccessToken,
	}

	// Run bootstrap
	return runBootstrapGo(cfg)
}
