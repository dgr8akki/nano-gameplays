---
description: "Task list for feature 001-video-picker-screen"
---

# Tasks: Initial UI — Video Picker & Game Identification

**Input**: Design documents from `specs/001-video-picker-screen/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

**Tests**: Test tasks are intentionally omitted. The user explicitly requested
"no tests needed for now" at planning time and the deviation is logged in
`plan.md` → Complexity Tracking. This deviation MUST be revisited before this
branch is merged to `main`.

**Organization**: Tasks are grouped by user story (US1=P1, US2=P2, US3=P3).
Each story phase is independently demonstrable per `spec.md`.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1, US2, US3)
- File paths are repo-relative

## Path Conventions

Single Go module at the repo root. Source under `cmd/` and `internal/` per
`plan.md` → Project Structure.

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization, dependency wiring, build tooling.

- [X] T001 Initialize Go module at repo root: create `go.mod` with module path `github.com/dgr8akki/nano-gameplays` and Go directive `1.22`
- [X] T002 Add primary dependencies in `go.mod` and `go.sum`: `github.com/charmbracelet/bubbletea`, `github.com/charmbracelet/lipgloss`, `github.com/charmbracelet/bubbles`, `github.com/anthropics/anthropic-sdk-go` (run `go get` for each, commit lockfile)
- [X] T003 [P] Create `.gitignore` at repo root excluding `bin/`, `*.test`, `*.out`, `.env`
- [X] T004 [P] Create `Makefile` at repo root with targets `build` (→ `bin/nano-gameplays`), `run`, `tidy`, `vet`, `fmt`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Cross-cutting types and the Bubble Tea program skeleton that
every user story phase depends on.

**⚠️ CRITICAL**: No user story work can begin until this phase is complete.

- [X] T005 Create `internal/handoff/payload.go` defining the `HandoffPayload`, `GameplayVideo`, `GameMetadata`, and `AnalysisRecord` structs exactly per `specs/001-video-picker-screen/contracts/handoff.md` and `data-model.md`, with `encoding/json` struct tags and `schema_version` pinned to `"1"`
- [X] T006 Create `internal/app/keys.go` defining the global `KeyMap` (quit, up, down, enter, back, confirm) using `github.com/charmbracelet/bubbles/key` and a `Help()` method exposing bindings to the footer
- [X] T007 Create `internal/app/model.go` with the top-level Bubble Tea `Model`, the `AppState` enum (`StatePicker`, `StateAnalyzing`, `StateEditing`, `StateDone`), terminal-size tracking via `tea.WindowSizeMsg`, and a preflight check that exits with code `2` and a stderr message when size < 100×30
- [X] T008 Create `internal/app/update.go` with the top-level `Update` function that routes `tea.Msg` values by current `AppState` to the appropriate pane handlers (stubs allowed for not-yet-implemented states)
- [X] T009 Create `internal/app/view.go` that composes the four-region layout (header, footer, left pane, right pane) using `lipgloss.JoinVertical` and `lipgloss.JoinHorizontal`, sized from the tracked window dimensions
- [X] T010 Create `cmd/nano-gameplays/main.go` with `flag` parsing for `--start-dir`, `--model`, `--no-tui` and an `identify` subcommand router (subcommand wiring may be a stub returning "not yet implemented" until US2)

**Checkpoint**: Skeleton compiles, launches the TUI showing four empty regions, and exits cleanly on `q` / `ctrl+c`.

---

## Phase 3: User Story 1 — Pick a gameplay video from disk (Priority: P1) 🎯 MVP

**Goal**: A user launches the app, navigates the filesystem with the
keyboard, selects a PlayStation gameplay video, and the app records the
selection and emits a minimal handoff payload to stdout. Spec: `Story 1`.

**Independent Test**: Launch in a directory containing at least one
`.mp4`/`.mov`/`.mkv`/`.webm` file. Navigate using only the keyboard, select a
video, press the confirm key. Observe a `HandoffPayload` JSON object on
stdout whose `video.path` matches the chosen file. No analysis needed.

### Implementation for User Story 1

- [X] T011 [P] [US1] Create `internal/ui/leftpane/picker.go` wrapping `github.com/charmbracelet/bubbles/filepicker.Model`, restricting selectable extensions to `.mp4`, `.mov`, `.mkv`, `.webm` (set `AllowedTypes`)
- [X] T012 [US1] In `internal/ui/leftpane/picker.go`, render non-selectable files with a dimmed lipgloss style so video files are visually distinguished (FR-010)
- [X] T013 [US1] In `internal/ui/leftpane/picker.go`, on a confirm keypress over a non-video file, emit a transient inline "not a supported video" message and keep focus in the picker (FR-011)
- [X] T014 [US1] In `internal/app/update.go`, route messages to the leftpane while `AppState == StatePicker` and on selection emit a `selectVideoMsg` carrying the absolute path
- [X] T015 [US1] In `internal/app/view.go`, render the leftpane in the left half during `StatePicker` and a placeholder string ("right pane") in the right half (the polished idle animation is US3)
- [X] T016 [US1] In `internal/ui/header/header.go` (new) and `internal/ui/footer/footer.go` (new), render a minimal header (literal `nano-gameplays`) and a minimal footer listing currently-bound keys via the keymap from T006 (the styled, context-aware version is US3)
- [X] T017 [US1] In `internal/app/update.go`, on `selectVideoMsg` build a minimal `HandoffPayload` (populated `video`, empty `metadata`, empty `analysis`) using `internal/handoff` and emit it to stdout via a `tea.Quit` command after the next render
- [X] T018 [US1] In `cmd/nano-gameplays/main.go`, wire `--start-dir` (default = `os.Getwd()`) into the leftpane initial directory; ensure the terminal-size preflight from T007 prints to stderr and exits `2` if the terminal is too small before the program loop starts

**Checkpoint**: User Story 1 fully functional. Picker navigates, filters,
rejects non-videos, confirms, and prints a `HandoffPayload` JSON with only
the `video` block populated.

---

## Phase 4: User Story 2 — Auto-fill game metadata from the selected video (Priority: P2)

**Goal**: After selection, the right pane shows "Analyzing…", the app
extracts 1–2 frames via `ffmpeg`, calls Claude with the versioned prompt,
and renders editable fields for game title and scene/level/mode plus a
confidence badge. The user edits if needed and confirms. Spec: `Story 2`.

**Independent Test**: With US1 working, select a video. Within 5 s of
selection (success path) the right pane shows three fields populated by
analysis. Edit a field, press the confirm key. Stdout `HandoffPayload`
includes both `metadata` and a non-empty `analysis` block whose
`prompt_version` is `identify-game.v1`. Force a Claude failure (e.g. unset
`ANTHROPIC_API_KEY`) and verify the form still appears empty/editable, the
user can type values manually, and `analysis.error` is populated in the
final payload.

### Implementation for User Story 2

- [X] T019 [P] [US2] Create `internal/analysis/extractor.go` with `ExtractFrames(ctx context.Context, videoPath string) ([]CapturedFrame, error)` that shells out to `ffmpeg` (`-ss <pct>` of duration via `ffprobe` first, `-frames:v 1 -f image2pipe -vcodec png -`) at offsets 10% and 60%, returning in-memory PNG byte slices and SHA-256 hashes; surface an `ExtractError` on non-zero exit, missing binary, or zero-byte output
- [X] T020 [P] [US2] Create `internal/analysis/prompt/identify_game_v1.go` exposing the prompt template constant exactly per `contracts/claude-prompt.md` (system instruction + required JSON response schema), with the version identifier `identify-game.v1`
- [X] T021 [P] [US2] Create `internal/analysis/claude.go` with `Identify(ctx context.Context, frames []CapturedFrame, modelID string) (GameMetadata, AnalysisRecord, error)` using `anthropic-sdk-go`, sending the frames as image content blocks alongside the prompt from T020, strictly JSON-unmarshalling the response into `{game_title, scene_or_level_or_mode, confidence}`, populating `AnalysisRecord` (model id, prompt version, frame hashes, RFC3339 timestamps, raw response, error)
- [X] T022 [US2] In `internal/analysis/claude.go`, coerce any `confidence` value other than `low`/`medium`/`high` to `low` and surface the coercion in `AnalysisRecord.error` (per `contracts/claude-prompt.md`)
- [X] T023 [US2] Create `internal/ui/rightpane/analyzing.go` showing the literal text "Analyzing…" alongside `bubbles/spinner.Model` (style `spinner.MiniDot` or `spinner.Pulse`) and rendered while `AppState == StateAnalyzing`
- [X] T024 [US2] Create `internal/ui/rightpane/metadata.go` rendering the editable form: two `bubbles/textinput.Model` fields (Game title, Scene/Level/Mode) plus a non-editable confidence badge that maps `low|medium|high` to a colored lipgloss style (red/yellow/green or icon equivalents); confirm key is disabled while game title is empty (data-model validation), with a footer hint indicating the requirement
- [X] T025 [US2] In `internal/app/update.go`, transition `StatePicker → StateAnalyzing` on `selectVideoMsg` and dispatch a `tea.Cmd` that runs T019 then T021 under a per-state `context.Context` stored on the model
- [X] T026 [US2] In `internal/app/update.go`, on a new `selectVideoMsg` while `StateAnalyzing` or `StateEditing`, call the stored cancel function and discard any late-arriving result whose video path does not match the current selection (FR-020)
- [X] T027 [US2] In `internal/app/update.go`, on `analysisErrMsg` (from `ExtractError` or Claude error) move to `StateEditing` with empty fields and `confidence = low`, and pass an error banner string into the metadata view to display (FR-018, FR-019)
- [X] T028 [US2] In `internal/app/update.go`, on the confirm key during `StateEditing`, build a fully-populated `HandoffPayload` including the `AnalysisRecord` from T021 (or the failure record from T027) and emit it to stdout via `tea.Quit`-then-print, mirroring the format produced in US1 T017
- [X] T029 [US2] In `cmd/nano-gameplays/main.go`, implement the `identify` subcommand: parse `--video PATH`, `--model`, `--json`; run the same extractor + Claude flow without launching Bubble Tea; emit the same `HandoffPayload` JSON on stdout (or a short human summary without `--json`); exit codes `0` payload emitted, `2` precondition failure (per `contracts/cli.md`)

**Checkpoint**: Both User Story 1 and User Story 2 work end-to-end. The
`identify` subcommand produces an identical payload shape to the TUI confirm
path. Failure paths (no API key, no network, ffmpeg missing, malformed
response) are recoverable.

---

## Phase 5: User Story 3 — Polished split-screen shell (Priority: P3)

**Goal**: Header, footer, left and right panes all rendered with
intentional visual style. Right pane shows a PS5-themed ASCII idle
animation while the user is in the picker, swaps to a polished animating
"Analyzing…" view during analysis, and the footer keys update with focus
context. Spec: `Story 3`.

**Independent Test**: Launch the app and, without performing any work,
visually confirm: header shows `nano-gameplays` in styled banner; right
pane shows a PS5 ASCII animation cycling smoothly (~12 fps); footer lists
the keys currently active in the focused region; on selection the right
pane swaps to the styled analyzing view; after metadata appears, the
confidence badge shows the right color/icon for its level.

### Implementation for User Story 3

- [X] T030 [P] [US3] Create `internal/ui/rightpane/idle.go` containing an 8-frame ASCII PS5 logo / pulsing-beams animation cycle, advanced by a `tea.Tick(83 * time.Millisecond)` ticker (~12 fps) and stopped/restarted on state transitions
- [X] T031 [P] [US3] Replace the minimal banner in `internal/ui/header/header.go` (from T016) with a styled lipgloss banner: bordered box, accent color, app name centered
- [X] T032 [US3] Upgrade `internal/ui/footer/footer.go` (from T016) to a context-aware help line driven by `bubbles/help` whose key set switches between picker bindings, editing bindings, and analyzing bindings as `AppState` changes (FR-003)
- [X] T033 [US3] In `internal/app/view.go`, render the idle animation from T030 in the right half whenever `AppState == StatePicker`, replacing the placeholder right pane introduced in T015
- [X] T034 [US3] In `internal/ui/rightpane/analyzing.go` (from T023), polish the layout: centered "Analyzing…" label above the spinner, fixed minimum width matched to the right-pane size, gentle accent color
- [X] T035 [US3] In `internal/ui/rightpane/metadata.go` (from T024), ensure the confidence badge has a clear icon-or-color cue for each of low/medium/high (e.g. `●` glyph in red/yellow/green) and that the badge's lipgloss style is consistent with the header's accent palette

**Checkpoint**: All three user stories shipped. The screen looks like the
spec describes end-to-end.

---

## Phase 6: Polish & Cross-Cutting Concerns

- [X] T036 Embed prompt template content into the binary using `//go:embed` in `internal/analysis/prompt/identify_game_v1.go` so the binary is self-contained
- [X] T037 [P] Add a top-level `README.md` at the repo root summarising the project, linking to `specs/001-video-picker-screen/{spec.md, plan.md, quickstart.md}` and the constitution at `.specify/memory/constitution.md`
- [X] T038 [P] Verify `HandoffPayload` JSON round-trips losslessly: write a short manual run script (`scripts/verify-roundtrip.sh`) that calls `nano-gameplays identify --video <fixture> --json` and pipes the output through `jq -e '.schema_version == "1" and (.analysis.prompt_version == "identify-game.v1")'`
- [X] T039 Walk through `quickstart.md` end-to-end on a real machine, capture a sample `HandoffPayload` to `specs/001-video-picker-screen/sample-payload.json`, and confirm SC-001..SC-007 pass on a representative environment

