# Contract — Handoff payload (unchanged)

The handoff payload contract is **unchanged** from
[`specs/001-video-picker-screen/contracts/handoff.md`](../../001-video-picker-screen/contracts/handoff.md).

`schema_version` stays pinned to `"1"`. Every field name, type, and
optionality rule is preserved. The only thing this feature changes is
the *producer* of the `analysis` block: the values are now sourced
from the `claude` CLI subprocess instead of the in-process SDK call,
but the field shapes do not change.

## Producer notes specific to this feature

- `analysis.model_id`: per FR-007 / R-5, this is set to the value
  passed to `claude --model` when the user (or `NANO_GAMEPLAYS_MODEL`)
  supplied one; otherwise the application reads it from the
  `--output-format json` envelope's model field if present, and
  otherwise leaves it empty and notes the omission in
  `analysis.error`.
- `analysis.raw_response`: the schema-validated JSON string the model
  emitted (extracted from the CLI envelope's `result` field). This is
  the same shape as the v1 raw response.
- `analysis.error`: any envelope-level error from the CLI (non-zero
  exit, schema-validation failure surfaced by `--json-schema`,
  malformed envelope, missing model id per FR-007).
- `analysis.frame_offsets_pct` and `analysis.frame_sha256`: produced
  by the existing extractor, untouched by this feature.

## Forward compatibility

- This feature does not introduce new top-level keys. Consumers that
  were already forward-compatible per the 001 contract continue to
  work without modification.
