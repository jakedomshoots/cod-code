#!/bin/sh
set -eu

root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
schema="${CEO_CLAUDE_MODEL_SCHEMA:-$root/scripts/codex-model-output.schema.json}"
tmp=$(mktemp -t ceo-claude-model.XXXXXX)
log=$(mktemp -t ceo-claude-model-log.XXXXXX)

cleanup() {
  rm -f "$tmp" "$log"
}
trap cleanup EXIT

if [ "${CEO_HARNESS_ADAPTER_PROBE:-}" = "version" ]; then
  claude --version
  exit 0
fi

if [ "${CEO_HARNESS_ADAPTER_PROBE:-}" = "dry-run" ]; then
  cat >/dev/null
  printf '%s\n' '{"status":"pass","summary":"claude oauth wrapper dry-run ready","confidence":1,"evidence":["wrapper: claude-model-command.sh","token_storage: none"],"tool_requests":[],"patches":[]}'
  exit 0
fi

prompt=$(cat)
kind="${CEO_MODEL_REQUEST_KIND:-unknown}"
agent="${CEO_AGENT_NAME:-unknown}"
role="${CEO_AGENT_ROLE:-unknown}"
context="${CEO_CONTEXT_MODE:-unknown}"

run_prompt=$(printf '%s\n' \
  "You are a real model backend for CEO Harness." \
  "Return one JSON object only. Do not include markdown." \
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
  "ceo_delegation -> choose the smallest useful set; for a narrow code edit usually select coder only." \
  "ceo_review -> set recommended_verdict to pass or fail and include a short summary." \
  "subagent work -> set status, summary, confidence, evidence, tool_requests, and patches." \
  "" \
  "Harness prompt:" \
  "$prompt")

schema_json=$(cat "$schema")

if [ -n "${CEO_CLAUDE_MODEL:-}" ]; then
  if ! claude \
    --print "$run_prompt" \
    --model "$CEO_CLAUDE_MODEL" \
    --output-format json \
    --no-session-persistence \
    --permission-mode plan \
    --json-schema "$schema_json" >"$tmp" 2>"$log"; then
    cat "$log" >&2
    exit 1
  fi
else
  if ! claude \
    --print "$run_prompt" \
    --output-format json \
    --no-session-persistence \
    --permission-mode plan \
    --json-schema "$schema_json" >"$tmp" 2>"$log"; then
    cat "$log" >&2
    exit 1
  fi
fi

if [ "${CEO_CLAUDE_MODEL_COMMAND_VERBOSE:-0}" = "1" ]; then
  cat "$log" >&2
fi

python3 - "$tmp" <<'PY'
import json
import re
import sys

raw = open(sys.argv[1], "r", encoding="utf-8").read().strip()

def emit(payload):
    print(json.dumps(payload, separators=(",", ":")))
    raise SystemExit(0)

def try_payload(text):
    text = text.strip()
    if not text:
        return
    if text.startswith("{"):
        try:
            parsed = json.loads(text)
            if isinstance(parsed, dict):
                emit(parsed)
        except json.JSONDecodeError:
            pass
    fenced = re.search(r"```(?:json)?\s*(\{.*?\})\s*```", text, re.DOTALL)
    if fenced:
        try:
            emit(json.loads(fenced.group(1).strip()))
        except json.JSONDecodeError:
            pass
    decoder = json.JSONDecoder()
    for index, char in enumerate(text):
        if char != "{":
            continue
        try:
            payload, _ = decoder.raw_decode(text[index:])
        except json.JSONDecodeError:
            continue
        if isinstance(payload, dict):
            emit(payload)

try:
    outer = json.loads(raw)
except json.JSONDecodeError:
    try_payload(raw)
    raise SystemExit("claude model command returned no JSON object")

if isinstance(outer, dict):
    for key in ("status", "recommended_verdict", "selected_subagents", "patches"):
        if key in outer:
            emit(outer)
    for key in ("result", "content", "text", "message"):
        value = outer.get(key)
        if isinstance(value, str):
            try_payload(value)
        if isinstance(value, list):
            joined = "\n".join(item.get("text", "") for item in value if isinstance(item, dict))
            try_payload(joined)

raise SystemExit("claude model command returned no JSON object")
PY