---

## Dependencies & User-Story Completion Order

```
Phase 1 (Setup)
   ↓
Phase 2 (Foundational)
   ↓
   ├─→ Phase 3 (US1 — MVP)            ← ship this first
   │       ↓
   │       └─→ Phase 4 (US2)          ← ship next
   │              ↓
   │              └─→ Phase 5 (US3)   ← ship last
   ↓
Phase 6 (Polish)
```

- US2 depends on US1 (it operates on the video US1 selects).
- US3 depends on US1 and US2 (it polishes the surfaces both stories
  render). T030 (idle animation) and T031 (header banner) are themselves
  parallelisable inside US3.

## Parallel Execution Examples

Within a phase, tasks marked `[P]` touch different files and can be
implemented in parallel:

- **Phase 1**: `T003` (gitignore) and `T004` (Makefile) in parallel.
- **Phase 4 (US2)**: `T019` (extractor), `T020` (prompt template), and
  `T021` (Claude client) all live in distinct files and can be developed
  in parallel by separate agents/sessions.
- **Phase 5 (US3)**: `T030` (idle animation), `T031` (header banner) in
  parallel.
- **Phase 6**: `T037` (README) and `T038` (round-trip script) in
  parallel.

## Implementation Strategy

1. Ship **US1** end-to-end first (Phases 1 → 2 → 3). This is the MVP. The
   product can already select a video and emit a minimal handoff JSON.
2. Layer **US2** on top (Phase 4). The product becomes intelligent: it
   pre-fills game metadata using Claude.
3. Polish with **US3** (Phase 5). The product becomes beautiful.
4. Apply Phase 6 polish before opening the merge PR. Re-introduce tests at
   that point per the deferral logged in `plan.md` → Complexity Tracking.
