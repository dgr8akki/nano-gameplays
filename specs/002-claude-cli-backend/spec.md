# Feature Specification: Use Local `claude` CLI Instead of Anthropic SDK

**Feature Branch**: `002-claude-cli-backend`
**Created**: 2026-05-02
**Status**: Draft
**Input**: User description: "Instead of anthropic sdk and manual key inputs, I want that my CLI is already having claude code access via 'claude' command access, I want to use that instead"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Identify a game without managing an API key (Priority: P1)

A developer launches `nano-gameplays` on a workstation that already has the
`claude` command-line tool installed and signed in. They pick a video, the
right pane shows "Analyzing…", and identification succeeds — without their
having set `ANTHROPIC_API_KEY` or any other Anthropic credential in the
environment.

**Why this priority**: This is the entire purpose of the change. The reason
the user filed this feature is that wiring an API key separately on top of
their already-authenticated `claude` CLI is friction; eliminating that
friction is the value.

**Independent Test**: From a shell where `ANTHROPIC_API_KEY` is unset and
`claude --version` returns successfully, run `nano-gameplays`, select a
gameplay clip, and confirm the metadata pane fills in within the same time
budget as today (within ~5 s on a reference connection).

**Acceptance Scenarios**:

1. **Given** a workstation where the `claude` CLI is installed and
   authenticated and `ANTHROPIC_API_KEY` is unset, **When** the user
   selects a gameplay video, **Then** the right pane fills in game title,
   scene/level/mode and confidence within the established time budget.
2. **Given** the `identify` non-interactive subcommand is invoked with
   `--video PATH --json` on the same workstation, **When** the call
   completes, **Then** stdout still emits a `HandoffPayload` with the
   same shape (`schema_version: "1"`) the previous SDK-based flow
   produced.

---

### User Story 2 - Get a clear, actionable error when the `claude` CLI is missing or broken (Priority: P2)

A developer who has not installed (or has not signed into) the `claude`
CLI runs `nano-gameplays`. They are told plainly that the `claude` CLI is
required and how to recover, instead of seeing a confusing API-side error
message that points at credentials they no longer manage directly.

**Why this priority**: Without a clear error path, removing the SDK leaves
new failure modes that are harder to diagnose than the old "set
`ANTHROPIC_API_KEY`" message. The recovery instructions matter.

**Independent Test**: From a shell where `claude` is not on `PATH`, launch
`nano-gameplays` and select a video; observe a clear inline failure that
names the missing tool. Repeat with `claude` present but unauthenticated;
observe the editable form opens with empty fields and an explanatory
banner (FR-018-style failure path), so the user can still type metadata
and confirm.

**Acceptance Scenarios**:

1. **Given** the `claude` CLI is not present on the system `PATH`,
   **When** the user launches the application, **Then** a clear preflight
   message tells them the `claude` CLI is required and the program exits
   without showing a half-functional analyzing state.
2. **Given** the `claude` CLI is present but its authentication is
   missing or expired, **When** the user selects a video, **Then** the
   metadata pane opens with empty editable fields and an explanatory
   banner; the user can still type values and emit a handoff payload
   whose `analysis.error` describes the auth failure.

---

### User Story 3 - Reproducibility record stays intact (Priority: P3)

A developer reviewing yesterday's handoff payload still sees which model
the call ran against, which version of the prompt template was used, and
the SHA-256 of every frame that was sent — just as before — even though
the call was routed through the `claude` CLI rather than the SDK.

**Why this priority**: The constitution's reproducibility requirement
(Principle III) is non-negotiable, but it is a quality bar not a primary
user goal, hence P3. Failing this story would block merge but not block a
demo of stories 1 and 2.

**Independent Test**: Inspect a `HandoffPayload` produced by the new flow
and confirm `analysis.prompt_version`, `analysis.model_id`,
`analysis.frame_offsets_pct`, and `analysis.frame_sha256` are all
populated and correct. Compare against an older SDK-produced payload to
confirm the schema is byte-compatible.

**Acceptance Scenarios**:

1. **Given** a successful identification through the `claude` CLI,
   **When** the resulting `HandoffPayload` JSON is inspected, **Then**
   `analysis.model_id` reflects the model the `claude` CLI actually used
   for the call (not invented), `analysis.prompt_version` is
   `identify-game.v1`, and `analysis.frame_sha256` lists one hash per
   frame sent.
