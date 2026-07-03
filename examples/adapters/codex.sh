#!/bin/sh

prompt=$(cat)

if [ -z "${CEO_CODEX_ADAPTER_COMMAND:-}" ]; then
	cat <<'JSON'
{"status":"needs_input","summary":"Codex CLI adapter setup is missing. Set CEO_CODEX_ADAPTER_COMMAND to a wrapper that supports CEO_HARNESS_ADAPTER_PROBE=version and dry-run.","confidence":0.4,"evidence":["adapter: codex","setup: docs/adapters/codex.md"]}
JSON
	exit 0
fi

printf '%s' "$prompt" | sh -c "$CEO_CODEX_ADAPTER_COMMAND"
