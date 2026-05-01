package app

import (
	"context"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

// AppState is the top-level state of the screen.
type AppState int

const (
	StatePicker AppState = iota
	StateAnalyzing
	StateEditing
	StateDone
)

// MinCols and MinRows are the hard minimum terminal dimensions (R-8 / SC-007).
const (
	MinCols = 100
	MinRows = 30
)

// Model is the top-level Bubble Tea model.
type Model struct {
	State  AppState
	Keys   KeyMap
	Width  int
	Height int

	StartDir string
	ModelID  string

	// Picker is the leftpane file picker. Set in NewModel.
	Picker PickerComponent

	// Right-pane components.
	Idle      IdleComponent
	Analyzing AnalyzingComponent
	Metadata  MetadataComponent

	// Header / footer.
	Header HeaderComponent
	Footer FooterComponent

	// Selected video path. Empty until the user confirms a selection.
	SelectedVideo string
	// ErrorBanner shown above the metadata form on FR-018 / FR-019.
	ErrorBanner string

	// AnalysisResult holds the most recent successful analysis record. Used
	// when emitting the final handoff payload from the editing state.
	AnalysisResult AnalysisResult

	// AnalysisCancel cancels the in-flight analysis goroutine. Nil when no
	// analysis is in flight.
	AnalysisCancel context.CancelFunc

	// Quitting is set when the program is about to exit via tea.Quit so that
	// the final handoff payload can be flushed in main.go.
	FinalPayload *FinalPayload
}

// PickerComponent is the leftpane interface (concrete type defined in
// internal/ui/leftpane).
type PickerComponent interface {
	Init() tea.Cmd
	Update(msg tea.Msg) (PickerComponent, tea.Cmd)
	View() string
	SetSize(width, height int)
	SetStartDir(dir string)
}

// Right-pane components are interfaces so internal/app can swap views without
// importing every concrete renderer's package directly.
type IdleComponent interface {
	Init() tea.Cmd
	Update(msg tea.Msg) (IdleComponent, tea.Cmd)
	View() string
	SetSize(width, height int)
}

type AnalyzingComponent interface {
	Init() tea.Cmd
	Update(msg tea.Msg) (AnalyzingComponent, tea.Cmd)
	View() string
	SetSize(width, height int)
}

type MetadataComponent interface {
	Update(msg tea.Msg) (MetadataComponent, tea.Cmd)
	View() string
	SetSize(width, height int)
	Focus() tea.Cmd
	GameTitle() string
	Scene() string
	Confidence() string
	HasGameTitle() bool
	SetFromAnalysis(title, scene, confidence string)
	SetErrorBanner(msg string)
}

type HeaderComponent interface {
	View(width int) string
}

type FooterComponent interface {
	View(width int, state AppState) string
}

// AnalysisResult wraps the successful (or failed) analysis output that will be
// embedded into the final handoff payload.
type AnalysisResult struct {
	Title       string
	Scene       string
	Confidence  string
	Record      any // *handoff.AnalysisRecord but kept opaque to avoid import cycles
	HasResult   bool
	HasError    bool
	ErrorString string
}

// FinalPayload signals main.go to print the JSON handoff after tea exits.
type FinalPayload struct {
	JSON []byte
}

// NewModel constructs a fresh model with the supplied configuration. Concrete
// components are wired in at construction time by the caller via the setters
// on the returned Model.
func NewModel(startDir, modelID string) Model {
	return Model{
		State:    StatePicker,
		Keys:     DefaultKeyMap(),
		StartDir: startDir,
		ModelID:  modelID,
	}
}

// Init is the Bubble Tea Init hook.
func (m Model) Init() tea.Cmd {
	cmds := []tea.Cmd{}
	if m.Picker != nil {
		cmds = append(cmds, m.Picker.Init())
	}
	if m.Idle != nil {
		cmds = append(cmds, m.Idle.Init())
	}
	return tea.Batch(cmds...)
}

// PreflightTerminalSize verifies the terminal is at least MinCols x MinRows.
// Returns false and writes a single-line stderr message on failure.
func PreflightTerminalSize(cols, rows int) bool {
	if cols < MinCols || rows < MinRows {
		fmt.Fprintf(os.Stderr, "terminal must be at least %dx%d (got %dx%d)\n",
			MinCols, MinRows, cols, rows)
		return false
	}
	return true
}
