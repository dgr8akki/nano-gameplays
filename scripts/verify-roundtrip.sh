#!/usr/bin/env bash
# Verify the HandoffPayload JSON round-trips losslessly through the identify
# subcommand. Requires `jq` and a video fixture at the path supplied as $1.
set -euo pipefail

if [[ $# -lt 1 ]]; then
  echo "usage: scripts/verify-roundtrip.sh <video-path>" >&2
  exit 2
fi

VIDEO="$1"
BIN="${BIN:-./bin/nano-gameplays}"

if [[ ! -x "$BIN" ]]; then
  echo "binary not found at $BIN; build first: go build -o $BIN ./cmd/nano-gameplays" >&2
  exit 2
fi

"$BIN" identify --video "$VIDEO" --json |
  jq -e '.schema_version == "1" and (.analysis.prompt_version == "identify-game.v1")'
echo "ok: schema_version=1, prompt_version=identify-game.v1"
