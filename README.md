# nano-gameplays

A keyboard-driven TUI for picking a PlayStation gameplay video and identifying
the game / scene with Claude vision so the next stage of the pipeline can
operate on a clean `(video, metadata)` payload.

## Status

First feature in active development:
[`specs/001-video-picker-screen/`](specs/001-video-picker-screen/).

- [Specification](specs/001-video-picker-screen/spec.md)
- [Implementation plan](specs/001-video-picker-screen/plan.md)
- [Quickstart](specs/001-video-picker-screen/quickstart.md)
- [Project constitution](.specify/memory/constitution.md)

## Quick build

```sh
go build -o bin/nano-gameplays ./cmd/nano-gameplays
./bin/nano-gameplays
```

See the quickstart linked above for prerequisites (Go 1.22+, ffmpeg on PATH,
`ANTHROPIC_API_KEY` exported, ≥ 100×30 terminal).
