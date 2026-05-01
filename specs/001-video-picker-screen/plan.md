# Implementation Plan: Initial UI — Video Picker & Game Identification

**Branch**: `001-video-picker-screen` | **Date**: 2026-05-02 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `specs/001-video-picker-screen/spec.md`

## Summary

Deliver the application's initial screen as a Bubble Tea (Go) TUI: a four-region
layout (header / footer / left pane / right pane) where the user keyboard-picks
a PlayStation gameplay video on the left, while the right pane shows a
PS5-themed idle animation. On selection, the right pane transitions to an
"Analyzing…" animation; the app extracts 1–2 representative frames from the
video via `ffmpeg`, sends them to Claude (vision-capable) for game
identification, and renders the three returned fields — game title,
scene/level/mode, confidence (low/med/high) — in editable form. The user
edits if needed and confirms, producing a `(video path, metadata)` handoff
payload for the next screen. Frames are held in memory only.

## Technical Context

**Language/Version**: Go 1.22+ (single static binary, easy distribution)
**Primary Dependencies**:
  - `github.com/charmbracelet/bubbletea` — TUI runtime (Elm architecture)
  - `github.com/charmbracelet/lipgloss` — layout primitives (header/footer/split)
  - `github.com/charmbracelet/bubbles` — pre-built `filepicker`, `spinner`, `help`, `textinput`
  - `github.com/anthropics/anthropic-sdk-go` — Claude API client (vision)
  - `ffmpeg` (system binary, invoked via `os/exec`) — frame extraction
**Storage**: None on disk for this screen. In-memory only (frames, in-flight metadata). Application config (e.g., API key) read from env / config file at startup is out of scope for this screen.
**Testing**: Deferred for v1 per user direction ("no tests needed for now"). See Complexity Tracking — this is a deliberate, time-boxed deviation from the project's general Dev Workflow gates and MUST be revisited before this branch is merged to `main`.
**Target Platform**: macOS and Linux terminals, ≥ 100 columns × 30 rows, truecolor or 256-color terminal.
**Project Type**: Single-project CLI/TUI application (single Go module).
**Performance Goals** (from spec SC-005, SC-006):
  - Filepicker navigation feels immediate up to 1,000 entries per directory.
  - Metadata visible within 5 s of selection on a reference connection.
  - Idle animation maintains a smooth visible frame rate (target ≥ 12 fps).
**Constraints**:
  - Captured frames MUST stay in memory only (FR-023).
  - Layout MUST refuse to render below 100×30 with a clear message (SC-007).
  - In-flight Claude calls MUST be cancellable on selection change (FR-020).
**Scale/Scope**: Single user, single video at a time, single in-flight Claude call at a time. One screen of an eventually multi-screen app.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

Reference: [../../.specify/memory/constitution.md](../../.specify/memory/constitution.md), v1.0.0.

| Principle | Compliance |
|-----------|------------|
| **I. CLI Interface** | Default invocation launches the TUI. A non-interactive equivalent — `nano-gameplays identify --video PATH [--start-dir DIR] --json` — MUST emit the same `(video path, metadata)` payload to stdout (see [contracts/cli.md](./contracts/cli.md)). All flags documented; exit codes follow convention; errors → stderr; `--json` toggles machine-readable output. **PASS.** |
| **II. Beautiful TUI** | Bubble Tea + Lip Gloss + Bubbles is the chosen stack precisely because it supports header/footer/split layouts, animations, contextual help, and the `filepicker` and `spinner` widgets the spec literally describes. Graceful degradation: under `--no-tui` (or non-TTY stdout), the program runs the non-interactive flow above. **PASS.** |
| **III. Intelligent Claude Integration** | Game identification is performed by a Claude vision call (see [research.md](./research.md), R-3). Each call's prompt template version, model ID, image hash(es), and response are included verbatim in the handoff payload's `analysis_record` field, satisfying the constitution's reproducibility requirement. Failures surface explicitly per FR-018; no silent heuristic fallback. **PASS.** |
| **IV. Simplicity of Use** | Zero required flags: `nano-gameplays` launches the TUI starting at the current working directory. All behaviour overridable by flags or config file (precedence: defaults → config → flags). Filepicker and editable-field widgets reused from `bubbles`, not reinvented. **PASS.** |
| **V. Virality-Aware Output** | Not directly produced on this screen (this screen produces metadata, not shorts). The screen's output schema includes the fields downstream virality scoring requires (game title + scene/level/mode), so this screen does not block the principle. **PASS (not directly invoked).** |
| **Performance Standards** | This screen processes one video at a time and does not yet exercise the batched-rendering performance bar; that bar binds to later screens. The screen does emit a per-call `analysis_record` with timing fields so future regressions are detectable. **PASS (scope-appropriate).** |
| **Dev Workflow & Quality Gates** | Constitution Check (this section) ✅. Reproducibility gate ✅ (prompt template versioned, see [contracts/claude-prompt.md](./contracts/claude-prompt.md)). CLI/TUI parity gate ✅ ([contracts/cli.md](./contracts/cli.md)). Performance gate: a reference-input perf record will be required before merge. **Test gate explicitly deferred — see Complexity Tracking.** |

