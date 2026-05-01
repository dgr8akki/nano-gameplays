package rightpane

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/dgr8akki/nano-gameplays/internal/app"
)

// Metadata renders the editable metadata form.
type Metadata struct {
	titleInput textinput.Model
	sceneInput textinput.Model
	confidence string
	errBanner  string
	focusIndex int
	width      int
	height     int
}

// NewMetadata constructs the metadata form.
func NewMetadata() *Metadata {
	title := textinput.New()
	title.Placeholder = "Game title"
	title.Prompt = "› "
	title.CharLimit = 100

	scene := textinput.New()
	scene.Placeholder = "Scene / level / mode"
	scene.Prompt = "› "
	scene.CharLimit = 200

	return &Metadata{
		titleInput: title,
		sceneInput: scene,
		confidence: "low",
	}
}

// Update implements app.MetadataComponent.
func (m *Metadata) Update(msg tea.Msg) (app.MetadataComponent, tea.Cmd) {
	if k, ok := msg.(tea.KeyMsg); ok {
		switch k.String() {
		case "tab", "shift+tab", "down", "up":
			m.toggleFocus()
			return m, nil
		}
	}

	var cmd tea.Cmd
	if m.focusIndex == 0 {
		m.titleInput, cmd = m.titleInput.Update(msg)
	} else {
		m.sceneInput, cmd = m.sceneInput.Update(msg)
	}
	return m, cmd
}

func (m *Metadata) toggleFocus() {
	if m.focusIndex == 0 {
		m.titleInput.Blur()
		m.sceneInput.Focus()
		m.focusIndex = 1
	} else {
		m.sceneInput.Blur()
		m.titleInput.Focus()
		m.focusIndex = 0
	}
}

// View implements app.MetadataComponent.
func (m *Metadata) View() string {
	rows := []string{}
	if m.errBanner != "" {
		rows = append(rows,
			lipgloss.NewStyle().
				Foreground(lipgloss.Color("203")).
				Bold(true).
				Render("! "+m.errBanner),
			"",
		)
	}
	labelStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("250"))
	rows = append(rows,
		labelStyle.Render("Game title"),
		m.titleInput.View(),
		"",
		labelStyle.Render("Scene / level / mode"),
		m.sceneInput.View(),
		"",
		labelStyle.Render("Confidence"),
		renderConfidenceBadge(m.confidence),
	)
	if !m.HasGameTitle() {
		rows = append(rows, "",
			lipgloss.NewStyle().
				Foreground(lipgloss.Color("245")).
				Italic(true).
				Render("game title required to confirm (ctrl+s)"),
		)
	}
	body := strings.Join(rows, "\n")
	w := m.width
	if w <= 0 {
		w = 40
	}
	h := m.height
	if h <= 0 {
		h = 10
	}
	return lipgloss.NewStyle().
		Width(w).
		Height(h).
		Padding(1, 2).
		Render(body)
}

// SetSize implements app.MetadataComponent.
func (m *Metadata) SetSize(width, height int) {
	m.width = width
	m.height = height
	usable := width - 6
	if usable < 10 {
		usable = 10
	}
	m.titleInput.Width = usable
	m.sceneInput.Width = usable
}

// Focus implements app.MetadataComponent.
func (m *Metadata) Focus() tea.Cmd {
	m.focusIndex = 0
	m.sceneInput.Blur()
	return m.titleInput.Focus()
}

// GameTitle implements app.MetadataComponent.
func (m *Metadata) GameTitle() string {
	return strings.TrimSpace(m.titleInput.Value())
}

// Scene implements app.MetadataComponent.
func (m *Metadata) Scene() string {
	return strings.TrimSpace(m.sceneInput.Value())
}

// Confidence implements app.MetadataComponent.
func (m *Metadata) Confidence() string { return m.confidence }

// HasGameTitle implements app.MetadataComponent.
func (m *Metadata) HasGameTitle() bool { return m.GameTitle() != "" }

// SetFromAnalysis implements app.MetadataComponent.
func (m *Metadata) SetFromAnalysis(title, scene, confidence string) {
	m.titleInput.SetValue(title)
	m.sceneInput.SetValue(scene)
	if confidence == "" {
		confidence = "low"
	}
	m.confidence = confidence
}

// SetErrorBanner implements app.MetadataComponent.
func (m *Metadata) SetErrorBanner(msg string) { m.errBanner = msg }

func renderConfidenceBadge(level string) string {
	color := lipgloss.Color("203") // red default
	switch level {
	case "high":
		color = lipgloss.Color("46") // green
	case "medium":
		color = lipgloss.Color("220") // yellow
	}
	style := lipgloss.NewStyle().Foreground(color).Bold(true)
	return style.Render("● " + level)
}
