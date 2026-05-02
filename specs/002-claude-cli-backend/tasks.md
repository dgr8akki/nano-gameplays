---
description: "Task list for feature 002-claude-cli-backend"
---

# Tasks: Use Local `claude` CLI Instead of Anthropic SDK

**Input**: Design documents from `specs/002-claude-cli-backend/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

**Tests**: Test tasks are intentionally omitted. The deferral inherited
from feature `001-video-picker-screen` (Complexity Tracking → "no
automated tests on this branch") still applies and is duplicated in
this plan's Complexity Tracking. The deferral MUST be removed before
either branch is merged to `main`.

**Organization**: Tasks are grouped by user story (US1=P1, US2=P2,
US3=P3). Each story phase is independently demonstrable per `spec.md`.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1, US2, US3)
- File paths are repo-relative

## Path Conventions

Single Go module at the repo root, structure inherited from
`001-video-picker-screen` and extended with `internal/analysis/claudecli/`
per `plan.md` → Project Structure.

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Tear down the SDK-era runtime and prepare a clean tree for
the `claude` CLI subprocess implementation.

- [X] T001 Delete the `github.com/anthropics/anthropic-sdk-go` import from `internal/analysis/claude.go` and remove every other runtime reference (types, helpers, error wrappers); run `go mod tidy` from repo root and commit the pruned `go.mod`/`go.sum`. Final tree MUST contain zero matches for `grep -RnE "anthropic-sdk-go|ANTHROPIC_API_KEY" cmd/ internal/ scripts/`.

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Land the new `internal/analysis/claudecli/` subpackage
skeleton, the embedded JSON Schema, and the program-start preflight for
the `claude` binary. Every user-story phase depends on these.

**⚠️ CRITICAL**: No user story work can begin until this phase is complete.

- [X] T002 [P] Create `internal/analysis/prompt/identify_schema_v1.json` with the canonical identify-game schema (per `contracts/claude-prompt.md` §"Response shape"): top-level object, `additionalProperties: false`, `required: [game_title, scene_or_level_or_mode, confidence]`, `confidence` enum `["low","medium","high"]`.
- [X] T003 [P] Create `internal/analysis/prompt/identify_schema_v1.go` exposing the schema via `//go:embed identify_schema_v1.json` as `IdentifySchemaV1() string` and a constant `IdentifySchemaVersion = "identify-game.v1"` mirroring `IdentifyGameVersion` so the schema and prompt bump together.
- [X] T004 [P] Create `internal/analysis/claudecli/tempframes.go` exposing `WriteFrames(frames []analysis.CapturedFrame) (dir string, paths []string, err error)` that creates `os.MkdirTemp("", "nano-gameplays-*")` (mode `0700`), writes each frame as `frame_<offsetPct>.png` (mode `0600`), and returns the dir + per-frame absolute paths. Avoid an import cycle with `internal/analysis` by accepting a slim local input type instead of importing `analysis.CapturedFrame` directly (e.g., `Frame{Bytes []byte; OffsetPct int}`).
- [X] T005 [P] Create `internal/analysis/claudecli/parse.go` defining `ResultEnvelope { Type string; Result string; IsError bool; Model string; Error string }` per `data-model.md` §`claudecli.ResultEnvelope`, plus `DecodeEnvelope(stdout []byte) (ResultEnvelope, error)` that JSON-unmarshals the `claude -p --output-format json` envelope and tolerates extra envelope keys.
- [X] T006 Add a `claude` binary preflight to `cmd/nano-gameplays/main.go`: call `exec.LookPath("claude")` before `tea.NewProgram(...)` (and at the top of `runIdentify` for the non-interactive path); on failure print `claude CLI not found in PATH; install Claude Code from https://claude.com/claude-code` to stderr and exit `2`. Mirror the existing `ffmpeg`-style preflight semantics from feature 001.

**Checkpoint**: `go build ./...` passes with the new subpackage in place,
the SDK is gone, and `nano-gameplays` exits cleanly with the new
preflight message when `claude` is unavailable.

---

## Phase 3: User Story 1 — Identify a game without managing an API key (Priority: P1) 🎯 MVP

