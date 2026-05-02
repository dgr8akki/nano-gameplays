// Package header renders the application header bar.
package header

import "github.com/charmbracelet/lipgloss"

// Model is the header view.
type Model struct{}

// New returns a fresh header.
func New() *Model { return &Model{} }

var bannerStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("231")).
	Background(lipgloss.Color("57")).
	Border(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("99")).
	Padding(0, 2)

// View renders the header at the supplied width.
func (m *Model) View(width int) string {
	banner := bannerStyle.Render("nano-gameplays")
	return lipgloss.NewStyle().
		Width(width).
		Align(lipgloss.Center).
		Render(banner)
}
