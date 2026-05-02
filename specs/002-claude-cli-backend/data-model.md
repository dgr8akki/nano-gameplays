# Phase 1 — Data Model

This feature does **not** change the cross-screen data model defined by
`specs/001-video-picker-screen/data-model.md`. The `HandoffPayload`,
`GameplayVideo`, `GameMetadata`, and `AnalysisRecord` shapes remain
identical, and `schema_version` stays pinned to `"1"`. This file
documents only the additions.

## Unchanged entities

- **`GameplayVideo`** — unchanged. See 001 §`GameplayVideo`.
- **`CapturedFrame`** *(transient, in-memory only)* — unchanged surface,
  but the lifecycle adds one step (see additions below).
- **`GameMetadata`** — unchanged. See 001 §`GameMetadata`.
- **`AnalysisRecord`** — unchanged. The producer changes; the shape
  does not. See 001 §`AnalysisRecord`.
- **`HandoffPayload`** — unchanged. See 001 §`HandoffPayload`.

## Additions

### `claudecli.Invocation` *(transient, internal)*

Represents the constructed argv + environment + working tempdir for one
`claude -p` subprocess. Lifetime spans exactly one identify call.

| Field | Type | Notes |
|---|---|---|
| `Argv` | `[]string` | The full `claude` argv (binary path + flags + final user message). |
| `TempDir` | `string` (absolute path) | Per-call temp dir under `os.TempDir()` (mode `0700`). Owns the PNG frames for the duration of the call. |
| `FramePaths` | `[]string` | Absolute paths to `frame_<pct>.png` files inside `TempDir`. Same length as `[]CapturedFrame`. |
| `ModelID` | `string` (optional) | Empty when the user supplied no `--model` and no `NANO_GAMEPLAYS_MODEL`. |

**Validation**:
- `TempDir` MUST be a child of `os.TempDir()` and MUST be created with
  mode `0700`.
- `FramePaths[i]` MUST be inside `TempDir` and MUST have mode `0600`.
- `Argv[0]` is always the absolute path returned by `exec.LookPath("claude")`.

**Lifecycle**: Constructed in `internal/analysis/claude.go` immediately
before the subprocess starts. The owning function calls
`defer os.RemoveAll(invocation.TempDir)`; cleanup runs on every exit
path (success, error, panic, parent-context cancellation).

---

### `claudecli.ResultEnvelope` *(transient, internal)*

The decoded shape of `claude -p --output-format json` stdout. Used only
to extract the model's identify-response payload and any envelope-level
error indication; not persisted.

| Field | Type | Notes |
|---|---|---|
| `Type` | `string` | Expected `"result"` for the single-result envelope. |
| `Result` | `string` | The schema-conforming JSON string the model emitted (subject of `--json-schema`). |
| `IsError` | `bool` | True when the call failed at envelope level (auth, rate limit, schema validation, etc.). |
| `Model` | `string` (optional) | Model identifier the CLI reports; copied verbatim into `AnalysisRecord.ModelID` when our own override was empty. May be missing on older CLI versions. |
| `Error` | `string` (optional) | Envelope-level error string. |

**Validation**:
- On `IsError == true`, the contents of `Result` MUST be preserved into
  `AnalysisRecord.RawResponse` and `Error` MUST be appended to
  `AnalysisRecord.Error`.
- On `IsError == false`, `Result` MUST be JSON-unmarshallable into the
  identify-response shape from `001`'s `data-model.md` §`GameMetadata`
  (game_title / scene_or_level_or_mode / confidence). Any unmarshal
  failure is itself a Claude-failure path (FR-005).

**Lifecycle**: Decoded once per call from the subprocess's captured
stdout, immediately consumed by the orchestrator, then dropped.

---

## State machine (right-pane) — unchanged

The state machine in `001`'s data-model is preserved verbatim:

```
[idle] --selectVideo--> [analyzing] --analysisOk--> [editing]
                              \--analysisErr-->     [editing(empty,err)]
[editing] --confirm--> [done]   (handoff emitted)
[*]       --selectVideo(other)--> [analyzing]   (cancels in-flight)
```

The only difference is what produces the `analysisOk` / `analysisErr`
message: the `internal/analysis/claudecli/` subpackage instead of the
embedded SDK client.
