// Package footer renders the contextual help footer.
package footer

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"

	"github.com/dgr8akki/nano-gameplays/internal/app"
)

// Model is the footer view; it adapts its bindings to the current AppState.
type Model struct {
	keys app.KeyMap
	help help.Model
}

// New returns a footer wired against the supplied keymap.
func New(keys app.KeyMap) *Model {
	h := help.New()
	h.ShortSeparator = "  •  "
	h.Styles.ShortKey = h.Styles.ShortKey.Foreground(lipgloss.Color("99"))
	h.Styles.ShortDesc = h.Styles.ShortDesc.Foreground(lipgloss.Color("245"))
	h.Styles.ShortSeparator = h.Styles.ShortSeparator.Foreground(lipgloss.Color("238"))
	return &Model{keys: keys, help: h}
}

// View renders the footer line at the supplied width and state.
func (m *Model) View(width int, state app.AppState) string {
	m.help.Width = width
	bindings := m.bindingsFor(state)
	return lipgloss.NewStyle().Width(width).Render(m.help.ShortHelpView(bindings))
}

func (m *Model) bindingsFor(state app.AppState) []key.Binding {
	switch state {
	case app.StatePicker:
		return []key.Binding{m.keys.Up, m.keys.Down, m.keys.Enter, m.keys.Quit}
	case app.StateAnalyzing:
		return []key.Binding{m.keys.Quit}
	case app.StateEditing:
		return []key.Binding{m.keys.Confirm, m.keys.Back, m.keys.Quit}
	}
	return []key.Binding{m.keys.Quit}
}
