package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/lipgloss"
)

// Progress message types
type progressMsg struct {
	percent float64
	label   string
}

type progressCompleteMsg struct {
	label string
}

// Progress model for displaying download/operation progress
type progressModel struct {
	progress progress.Model
	label    string
	percent  float64
	width    int
}

func newProgressModel(label string) progressModel {
	p := progress.New(progress.WithDefaultGradient())
	return progressModel{
		progress: p,
		label:    label,
		percent:  0.0,
		width:    40,
	}
}

func (m progressModel) Init() tea.Cmd {
	return nil
}

func (m progressModel) Update(msg tea.Msg) (progressModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width - 4
		m.progress.Width = m.width
		return m, nil

	case progressMsg:
		m.percent = msg.percent
		if msg.label != "" {
			m.label = msg.label
		}
		// Update the underlying progress bar
		// progress.Model.Update returns (tea.Model, tea.Cmd), need to assert back
		var cmd tea.Cmd
		updated, cmd := m.progress.Update(msg.percent)
		if p, ok := updated.(progress.Model); ok {
			m.progress = p
		}
		return m, cmd

	case progressCompleteMsg:
		m.percent = 1.0
		var cmd tea.Cmd
		updated, cmd := m.progress.Update(1.0)
		if p, ok := updated.(progress.Model); ok {
			m.progress = p
		}
		return m, cmd

	default:
		var cmd tea.Cmd
		updated, cmd := m.progress.Update(msg)
		if p, ok := updated.(progress.Model); ok {
			m.progress = p
		}
		return m, cmd
	}
}

func (m progressModel) View() string {
	if m.percent <= 0 {
		return ""
	}

	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252")).
		MarginBottom(1)

	percentStr := fmt.Sprintf("%.1f%%", m.percent*100)
	percentStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("205")).
		Bold(true)

	bar := m.progress.View()
	
	return fmt.Sprintf("%s\n%s %s\n",
		labelStyle.Render(m.label),
		bar,
		percentStyle.Render(percentStr),
	)
}

// Parse progress from wget output
// wget --progress=bar shows: "100%[=========>] 1,234,567  1.23M/s  in 5s"
func parseWgetProgress(line string) (float64, bool) {
	// Look for percentage pattern: "XX%" or "XXX%"
	if idx := strings.Index(line, "%"); idx > 0 {
		// Extract number before %
		start := idx - 1
		for start > 0 && (line[start] >= '0' && line[start] <= '9' || line[start] == '.') {
			start--
		}
		if start < idx-1 {
			var percent float64
			if _, err := fmt.Sscanf(line[start+1:idx], "%f", &percent); err == nil {
				if percent >= 0 && percent <= 100 {
					return percent / 100.0, true
				}
			}
		}
	}
	return 0, false
}

// Parse progress from curl output
// curl --progress-bar shows: "########## 100.0%"
func parseCurlProgress(line string) (float64, bool) {
	// Look for percentage pattern at end of line
	if idx := strings.LastIndex(line, "%"); idx > 0 {
		// Extract number before %
		start := idx - 1
		for start > 0 && (line[start] >= '0' && line[start] <= '9' || line[start] == '.') {
			start--
		}
		if start < idx-1 {
			var percent float64
			if _, err := fmt.Sscanf(line[start+1:idx], "%f", &percent); err == nil {
				if percent >= 0 && percent <= 100 {
					return percent / 100.0, true
				}
			}
		}
	}
	return 0, false
}

// Parse progress from rsync output
// rsync --info=progress2 shows: "1,234,567  50%  123.45M/s    0:00:05"
func parseRsyncProgress(line string) (float64, bool) {
	// Look for percentage pattern
	parts := strings.Fields(line)
	for _, part := range parts {
		if strings.HasSuffix(part, "%") {
			var percent float64
			if _, err := fmt.Sscanf(part, "%f%%", &percent); err == nil {
				if percent >= 0 && percent <= 100 {
					return percent / 100.0, true
				}
			}
		}
	}
	return 0, false
}

// Parse generic percentage from any line
func parseGenericProgress(line string) (float64, bool) {
	// Try wget format first
	if p, ok := parseWgetProgress(line); ok {
		return p, true
	}
	// Try curl format
	if p, ok := parseCurlProgress(line); ok {
		return p, true
	}
	// Try rsync format
	if p, ok := parseRsyncProgress(line); ok {
		return p, true
	}
	return 0, false
}
