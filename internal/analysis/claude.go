package analysis

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/anthropics/anthropic-sdk-go"

	"github.com/dgr8akki/nano-gameplays/internal/analysis/prompt"
	"github.com/dgr8akki/nano-gameplays/internal/handoff"
)

const defaultMaxTokens int64 = 512

type identifyResponse struct {
	GameTitle          string `json:"game_title"`
	SceneOrLevelOrMode string `json:"scene_or_level_or_mode"`
	Confidence         string `json:"confidence"`
}

// Identify sends the captured frames to Claude with the versioned prompt and
// returns the parsed metadata plus a populated AnalysisRecord. The record is
// always returned (even on error) so callers can persist the failure trace per
// FR-018.
func Identify(ctx context.Context, frames []CapturedFrame, modelID string) (handoff.GameMetadata, handoff.AnalysisRecord, error) {
	rec := handoff.AnalysisRecord{
		PromptVersion:    prompt.IdentifyGameVersion,
		ModelID:          modelID,
		FrameOffsetsPct:  framesOffsets(frames),
		FrameSHA256:      framesHashes(frames),
		RequestStartedAt: nowRFC3339(),
	}
	if len(frames) == 0 {
		rec.RequestFinishedAt = nowRFC3339()
		rec.Error = "no frames extracted"
		return handoff.GameMetadata{Confidence: "low"}, rec, errors.New(rec.Error)
	}

	client := anthropic.NewClient()

	contentBlocks := make([]anthropic.ContentBlockParamUnion, 0, len(frames)+1)
	for _, f := range frames {
		encoded := base64.StdEncoding.EncodeToString(f.Bytes)
		contentBlocks = append(contentBlocks, anthropic.NewImageBlockBase64("image/png", encoded))
	}
	contentBlocks = append(contentBlocks, anthropic.NewTextBlock("Identify the game shown in these frames and respond per the schema."))

	params := anthropic.MessageNewParams{
		MaxTokens: defaultMaxTokens,
		Model:     anthropic.Model(modelID),
		System: []anthropic.TextBlockParam{
			{Text: prompt.IdentifyGameV1()},
		},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(contentBlocks...),
		},
	}

	resp, err := client.Messages.New(ctx, params)
	rec.RequestFinishedAt = nowRFC3339()
	if err != nil {
		rec.Error = err.Error()
		return handoff.GameMetadata{Confidence: "low"}, rec, err
	}

	rawText := extractTextResponse(resp)
	rec.RawResponse = rawText
	parsed, parseErr := parseIdentifyResponse(rawText)
	if parseErr != nil {
		rec.Error = parseErr.Error()
		return handoff.GameMetadata{Confidence: "low"}, rec, parseErr
	}

	conf, coerced := coerceConfidence(parsed.Confidence)
	if coerced {
		rec.Error = fmt.Sprintf("coerced confidence value %q to low", parsed.Confidence)
	}

	return handoff.GameMetadata{
		GameTitle:          strings.TrimSpace(parsed.GameTitle),
		SceneOrLevelOrMode: strings.TrimSpace(parsed.SceneOrLevelOrMode),
		Confidence:         conf,
	}, rec, nil
}

func parseIdentifyResponse(raw string) (identifyResponse, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return identifyResponse{}, errors.New("empty model response")
	}
	// Defensive: tolerate a single ```json fence even though the prompt forbids it.
	trimmed = strings.TrimPrefix(trimmed, "```json")
	trimmed = strings.TrimPrefix(trimmed, "```")
	trimmed = strings.TrimSuffix(trimmed, "```")
	trimmed = strings.TrimSpace(trimmed)

	var out identifyResponse
	if err := json.Unmarshal([]byte(trimmed), &out); err != nil {
		return identifyResponse{}, fmt.Errorf("decode model JSON: %w", err)
	}
	return out, nil
}

// coerceConfidence enforces the low/medium/high enum (FR-014, contract). Returns
// the canonical value plus whether coercion happened.
func coerceConfidence(value string) (string, bool) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "low":
		return "low", false
	case "medium":
		return "medium", false
	case "high":
		return "high", false
	}
	return "low", true
}

func extractTextResponse(msg *anthropic.Message) string {
	if msg == nil {
		return ""
	}
	var b strings.Builder
	for _, block := range msg.Content {
		if block.Type == "text" {
			b.WriteString(block.Text)
		}
	}
	return b.String()
}

func framesOffsets(frames []CapturedFrame) []int {
	out := make([]int, len(frames))
	for i, f := range frames {
		out[i] = f.OffsetPct
	}
	return out
}

func framesHashes(frames []CapturedFrame) []string {
	out := make([]string, len(frames))
	for i, f := range frames {
		out[i] = f.SHA256
	}
	return out
}

func nowRFC3339() string { return time.Now().UTC().Format(time.RFC3339) }
