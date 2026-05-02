# Phase 0 — Research

Open technical questions implied by this feature spec, resolved.

---

## R-1: How to invoke the local `claude` CLI non-interactively

**Decision**: Spawn the binary as
`claude -p --output-format json --system-prompt <embedded prompt> --json-schema <schema> --model <model-or-empty> --add-dir <tempdir> --allowedTools Read --disallowedTools Bash --disable-slash-commands --no-session-persistence "<user-message>"`,
captured via `exec.CommandContext` so the parent context drives
cancellation. The binary is on `PATH` and is owned by the user.

**Rationale**:
- `-p` / `--print` is the documented non-interactive shape and the only
  one that emits a result and exits without a pty (Claude Code
  CLI 2.1.x).
- `--output-format json` returns a single JSON envelope (result string,
  session id, usage) which is trivial to parse from Go.
- `--system-prompt` (not `--append-system-prompt`) replaces the default
  system prompt entirely with our `identify-game.v1` template, removing
  any per-machine drift in the upstream prompt.
- `--json-schema` constrains the model's `result` payload to our exact
  identify-response shape. This eliminates the brittle prose-parsing
  risk class (FR-018 in 001 was driven by free-form responses; with
  `--json-schema`, malformed-output failures collapse to a single
  generic error path).
- `--add-dir <tempdir>` is the only way to grant file-tool access to a
  directory the model needs to read; without it the `Read` tool would
  refuse the per-call frame paths.
- `--allowedTools Read --disallowedTools Bash --disable-slash-commands
  --no-session-persistence` together produce a deterministic, sandboxed,
  non-persistent invocation. No skills, no shell, no resumable session
  state on disk.
- `--bare` was **rejected**: per its own help text, `--bare` "skips
  keychain reads" and forces auth via `ANTHROPIC_API_KEY` only. That
  is the literal opposite of this feature's intent (use the user's
  signed-in OAuth/keychain credential).

**Alternatives considered**:
- **Pipe a JSON envelope on stdin (`--input-format stream-json`)**.
  Powerful, but overkill for a single-message identify call. Rejected
  on simplicity grounds.
- **Use `--append-system-prompt` and rely on the default system prompt
  for context**. Rejected — defeats reproducibility (the default
  system prompt varies across versions / users).

---

## R-2: How to deliver PNG frames to the `claude` CLI

**Decision**: Per call, create a private temporary directory via
`os.MkdirTemp("", "nano-gameplays-")` (mode `0700`), write each frame
as `frame_<offsetPct>.png` (mode `0600`) into that directory, pass the
directory to `claude` via `--add-dir`, reference each frame by absolute
path in the user message, and `defer os.RemoveAll(tempdir)` so the
directory is gone before the analysis goroutine returns — on success,
on error, on context cancellation, and on panic.

**Rationale**:
- The `claude` CLI does not document a binary-stdin path for images.
  `--file <id:relpath>` targets the Anthropic Files API (pre-uploaded
  resources by file ID), not local PNG bytes.
- The `Read` tool with `--add-dir` is the documented path for granting
  the model access to local files for one call.
- Writing under `os.TempDir()` (`/tmp` on Linux, `$TMPDIR` on macOS)
  with mode `0700` plus per-file `0600` keeps the on-disk window
  user-private; deletion via deferred `os.RemoveAll` runs on every
  exit path the goroutine can take.

**Alternatives considered**:
- **Embed PNGs base64 in the prompt text**. The model would not see
  them as images; it would see them as text and the identification
  would degrade severely. Rejected.
- **Write PNGs into the project's working directory and rely on
  `claude`'s default cwd-trust**. Pollutes the user's repo with
  transient files; non-private; rejected.

**Tradeoff**: a brief on-disk window for the frames. This is the
single Complexity Tracking entry on `plan.md` for this feature; the
intent of 001 FR-023 is preserved by the user-only permissions and
the deferred cleanup, even though the bytes are no longer
strictly memory-only.

---

## R-3: Constraining the model's response shape

**Decision**: Pass the `identify-game.v1` JSON Schema to the CLI via
`--json-schema <schema-string>`. The schema describes:

```json
{
  "type": "object",
  "additionalProperties": false,
  "properties": {
    "game_title":             { "type": "string" },
    "scene_or_level_or_mode": { "type": "string" },
    "confidence":             { "type": "string", "enum": ["low","medium","high"] }
  },
  "required": ["game_title", "scene_or_level_or_mode", "confidence"]
}
```

The schema is embedded in the binary alongside the prompt template via
`//go:embed`, and bumped together with the prompt under the
`identify-game.vN` versioning rules.

**Rationale**:
- Schema-constrained output eliminates the "tolerate a stray ```json
  fence" defensive parsing in `internal/analysis/claude.go` today.
- A schema mismatch is now reported by the CLI itself (the call's
  `result.is_error` is `true` and the envelope describes the
  validation failure), which the application surfaces verbatim into
  `analysis.error` per FR-005.
- Confidence enum coercion (the `coerceConfidence` helper from 001) is
  no longer needed at runtime; the schema makes any other value
  impossible to receive. The helper MAY remain as defensive code or
  MAY be removed; preference is removal once the schema-enforced path
  is verified.

