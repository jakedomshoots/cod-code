#!/bin/sh
set -eu

tmp=$(mktemp -t ceo-goose-model.XXXXXX)
log=$(mktemp -t ceo-goose-model-log.XXXXXX)

cleanup() {
  rm -f "$tmp" "$log"
}
trap cleanup EXIT

if [ "${CEO_HARNESS_ADAPTER_PROBE:-}" = "version" ]; then
  goose --version
  exit 0
fi

if [ "${CEO_HARNESS_ADAPTER_PROBE:-}" = "dry-run" ]; then
  cat >/dev/null
  printf '%s\n' '{"status":"pass","summary":"goose oauth wrapper dry-run ready","confidence":1,"evidence":["wrapper: goose-model-command.sh","token_storage: none"],"tool_requests":[],"patches":[]}'
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
  "ceo_delegation -> {\"selected_subagents\":[\"coder\"],\"summary\":\"short reason\"}" \
  "ceo_review -> {\"recommended_verdict\":\"pass|fail\",\"summary\":\"short reason\"}" \
  "subagent work -> {\"status\":\"pass|fail|needs_input\",\"summary\":\"short result\",\"confidence\":0.0,\"evidence\":[\"item\"],\"tool_requests\":[],\"patches\":[]}" \
  "" \
  "Harness prompt:" \
  "$prompt")

if [ -n "${CEO_GOOSE_PROVIDER:-}" ] && [ -n "${CEO_GOOSE_MODEL:-}" ]; then
  if ! goose run --text "$run_prompt" --no-session --quiet --output-format json --max-turns 1 --provider "$CEO_GOOSE_PROVIDER" --model "$CEO_GOOSE_MODEL" >"$tmp" 2>"$log"; then
    cat "$log" >&2
    exit 1
  fi
elif [ -n "${CEO_GOOSE_MODEL:-}" ]; then
  if ! goose run --text "$run_prompt" --no-session --quiet --output-format json --max-turns 1 --model "$CEO_GOOSE_MODEL" >"$tmp" 2>"$log"; then
    cat "$log" >&2
    exit 1
  fi
else
  if ! goose run --text "$run_prompt" --no-session --quiet --output-format json --max-turns 1 >"$tmp" 2>"$log"; then
    cat "$log" >&2
    exit 1
  fi
fi

if [ "${CEO_GOOSE_MODEL_COMMAND_VERBOSE:-0}" = "1" ]; then
  cat "$log" >&2
fi

python3 - "$tmp" <<'PY'
import json
import re
import sys

raw = open(sys.argv[1], "r", encoding="utf-8").read()

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

def collect_text(value):
    if isinstance(value, str):
        return value
    if isinstance(value, list):
        parts = []
        for item in value:
            if isinstance(item, str):
                parts.append(item)
            elif isinstance(item, dict):
                for key in ("text", "content", "message"):
                    if isinstance(item.get(key), str):
                        parts.append(item[key])
        return "\n".join(parts)
    if isinstance(value, dict):
        for key in ("text", "content", "message", "result"):
            if key in value:
                return collect_text(value[key])
    return ""

try:
    outer = json.loads(raw)
except json.JSONDecodeError:
    try_payload(raw)
    raise SystemExit("goose model command returned no JSON object")

if isinstance(outer, dict):
    for key in ("status", "recommended_verdict", "selected_subagents", "patches"):
        if key in outer:
            emit(outer)
    text = collect_text(outer)
    if text:
        try_payload(text)

for line in raw.splitlines():
    try:
        event = json.loads(line)
    except json.JSONDecodeError:
        continue
    if isinstance(event, dict):
        text = collect_text(event)
        if text:
            try_payload(text)

raise SystemExit("goose model command returned no JSON object")
PY
