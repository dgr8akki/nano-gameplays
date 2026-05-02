// Package analysis provides ffmpeg frame extraction and the Claude vision
// identification call. Frames are kept in memory only (FR-023).
package analysis

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// CapturedFrame is a single in-memory PNG frame plus reproducibility metadata.
type CapturedFrame struct {
	Bytes     []byte
	OffsetPct int
	SHA256    string
}

// ExtractError is returned by ExtractFrames when extraction fails.
type ExtractError struct {
	Stage string
	Err   error
}

func (e *ExtractError) Error() string {
	if e == nil {
		return ""
	}
	return fmt.Sprintf("ffmpeg extraction failed during %s: %v", e.Stage, e.Err)
}

func (e *ExtractError) Unwrap() error { return e.Err }

// FrameOffsetsPct is the deterministic offset list (R-2): 10% and 60%.
var FrameOffsetsPct = []int{10, 60}

// ExtractFrames pulls one frame per offset in FrameOffsetsPct from the supplied
// video. It shells out to ffprobe (for duration) and ffmpeg (for the frame
// itself), decodes via stdout to keep frames off disk.
//
// On any failure it returns the partial slice gathered so far (may be empty)
// alongside an *ExtractError so callers can still emit a failure record.
func ExtractFrames(ctx context.Context, videoPath string) ([]CapturedFrame, error) {
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		return nil, &ExtractError{Stage: "lookup-ffmpeg", Err: errors.New("ffmpeg not found in PATH")}
	}
	if _, err := exec.LookPath("ffprobe"); err != nil {
		return nil, &ExtractError{Stage: "lookup-ffprobe", Err: errors.New("ffprobe not found in PATH")}
	}

	duration, err := probeDuration(ctx, videoPath)
	if err != nil {
		return nil, &ExtractError{Stage: "probe", Err: err}
	}
	if duration <= 0 {
		return nil, &ExtractError{Stage: "probe", Err: fmt.Errorf("non-positive duration %.2fs", duration)}
	}

	out := make([]CapturedFrame, 0, len(FrameOffsetsPct))
	var firstErr error
	for _, pct := range FrameOffsetsPct {
		seek := duration * float64(pct) / 100.0
		buf, err := extractOne(ctx, videoPath, seek)
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		hash := sha256.Sum256(buf)
		out = append(out, CapturedFrame{
			Bytes:     buf,
			OffsetPct: pct,
			SHA256:    hex.EncodeToString(hash[:]),
		})
	}
	if len(out) == 0 {
		return nil, &ExtractError{Stage: "extract", Err: firstErr}
	}
	return out, nil
}

func probeDuration(ctx context.Context, videoPath string) (float64, error) {
	cmd := exec.CommandContext(ctx,
		"ffprobe",
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		videoPath,
	)
	out, err := cmd.Output()
	if err != nil {
		return 0, err
	}
	s := strings.TrimSpace(string(out))
	d, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, fmt.Errorf("parse duration %q: %w", s, err)
	}
	return d, nil
}

func extractOne(ctx context.Context, videoPath string, seek float64) ([]byte, error) {
	var stdout, stderr bytes.Buffer
	cmd := exec.CommandContext(ctx,
		"ffmpeg",
		"-loglevel", "error",
		"-ss", strconv.FormatFloat(seek, 'f', 3, 64),
		"-i", videoPath,
		"-frames:v", "1",
		"-f", "image2pipe",
		"-vcodec", "png",
		"-",
	)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("%w (stderr: %s)", err, strings.TrimSpace(stderr.String()))
	}
	if stdout.Len() == 0 {
		return nil, errors.New("zero-byte output")
	}
	return stdout.Bytes(), nil
}
