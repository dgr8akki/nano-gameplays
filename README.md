# nano-gameplays

A keyboard-driven TUI for picking a PlayStation gameplay video and identifying
the game / scene with Claude vision so the next stage of the pipeline can
operate on a clean `(video, metadata)` payload.

## Status

Active feature: [`specs/002-claude-cli-backend/`](specs/002-claude-cli-backend/).

- [Specification](specs/002-claude-cli-backend/spec.md)
- [Implementation plan](specs/002-claude-cli-backend/plan.md)
- [Quickstart](specs/002-claude-cli-backend/quickstart.md)

Predecessor feature: [`specs/001-video-picker-screen/`](specs/001-video-picker-screen/)
(picker UI, frame extraction, payload schema). Still load-bearing for everything
except how the Claude call is made.

- [Project constitution](.specify/memory/constitution.md)

## Quick build

```sh
go build -o bin/nano-gameplays ./cmd/nano-gameplays
./bin/nano-gameplays
```

## Prerequisites

- Go 1.22+
- `ffmpeg` on `$PATH`
- The **`claude` CLI** (Claude Code) on `$PATH` and signed in. Verify with
  `claude --version` and `claude auth`.
- A terminal at least 100 cols × 30 rows.

`ANTHROPIC_API_KEY` is **not** required by `nano-gameplays`. The identify call
is routed through your `claude` CLI's own credential. If you have the env var
set, `nano-gameplays` itself ignores it; whether your `claude` install
consults it as an auth fallback is the CLI's own decision.
