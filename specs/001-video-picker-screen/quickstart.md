# Quickstart

A working setup for someone who just cloned this repo and wants to run
the video picker screen end-to-end.

## Prerequisites

- **Go**: 1.22 or newer (`go version`).
- **ffmpeg**: on `$PATH` (`ffmpeg -version`). macOS: `brew install
  ffmpeg`. Linux: distro package manager.
- **Anthropic API key**: export `ANTHROPIC_API_KEY` in your shell.
- **Terminal**: ≥ 100 columns × 30 rows, truecolor or 256-color.

## Build & run

```sh
go build -o bin/nano-gameplays ./cmd/nano-gameplays

# Interactive: launch the picker in the current directory.
./bin/nano-gameplays

# Interactive: launch the picker rooted somewhere else.
./bin/nano-gameplays --start-dir ~/Movies/PS5

# Non-interactive: identify a single video, emit handoff JSON.
./bin/nano-gameplays identify --video ~/Movies/PS5/clip.mp4 --json
```

## What you should see (interactive)

1. Header bar shows `nano-gameplays`.
2. Left pane: file picker rooted at the start directory; videos shown
   normally, other files dimmed.
3. Right pane: PS5-themed ASCII animation looping.
4. Footer: a single line of key hints (move / open / select / quit).
5. Pick a video file and confirm with `enter`.
6. Right pane swaps to a spinner with `Analyzing…`.
7. Within ~5 s, three editable fields appear: game title, scene /
   level / mode, and a confidence badge (low / medium / high).
8. Edit anything that's wrong; press `ctrl+s` to confirm.
9. Process exits and prints the handoff JSON to stdout.

## What you should see (non-interactive)

```sh
$ ./bin/nano-gameplays identify --video ~/Movies/PS5/clip.mp4 --json
{"schema_version":"1","video":{"path":"/Users/.../clip.mp4", ...}, ...}
```

## Common failure cases (and what they look like)

- **Terminal too small** → exit 2, `terminal must be at least 100x30`
  on stderr.
- **ffmpeg missing** → exit 2, `ffmpeg not found in PATH` on stderr.
- **No `ANTHROPIC_API_KEY`** → TUI starts; on selection the right pane
  shows the failure message (FR-018) with empty editable fields. The
  user can still type metadata and confirm.
- **Network failure during analysis** → same FR-018 path. The handoff
  payload's `analysis.error` is non-empty.

## Where to look in the code

- `cmd/nano-gameplays/main.go` — flag parsing, picks TUI vs. identify
  path.
- `internal/app/` — Bubble Tea model, update, view.
- `internal/ui/leftpane/` — file picker (wraps `bubbles/filepicker`).
- `internal/ui/rightpane/` — three sub-views: idle, analyzing,
  metadata-edit.
- `internal/analysis/` — ffmpeg wrapper, Claude call, AnalysisRecord.
- `internal/handoff/` — handoff payload schema (Go struct mirroring
  `contracts/handoff.md`).
