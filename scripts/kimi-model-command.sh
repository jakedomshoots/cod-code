#!/bin/sh
set -eu

tmp=$(mktemp -t ceo-kimi-model.XXXXXX)
log=$(mktemp -t ceo-kimi-model-log.XXXXXX)

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
  "You are a real model backend for CEO Harness, not the outer coding agent." \
  "Return one JSON object only. Do not include markdown unless you cannot avoid it." \
  "Do not edit files directly. If an edit is needed, propose it in the JSON patches array." \
  "For existing files, every patch must include path, exact old text, and new text. Use content only when creating a new file." \
  "If the harness prompt lists Required diff terms, the changed file must include those exact terms." \
  "If you need more workspace content, request it with tool_requests instead of guessing." \
  "" \
  "Request metadata:" \
  "kind: $kind" \
  "agent: $agent" \
  "role: $role" \
  "context: $context" \
  "" \
  "JSON contracts:" \
  "ceo_delegation -> {\"selected_subagents\":[\"coder\"],\"summary\":\"short reason\"}; choose the smallest useful set." \
  "ceo_review -> {\"recommended_verdict\":\"pass|fail\",\"summary\":\"short reason\"}" \
  "subagent work -> {\"status\":\"pass|fail|needs_input\",\"summary\":\"short result\",\"confidence\":0.0,\"evidence\":[\"item\"],\"tool_requests\":[],\"patches\":[{\"path\":\"existing.txt\",\"old\":\"exact old text\",\"new\":\"replacement text\"}]}" \
  "" \
  "Harness prompt:" \
  "$prompt")

if ! kimi -p "$run_prompt" --output-format stream-json >"$tmp" 2>"$log"; then
  cat "$log" >&2
  exit 1
fi

python3 - "$tmp" <<'PY'
import json
import re
import sys

content = None
with open(sys.argv[1], "r", encoding="utf-8") as handle:
    for raw_line in handle:
        line = raw_line.strip()
        if not line:
            continue
        try:
            event = json.loads(line)
        except json.JSONDecodeError:
            continue
        if event.get("role") == "assistant" and isinstance(event.get("content"), str):
            content = event["content"].strip()

if not content:
    raise SystemExit("kimi model command returned no assistant content")

if content.startswith("{"):
    print(content)
    raise SystemExit(0)

fenced = re.search(r"```(?:json)?\s*(\{.*?\})\s*```", content, re.DOTALL)
if fenced:
    print(fenced.group(1).strip())
    raise SystemExit(0)

decoder = json.JSONDecoder()
for index, char in enumerate(content):
    if char != "{":
        continue
    try:
        payload, _ = decoder.raw_decode(content[index:])
    except json.JSONDecodeError:
        continue
    print(json.dumps(payload, separators=(",", ":")))
    raise SystemExit(0)

raise SystemExit("kimi model command returned no JSON object")
PY
