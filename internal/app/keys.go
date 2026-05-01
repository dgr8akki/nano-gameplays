package app

import "github.com/charmbracelet/bubbles/key"

// KeyMap holds the global key bindings for the application.
type KeyMap struct {
	Quit    key.Binding
	Up      key.Binding
	Down    key.Binding
	Enter   key.Binding
	Back    key.Binding
	Confirm key.Binding
}

// DefaultKeyMap returns the standard bindings used across panes.
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "open/select"),
		),
		Back: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "back"),
		),
		Confirm: key.NewBinding(
			key.WithKeys("ctrl+s"),
			key.WithHelp("ctrl+s", "confirm"),
		),
	}
}

// Help returns the key bindings exposed in the footer in their default order.
func (k KeyMap) Help() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Enter, k.Confirm, k.Back, k.Quit}
}
