# Contract — `claude` CLI subprocess invocation

This contract is **internal**: it governs how `internal/analysis/claudecli/`
talks to the `claude` binary. Bumping any rule here is a v2 patch-level
change unless it changes the shape of the `analysis_record` (in which
case it is a feature-level change — see `claude-prompt.md` and
`handoff.md`).

## Argv shape (one identify call)

```
claude -p \
    --output-format json \
    --system-prompt    <embedded identify-game.v1 body> \
    --json-schema      <embedded identify-game.v1 schema> \
    [--model           <user-supplied or NANO_GAMEPLAYS_MODEL>] \
    --add-dir          <per-call tempdir absolute path> \
    --allowedTools     Read \
    --disallowedTools  Bash \
    --disable-slash-commands \
    --no-session-persistence \
    -- \
    "Identify the game shown in the frames at <path1> and <path2>. Respond per the schema."
```

**Rules**:
- `claude` is resolved via `exec.LookPath("claude")` once at program
  start; argv[0] is the absolute path returned.
- `--model` is included only when the user supplied a value via
  `--model` or `NANO_GAMEPLAYS_MODEL`. When omitted, `claude`'s own
  default model is used and `analysis.model_id` is filled from the
  envelope (or left empty per FR-007).
- The user message string lists the frame paths in offset order
  (`frame_10.png` then `frame_60.png` for two frames; degraded to one
  if extraction lost the second). The exact message wording is
  governed by `claude-prompt.md` and may be tweaked under PATCH-level
  rules.

## Stdin / stdout / stderr

- **Stdin**: closed immediately (no input).
- **Stdout**: a single JSON envelope from `--output-format json`. Read
  to EOF, decode into `claudecli.ResultEnvelope`, then close.
- **Stderr**: captured to a bounded in-memory buffer (cap ~64 KB) and
  preserved into `analysis.error` only when the call failed (non-zero
  exit, malformed envelope, schema-validation failure).

## Environment

- `cmd.Env` defaults to the parent environment unmodified.
  Specifically, **`ANTHROPIC_API_KEY` is neither set nor scrubbed** by
  `nano-gameplays`; whether `claude` consults it as an auth fallback
  is `claude`-side behaviour outside this contract's scope.
- **No** working directory override (`cmd.Dir` unset; inherits parent
  cwd).

## Cancellation

- Built via `exec.CommandContext(ctx, ...)` so the context the model
  passes to `analysis.Run` drives subprocess termination.
- On context cancellation the Go runtime sends `SIGKILL`. The
  application MUST NOT wait for graceful shutdown — the call is
  short-lived and idempotent, and a fast kill keeps selection
  switching responsive.

## Exit codes

| Exit | Treatment |
|------|-----------|
| `0`, envelope `is_error=false`, schema-valid `result`         | Success. Emit `AnalysisOKMsg`. |
| `0`, envelope `is_error=true` OR `result` not schema-valid    | Treated as Claude failure (FR-005). Emit `AnalysisErrMsg` carrying envelope error / schema-validation error string. |
| Non-zero, killed via context cancellation                     | Silently dropped (the user already moved on). |
| Non-zero, any other reason                                    | Treated as Claude failure (FR-005). Emit `AnalysisErrMsg` carrying the captured stderr (truncated to the in-memory cap). |

## Tempdir lifecycle

- Created once per call by
  `os.MkdirTemp("", "nano-gameplays-*")` (mode `0700`).
- One PNG per `CapturedFrame` written with `0600` and a deterministic
  filename `frame_<offsetPct>.png`.
- `defer os.RemoveAll(tempdir)` is the **only** cleanup mechanism
  required; it runs on every exit path the analysis goroutine can
  take (success, parse error, context cancel, panic).
- The application MUST NOT rely on the OS reaping `os.TempDir()` later;
  cleanup is mandatory and synchronous on the call's exit path.

## Reproducibility

- `analysis.prompt_version` ← `identify-game.v1` (constant in
  `internal/analysis/prompt`).
- `analysis.model_id` ← per R-5: user override if any, else envelope
  `Model`, else empty + error note.
- `analysis.frame_offsets_pct`, `analysis.frame_sha256` ← from the
  extractor (unchanged).
- `analysis.raw_response` ← envelope `result` string verbatim.
- `analysis.error` ← envelope `error` string OR captured stderr OR a
  short reason string the application generated (e.g. `model id
  unavailable`).
- `analysis.request_started_at` / `request_finished_at` ← RFC3339 UTC
  timestamps captured by the application around the subprocess call.
