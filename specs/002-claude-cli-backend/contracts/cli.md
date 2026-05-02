# Contract — CLI surface (delta from 001)

The user-facing CLI surface is **unchanged** from
[`specs/001-video-picker-screen/contracts/cli.md`](../../001-video-picker-screen/contracts/cli.md).
This file records only the deltas this feature introduces.

## What does NOT change

- The two entry forms — `nano-gameplays` (TUI) and `nano-gameplays
  identify --video PATH [--model MODEL] [--json]` — and every flag
  documented for them remain in place with identical names and
  semantics from a user's perspective.
- Exit codes are unchanged (`0`/`1`/`2`).
- `--json` output is governed by
  [`specs/001-video-picker-screen/contracts/handoff.md`](../../001-video-picker-screen/contracts/handoff.md)
  and `schema_version` stays `"1"`.

## What changes (user-visible)

1. **Required system dependencies**:
   - **Was**: `ffmpeg` on PATH, `ANTHROPIC_API_KEY` set in env.
   - **Is**: `ffmpeg` on PATH, **`claude` on PATH and signed in**.
   - `ANTHROPIC_API_KEY` is no longer consulted by `nano-gameplays`
     itself. Whether the user's `claude` install consults it as a
     fallback for its own auth is `claude`'s decision and outside the
     scope of `nano-gameplays`.

2. **Preflight failure messages**:
   - The existing `ffmpeg not found in PATH` preflight message is
     joined by `claude CLI not found in PATH; install Claude Code
     from https://claude.com/claude-code`. Both produce exit code
     `2`.

3. **`--model` flag semantics**:
   - **Was**: passed to the Anthropic SDK as the model id.
   - **Is**: passed verbatim to `claude --model <value>`. The accepted
     value set is whatever the user's `claude` install accepts (model
     aliases like `sonnet`/`opus` or full ids like `claude-sonnet-4-7`).
   - **Behavioural change**: when `--model` and `NANO_GAMEPLAYS_MODEL`
     are both unset, `nano-gameplays` no longer injects a default
     model id; it omits `--model` from the `claude` argv and lets the
     `claude` CLI use its own default. `analysis.model_id` is left
     empty in that case (per FR-007).

## Stability

- This is a v1.x change to the CLI: a flag was not added or removed
  but the *meaning* of `--model`'s default was relaxed (no
  application-side default any more). Treat as a MINOR change.
- The `--json` payload schema is byte-compatible with v1; no consumer
  needs to change.
