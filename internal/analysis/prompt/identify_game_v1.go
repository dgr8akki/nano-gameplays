// Package prompt holds the versioned Claude prompt templates used by the
// analysis package. See specs/001-video-picker-screen/contracts/claude-prompt.md.
package prompt

import _ "embed"

// IdentifyGameVersion is the identifier embedded in AnalysisRecord.prompt_version.
const IdentifyGameVersion = "identify-game.v1"

//go:embed identify_game_v1.txt
var identifyGameV1 string

// IdentifyGameV1 returns the verbatim system instruction sent on the identify
// call. Edits to the embedded file require a new version (identify-game.v2,
// …); never in-place per the prompt contract's versioning rules.
func IdentifyGameV1() string { return identifyGameV1 }
