// Package leftpane wraps bubbles/filepicker to satisfy app.PickerComponent.
package leftpane

import (
	"time"

	"github.com/charmbracelet/bubbles/filepicker"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/dgr8akki/nano-gameplays/internal/app"
)

var allowedExtensions = []string{".mp4", ".mov", ".mkv", ".webm"}

// Model is the leftpane file picker.
type Model struct {
	picker      filepicker.Model
	width       int
	height      int
	notice      string
	noticeUntil time.Time
}

type clearNoticeMsg struct{ at time.Time }

// New constructs a new picker rooted at startDir.
func New(startDir string) *Model {
	fp := filepicker.New()
	fp.AllowedTypes = allowedExtensions
	fp.CurrentDirectory = startDir
	fp.AutoHeight = false
	fp.ShowHidden = false
	fp.ShowSize = true
	fp.ShowPermissions = false
	fp.DirAllowed = false
	fp.FileAllowed = true

	// Visually dim non-selectable files (FR-010).
	fp.Styles.DisabledFile = fp.Styles.DisabledFile.
		Foreground(lipgloss.Color("240")).
		Faint(true)

	return &Model{picker: fp}
}

// Init implements app.PickerComponent.
func (m *Model) Init() tea.Cmd { return m.picker.Init() }

// Update implements app.PickerComponent.
func (m *Model) Update(msg tea.Msg) (app.PickerComponent, tea.Cmd) {
	if cm, ok := msg.(clearNoticeMsg); ok {
		if !cm.at.Before(m.noticeUntil) {
			m.notice = ""
		}
		return m, nil
	}

	if didSelect, path := m.picker.DidSelectFile(msg); didSelect {
		// Allowed video file selection.
		return m, func() tea.Msg { return app.SelectVideoMsg{Path: path} }
	}
	if didSelect, _ := m.picker.DidSelectDisabledFile(msg); didSelect {
		// Show transient inline message; keep focus in picker (FR-011).
		m.notice = "not a supported video"
		until := time.Now().Add(2 * time.Second)
		m.noticeUntil = until
		return m, tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
			return clearNoticeMsg{at: t}
		})
	}

	var cmd tea.Cmd
	m.picker, cmd = m.picker.Update(msg)
	return m, cmd
}

// View implements app.PickerComponent.
func (m *Model) View() string {
	view := m.picker.View()
	if m.notice != "" {
		notice := lipgloss.NewStyle().
			Foreground(lipgloss.Color("203")).
			Render("! " + m.notice)
		view += "\n" + notice
	}
	return view
}

// SetSize implements app.PickerComponent.
func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
	if height > 2 {
		m.picker.SetHeight(height - 2)
	}
}

// SetStartDir implements app.PickerComponent.
func (m *Model) SetStartDir(dir string) {
	m.picker.CurrentDirectory = dir
}
