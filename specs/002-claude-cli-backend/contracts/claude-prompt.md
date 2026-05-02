# Contract — Claude prompt template (delta from 001)

The prompt template itself is **unchanged**. The transport changes.
This file documents the delta from
[`specs/001-video-picker-screen/contracts/claude-prompt.md`](../../001-video-picker-screen/contracts/claude-prompt.md).

## Identifier (unchanged)

`prompt_version: identify-game.v1`

The literal text in
`internal/analysis/prompt/identify_game_v1.txt` is unchanged. The
versioning rules (MAJOR/MINOR/PATCH) from the 001 contract apply
unchanged.

## Transport (changed)

- **Was**: prompt body sent as the `system` block in the Anthropic
  Messages API call from `anthropic-sdk-go`. Frames sent as
  base64-encoded `image` content blocks in a single user message.
- **Is**: prompt body passed via `claude --system-prompt <body>` to a
  subprocess of the local `claude` CLI. Frames written to a per-call
  tempdir, made readable via `claude --add-dir <tempdir>`, and
  referenced by absolute path in the user message. The `Read` tool
  is the only allowed tool for the call.

## Response shape (changed mechanism, unchanged shape)

- **Was**: the model returned a JSON object that the application
  decoded with `encoding/json`, with defensive trimming for stray
  ```json fences and runtime coercion of unknown `confidence` values
  to `low`.
- **Is**: the same JSON object shape, but enforced by the CLI via
  `claude --json-schema <schema>`. The schema is the canonical source
  of truth and is embedded in the binary alongside the prompt:

```json
{
  "type": "object",
  "additionalProperties": false,
  "properties": {
    "game_title":             { "type": "string" },
    "scene_or_level_or_mode": { "type": "string" },
    "confidence":             { "type": "string", "enum": ["low","medium","high"] }
  },
  "required": ["game_title", "scene_or_level_or_mode", "confidence"]
}
```

The schema is bumped together with the prompt (`identify-game.v2`
bumps both the `.txt` and the `.json` artifacts). PATCH-level wording
edits to the prompt body MAY leave the schema untouched.

## Client behaviour (changed)

- **Decoder**: the application no longer trims `\`\`\`json` fences
  itself; on parse failure, treat as Claude failure (FR-005),
  preserve the raw envelope contents in `AnalysisRecord.raw_response`,
  set `error` to a short human-readable reason.
- **Schema drift**: the CLI rejects extra keys at the schema layer
  (`additionalProperties: false`); the application does not need to
  ignore unknowns in the response.
- **Confidence enum drift**: impossible at runtime when the schema
  enforces the enum; the historical `coerceConfidence` helper from
  001 is removed.

## Versioning (unchanged)

Bumping rules mirror the 001 contract:
- **MAJOR** — remove a required output field, change an enum value's
  meaning.
- **MINOR** — add a new output field. Note: if the schema gains a
  field, the schema's `required` and `additionalProperties` MUST be
  updated together.
- **PATCH** — wording-only edits to the prompt body that do not
  change the response schema.
