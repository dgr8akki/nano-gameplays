package analysis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/dgr8akki/nano-gameplays/internal/analysis/claudecli"
	"github.com/dgr8akki/nano-gameplays/internal/analysis/prompt"
	"github.com/dgr8akki/nano-gameplays/internal/handoff"
)

type identifyResponse struct {
	GameTitle          string `json:"game_title"`
	SceneOrLevelOrMode string `json:"scene_or_level_or_mode"`
	Confidence         string `json:"confidence"`
}

// Identify orchestrates one identification call: write captured frames
// to a private per-call tempdir, invoke the local `claude` CLI with the
// versioned prompt + schema, decode the structured response, and
// produce the (GameMetadata, AnalysisRecord) pair the application
// emits in its handoff payload. Cleanup of the tempdir is mandatory and
// runs on every exit path (success, error, context cancellation) via a
// deferred RemoveAll.
//
// See specs/002-claude-cli-backend/{plan.md,research.md,contracts/}.
func Identify(ctx context.Context, frames []CapturedFrame, modelID string) (handoff.GameMetadata, handoff.AnalysisRecord, error) {
	rec := handoff.AnalysisRecord{
		PromptVersion:   prompt.IdentifyGameVersion,
		ModelID:         modelID,
		FrameOffsetsPct: framesOffsets(frames),
		FrameSHA256:     framesHashes(frames),
	}

	if len(frames) == 0 {
		rec.RequestStartedAt = nowRFC3339()
		rec.RequestFinishedAt = rec.RequestStartedAt
		rec.Error = "no frames extracted"
		return handoff.GameMetadata{Confidence: "low"}, rec, errors.New(rec.Error)
	}

	cliFrames := make([]claudecli.Frame, len(frames))
	for i, f := range frames {
		cliFrames[i] = claudecli.Frame{Bytes: f.Bytes, OffsetPct: f.OffsetPct}
	}
	tempDir, framePaths, err := claudecli.WriteFrames(cliFrames)
	if err != nil {
		rec.RequestStartedAt = nowRFC3339()
		rec.RequestFinishedAt = rec.RequestStartedAt
		rec.Error = fmt.Sprintf("tempdir / frame write: %v", err)
		return handoff.GameMetadata{Confidence: "low"}, rec, err
	}
	defer os.RemoveAll(tempDir)

	opts := claudecli.InvokeOpts{
		SystemPrompt: prompt.IdentifyGameV1(),
		JSONSchema:   prompt.IdentifySchemaV1(),
		ModelID:      modelID,
		TempDir:      tempDir,
		UserMessage:  buildUserMessage(framePaths),
	}

	rec.RequestStartedAt = nowRFC3339()
	runResult, runErr := claudecli.Run(ctx, opts)
	rec.RequestFinishedAt = nowRFC3339()
	rec.RawResponse = pickRawResponse(runResult.Envelope)

	if runErr != nil {
		rec.Error = fmt.Sprintf("claude invoke: %v", runErr)
		return handoff.GameMetadata{Confidence: "low"}, rec, runErr
	}
	if runResult.ExitErr != nil {
		// Context cancellations are surfaced by exec.CommandContext as a
		// non-zero exit; let the orchestrator distinguish them via ctx.Err().
		if ctx.Err() != nil {
			rec.Error = fmt.Sprintf("claude cancelled: %v", ctx.Err())
			return handoff.GameMetadata{Confidence: "low"}, rec, ctx.Err()
		}
		rec.Error = formatCallError("claude exit", runResult)
		return handoff.GameMetadata{Confidence: "low"}, rec, runResult.ExitErr
	}
	if runResult.Envelope.IsError {
		rec.Error = formatCallError("envelope-error", runResult)
		return handoff.GameMetadata{Confidence: "low"}, rec, errors.New(rec.Error)
	}
	parsed, decodeErr := decodeIdentifyResponse(runResult.Envelope)
	if decodeErr != nil {
		rec.Error = fmt.Sprintf("decode model JSON: %v", decodeErr)
		return handoff.GameMetadata{Confidence: "low"}, rec, decodeErr
	}

	rec.ModelID = resolveModelID(modelID, runResult.Envelope.Model, &rec)

	return handoff.GameMetadata{
		GameTitle:          strings.TrimSpace(parsed.GameTitle),
		SceneOrLevelOrMode: strings.TrimSpace(parsed.SceneOrLevelOrMode),
		Confidence:         parsed.Confidence, // schema-enforced enum, no coercion needed
	}, rec, nil
}

// pickRawResponse returns the verbatim model output for the AnalysisRecord:
// the structured_output JSON when present (the canonical source of truth for
// schema-constrained calls), otherwise the textual result.
func pickRawResponse(env claudecli.ResultEnvelope) string {
	if len(env.StructuredOutput) > 0 {
		return string(env.StructuredOutput)
	}
	return env.Result
}

// decodeIdentifyResponse prefers the schema-validated structured_output
// from the envelope when present; falls back to JSON-decoding the textual
// result for compatibility with envelopes that omit structured_output.
func decodeIdentifyResponse(env claudecli.ResultEnvelope) (identifyResponse, error) {
	var out identifyResponse
	if len(env.StructuredOutput) > 0 {
		if err := json.Unmarshal(env.StructuredOutput, &out); err != nil {
			return out, err
		}
		return out, nil
	}
	if env.Result == "" {
		return out, errors.New("empty model response")
	}
	if err := json.Unmarshal([]byte(env.Result), &out); err != nil {
		return out, err
	}
	return out, nil
}

// resolveModelID applies the precedence from
// specs/002-claude-cli-backend/research.md §R-5: explicit override wins,
// then envelope-reported model, then empty + an `analysis.error` note.
func resolveModelID(override, envelope string, rec *handoff.AnalysisRecord) string {
	if override != "" {
		return override
	}
	if envelope != "" {
		return envelope
	}
	rec.Error = appendErr(rec.Error, "model id unavailable")
	return ""
}

// formatCallError produces a single-line, descriptive error string drawn
// from (in priority order) the envelope's own error, the captured stderr,
// and the stage label.
func formatCallError(stage string, r claudecli.RunResult) string {
	if r.Envelope.Error != "" {
		return fmt.Sprintf("%s: %s", stage, r.Envelope.Error)
	}
	if s := claudeclireadStderr(r.Stderr); s != "" {
		return fmt.Sprintf("%s: %s", stage, s)
	}
	if r.ExitErr != nil {
		return fmt.Sprintf("%s: %v", stage, r.ExitErr)
	}
	return stage
}

func appendErr(existing, more string) string {
	if existing == "" {
		return more
	}
	return existing + "; " + more
}

// claudeclireadStderr is a small indirection so this package can reuse the
// claudecli stderr-trimming helper without exporting it broadly.
func claudeclireadStderr(b []byte) string {
	s := strings.TrimSpace(string(b))
	if len(s) > 1024 {
		s = s[:1024] + "…"
	}
	return strings.ReplaceAll(s, "\n", " ")
}

func buildUserMessage(framePaths []string) string {
	switch len(framePaths) {
	case 0:
		return "Identify the game shown. Respond per the schema."
	case 1:
		return fmt.Sprintf("Identify the game shown in the frame at %s. Respond per the schema.", framePaths[0])
	default:
		var b strings.Builder
		b.WriteString("Identify the game shown in the frames at ")
		for i, p := range framePaths {
			if i > 0 {
				if i == len(framePaths)-1 {
					b.WriteString(" and ")
				} else {
					b.WriteString(", ")
				}
			}
			b.WriteString(p)
		}
		b.WriteString(". Respond per the schema.")
		return b.String()
	}
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
