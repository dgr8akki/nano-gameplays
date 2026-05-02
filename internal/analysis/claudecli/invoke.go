package claudecli

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
)

const (
	binaryName       = "claude"
	stderrCapBytes   = 64 * 1024
	allowedToolsList = "Read"
	denyToolsList    = "Bash"
)

// InvokeOpts is the input shape for Run.
type InvokeOpts struct {
	SystemPrompt string
	JSONSchema   string
	ModelID      string
	TempDir      string
	UserMessage  string
}

// BuildArgv assembles the canonical claude argv per
// specs/002-claude-cli-backend/contracts/claude-cli.md §"Argv shape".
// `--model` is omitted entirely when ModelID is empty so the user's
// claude install picks its own default (FR-008 / R-5).
func BuildArgv(opts InvokeOpts) []string {
	argv := []string{
		"-p",
		"--output-format", "json",
		"--system-prompt", opts.SystemPrompt,
		"--json-schema", opts.JSONSchema,
	}
	if opts.ModelID != "" {
		argv = append(argv, "--model", opts.ModelID)
	}
	argv = append(argv,
		"--add-dir", opts.TempDir,
		"--allowedTools", allowedToolsList,
		"--disallowedTools", denyToolsList,
		"--disable-slash-commands",
		"--no-session-persistence",
		"--",
		opts.UserMessage,
	)
	return argv
}

// RunResult bundles everything the orchestrator needs to populate the
// AnalysisRecord.
type RunResult struct {
	Envelope ResultEnvelope
	Stderr   []byte
	ExitErr  error // nil on exit-zero
}

// Run executes one `claude -p` subprocess under the supplied context.
// Cancellation propagates to the subprocess via exec.CommandContext.
//
// The function does NOT manage the temp directory; the caller is
// responsible for `defer os.RemoveAll(opts.TempDir)`. Run only consumes
// the directory via the argv passed to claude.
func Run(ctx context.Context, opts InvokeOpts) (RunResult, error) {
	if opts.TempDir == "" {
		return RunResult{}, errors.New("InvokeOpts.TempDir is required")
	}
	if opts.SystemPrompt == "" || opts.JSONSchema == "" || opts.UserMessage == "" {
		return RunResult{}, errors.New("InvokeOpts: SystemPrompt, JSONSchema, and UserMessage are required")
	}

	argv := BuildArgv(opts)

	cmd := exec.CommandContext(ctx, binaryName, argv...)
	cmd.Stdin = nil

	var stdout bytes.Buffer
	stderr := &boundedBuffer{cap: stderrCapBytes}
	cmd.Stdout = &stdout
	cmd.Stderr = stderr

	exitErr := cmd.Run()

	out := RunResult{Stderr: stderr.Bytes(), ExitErr: exitErr}

	// Decode the envelope when stdout is non-empty regardless of exit code:
	// `claude -p --output-format json` emits a structured error envelope on
	// some failure modes too (auth, rate limit, schema validation), and we
	// want to surface the structured error rather than swallowing it.
	if stdout.Len() > 0 {
		env, decodeErr := DecodeEnvelope(stdout.Bytes())
		if decodeErr != nil {
			// Fall through with empty envelope; orchestrator treats this as a
			// malformed-output failure path (FR-005). Preserve the raw bytes
			// so the orchestrator can fold them into the record.
			out.Envelope = ResultEnvelope{Result: stdout.String(),
				IsError: true,
				Error:   fmt.Sprintf("envelope decode: %v", decodeErr),
			}
			return out, nil
		}
		out.Envelope = env
	}
	return out, nil
}

// boundedBuffer is an io.Writer that retains at most `cap` bytes.
type boundedBuffer struct {
	cap int
	buf bytes.Buffer
}

func (b *boundedBuffer) Write(p []byte) (int, error) {
	remaining := b.cap - b.buf.Len()
	if remaining <= 0 {
		return len(p), nil
	}
	if len(p) > remaining {
		b.buf.Write(p[:remaining])
		return len(p), nil
	}
	b.buf.Write(p)
	return len(p), nil
}

func (b *boundedBuffer) Bytes() []byte { return b.buf.Bytes() }

var _ io.Writer = (*boundedBuffer)(nil)
