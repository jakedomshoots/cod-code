#!/bin/sh
set -eu

workspace=$(pwd)
root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
schema="${CEO_CODEX_MODEL_SCHEMA:-$root/scripts/codex-model-output.schema.json}"
tmp=$(mktemp -t ceo-codex-model.XXXXXX)
log=$(mktemp -t ceo-codex-model-log.XXXXXX)

cleanup() {
  rm -f "$tmp" "$log"
}
trap cleanup EXIT

prompt=$(cat)
kind="${CEO_MODEL_REQUEST_KIND:-unknown}"
agent="${CEO_AGENT_NAME:-unknown}"
role="${CEO_AGENT_ROLE:-unknown}"
context="${CEO_CONTEXT_MODE:-unknown}"

run_prompt=$(printf '%s\n' \
  "You are a real model backend for Cod Code." \
  "Return one JSON object only. Do not include markdown." \
  "The schema is strict: include every top-level key; use null, [], or {} for unused fields." \
  "Do not edit files directly. If an edit is needed, propose it in the JSON patches array." \
  "If you need more workspace content, request it with tool_requests instead of guessing." \
  "" \
  "Request metadata:" \
  "kind: $kind" \
  "agent: $agent" \
  "role: $role" \
  "context: $context" \
  "" \
  "JSON contracts:" \
  "ceo_delegation -> choose the smallest useful set; for a narrow code edit usually select coder only; keep assignments as {}; set status/recommended_verdict to null." \
  "ceo_review -> set recommended_verdict and summary; set status to null." \
  "subagent work -> set status, summary, confidence, evidence, tool_requests, and patches; set recommended_verdict to null." \
  "" \
  "Harness prompt:" \
  "$prompt")

if [ -n "${CEO_CODEX_MODEL:-}" ]; then
  if ! printf '%s' "$run_prompt" | codex exec \
    --ephemeral \
    --ignore-user-config \
    --ignore-rules \
    --sandbox read-only \
    --skip-git-repo-check \
    --cd "$workspace" \
    --model "$CEO_CODEX_MODEL" \
    --output-schema "$schema" \
    -o "$tmp" \
    - >"$log" 2>&1; then
    cat "$log" >&2
    exit 1
  fi
else
  if ! printf '%s' "$run_prompt" | codex exec \
    --ephemeral \
    --ignore-user-config \
    --ignore-rules \
    --sandbox read-only \
    --skip-git-repo-check \
    --cd "$workspace" \
    --output-schema "$schema" \
    -o "$tmp" \
    - >"$log" 2>&1; then
    cat "$log" >&2
    exit 1
  fi
fi

if [ "${CEO_CODEX_MODEL_COMMAND_VERBOSE:-0}" = "1" ]; then
  cat "$log" >&2
fi

cat "$tmp"
