// Package rightpane renders the right-half views: idle, analyzing, metadata.
package rightpane

import (
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/dgr8akki/nano-gameplays/internal/app"
)

// Analyzing renders the spinner and "Analyzing…" label shown while a Claude
// call is in flight.
type Analyzing struct {
	spinner spinner.Model
	width   int
	height  int
}

// NewAnalyzing constructs the analyzing view.
func NewAnalyzing() *Analyzing {
	sp := spinner.New(spinner.WithSpinner(spinner.Pulse))
	sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("99"))
	return &Analyzing{spinner: sp}
}

// Init implements app.AnalyzingComponent.
func (a *Analyzing) Init() tea.Cmd { return a.spinner.Tick }

// Update implements app.AnalyzingComponent.
func (a *Analyzing) Update(msg tea.Msg) (app.AnalyzingComponent, tea.Cmd) {
	var cmd tea.Cmd
	a.spinner, cmd = a.spinner.Update(msg)
	return a, cmd
}

// View implements app.AnalyzingComponent.
func (a *Analyzing) View() string {
	label := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("99")).
		Render("Analyzing…")
	body := lipgloss.JoinVertical(lipgloss.Center, label, a.spinner.View())
	w := a.width
	h := a.height
	if w <= 0 {
		w = 40
	}
	if h <= 0 {
		h = 10
	}
	return lipgloss.NewStyle().
		Width(w).
		Height(h).
		Align(lipgloss.Center, lipgloss.Center).
		Render(body)
}

// SetSize implements app.AnalyzingComponent.
func (a *Analyzing) SetSize(width, height int) {
	a.width = width
	a.height = height
}
