# Phase 1 — Data Model

This screen is in-memory only. No persistence. The model below is the shape
held by the Bubble Tea program and the shape emitted at handoff.

## Entities

### `GameplayVideo`

The user's selected source file.

| Field | Type | Notes |
|---|---|---|
| `path` | `string` (absolute filesystem path) | Identity. Set when user confirms a selection in the picker. |
| `size_bytes` | `int64` | Captured at selection time; surfaced in failure messages. |
| `extension` | `string` | Lowercased; one of the supported set (see Validation). |

**Validation**:
- `extension` MUST be in the supported set: `mp4`, `mov`, `mkv`, `webm`.
  Anything else is non-selectable in the picker (FR-010, FR-011).
- `path` MUST exist and be readable at confirmation time; otherwise the
  picker rejects the selection with a non-blocking message.

**Lifecycle**: created on confirm, replaced on a new confirm, never
mutated.

---

### `CapturedFrame` *(transient, in-memory only)*

| Field | Type | Notes |
|---|---|---|
| `bytes` | `[]byte` | PNG payload returned by ffmpeg over stdout. |
| `offset_pct` | `int` | 10 or 60 (R-2). |
| `sha256` | `string` (hex) | Computed once; carried into the AnalysisRecord. |

**Validation**:
- `len(bytes) > 0`. Zero-byte payload → ExtractError, FR-019 path.

**Lifecycle**: created during analysis kickoff, dropped immediately after
the Claude call finishes (FR-023). The `sha256` is retained on the
`AnalysisRecord`; the bytes are not.

---

### `GameMetadata`

The fixed three-field result described in spec FR-014.

| Field | Type | Notes |
|---|---|---|
| `game_title` | `string` | Editable in the TUI; freeform 1–100 chars. |
| `scene_or_level_or_mode` | `string` | Editable; freeform 0–200 chars. |
| `confidence` | `enum{"low","medium","high"}` | Not user-editable (decision in plan); set from Claude response. |

**Validation**:
- `game_title` MUST be non-empty at confirm time. If Claude returned
  nothing and the user did not enter anything, the confirm key is
  inactive and the footer shows a hint.
- `scene_or_level_or_mode` MAY be empty at confirm time.
- `confidence` MUST be one of the three enum values; on FR-018 path
  (Claude failure) it defaults to `low` and the failure is surfaced in
  the right pane.

**State transitions** (right-pane state machine):

```
[idle] --selectVideo--> [analyzing] --analysisOk--> [editing]
                              \--analysisErr-->     [editing(empty,err)]
[editing] --confirm--> [done]   (handoff emitted)
[*]       --selectVideo(other)--> [analyzing]   (cancels in-flight)
```

---

### `AnalysisRecord`

The reproducibility record (Constitution Principle III, R-3). Carried
inside the handoff payload.

| Field | Type | Notes |
|---|---|---|
| `prompt_version` | `string` | e.g. `identify-game.v1`. |
| `model_id` | `string` | e.g. `claude-sonnet-4-6`. |
| `frame_offsets_pct` | `[]int` | The offsets used by the extractor. |
| `frame_sha256` | `[]string` | Per-frame hash of the bytes sent to Claude. |
| `request_started_at` | `string` (RFC 3339) | |
| `request_finished_at` | `string` (RFC 3339) | |
| `raw_response` | `string` | Verbatim model output (the JSON it returned, or empty on error). |
| `error` | `string` (optional) | Empty on success. |

**Validation**: required on every emitted handoff payload, even on the
FR-018 failure path (where `error` is non-empty and `game_title` /
`scene_or_level_or_mode` may be empty). The presence of an
`AnalysisRecord` proves the call was attempted and is replayable.

---

### `HandoffPayload`

The JSON object emitted to the next screen (TUI path) and to stdout
(`identify --json` path). Defined in
[contracts/handoff.md](./contracts/handoff.md). Composed of:

- `video`: `GameplayVideo`
- `metadata`: `GameMetadata`
- `analysis`: `AnalysisRecord`
- `schema_version`: `string` — pinned to `1`.