**Goal**: A developer with the `claude` CLI signed in can run
`nano-gameplays`, pick a video, and see the metadata pane fill in —
without ever exporting `ANTHROPIC_API_KEY`. Spec: `Story 1`.

**Independent Test**: From a shell with `unset ANTHROPIC_API_KEY` and
`claude --version` succeeding, run `./bin/nano-gameplays identify
--video <fixture> --json` and observe a `HandoffPayload` with
`schema_version: "1"`, populated `metadata.game_title`, and
`analysis.error == ""`.

### Implementation for User Story 1

- [X] T007 [US1] In `internal/analysis/claudecli/invoke.go`, implement `BuildArgv(opts InvokeOpts) []string` that assembles the canonical argv from `contracts/claude-cli.md`: `claude -p --output-format json --system-prompt <body> --json-schema <schema> [--model <id>] --add-dir <tempdir> --allowedTools Read --disallowedTools Bash --disable-slash-commands --no-session-persistence -- "<user message>"`. The `--model` flag MUST be omitted from argv when `opts.ModelID == ""` (FR-008).
- [X] T008 [US1] In `internal/analysis/claudecli/invoke.go`, implement `Run(ctx context.Context, opts InvokeOpts) (ResultEnvelope, []byte /*stderr*/, error)` that wraps `exec.CommandContext`, closes stdin, captures stdout to a buffer, captures stderr to a bounded buffer (cap 64 KiB), waits, and returns the decoded envelope (via the parser from T005) plus the captured stderr bytes for failure paths.
- [X] T009 [US1] Rewrite `internal/analysis/claude.go` so `Identify(ctx, frames, modelID)` orchestrates: (a) `claudecli.WriteFrames(frames)` then `defer os.RemoveAll(dir)`, (b) build `InvokeOpts` carrying the embedded prompt body (`prompt.IdentifyGameV1()`), the embedded schema (`prompt.IdentifySchemaV1()`), the resolved `modelID`, the tempdir, and the user message ("Identify the game shown in the frames at <path1> and <path2>. Respond per the schema."), (c) call `claudecli.Run(ctx, opts)`, (d) JSON-unmarshal `envelope.Result` into the identify-response struct, (e) return `(handoff.GameMetadata, handoff.AnalysisRecord, error)`. Delete the SDK code path entirely. The function MUST keep its existing signature so callers in `runner.go` and `cmd/nano-gameplays/identify.go` are unchanged.
- [X] T010 [US1] In `internal/analysis/claude.go`, capture `request_started_at` immediately before the subprocess starts and `request_finished_at` immediately after `claudecli.Run` returns; populate the rest of the `AnalysisRecord` (prompt_version from `prompt.IdentifyGameVersion`, frame_offsets_pct + frame_sha256 from the input slice, raw_response from `envelope.Result`, error empty on the success path).
- [X] T011 [US1] Smoke-verify the end-to-end MVP: build the binary, run `./bin/nano-gameplays identify --video <fixture> --json` in a shell where `ANTHROPIC_API_KEY` is unset, observe stdout is a single-line `HandoffPayload` JSON with `schema_version=1` and `analysis.prompt_version=identify-game.v1`. Pipe the output through `./scripts/verify-roundtrip.sh <fixture>` to confirm the same.

**Checkpoint**: User Story 1 is fully functional. The Anthropic SDK is
not reachable from any code path. A developer with `claude` signed in
can identify a video without setting `ANTHROPIC_API_KEY`.

---

## Phase 4: User Story 2 — Clear, actionable errors when `claude` is missing or broken (Priority: P2)

**Goal**: Failure paths are diagnosable. Missing binary fails fast at
preflight; signed-out / errored calls leave the editable form open with
a banner so the user can still emit a payload. Spec: `Story 2`.

**Independent Test**: Three shells: (a) `PATH=/usr/bin nano-gameplays`
(no `claude` binary) exits `2` within ~1 s with the install-hint
message; (b) with `claude` present but `claude auth` cleared, run the
TUI, select a video, see the editable form with a non-empty banner;
(c) `nano-gameplays identify --video <fixture> --json` in the same
auth-cleared environment emits a payload whose `analysis.error` is
non-empty.