**Alternatives considered**:
- **Keep prose-parsing on the client side**. Rejected — schema
  enforcement is strictly stronger and is offered for free by the CLI.

---

## R-4: Cancellation when the user picks a different video

**Decision**: `exec.CommandContext(ctx, "claude", argv...)` with the
existing per-state `context.Context` from `internal/app/model.go`. On a
new `selectVideoMsg` for a different video, the model calls the stored
cancel func; the Go runtime sends `SIGKILL` to the `claude`
subprocess; any late stdout from the killed call is dropped on
arrival because the message handler checks the video path against the
current selection (path mismatch ⇒ stale ⇒ ignored).

**Rationale**:
- Identical semantics to the SDK path (HTTP request abort), achieved
  via `exec.CommandContext`'s built-in cancellation. No new
  cancellation primitive needed.
- The "drop stale by path mismatch" guard in
  `internal/app/update.go` already exists and applies as-is.

**Alternatives considered**: managing the subprocess via
`Process.Signal(SIGTERM)` for a graceful shutdown. Rejected — the
identify call is short-lived and idempotent; SIGKILL is fine and
removes a wait-for-graceful-exit window that would slow up the next
selection.

---

## R-5: Surfacing the model id in the AnalysisRecord

**Decision**:

- If the user (or `NANO_GAMEPLAYS_MODEL`) supplied a value, that exact
  string is what we passed to `--model` and what we record in
  `analysis.model_id`.
- If neither was supplied, we omit `--model` from the argv (defer to
  the `claude` CLI's own default). In that case we attempt to pull
  the resolved model from the `--output-format json` envelope; if
  the CLI does not surface it, `analysis.model_id` is left empty and
  `analysis.error` notes that the model identifier was unavailable
  (per FR-007).

**Rationale**:
- Recording an invented or stale default like `claude-sonnet-4-6`
  when the CLI actually used `claude-sonnet-4-7` would defeat
  Constitution Principle III's reproducibility bar.
- Empty + `analysis.error` describing the omission is honest and
  re-derivable: a re-run with the same `--model` value (in this case,
  none) on the same CLI version reproduces the call.

**Alternatives considered**: hard-default to `claude-sonnet-4-6` when
no override is given. Rejected — couples the application to a model
name that the `claude` CLI may have moved past.

---

## R-6: Removing the `anthropic-sdk-go` runtime dependency

**Decision**: Delete the import from `internal/analysis/claude.go`,
delete the SDK from `go.mod` via `go mod tidy`, and delete every
runtime reference (types, error wrappers, env-var probes). The only
allowed remaining mention is in retired-history files
(`specs/001-...`/`tasks.md` notes, `CHANGELOG`-equivalents in commit
messages); none in the runtime tree.

**Rationale**:
- FR-001 / FR-011 / SC-007 explicitly require zero remaining runtime
  references. The most reliable way to enforce that is to delete the
  import, run `go mod tidy`, and grep the runtime tree for the
  package name.

**Verification**: a one-line shell check (added to
`scripts/verify-roundtrip.sh` or a new
`scripts/verify-no-sdk.sh`):

```sh
! grep -RnE "anthropic-sdk-go|ANTHROPIC_API_KEY" cmd/ internal/ scripts/
```

The command MUST exit 0 (no matches) on the post-merge tree.

---

## R-7: Behaviour when `ANTHROPIC_API_KEY` is set in the user's env

**Decision**: Ignore it. We do **not** pass it through to the
subprocess explicitly (the user's environment leaks through by
default; the `claude` CLI may or may not consult it depending on
whether OAuth is also signed in). We document in the README that for
this feature the variable is no longer consulted by `nano-gameplays`
itself; whether `claude` consults it as a fallback is a `claude`-side
behaviour outside our scope.

**Rationale**:
- Removing the SDK is enough to satisfy FR-011 (no silent re-enable
  of the SDK path). What `claude` does with the variable is defined
  by `claude`, not by `nano-gameplays`.

**Alternatives considered**: actively scrub `ANTHROPIC_API_KEY` from
`cmd.Env` before exec. Rejected — surprising; the user might *want*
it consulted as a fallback by their `claude` install.

---

## R-8: Preflight check for the `claude` binary

**Decision**: At program start (before the Bubble Tea program loop),
call `exec.LookPath("claude")`. If it returns an error, print a
single-line stderr message —
`claude CLI not found in PATH; install Claude Code from https://claude.com/claude-code` —
and exit `2` (precondition failure, matching the existing exit-code
contract for the `ffmpeg` preflight from 001 R-2).

**Rationale**:
- Failing at launch is far better UX than failing at the moment the
  user picks a video (which surfaces as the empty FR-018-style form
  and looks like a Claude failure rather than a missing binary).
- The wording mirrors the existing `ffmpeg not found in PATH` message
  for consistency.

**Alternatives considered**: try-then-fail-on-first-call. Rejected —
the failure looks like an intermittent Claude error, which is
diagnostically misleading.
