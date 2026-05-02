package app

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/dgr8akki/nano-gameplays/internal/handoff"
)

// SelectVideoMsg is emitted when the picker confirms a video selection.
type SelectVideoMsg struct {
	Path string
}

// AnalysisOKMsg is emitted when Claude returns a successful identification.
type AnalysisOKMsg struct {
	VideoPath  string
	Title      string
	Scene      string
	Confidence string
	Record     handoff.AnalysisRecord
}

// AnalysisErrMsg is emitted when extraction or Claude fails.
type AnalysisErrMsg struct {
	VideoPath string
	Err       string
	Record    handoff.AnalysisRecord
}

// Update is the top-level Bubble Tea Update function.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		if !PreflightTerminalSize(msg.Width, msg.Height) {
			return m, tea.Quit
		}
		m.layoutChildren()
		return m, nil

	case tea.KeyMsg:
		if key := msg.String(); key == "ctrl+c" || (m.State == StatePicker && key == "q") {
			return m, tea.Quit
		}
	}

	switch m.State {
	case StatePicker:
		return m.updatePicker(msg)
	case StateAnalyzing:
		return m.updateAnalyzing(msg)
	case StateEditing:
		return m.updateEditing(msg)
	case StateDone:
		return m, tea.Quit
	}
	return m, nil
}

func (m *Model) layoutChildren() {
	if m.Width <= 0 || m.Height <= 0 {
		return
	}
	headerH, footerH := 3, 1
	bodyH := m.Height - headerH - footerH
	if bodyH < 1 {
		bodyH = 1
	}
	leftW := m.Width / 2
	rightW := m.Width - leftW
	if m.Picker != nil {
		m.Picker.SetSize(leftW, bodyH)
	}
	if m.Idle != nil {
		m.Idle.SetSize(rightW, bodyH)
	}
	if m.Analyzing != nil {
		m.Analyzing.SetSize(rightW, bodyH)
	}
	if m.Metadata != nil {
		m.Metadata.SetSize(rightW, bodyH)
	}
}

func (m Model) updatePicker(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case SelectVideoMsg:
		return m.beginAnalysis(msg.Path)
	}
	if m.Picker == nil {
		return m, nil
	}
	var cmd tea.Cmd
	m.Picker, cmd = m.Picker.Update(msg)
	return m, cmd
}

func (m Model) updateAnalyzing(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case AnalysisOKMsg:
		if msg.VideoPath != m.SelectedVideo {
			return m, nil // stale, dropped per FR-020
		}
		m.AnalysisResult = AnalysisResult{
			Title:      msg.Title,
			Scene:      msg.Scene,
			Confidence: msg.Confidence,
			Record:     msg.Record,
			HasResult:  true,
		}
		m.ErrorBanner = ""
		if m.Metadata != nil {
			m.Metadata.SetFromAnalysis(msg.Title, msg.Scene, msg.Confidence)
			m.Metadata.SetErrorBanner("")
		}
		m.State = StateEditing
		var cmd tea.Cmd
		if m.Metadata != nil {
			cmd = m.Metadata.Focus()
		}
		return m, cmd
	case AnalysisErrMsg:
		if msg.VideoPath != m.SelectedVideo {
			return m, nil
		}
		m.AnalysisResult = AnalysisResult{
			Title:       "",
			Scene:       "",
			Confidence:  "low",
			Record:      msg.Record,
			HasResult:   true,
			HasError:    true,
			ErrorString: msg.Err,
		}
		m.ErrorBanner = msg.Err
		if m.Metadata != nil {
			m.Metadata.SetFromAnalysis("", "", "low")
			m.Metadata.SetErrorBanner(msg.Err)
		}
		m.State = StateEditing
		var cmd tea.Cmd
		if m.Metadata != nil {
			cmd = m.Metadata.Focus()
		}
		return m, cmd
	case SelectVideoMsg:
		return m.beginAnalysis(msg.Path)
	}
	if m.Analyzing != nil {
		var cmd tea.Cmd
		m.Analyzing, cmd = m.Analyzing.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m Model) updateEditing(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if key.Matches(msg, m.Keys.Confirm) && m.Metadata != nil && m.Metadata.HasGameTitle() {
			return m.confirmEditing()
		}
		if key.Matches(msg, m.Keys.Back) {
			m.cancelInFlight()
			m.State = StatePicker
			m.ErrorBanner = ""
			m.SelectedVideo = ""
			return m, nil
		}
	case SelectVideoMsg:
		return m.beginAnalysis(msg.Path)
	}
	if m.Metadata != nil {
		var cmd tea.Cmd
		m.Metadata, cmd = m.Metadata.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m Model) beginAnalysis(path string) (tea.Model, tea.Cmd) {
	m.cancelInFlight()
	m.SelectedVideo = path
	m.ErrorBanner = ""
	m.AnalysisResult = AnalysisResult{}
	m.State = StateAnalyzing
	cmd, cancel := startAnalysisCmd(path, m.ModelID)
	m.AnalysisCancel = cancel
	cmds := []tea.Cmd{cmd}
	if m.Analyzing != nil {
		cmds = append(cmds, m.Analyzing.Init())
	}
	return m, tea.Batch(cmds...)
}

func (m *Model) cancelInFlight() {
	if m.AnalysisCancel != nil {
		m.AnalysisCancel()
		m.AnalysisCancel = nil
	}
}

func (m Model) confirmEditing() (tea.Model, tea.Cmd) {
	rec, _ := m.AnalysisResult.Record.(handoff.AnalysisRecord)
	payload := handoff.HandoffPayload{
		SchemaVersion: handoff.SchemaVersion,
		Video:         buildVideo(m.SelectedVideo),
		Metadata: handoff.GameMetadata{
			GameTitle:          m.Metadata.GameTitle(),
			SceneOrLevelOrMode: m.Metadata.Scene(),
			Confidence:         m.Metadata.Confidence(),
		},
		Analysis: rec,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		fmt.Fprintf(os.Stderr, "encode handoff payload: %v\n", err)
		return m, tea.Quit
	}
	m.FinalPayload = &FinalPayload{JSON: data}
	m.State = StateDone
	return m, tea.Quit
}

// buildVideo populates GameplayVideo for the supplied path. Best-effort: stat
// failures are tolerated (size_bytes will be zero and extension will still be
// computed from the path).
func buildVideo(path string) handoff.GameplayVideo {
	out := handoff.GameplayVideo{Path: path}
	if info, err := os.Stat(path); err == nil {
		out.SizeBytes = info.Size()
	}
	out.Extension = lowerExt(path)
	return out
}

func lowerExt(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '.' {
			ext := path[i+1:]
			b := make([]byte, len(ext))
			for j := 0; j < len(ext); j++ {
				c := ext[j]
				if c >= 'A' && c <= 'Z' {
					c += 32
				}
				b[j] = c
			}
			return string(b)
		}
		if path[i] == '/' {
			break
		}
	}
	return ""
}
