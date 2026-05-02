package app

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// View renders the four-region layout: header on top, footer on bottom,
// left/right panes filling the body.
func (m Model) View() string {
	if m.Width == 0 || m.Height == 0 {
		return ""
	}

	header := ""
	if m.Header != nil {
		header = m.Header.View(m.Width)
	}
	footer := ""
	if m.Footer != nil {
		footer = m.Footer.View(m.Width, m.State)
	}

	headerH := lipgloss.Height(header)
	footerH := lipgloss.Height(footer)
	bodyH := m.Height - headerH - footerH
	if bodyH < 1 {
		bodyH = 1
	}
	leftW := m.Width / 2
	rightW := m.Width - leftW

	leftBody := ""
	if m.Picker != nil {
		leftBody = m.Picker.View()
	}
	leftBlock := lipgloss.NewStyle().Width(leftW).Height(bodyH).Render(leftBody)

	rightBody := ""
	switch m.State {
	case StatePicker:
		if m.Idle != nil {
			rightBody = m.Idle.View()
		} else {
			rightBody = "right pane"
		}
	case StateAnalyzing:
		if m.Analyzing != nil {
			rightBody = m.Analyzing.View()
		}
	case StateEditing, StateDone:
		if m.Metadata != nil {
			rightBody = m.Metadata.View()
		}
	}
	rightBlock := lipgloss.NewStyle().Width(rightW).Height(bodyH).Render(rightBody)

	body := lipgloss.JoinHorizontal(lipgloss.Top, leftBlock, rightBlock)
	parts := []string{}
	if header != "" {
		parts = append(parts, header)
	}
	parts = append(parts, body)
	if footer != "" {
		parts = append(parts, footer)
	}
	return strings.Join(parts, "\n")
}
