package analysis

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/dgr8akki/nano-gameplays/internal/app"
)

// Run executes the full extract → identify pipeline for a single video and
// returns the appropriate tea.Msg (AnalysisOKMsg or AnalysisErrMsg) for the
// Bubble Tea update loop.
func Run(ctx context.Context, videoPath, modelID string) tea.Msg {
	frames, err := ExtractFrames(ctx, videoPath)
	if err != nil {
		return app.AnalysisErrMsg{
			VideoPath: videoPath,
			Err:       err.Error(),
		}
	}
	meta, rec, err := Identify(ctx, frames, modelID)
	if err != nil {
		return app.AnalysisErrMsg{
			VideoPath: videoPath,
			Err:       err.Error(),
			Record:    rec,
		}
	}
	return app.AnalysisOKMsg{
		VideoPath:  videoPath,
		Title:      meta.GameTitle,
		Scene:      meta.SceneOrLevelOrMode,
		Confidence: meta.Confidence,
		Record:     rec,
	}
}
