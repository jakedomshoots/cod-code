#!/bin/sh
set -eu
root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)


tmp=$(mktemp -t ceo-opencode-model.XXXXXX)
log=$(mktemp -t ceo-opencode-model-log.XXXXXX)

cleanup() {
  rm -f "$tmp" "$log"
}
trap cleanup EXIT

if [ "${CEO_HARNESS_ADAPTER_PROBE:-}" = "version" ]; then
  opencode --version
  exit 0
fi

if [ "${CEO_HARNESS_ADAPTER_PROBE:-}" = "dry-run" ]; then
  cat >/dev/null
  printf '%s\n' '{"status":"pass","summary":"opencode oauth wrapper dry-run ready","confidence":1,"evidence":["wrapper: opencode-model-command.sh","token_storage: none"],"tool_requests":[],"patches":[]}'
  exit 0
fi

prompt=$(cat)
kind="${CEO_MODEL_REQUEST_KIND:-unknown}"
agent="${CEO_AGENT_NAME:-unknown}"
role="${CEO_AGENT_ROLE:-unknown}"
context="${CEO_CONTEXT_MODE:-unknown}"
workspace_context=$(python3 "$root/scripts/model-command-context.py" "$prompt")

run_prompt=$(printf '%s\n' \
  "You are a real model backend for Cod Code, not the outer coding agent." \
  "You are running in an isolated temporary directory. The workspace snippets below are the authoritative file contents." \
  "Return one JSON object only. Do not include markdown unless you cannot avoid it." \
  "Do not edit files directly. If an edit is needed, propose it in the JSON patches array." \
  "Do not use shell, filesystem, or agentic actions. Only return the JSON object for the requested contract." \
  "For existing files, every patch must include path, exact old text, and new text. Use content only when creating a new file." \
  "If the harness prompt lists Required diff terms, the changed file must include those exact terms." \
  "If you need more workspace content, request it with tool_requests instead of guessing." \
  "Return needs_input only for a missing user decision, never just because you want to inspect files." \
  "" \
  "Request metadata:" \
  "kind: $kind" \
  "agent: $agent" \
  "role: $role" \
  "context: $context" \
  "" \
  "JSON contracts:" \
  "ceo_delegation -> {\"selected_subagents\":[\"coder\"],\"summary\":\"short reason\"}; choose the smallest useful set." \
  "ceo_delegation must include a non-empty selected_subagents array using candidate names from the harness prompt." \
  "ceo_review -> {\"recommended_verdict\":\"pass|fail\",\"summary\":\"short reason\"}" \
  "For ceo_review, guard_verdict, checks, changed_files, patch_results, and workspace snippets are observed facts." \
  "Do not contradict guard_verdict, checks, patch_results, or workspace snippets. Recommend fail only for a concrete unmet requirement visible in those facts." \
  "subagent work -> {\"status\":\"pass|fail|needs_input\",\"summary\":\"short result\",\"confidence\":0.0,\"evidence\":[\"item\"],\"tool_requests\":[],\"patches\":[{\"path\":\"existing.txt\",\"old\":\"exact old text\",\"new\":\"replacement text\"}]}" \
  "" \
  "$workspace_context" \
  "" \
  "Harness prompt:" \
  "$prompt")

if [ -n "${CEO_OPENCODE_MODEL:-}" ]; then
  if ! opencode run --format json --model "$CEO_OPENCODE_MODEL" "$run_prompt" >"$tmp" 2>"$log"; then
    cat "$log" >&2
    exit 1
  fi
else
  if ! opencode run --format json "$run_prompt" >"$tmp" 2>"$log"; then
    cat "$log" >&2
    exit 1
  fi
fi

if [ "${CEO_OPENCODE_MODEL_COMMAND_VERBOSE:-0}" = "1" ]; then
  cat "$log" >&2
fi

python3 "$root/scripts/model-command-normalize.py" "$tmp" "$prompt"