### Implementation for User Story 2

- [X] T012 [US2] Refine the preflight from T006 in `cmd/nano-gameplays/main.go` to centralise the message strings (e.g. a `preflightClaudeBinary()` helper) so both `runTUI` and `runIdentify` emit byte-identical stderr. Add the same preflight wording to `runIdentify` so `identify` exits `2` with the install hint when the binary is gone.
- [X] T013 [US2] In `internal/analysis/claude.go`, branch on the result of `claudecli.Run`: non-zero exit OR `envelope.IsError == true` OR `json.Unmarshal(envelope.Result, &resp) != nil` → return `(GameMetadata{Confidence:"low"}, AnalysisRecord{Error: <descriptive>}, err)`. The error string must include (in priority order) the envelope error, the unmarshal failure, or the captured stderr (truncated to 1 KiB), prefixed with what stage failed (`claude exit / envelope-error / schema-validation / decode`).
- [X] T014 [US2] In `internal/analysis/claudecli/invoke.go`, ensure `defer os.RemoveAll(opts.TempDir)` runs in `Run` even when `cmd.Run()` returns context.Canceled or panics; add a `recover()` wrapper if needed. Add a unit-style assertion (a small main-style scratch program OR an `_example_test.go` is fine; no full test framework) confirming the dir is gone after a forced cancel.
- [X] T015 [US2] In `internal/analysis/claude.go`, ensure context cancellation between subprocess start and parse cleanup never leaks the tempdir: prefer ownership at the `Identify` layer (it created the dir, it owns the deferred cleanup); the `claudecli` package only writes into the dir.

**Checkpoint**: Stories 1 and 2 work end-to-end. Failure modes match
the spec's three failure-class scenarios (missing binary, signed-out,
malformed output) and produce diagnosable messages or a recoverable
editable form.

---

## Phase 5: User Story 3 — Reproducibility record stays intact (Priority: P3)

**Goal**: Every `HandoffPayload` continues to carry the
constitution-mandated reproducibility surface, sourced now from the
`claude` CLI. Spec: `Story 3`.

**Independent Test**: Inspect a `HandoffPayload` produced by a
successful run; confirm `analysis.prompt_version == "identify-game.v1"`,
`analysis.frame_sha256` length matches `analysis.frame_offsets_pct`,
`analysis.raw_response` is the verbatim JSON the model emitted, and
`analysis.model_id` either equals the value passed to `--model` or is
empty with a corresponding note in `analysis.error` (per FR-007).

### Implementation for User Story 3

- [X] T016 [US3] In `internal/analysis/claude.go`, populate `AnalysisRecord.ModelID` with the explicit precedence per `research.md` R-5: (1) the `modelID` argument the application passed to `--model` if non-empty; (2) otherwise `envelope.Model` from the parsed CLI envelope if non-empty; (3) otherwise empty string. When the empty branch is taken, append `"model id unavailable"` to `AnalysisRecord.Error` (preserving any earlier error string, semicolon-separated).
- [X] T017 [US3] In `internal/analysis/claude.go`, ensure `AnalysisRecord.RawResponse` is set to `envelope.Result` verbatim on every code path (success AND envelope-level error), so a re-run is always replayable from the record alone.
- [X] T018 [US3] In `internal/analysis/claude.go`, delete the legacy `coerceConfidence` helper and the defensive `\`\`\`json` fence trimming from the SDK era — both are dead code now that `--json-schema` enforces the response shape. The decoder MUST be a single `json.Unmarshal` and MUST treat any failure as a Claude failure (not a recoverable coercion).

**Checkpoint**: All three user stories shipped. The `HandoffPayload`
schema stays byte-compatible with feature 001's; reproducibility is
honest (no invented model ids, no silent coercions).

---

## Phase 6: Polish & Cross-Cutting Concerns

