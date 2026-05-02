#!/usr/bin/env bash
# Pre-merge gate for feature 002-claude-cli-backend SC-007: assert that the
# runtime source tree (cmd/, internal/) contains zero references to the
# deleted Anthropic SDK or to the ANTHROPIC_API_KEY environment variable.
# Exits 0 on success (no matches), non-zero on failure.
#
# scripts/ is intentionally excluded — this verifier itself names the
# strings it forbids; spec migration notes / changelogs may also reference
# them.
set -euo pipefail

NEEDLE='anthropic-sdk-go|ANTHROPIC_API_KEY'

if grep -RnE "$NEEDLE" cmd/ internal/ 2>/dev/null; then
  echo "FAIL: forbidden strings found in runtime tree (see matches above)" >&2
  exit 1
fi
echo "ok: no anthropic-sdk-go or ANTHROPIC_API_KEY references in cmd/ internal/"