**Result**: Initial Constitution Check **PASS** with one tracked deviation
(tests deferred). Re-evaluated post-design at the bottom of Phase 1; result
**PASS** unchanged.

## Project Structure

### Documentation (this feature)

```text
specs/001-video-picker-screen/
├── plan.md                  # This file
├── research.md              # Phase 0 output
├── data-model.md            # Phase 1 output
├── quickstart.md            # Phase 1 output
├── contracts/               # Phase 1 output
│   ├── cli.md               # Non-interactive CLI surface (parity with TUI)
│   ├── handoff.md           # Payload handed to the next screen / stdout
│   └── claude-prompt.md     # Prompt template + reproducibility record schema
├── checklists/
│   └── requirements.md      # From /speckit-specify
└── tasks.md                 # Generated by /speckit-tasks (not by this command)
```

### Source Code (repository root)

Single-project Go module:

```text
nano-gameplays/
├── cmd/
│   └── nano-gameplays/
│       └── main.go              # Entry point: parses flags, launches TUI or
│                                # non-interactive `identify` flow
├── internal/
│   ├── app/
│   │   ├── model.go             # Top-level Bubble Tea model: holds the
│   │   │                        # four-region layout state machine
│   │   ├── update.go            # Update loop: routes msgs to panes
│   │   ├── view.go              # Compose header / footer / left / right
│   │   └── keys.go              # Keymap + footer help binding
│   ├── ui/
│   │   ├── header/              # Header view (app name)
│   │   ├── footer/              # Context-aware key hints (uses bubbles/help)
│   │   ├── leftpane/            # File picker (wraps bubbles/filepicker,
│   │   │                        # adds video-extension filter + dim non-video)
│   │   └── rightpane/
│   │       ├── idle.go          # PS5 ASCII idle animation (tea.Tick driven)
│   │       ├── analyzing.go     # "Analyzing…" + spinner animation
│   │       └── metadata.go      # Editable metadata form (textinput x2 +
│   │                            # confidence badge)
│   ├── analysis/
│   │   ├── extractor.go         # ffmpeg wrapper: pick 1-2 representative
│   │   │                        # frames into in-memory []byte
│   │   ├── claude.go            # anthropic-sdk-go vision call,
│   │   │                        # cancellable via context.Context
│   │   └── record.go            # AnalysisRecord struct (reproducibility)
│   └── handoff/
│       └── payload.go           # Handoff payload schema (used by both TUI
│                                # confirm path and non-interactive --json)
└── go.mod
```

**Structure Decision**: Single Go module under the repo root. `cmd/` holds the
binary entry point; `internal/` holds non-exported packages organized by
concern (`app` = Bubble Tea program, `ui/*` = view layers per pane,
`analysis` = frame extraction + Claude, `handoff` = the cross-screen payload
shape that is also what `--json` emits). No `tests/` tree until tests are
re-introduced (see Complexity Tracking).

## Complexity Tracking

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| **No automated tests on this branch.** | Project is at the prototyping stage; the user explicitly requested no tests for now to keep iteration speed maximal while the UX shape stabilises. | Adding the test suite first would slow down the still-volatile UX shape (filepicker layout, animation timing, Claude prompt) and would be discarded with each iteration. Mitigations: (a) the non-interactive `identify` CLI surface is testable later without TUI mocks; (b) the reproducibility record makes Claude responses replayable; (c) **this deviation MUST be removed before this branch is merged to `main`** — at minimum, contract tests for the handoff payload and a smoke test for the `identify` CLI must be added. |