2. **Given** the `claude` CLI returns output that is not valid JSON for
   the identify schema, **When** the failure is recorded, **Then**
   `analysis.raw_response` contains the verbatim CLI output and
   `analysis.error` describes the parse failure, mirroring today's SDK
   failure path.

---

### Edge Cases

- The `claude` CLI is present but a different incompatible major version
  is installed → preflight should still pass binary detection but
  identification may fail; failure surfaces on the FR-018-style path with
  a descriptive error.
- The `claude` CLI takes substantially longer than the previous SDK call
  (e.g., due to interactive auth refresh on first use) → cancellation on
  new selection MUST still abandon the in-flight call cleanly.
- The user supplies `--model X` but the `claude` CLI cannot honour that
  model (unknown, gated, or unsupported by the local install) → failure
  surfaces with a clear message naming the model.
- The user has both the SDK environment variables set *and* the `claude`
  CLI installed → behaviour MUST be predictable and documented; the
  environment variables MUST NOT silently re-enable the old SDK path.
- The `claude` CLI emits warnings/diagnostics on stderr while the
  identification JSON is on stdout → only stdout is treated as the
  response; stderr is preserved into the analysis record only when the
  call fails.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST perform every game-identification call through
  the local `claude` command-line tool. The Anthropic SDK and any direct
  Anthropic HTTP client MUST be removed from the runtime path of the
  identify flow.
- **FR-002**: System MUST NOT require `ANTHROPIC_API_KEY` (or any other
  Anthropic-credential environment variable) to be set in order for
  identification to succeed.
- **FR-003**: System MUST detect at preflight time whether the `claude`
  CLI is available on `PATH`; if not, the program MUST exit with a
  preflight failure (matching the existing exit-code-2 contract for
  precondition failures) and a single-line message naming the missing
  tool and pointing the user at the recovery action (install / sign in).
- **FR-004**: System MUST cancel an in-flight `claude` invocation when
  the user picks a different video before the previous identification
  completes (preserves FR-020 from feature 001-video-picker-screen).
- **FR-005**: System MUST surface failures from the `claude` CLI
  (non-zero exit, malformed output, output that does not match the
  identify response schema) on the same editable-form-with-error path
  defined by feature 001-video-picker-screen (FR-018 / FR-019). The user
  MUST be able to manually enter metadata and still emit a handoff
  payload whose `analysis.error` is non-empty.
- **FR-006**: System MUST continue to populate the existing reproducibility
  fields on every emitted `HandoffPayload`: `analysis.prompt_version`,
  `analysis.model_id`, `analysis.frame_offsets_pct`,
  `analysis.frame_sha256`, `analysis.request_started_at`,
  `analysis.request_finished_at`, `analysis.raw_response`,
  `analysis.error`. The schema (`schema_version: "1"`) MUST not change.
- **FR-007**: `analysis.model_id` MUST reflect the model actually used by
  the `claude` CLI for that call. If the CLI does not surface that
  information, `analysis.model_id` MUST be empty rather than guessed; in
  that case `analysis.error` MUST note that the model identifier was
  unavailable.
- **FR-008**: The user-facing `--model` flag and the
  `NANO_GAMEPLAYS_MODEL` environment variable MUST continue to be
  accepted. When supplied, the value MUST be passed through to the
  `claude` CLI in whichever form that CLI accepts for selecting a model;
  when not supplied, the application MUST defer to the `claude` CLI's
  own default and MUST NOT inject its own model fallback string.
- **FR-009**: The captured frames sent to the `claude` CLI MUST not
  persist on disk after the call returns. If the chosen transport
  mechanism requires touching the filesystem (for example, writing PNGs
  to a temporary directory because the `claude` CLI needs file paths),
  the temporary files MUST be created with user-only permissions, MUST
  live under the operating system's per-user temporary directory, and
  MUST be deleted before the analysis goroutine exits — including on
  cancellation, error, and panic. The intent of feature 001's FR-023
  (frame privacy) MUST be preserved.
