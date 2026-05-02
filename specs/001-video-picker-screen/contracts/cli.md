# Contract — CLI surface

The application's command-line surface for this screen. Two entry forms,
both backed by the same internal flow. Required by Constitution
Principle I and the CLI/TUI parity gate.

## `nano-gameplays` (default — interactive TUI)

```
nano-gameplays [--start-dir PATH] [--model MODEL] [--no-tui]
```

**Flags**:
- `--start-dir PATH` — directory where the file picker opens. Default: the
  process's current working directory.
- `--model MODEL` — Claude model ID for analysis. Default:
  `claude-sonnet-4-6`. Also overridable via env `NANO_GAMEPLAYS_MODEL`
  (flag wins).
- `--no-tui` — refuse to start the TUI; behave as `identify` even if
  `--video` was not given (in which case prints usage and exits non-zero).

**Behaviour**:
- Stdout: nothing during the TUI session. On user confirm, prints the
  `HandoffPayload` JSON to stdout (one object, terminated with newline)
  and exits `0`.
- Stderr: only diagnostics and failure messages.
- Exit codes: `0` success/confirm, `1` user quit before confirm, `2`
  preflight failure (terminal too small, ffmpeg missing, etc.).

## `nano-gameplays identify` (non-interactive parity flow)

```
nano-gameplays identify --video PATH [--model MODEL] [--json]
```

**Flags**:
- `--video PATH` (required) — path to a gameplay video.
- `--model MODEL` — same semantics as above.
- `--json` — emit `HandoffPayload` JSON to stdout. Without `--json`, emit
  a short human-readable summary.

**Behaviour**:
- Skips the TUI entirely.
- Runs frame extraction → Claude → builds the same `HandoffPayload`.
- On Claude/extraction failure: still emits a `HandoffPayload` whose
  `metadata` fields are empty and whose `analysis.error` is set; exits
  `0` (the call was attempted; the caller can read `analysis.error`).
- Exit codes: `0` payload emitted, `2` precondition failure (file
  missing, unreadable, unsupported extension, `ffmpeg` missing).

## Stability

- The flag set above is the v1 contract. Removing or renaming a flag is
  a MAJOR change for this feature; adding a flag is MINOR.
- `--json` output schema is governed by [handoff.md](./handoff.md) and
  carries its own `schema_version`.
