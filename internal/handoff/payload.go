// Package handoff defines the cross-screen payload schema emitted both by the
// TUI confirm path and by `nano-gameplays identify --json`. See
// specs/001-video-picker-screen/contracts/handoff.md.
package handoff

// SchemaVersion is the pinned version for the v1 handoff payload.
const SchemaVersion = "1"

// HandoffPayload is the single JSON object handed to the next screen.
type HandoffPayload struct {
	SchemaVersion string         `json:"schema_version"`
	Video         GameplayVideo  `json:"video"`
	Metadata      GameMetadata   `json:"metadata"`
	Analysis      AnalysisRecord `json:"analysis"`
}

// GameplayVideo is the user's selected source file.
type GameplayVideo struct {
	Path      string `json:"path"`
	SizeBytes int64  `json:"size_bytes"`
	Extension string `json:"extension"`
}

// GameMetadata is the fixed three-field result described in spec FR-014.
type GameMetadata struct {
	GameTitle          string `json:"game_title"`
	SceneOrLevelOrMode string `json:"scene_or_level_or_mode"`
	Confidence         string `json:"confidence"`
}

// AnalysisRecord is the reproducibility record (Constitution Principle III).
type AnalysisRecord struct {
	PromptVersion     string   `json:"prompt_version"`
	ModelID           string   `json:"model_id"`
	FrameOffsetsPct   []int    `json:"frame_offsets_pct"`
	FrameSHA256       []string `json:"frame_sha256"`
	RequestStartedAt  string   `json:"request_started_at"`
	RequestFinishedAt string   `json:"request_finished_at"`
	RawResponse       string   `json:"raw_response"`
	Error             string   `json:"error"`
}
