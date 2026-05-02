# Quickstart

A working setup for someone who just pulled the `002-claude-cli-backend`
branch and wants to run the screen end-to-end.

## Prerequisites

- **Go**: 1.22 or newer (`go version`).
- **ffmpeg**: on `$PATH` (`ffmpeg -version`). macOS:
  `brew install ffmpeg`. Linux: distro package manager.
- **`claude` (Claude Code CLI)**: on `$PATH` and signed in. Verify:
  - `claude --version` prints a version (≥ 2.1.0).
  - `claude auth` (interactive) shows a signed-in account.
- **Terminal**: ≥ 100 columns × 30 rows, truecolor or 256-color.

## What's *not* required any more

- `ANTHROPIC_API_KEY` does not need to be set in your shell. If it is
  set, `nano-gameplays` will not consult it. Whether your `claude`
  install consults it as an auth fallback is its own decision; remove
  it if you want to be sure the call is going through your `claude`
  signed-in credential.

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
2. Left pane: file picker; videos shown normally, other files dimmed.
3. Right pane: PS5-themed ASCII animation looping.
4. Footer: a single line of key hints (move / open / select / quit).
5. Pick a video and confirm with `enter`.
6. Right pane swaps to a spinner with `Analyzing…`.
7. Within ~5 s, three editable fields appear: game title, scene /
   level / mode, and a confidence badge.
8. Press `ctrl+s` to confirm.
9. Process exits and prints the handoff JSON to stdout.

## What you should see (non-interactive)

```sh
$ ./bin/nano-gameplays identify --video ~/Movies/PS5/clip.mp4 --json
{"schema_version":"1","video":{"path":"/Users/.../clip.mp4", ...}, ...}
```

The payload shape is byte-compatible with feature 001's; only the
`analysis.model_id` field's *defaulting* changed (now empty when no
override was supplied — see [`contracts/cli.md`](./contracts/cli.md)).

## Common failure cases (and what they look like)

- **`claude` missing** → exit `2`, single-line stderr message naming
  the missing binary and pointing at the install URL.
- **`claude` present but not signed in** → TUI starts; on selection
  the right pane opens the editable form with empty fields and a
  banner describing the auth failure (FR-018-style path inherited
  from 001). The user can still type metadata and confirm; the
  emitted payload's `analysis.error` is non-empty.
- **`claude` returns a malformed / non-schema-valid response** →
  same FR-018-style path; `analysis.error` describes the schema
  validation failure.
- **`ffmpeg` missing** → exit `2`, `ffmpeg not found in PATH` on
  stderr (unchanged from 001).
- **Terminal too small** → exit `2`, `terminal must be at least
  100x30` on stderr (unchanged from 001).

## Where to look in the code

- `cmd/nano-gameplays/main.go` — flag parsing, picks TUI vs.
  identify path *(unchanged)*.
- `internal/app/` — Bubble Tea model, update, view *(unchanged)*.
- `internal/ui/{header,footer,leftpane,rightpane}/` *(unchanged)*.
- `internal/analysis/extractor.go` — ffmpeg wrapper *(unchanged)*.
- `internal/analysis/claude.go` — orchestrator: extract frames,
  hand off to `claudecli`, build the AnalysisRecord *(REWRITTEN)*.
- `internal/analysis/claudecli/invoke.go` — argv assembly + subprocess
  exec *(NEW)*.
- `internal/analysis/claudecli/tempframes.go` — per-call tempdir +
  PNG write + deferred cleanup *(NEW)*.
- `internal/analysis/claudecli/parse.go` — `--output-format json`
  envelope decoder *(NEW)*.
- `internal/analysis/prompt/identify_game_v1.{go,txt}` — prompt body
  *(unchanged)*.
- `internal/analysis/prompt/identify_schema_v1.go` — JSON Schema for
  `--json-schema` *(NEW)*.
- `internal/handoff/payload.go` — handoff payload schema *(unchanged)*.

## Verifying SDK removal

After a successful build:

```sh
! grep -RnE "anthropic-sdk-go|ANTHROPIC_API_KEY" cmd/ internal/ scripts/
```

Should exit `0` (no matches). The `anthropic-sdk-go` import path and
the `ANTHROPIC_API_KEY` literal must not appear anywhere in the
runtime tree.
