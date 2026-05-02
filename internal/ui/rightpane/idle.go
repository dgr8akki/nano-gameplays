package rightpane

import (
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/dgr8akki/nano-gameplays/internal/app"
)

// Idle renders the PS5-themed ASCII animation shown while the user is in the
// picker. The animation cycles at ~12 fps via a tea.Tick ticker.
type Idle struct {
	frame  int
	width  int
	height int
}

type idleTickMsg time.Time

const idleTickInterval = 83 * time.Millisecond

// NewIdle constructs the idle view.
func NewIdle() *Idle { return &Idle{} }

// Init implements app.IdleComponent.
func (i *Idle) Init() tea.Cmd { return idleTickCmd() }

// Update implements app.IdleComponent.
func (i *Idle) Update(msg tea.Msg) (app.IdleComponent, tea.Cmd) {
	if _, ok := msg.(idleTickMsg); ok {
		i.frame = (i.frame + 1) % len(idleFrames)
		return i, idleTickCmd()
	}
	return i, nil
}

// View implements app.IdleComponent.
func (i *Idle) View() string {
	w := i.width
	h := i.height
	if w <= 0 {
		w = 40
	}
	if h <= 0 {
		h = 10
	}
	body := lipgloss.NewStyle().
		Foreground(lipgloss.Color(beamColor(i.frame))).
		Bold(true).
		Render(idleFrames[i.frame])
	return lipgloss.NewStyle().
		Width(w).
		Height(h).
		Align(lipgloss.Center, lipgloss.Center).
		Render(body)
}

// SetSize implements app.IdleComponent.
func (i *Idle) SetSize(width, height int) {
	i.width = width
	i.height = height
}

func idleTickCmd() tea.Cmd {
	return tea.Tick(idleTickInterval, func(t time.Time) tea.Msg {
		return idleTickMsg(t)
	})
}

// beamColor cycles the accent color so the frames feel alive even when the
// underlying ASCII is a small set.
func beamColor(frame int) string {
	palette := []string{"81", "75", "69", "63", "99", "135", "171", "207"}
	return palette[frame%len(palette)]
}

// idleFrames is an 8-frame ASCII PS5 logo / pulsing-beams cycle. Crafted by
// hand to telegraph the PlayStation brand without using copyrighted glyphs.
var idleFrames = []string{
	idleFrame(`
        ┌─────────────────┐
        │                 │
        │     ░ PS5 ░     │
        │                 │
        └─────────────────┘
            · · · · ·
                 ·
`),
	idleFrame(`
        ┌─────────────────┐
        │       ░         │
        │     ▒ PS5 ▒     │
        │       ░         │
        └─────────────────┘
            ·  ·  ·  ·
                 ·
`),
	idleFrame(`
        ┌─────────────────┐
        │      ░ ░        │
        │    ▒▒ PS5 ▒▒    │
        │      ░ ░        │
        └─────────────────┘
           · ·   · · ·
                 · ·
`),
	idleFrame(`
        ┌─────────────────┐
        │     ░ ░ ░       │
        │   ▓▓▒ PS5 ▒▓▓   │
        │     ░ ░ ░       │
        └─────────────────┘
          · · · · · · ·
                · · ·
`),
	idleFrame(`
        ┌─────────────────┐
        │    ░ ░ ░ ░      │
        │  ▓▓▓▒ PS5 ▒▓▓▓  │
        │    ░ ░ ░ ░      │
        └─────────────────┘
        ·  ·  ·  ·  ·  ·
              · · · · ·
`),
	idleFrame(`
        ┌─────────────────┐
        │     ░ ░ ░       │
        │   ▓▓▒ PS5 ▒▓▓   │
        │     ░ ░ ░       │
        └─────────────────┘
          · · · · · · ·
                · · ·
`),
	idleFrame(`
        ┌─────────────────┐
        │      ░ ░        │
        │    ▒▒ PS5 ▒▒    │
        │      ░ ░        │
        └─────────────────┘
           · ·   · · ·
                 · ·
`),
	idleFrame(`
        ┌─────────────────┐
        │       ░         │
        │     ▒ PS5 ▒     │
        │       ░         │
        └─────────────────┘
            ·  ·  ·  ·
                 ·
`),
}

// idleFrame trims a leading newline so frames can be authored as multi-line
// raw strings beginning on a fresh line.
func idleFrame(s string) string { return strings.TrimPrefix(s, "\n") }
