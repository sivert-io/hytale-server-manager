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
	label    string
	value    string
	input    textinput.Model
	kind     fieldKind
	required bool
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
				label:    "Number of Servers",
				value:    fmt.Sprintf("%d", hytale.DefaultNumServers),
				kind:     fieldNumber,
				required: true,
			},
			{
				label:    "Base Port",
				value:    fmt.Sprintf("%d", hytale.DefaultBasePort),
				kind:     fieldNumber,
				required: true,
			},
			{
				label:    "Hostname Prefix",
				value:    hytale.DefaultHostnamePrefix,
				kind:     fieldText,
				required: true,
			},
			{
				label:    "Max Players",
				value:    fmt.Sprintf("%d", hytale.DefaultMaxPlayers),
				kind:     fieldNumber,
				required: true,
			},
			{
				label:    "JVM Arguments",
				value:    hytale.DefaultJVMArgs,
				kind:     fieldText,
				required: false,
			},
			{
				label:    "Admin Password (optional)",
				value:    "",
				kind:     fieldPassword,
				required: false,
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
	}

	return w
}

func (w installWizard) Update(msg tea.Msg) (installWizard, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if w.cursor > 0 {
				w.cursor--
			}

		case "down", "j":
			if w.cursor < len(w.fields)-1 {
				w.cursor++
			}

		case "enter":
			if w.cursor == len(w.fields) {
				// Submit button
				return w, w.submit()
			} else {
				// Move to next field
				if w.cursor < len(w.fields)-1 {
					w.cursor++
				}
			}

		case "esc":
			// Cancel - signal to return to main menu
			// We'll handle this in the model.Update() by checking for wizardCancelMsg
			return w, func() tea.Msg {
				return wizardCancelMsg{}
			}

		default:
			// Update current field
			if w.cursor < len(w.fields) {
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

		value := field.input.View()
		if i == w.cursor {
			value = selectedStyle.Render(value)
		}

		s += fmt.Sprintf("%s%s: %s\n", cursor, label, value)
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
	numServers, _ := strconv.Atoi(w.fields[0].value)
	basePort, _ := strconv.Atoi(w.fields[1].value)
	hostnamePrefix := w.fields[2].value
	maxPlayers, _ := strconv.Atoi(w.fields[3].value)
	jvmArgs := w.fields[4].value
	adminPassword := w.fields[5].value

	// Create bootstrap config
	cfg := hytale.BootstrapConfig{
		HytaleUser:     hytale.DefaultHytaleUser,
		NumServers:     numServers,
		BasePort:       basePort,
		QueryPort:      basePort + 1,
		HostnamePrefix: hostnamePrefix,
		AdminPassword:  adminPassword,
		MaxPlayers:     maxPlayers,
		JVMArgs:        jvmArgs,
	}

	// Run bootstrap
	return runBootstrapGo(cfg)
}
