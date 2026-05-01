# Phase 0 — Research

Resolves the open technical questions implied by the spec and by the user's
direction "use the best tool available for CLI". One entry per decision.

---

## R-1: TUI framework

**Decision**: Bubble Tea + Lip Gloss + Bubbles (the Charm stack), in Go.

**Rationale**:
- Bubble Tea uses the Elm architecture (model / update / view) which maps
  cleanly onto the spec's pane state machine (idle → analyzing → metadata
  edit) and onto FR-020's "abandon in-flight on new selection" via message
  cancellation.
- Lip Gloss provides ergonomic primitives for exactly the four-region layout
  the spec demands (header, footer, left/right split).
- The `bubbles` package ships a working `filepicker` (FR-006…FR-010), a
  `spinner` for the analyzing state (FR-005), `help` for the contextual
  footer (FR-003), and `textinput` for editable metadata fields (FR-016).
  Each is a direct match to a spec requirement.
- Compiles to a single static binary, satisfying Principle IV (Simplicity of
  Use) at distribution time.

**Alternatives considered**:
- **Python + Textual** — also an excellent TUI framework. Rejected because
  ffmpeg + Claude SDK + a system Python add a runtime install burden that
  hurts the simplicity principle, and animation smoothness in Textual is
  good but not better than Bubble Tea's `tea.Tick`-driven loop.
- **Rust + ratatui** — equivalent capability, but ratatui requires more
  hand-rolled widgets (no `filepicker` of comparable quality off-the-shelf),
  which adds scope on this screen for no user-visible gain.
- **Node + ink** — animations and layout less polished; weaker filesystem
  navigation widgets; ffmpeg integration is via shell anyway, no advantage.

---

## R-2: Video frame extraction

**Decision**: Shell out to `ffmpeg` via `os/exec`, requesting two frames at
fixed-fraction offsets (10% and 60% of duration) decoded to in-memory PNG
bytes via `-f image2pipe -vcodec png -`.

**Rationale**:
- `ffmpeg` is universally available on developer machines and trivially
  installable everywhere this app runs (macOS / Linux). Linking a Go
  ffmpeg binding adds C-toolchain build complexity for a one-call use case.
- Two frames at well-separated offsets defends against an unrepresentative
  single frame (loading screen, all-black cutscene, menu — see spec Edge
  Cases). Two frames is also the upper bound the user described
  ("a screenshot or maybe 2").
- Reading via stdout pipe satisfies FR-023: frames never touch disk.

**Alternatives considered**:
- **Single mid-point frame** — simpler but fragile against menu/cutscene
  hits at the midpoint.
- **CGo binding to libavcodec** — best performance, but enormous build
  complexity for a tiny call site. Rejected.
- **Random offsets** — deterministic offsets are easier to reproduce in the
  analysis record (R-3).

**Failure handling**: ffmpeg non-zero exit, missing binary, or zero-byte
output → bubble up as a structured `ExtractError` to the model, which
satisfies FR-019 by surfacing the failure and keeping fields manually
editable.

---

## R-3: Claude vision call + reproducibility record

**Decision**: Use `github.com/anthropics/anthropic-sdk-go` to call the
current default Sonnet vision model (`claude-sonnet-4-6`, configurable via
`--model` and `NANO_GAMEPLAYS_MODEL` env var) with a versioned prompt
template `prompts/identify-game/v1.txt`, returning a strictly-structured
JSON response. Every call appends an `AnalysisRecord` to the in-memory
state, copied verbatim into the handoff payload.

**Prompt template (v1) requirements** (see
[contracts/claude-prompt.md](./contracts/claude-prompt.md)):
- Asks Claude to return JSON `{game_title, scene_or_level_or_mode,
  confidence: "low"|"medium"|"high"}`. No free text.
- Includes both extracted frames as image inputs.
- Pins `prompt_version: "identify-game.v1"`.

**AnalysisRecord fields**:
- `prompt_version` (string, e.g. `identify-game.v1`)
- `model_id` (string, e.g. `claude-sonnet-4-6`)
- `frame_offsets_pct` (`[10, 60]`)
- `frame_sha256` (`[hex, hex]`) — proves which images were sent without
  retaining the images
