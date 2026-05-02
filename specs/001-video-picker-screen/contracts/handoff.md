# Contract — Handoff payload

Single JSON object. Produced both when the user confirms in the TUI and
when `nano-gameplays identify --json` is invoked. Consumed by the next
screen and by any caller of `identify --json`.

## Schema

```json
{
  "schema_version": "1",
  "video": {
    "path": "/abs/path/to/clip.mp4",
    "size_bytes": 1234567890,
    "extension": "mp4"
  },
  "metadata": {
    "game_title": "Spider-Man 2",
    "scene_or_level_or_mode": "Symbiote suit chase, Times Square",
    "confidence": "high"
  },
  "analysis": {
    "prompt_version": "identify-game.v1",
    "model_id": "claude-sonnet-4-6",
    "frame_offsets_pct": [10, 60],
    "frame_sha256": [
      "9b74c9897bac770ffc029102a200c5de",
      "e3b0c44298fc1c149afbf4c8996fb924"
    ],
    "request_started_at": "2026-05-02T18:31:04Z",
    "request_finished_at": "2026-05-02T18:31:07Z",
    "raw_response": "{\"game_title\":\"Spider-Man 2\",\"scene_or_level_or_mode\":\"Symbiote suit chase, Times Square\",\"confidence\":\"high\"}",
    "error": ""
  }
}
```

## Field rules

- `schema_version`: pinned to `"1"` for this feature; bumped if any
  required field is removed or renamed.
- `video.path`: absolute path; the only field uniquely identifying the
  source.
- `video.extension`: lowercased, no leading dot.
- `metadata.game_title`: non-empty. (On FR-018 failure path, the user
  must type a value before confirm becomes active.)
- `metadata.scene_or_level_or_mode`: may be empty.
- `metadata.confidence`: enum `low|medium|high`. On FR-018 failure path,
  defaults to `low`.
- `analysis.error`: empty on success path; non-empty on FR-018 /
  FR-019. The payload is still emitted in either case.
- `analysis.frame_sha256`: length matches `frame_offsets_pct`; usually
  2 entries, but the contract allows 1 if the second extraction failed
  while the first succeeded.

## Producer/consumer expectations

- **Producers** MUST emit valid UTF-8 JSON on a single output, with one
  trailing newline.
- **Consumers** MUST treat unknown fields as forward-compatible (ignore,
  do not error) — to allow MINOR additions in future versions.
