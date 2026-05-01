package main

import (
	"github.com/dgr8akki/nano-gameplays/internal/analysis"
	"github.com/dgr8akki/nano-gameplays/internal/app"
	"github.com/dgr8akki/nano-gameplays/internal/ui/footer"
	"github.com/dgr8akki/nano-gameplays/internal/ui/header"
	"github.com/dgr8akki/nano-gameplays/internal/ui/leftpane"
	"github.com/dgr8akki/nano-gameplays/internal/ui/rightpane"
)

// wireComponents installs concrete implementations of every component the
// app.Model depends on and the analysis runner the model invokes per
// selection.
func wireComponents(m *app.Model, startDir, _ string) {
	m.Picker = leftpane.New(startDir)
	m.Header = header.New()
	m.Footer = footer.New(m.Keys)
	m.Idle = rightpane.NewIdle()
	m.Analyzing = rightpane.NewAnalyzing()
	m.Metadata = rightpane.NewMetadata()

	app.AnalysisRunner = analysis.Run
}