- `request_started_at`, `request_finished_at` (RFC 3339)
- `raw_response` (string) — the verbatim model output
- `error` (string, optional)

**Rationale**: Constitution Principle III mandates reproducibility (prompt
template, model ID, inputs recorded with each generated artifact).
Hashing-not-storing the frames preserves reproducibility while honouring
FR-023's no-disk rule. Strict JSON output simplifies confidence-bucket
mapping (R-6) and removes the need for parsing free-form English.

**Alternatives considered**:
- **Free-form English response, parse on client** — rejected; brittle and
  loses Principle III's "explicit failure" property.
- **Store frames on disk for replay** — rejected; would break FR-023.
  Hashing is sufficient for proof-of-input.

**Failure handling**: any error from the SDK → `AnalysisRecord.error` set
and FR-018's empty-but-editable form shown.

---

## R-4: Cancellation of in-flight analysis

**Decision**: Wrap each Claude call in a `context.Context` that the model's
update loop holds. On `selectVideoMsg` for a *different* path, the model
calls the cancel func before issuing a new analysis command.

**Rationale**: Direct, idiomatic Go, no extra dependency. Maps to FR-020.
Bubble Tea's `tea.Cmd` returns a message; the cancellation just causes the
in-flight goroutine's HTTP request to be aborted and its result to be
dropped on arrival (the message handler checks the path against current
state).

**Alternatives considered**: result-channel select with a "version" counter.
Equivalent semantics, more code. Rejected.

---

## R-5: Idle animation cadence

**Decision**: 12 frames-per-second tick (`time.NewTicker(83ms)` / `tea.Tick`)
for the PS5 ASCII idle animation and the analyzing animation. Idle
animation is a hand-authored 8-frame cycle of the PS5 logo and pulsing
beams; the analyzing animation reuses `bubbles/spinner` with a tasteful
spinner style (`spinner.MiniDot` or `spinner.Pulse`).

**Rationale**: 12 fps is plenty to read as motion in a terminal, costs
negligible CPU, and avoids audible-fan terminal redraw at higher rates.
Separate animation states (FR-004 vs FR-005) are state-machine transitions
on the same ticker.

---

## R-6: Mapping Claude's confidence to low/med/high

**Decision**: Have Claude itself emit a single categorical value `low`,
`medium`, or `high` directly in the JSON response (per R-3). No client-side
threshold mapping.

**Rationale**: Avoids inventing a numeric→bucket threshold the user can't
reason about. Aligns with the clarification (Q4 → B). The prompt
template's response schema enforces the value set.

**Alternatives considered**: ask for a 0–100 numeric and bucket on the
client. Rejected — adds a configurable threshold that nobody asked for.

---

## R-7: CLI/TUI parity (non-interactive flow)

**Decision**: Provide `nano-gameplays identify --video PATH [--start-dir
DIR] [--model ID] --json` as the non-interactive equivalent of the screen.
Output is the same `HandoffPayload` JSON (see
[contracts/handoff.md](./contracts/handoff.md)) the TUI confirm path
produces, with the only difference being that the `metadata` fields are
not user-edited.

**Rationale**: Constitution Principle I and the Dev Workflow CLI/TUI
parity gate require this. It is also the foothold for re-introducing
tests later (smoke test for `identify`, contract test for the JSON
schema) without TUI test infrastructure.

**Alternatives considered**: omit the non-interactive path until the next
screen is also defined. Rejected — the constitution gate is per-feature,
and skipping would create a parity debt that compounds.

---

## R-8: Terminal-size minimum

**Decision**: Hard minimum **100 columns × 30 rows** (matches SC-007).
Below that, Bubble Tea program exits with a single-line message to stderr
and exit code `2`.

**Rationale**: Pre-flight check is trivial via `tea.WindowSizeMsg` on
program start; failing fast with a clear message is far better UX than a
broken half-rendered shell.