- **FR-010**: The `identify` non-interactive subcommand MUST continue to
  exist with the same flag surface (`--video`, `--model`, `--json`),
  same exit-code semantics (`0` payload emitted, `2` precondition
  failure), and same payload shape. No CLI-surface flag may be removed
  by this feature.
- **FR-011**: The presence of `ANTHROPIC_API_KEY` in the user's
  environment MUST NOT silently re-enable a direct-SDK code path. Any
  remaining behavioural difference from setting that variable MUST be
  documented; preferred behaviour is "ignored".
- **FR-012**: All quickstart and README references to obtaining /
  exporting `ANTHROPIC_API_KEY` MUST be replaced with a reference to
  installing and signing into the `claude` CLI.

### Key Entities *(include if feature involves data)*

- **Identify invocation**: a single call from the application to the
  `claude` CLI carrying the prompt template and the captured frames, and
  receiving a JSON response describing the identified game. Identity is
  the (video path, frame hashes) tuple; lifetime is one call.
- **AnalysisRecord** *(unchanged from feature 001)*: the per-call
  reproducibility record embedded into every emitted `HandoffPayload`.
  This feature changes the *producer* of the record but not its shape.
- **`claude` CLI dependency**: a system-level prerequisite (the `claude`
  binary on `PATH`, authenticated for the user). Detected at preflight
  time; failures are user-actionable.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A new developer can run `nano-gameplays`, select a video,
  and see populated metadata fields without ever setting an Anthropic
  credential environment variable, on a workstation where the `claude`
  CLI is signed in.
- **SC-002**: 100% of `HandoffPayload` JSON objects produced by the new
  flow validate against the existing v1 schema (same field names, same
  types, same `schema_version` value) — verified by a payload
  round-trip check on a reference fixture.
- **SC-003**: When the `claude` CLI is missing from `PATH`, the
  application exits within 1 second of launch with a single, clear
  message that names the missing tool and a one-line recovery hint.
- **SC-004**: When the `claude` CLI is present but the call fails (for
  any reason — auth, timeout, malformed JSON, model unavailable), the
  user can still complete the screen by typing metadata and confirming,
  and the emitted payload's `analysis.error` field is non-empty.
- **SC-005**: 0 captured-frame PNG files remain on disk in the per-user
  temporary directory after `nano-gameplays` exits, on every supported
  exit path (clean confirm, user quit, ffmpeg failure, `claude` CLI
  failure, ctrl+c during analysis). Verified by snapshotting the temp
  directory before and after a representative session.
- **SC-006**: The end-to-end identify time (selection → metadata
  visible) on a reference connection stays within the same ≤ 5 s budget
  established by feature 001 (SC-006). The change of backend MUST NOT
  introduce a ≥ 1 s regression in the median case.
- **SC-007**: The repository contains zero remaining references to
  `ANTHROPIC_API_KEY`, the `anthropic-sdk-go` import path, or
  SDK-specific types in the runtime source tree (excluding migration
  notes / changelog entries).

## Assumptions

- The `claude` CLI is the Claude Code command-line tool (a separate
  binary on the user's `PATH`) and the user has already authenticated
  it (e.g. via the tool's own login flow). This feature does not
  introduce its own login UI.
- The `claude` CLI exposes some way to (a) supply a system /
  instruction prompt verbatim, (b) supply one or more PNG images as
  inputs, and (c) request a non-streaming text response. Concrete flag
  names will be confirmed during planning; this spec asserts only that
  *some* such surface exists.
- The reasonable transport for PNG frames is via short-lived files in
  the operating system's per-user temporary directory, deleted by the
  application before the analysis goroutine exits. This preserves the
  spirit of FR-023 (no long-lived on-disk frames) without depending on
  the `claude` CLI accepting binary stdin for image input.
- The `claude` CLI's exit code is meaningful: a non-zero exit indicates
  the call failed, regardless of stderr content. Zero exit with empty
  or non-JSON stdout is treated as a malformed-response failure on the
  same path as a non-zero exit.
- Existing constitution Principle III (reproducibility) is preserved:
  whichever model the `claude` CLI used must be discoverable from its
  output or its flags; if not, the field is left empty per FR-007.
- This feature does not change the rest of the screen — left-pane
  picker, right-pane idle animation, footer help, terminal-size
  preflight, payload shape — beyond removing SDK references.
