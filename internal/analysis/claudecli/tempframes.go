// Package claudecli owns every detail of talking to the local `claude`
// command-line tool: argv assembly, per-call temp directory lifecycle,
// subprocess execution, and `--output-format json` envelope parsing. See
// specs/002-claude-cli-backend/contracts/claude-cli.md.
package claudecli

import (
	"fmt"
	"os"
	"path/filepath"
)

// Frame is a slim, package-local view of an extracted PNG frame. It exists
// so tempframes can avoid importing internal/analysis (and creating an
// import cycle); the orchestrator translates from analysis.CapturedFrame
// at the call site.
type Frame struct {
	Bytes     []byte
	OffsetPct int
}

// WriteFrames creates a private per-call temporary directory and writes one
// PNG file per frame into it. Returns the absolute tempdir path and the
// per-frame absolute paths in the same order as the input.
//
// The caller MUST `defer os.RemoveAll(dir)` immediately after a successful
// return; cleanup is the caller's responsibility (see contracts/claude-cli.md
// §"Tempdir lifecycle"). On error the directory is removed before return so
// the caller does not have to handle partial state.
func WriteFrames(frames []Frame) (dir string, paths []string, err error) {
	if len(frames) == 0 {
		return "", nil, fmt.Errorf("no frames to write")
	}

	dir, err = os.MkdirTemp("", "nano-gameplays-")
	if err != nil {
		return "", nil, fmt.Errorf("create tempdir: %w", err)
	}
	// MkdirTemp uses 0700 on Unix; explicit chmod for clarity / paranoia.
	if err := os.Chmod(dir, 0o700); err != nil {
		os.RemoveAll(dir)
		return "", nil, fmt.Errorf("chmod tempdir: %w", err)
	}

	paths = make([]string, 0, len(frames))
	for _, f := range frames {
		name := fmt.Sprintf("frame_%d.png", f.OffsetPct)
		full := filepath.Join(dir, name)
		if err := os.WriteFile(full, f.Bytes, 0o600); err != nil {
			os.RemoveAll(dir)
			return "", nil, fmt.Errorf("write %s: %w", name, err)
		}
		paths = append(paths, full)
	}
	return dir, paths, nil
}
