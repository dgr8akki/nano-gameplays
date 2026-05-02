# Contract — Claude prompt template

Versioned prompt used to identify a PlayStation gameplay clip from 1–2
extracted frames. Reproducibility surface required by Constitution
Principle III.

## Identifier

`prompt_version: identify-game.v1`

This identifier MUST be embedded in the call's `AnalysisRecord.prompt_version`
field. Any edit to the prompt body, schema, or input shape requires a new
version (`identify-game.v2`, …) — MINOR change for this feature, never
in-place.

## Inputs

- 1 or 2 PNG images (the captured frames), as image content blocks.
- No other context messages on this call.

## System / instruction (verbatim, v1)

> You identify PlayStation gameplay from one or two still frames. Output
> **only** a single JSON object that conforms exactly to the schema below.
> No prose, no code fence, no commentary. If you cannot identify the
> game with reasonable evidence from the frames, return `confidence:
> "low"` and your best guess for `game_title` (or an empty string).

## Required response schema

```json
{
  "game_title": "<string, may be empty>",
  "scene_or_level_or_mode": "<string, may be empty>",
  "confidence": "low" | "medium" | "high"
}
```

## Confidence semantics

- `high` — Distinctive UI/HUD/character is visible and the model is
  certain.
- `medium` — Likely identification from environment / character cues,
  but not from explicit UI.
- `low` — Generic content (loading screen, menu, common environment) or
  the model is unsure.

## Client behaviour

- **Decoder**: strict JSON unmarshal; on parse failure, treat as Claude
  failure (FR-018 path), preserve the raw bytes in
  `AnalysisRecord.raw_response`, set `error` to a short human-readable
  reason.
- **Schema drift**: extra top-level keys MUST be ignored (forward
  compat).
- **Confidence enum drift**: any value other than `low|medium|high` MUST
  be coerced to `low` and logged in `error`.

## Versioning

Bumping rules (mirror constitution semver):
- **MAJOR** for the prompt — remove a required output field, change an
  enum value's meaning.
- **MINOR** — add a new output field (consumers must ignore unknowns,
  per handoff.md).
- **PATCH** — wording-only edits inside the system instruction that do
  not change the response schema.
