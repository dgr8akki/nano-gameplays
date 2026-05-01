package app

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
)

// AnalysisRunner is the function the model uses to run a single analysis.
// It is replaced at program-construction time with the real implementation
// from internal/analysis (kept as a package-level var so internal/app does
// not import internal/analysis directly and create a cycle).
var AnalysisRunner func(ctx context.Context, videoPath, modelID string) tea.Msg

// startAnalysisCmd dispatches the configured AnalysisRunner under a
// cancellable context and returns the command + cancel func to the caller.
func startAnalysisCmd(videoPath, modelID string) (tea.Cmd, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())
	cmd := func() tea.Msg {
		if AnalysisRunner == nil {
			return AnalysisErrMsg{
				VideoPath: videoPath,
				Err:       "analysis runner not configured",
			}
		}
		return AnalysisRunner(ctx, videoPath, modelID)
	}
	return cmd, cancel
}
