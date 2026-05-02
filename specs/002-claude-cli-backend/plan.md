# Implementation Plan: Use Local `claude` CLI Instead of Anthropic SDK

**Branch**: `002-claude-cli-backend` | **Date**: 2026-05-02 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `specs/002-claude-cli-backend/spec.md`

## Summary

Replace the in-process Anthropic SDK call (the only path that today reads
`ANTHROPIC_API_KEY`) with a subprocess invocation of the local `claude`
command-line tool that the developer has already authenticated. The
identify flow keeps the same pre-call shape (ffmpeg-extracted PNG frames,
the `identify-game.v1` prompt template, cancellable per-selection), the
same emitted `HandoffPayload` schema, and the same constitution-mandated
reproducibility record. The only thing that changes is *who* makes the
HTTP call to Anthropic — the user's `claude` binary instead of our own
SDK client embedded in `nano-gameplays`. The frame transport switches
from in-process base64 to a per-call temporary directory under
`os.TempDir()` that is created with user-only permissions and deleted
before the analysis goroutine returns, preserving the spirit of
`001-video-picker-screen` FR-023.

## Technical Context

**Language/Version**: Go 1.22+ (single static binary, unchanged from 001).
**Primary Dependencies**:
  - `github.com/charmbracelet/bubbletea` — TUI runtime *(unchanged)*
  - `github.com/charmbracelet/lipgloss` — layout primitives *(unchanged)*
  - `github.com/charmbracelet/bubbles` — `filepicker`, `spinner`, `help`,
    `textinput` *(unchanged)*
  - `ffmpeg` (system binary) — frame extraction *(unchanged)*
  - **`claude` (system binary, ≥ 2.1.0)** — vision-capable Claude Code CLI
    invoked via `os/exec`. **NEW.**
  - **`github.com/anthropics/anthropic-sdk-go` — REMOVED.**
**Storage**: In-memory for the program's working state (unchanged). For
each identify call, the captured PNG frames are written briefly to a
per-call subdirectory under `os.TempDir()` (e.g. `nano-gameplays-XXXX/`)
created with mode `0700`, and the entire directory is removed via
`defer` before the analysis goroutine returns — including on
cancellation, error, and panic. No frames persist after the call.
**Testing**: Still deferred for v2 per the deviation logged in
`001-video-picker-screen/plan.md` → Complexity Tracking. The same
deferral rationale applies: rapid iteration on the integration shape.
The deferral MUST be removed before the merge to `main` of either
branch (whichever lands first will own the test scaffolding).
**Target Platform**: macOS and Linux terminals, ≥ 100 columns × 30 rows
*(unchanged)*. Adds an additional system-binary prerequisite: a
working, signed-in `claude` CLI on `PATH`.
**Project Type**: Single-project CLI/TUI Go module *(unchanged)*.
**Performance Goals** (inherited from spec 001 SC-005, SC-006; this spec
SC-006):
  - Filepicker navigation feels immediate up to 1,000 entries.
  - Metadata visible within 5 s of selection on a reference connection.
    The `claude` subprocess invocation overhead MUST NOT regress this
    budget by ≥ 1 s in the median case.
  - Idle animation maintains ≥ 12 fps.
**Constraints**:
  - Captured frames MUST not persist on disk after the analysis
    goroutine exits (this spec FR-009; preserves 001 FR-023 intent).
  - In-flight `claude` subprocess MUST be cancellable on selection
    change (this spec FR-004; preserves 001 FR-020).
  - Layout MUST refuse to render below 100×30 *(inherited from 001)*.
  - The `ANTHROPIC_API_KEY` environment variable MUST NOT silently
    re-enable a direct-SDK code path *(this spec FR-011)*; the SDK is
    deleted from the source tree, so this is enforced by absence.
**Scale/Scope**: Single user, single video at a time, single `claude`
subprocess at a time *(unchanged from 001)*.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

Reference: [../../.specify/memory/constitution.md](../../.specify/memory/constitution.md), v1.0.0.

| Principle | Compliance |
|-----------|------------|
| **I. CLI Interface** | The user-facing CLI surface (`nano-gameplays`, `nano-gameplays identify --video PATH --json`) is unchanged: same flags, same exit codes, same `--json` payload. The new dependency on the `claude` binary is internal — user observable only as a preflight error if it is missing. **PASS.** |
| **II. Beautiful TUI** | No TUI changes; the analyzing-state spinner, idle animation, and metadata form are untouched. The `claude` subprocess runs off the UI goroutine via the existing `tea.Cmd` mechanism. **PASS.** |
| **III. Intelligent Claude Integration** | This is the *intent* of the change: route Claude calls through the user's own authenticated `claude` CLI rather than a separately-keyed SDK. Reproducibility is preserved — the prompt template is still embedded in the binary at version `identify-game.v1` and emitted verbatim into `analysis.prompt_version`; `analysis.model_id` is set to the `--model` value we passed to `claude` (or empty when no override was given, per FR-007); `analysis.frame_sha256` and `analysis.raw_response` are populated as before; `analysis.error` carries any failure verbatim. The model output is constrained by passing the schema through the `claude` CLI's `--json-schema` flag, eliminating the prose-parsing risk class. **PASS.** |
| **IV. Simplicity of Use** | The change *removes* a required environment variable (`ANTHROPIC_API_KEY`) for users who already have `claude` signed in — strictly an improvement on Principle IV. The new prerequisite (a signed-in `claude` CLI) is reasonable for the target user (a developer working with Claude Code) and is detected at preflight with a clear actionable message. **PASS.** |
| **V. Virality-Aware Output** | Not invoked on this screen *(unchanged)*. **PASS (not directly invoked).** |
| **Performance Standards** | The subprocess overhead for `claude -p --output-format json` on a vision call is dominated by the upstream HTTP latency, which is the same path the SDK took. SC-006 holds the line at ≤ 1 s additional median latency. **PASS (with SC-006 as the gate).** |
| **Dev Workflow & Quality Gates** | Constitution Check (this section) ✅. Reproducibility gate ✅ (prompt template version unchanged, model id propagated, frames hashed). CLI/TUI parity gate ✅ (no flag removals; `identify` subcommand still produces the v1 payload). Performance gate: a before/after measurement on a reference fixture is required before merge (recorded in Phase 6 polish). **Test gate inherits the deferral from 001-video-picker-screen — MUST be removed before merge to `main`.** |

