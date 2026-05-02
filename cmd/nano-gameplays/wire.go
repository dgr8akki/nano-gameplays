package main

import (
	"github.com/dgr8akki/nano-gameplays/internal/app"
)

// buildModel returns the fully-wired top-level Model. Concrete components are
// attached in US1/US2/US3 implementation phases.
func buildModel(startDir, modelID string) app.Model {
	m := app.NewModel(startDir, modelID)
	wireComponents(&m, startDir, modelID)
	return m
}