- [X] T019 [P] Update `README.md` at the repo root: replace the `ANTHROPIC_API_KEY` prerequisite with "the `claude` CLI on `PATH` and signed in"; link to `specs/002-claude-cli-backend/{spec.md, plan.md, quickstart.md}`.
- [X] T020 [P] Create `scripts/verify-no-sdk.sh` (mode `0755`) that runs `! grep -RnE "anthropic-sdk-go|ANTHROPIC_API_KEY" cmd/ internal/ scripts/` and exits non-zero on any match. The pre-merge gate per SC-007 is "this script returns 0".
- [X] T021 [P] Update `specs/001-video-picker-screen/quickstart.md` with a top-of-file admonition pointing readers at `specs/002-claude-cli-backend/quickstart.md` for the current build/run path; do NOT delete the 001 quickstart (it remains the canonical reference for the picker).
- [X] T022 Walk through `specs/002-claude-cli-backend/quickstart.md` end-to-end on a real machine (real `claude` signed in, real video fixture). Capture a sample `HandoffPayload` to `specs/002-claude-cli-backend/sample-payload.json` and confirm SC-001..SC-005 + SC-007 pass on a representative environment. SC-006 (≤5 s budget) is captured separately in T023.
- [X] T023 Capture before/after timing for the identify call on a reference fixture (the same fixture used for `specs/001-video-picker-screen/sample-payload.json` if available, else a fresh ffmpeg `testsrc` clip): record the median and p95 wall time of `analysis.request_finished_at - analysis.request_started_at` over 5 runs against (a) the SDK code at commit `6527efd` and (b) HEAD of `002-claude-cli-backend`. Append the numbers to `specs/002-claude-cli-backend/quickstart.md` as a "Performance gate evidence" section. The gate is met if the median delta is `< 1 s` (this spec SC-006).
- [X] T024 Confirm the constitution Reproducibility gate: open the captured `sample-payload.json` and verify by inspection that `analysis.prompt_version`, `analysis.model_id` (per FR-007 semantics), `analysis.frame_offsets_pct`, `analysis.frame_sha256`, `analysis.request_started_at`, `analysis.request_finished_at`, `analysis.raw_response` are all populated; record the verification in the same "Performance gate evidence" section as T023.

---

## Dependencies & User-Story Completion Order

```
Phase 1 (Setup: SDK removal)
   ↓
Phase 2 (Foundational: claudecli skeleton + preflight)
   ↓
   ├─→ Phase 3 (US1 — MVP: identify w/o API key)        ← ship this first
   │       ↓
   │       └─→ Phase 4 (US2: error paths)               ← ship next
   │              ↓
   │              └─→ Phase 5 (US3: reproducibility)    ← ship last
   ↓
Phase 6 (Polish + perf gate)
```

- US2 depends on US1 (it polishes the failure surfaces US1 introduced).
- US3 depends on US1 and US2 (it asserts on the AnalysisRecord that US1
  produces and the failure paths US2 wires up). Inside US3, T016-T018
  touch the same file (`internal/analysis/claude.go`) so they MUST run
  sequentially.

## Parallel Execution Examples

Within a phase, tasks marked `[P]` touch different files and can be
implemented in parallel:

- **Phase 2**: T002 (`identify_schema_v1.json`), T003 (`identify_schema_v1.go`), T004 (`tempframes.go`), and T005 (`parse.go`) all live in distinct files and can be developed in parallel.
- **Phase 6**: T019 (`README.md`), T020 (`scripts/verify-no-sdk.sh`), and T021 (`specs/001-.../quickstart.md`) live in distinct files and can be done in parallel.

T009 / T010 / T013 / T015 / T016 / T017 / T018 all touch
`internal/analysis/claude.go` and MUST run sequentially in the order
listed.

## Implementation Strategy

1. Ship **US1** end-to-end first (Phases 1 → 2 → 3). This is the MVP:
   the application runs without the SDK and without `ANTHROPIC_API_KEY`.
2. Layer **US2** on top (Phase 4). The failure modes become diagnosable.
3. Polish with **US3** (Phase 5). The AnalysisRecord becomes faithful
   to the new producer.
4. Apply Phase 6 polish before opening the merge PR. Re-introduce the
   tests deferral remediation (per the inherited Complexity Tracking
   item) at the same time the comparable item from feature 001 is
   addressed; the two deferrals are linked.