**Result**: Constitution Check **PASS** with the same single inherited
deviation (tests deferred). Re-evaluated post-design at the bottom of
Phase 1; result **PASS** unchanged.

## Project Structure

### Documentation (this feature)

```text
specs/002-claude-cli-backend/
├── plan.md                  # This file
├── research.md              # Phase 0 output (this command)
├── data-model.md            # Phase 1 output (this command)
├── quickstart.md            # Phase 1 output (this command)
├── contracts/               # Phase 1 output (this command)
│   ├── claude-cli.md        # Subprocess invocation contract (NEW for v2)
│   ├── claude-prompt.md     # Prompt template + JSON schema (delta from 001)
│   ├── handoff.md           # Reference to unchanged 001 contract
│   └── cli.md               # Reference to unchanged 001 contract
├── checklists/
│   └── requirements.md      # From /speckit-specify
└── tasks.md                 # Generated by /speckit-tasks (not by this command)
```

### Source Code (repository root)

The source tree from `001-video-picker-screen` is preserved; this
feature swaps the implementation of `internal/analysis/claude.go` and
deletes the SDK dependency:

```text
nano-gameplays/
├── cmd/
│   └── nano-gameplays/
│       ├── main.go              # unchanged
│       ├── components.go        # unchanged (still wires analysis.Run as runner)
│       ├── identify.go          # unchanged (still calls analysis.ExtractFrames + Identify)
│       └── wire.go              # unchanged
├── internal/
│   ├── app/
│   │   └── *.go                 # unchanged
│   ├── ui/
│   │   ├── header/, footer/, leftpane/, rightpane/
│   │   └── *                    # all unchanged
│   ├── analysis/
│   │   ├── extractor.go         # unchanged (still produces in-memory CapturedFrame)
│   │   ├── claude.go            # REWRITTEN: shells out to `claude` CLI
│   │   ├── claudecli/
│   │   │   ├── invoke.go        # NEW: `claude -p` invocation builder + executor
│   │   │   ├── tempframes.go    # NEW: per-call tempdir for PNGs (mode 0700, defer cleanup)
│   │   │   └── parse.go         # NEW: --output-format json envelope parser
│   │   ├── prompt/
│   │   │   ├── identify_game_v1.go     # unchanged (still //go:embed)
│   │   │   ├── identify_game_v1.txt    # unchanged
│   │   │   └── identify_schema_v1.go   # NEW: JSON Schema constant for --json-schema
│   │   └── runner.go            # unchanged signature; calls Identify under the hood
│   └── handoff/
│       └── payload.go           # unchanged (schema_version remains "1")
├── scripts/
│   └── verify-roundtrip.sh      # unchanged
├── go.mod                       # `anthropic-sdk-go` removed; `go mod tidy` will prune
│                                # all transitive Anthropic deps
├── README.md                    # updated: drop ANTHROPIC_API_KEY, add claude CLI
└── specs/                       # /specs/...
```

**Structure Decision**: Keep the existing single Go module. Introduce a
new `internal/analysis/claudecli/` subpackage that owns every detail of
talking to the `claude` binary (argv assembly, tempdir lifecycle, output
envelope parsing). `claude.go` becomes a thin orchestrator: extract
frames → write to tempdir → invoke claudecli → cleanup → return
`(GameMetadata, AnalysisRecord, error)`. This keeps the call-site in
`runner.go` and the contract-with-the-app boundary intact.

## Complexity Tracking

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| **No automated tests on this branch (inherited).** | Same rationale as `001-video-picker-screen` — the integration shape with the `claude` CLI is still being shaken out and tests written today against argv would churn. | Adding tests first would slow validation of the chosen invocation form. Mitigations: (a) `internal/analysis/claudecli/` is a small surface (argv builder, parser) easy to unit-test once the shape stabilises; (b) the round-trip script (`scripts/verify-roundtrip.sh`) already smoke-tests the `identify --json` path end-to-end. **This deviation MUST be removed before either feature 001 or 002 merges to `main`.** |
| **Brief on-disk window for captured frames.** | The `claude` CLI's vision input flow needs file paths reachable via `--add-dir`; binary stdin for image content is not in the CLI's documented surface. | We considered streaming stdin (not supported), uploading via `--file` (that flag targets pre-uploaded Anthropic Files API resources, not local files), or skipping frame inputs entirely (loses identification accuracy). Mitigation: tempdir is created with mode `0700` under `os.TempDir()`, removed via `defer os.RemoveAll(...)` in the analysis goroutine — including on context cancellation and panic. Net effect: PNGs are filesystem-visible only for the duration of one identify call. |
