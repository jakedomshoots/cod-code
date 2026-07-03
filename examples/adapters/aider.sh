#!/bin/sh

prompt=$(cat)

if [ -z "${CEO_AIDER_ADAPTER_COMMAND:-}" ]; then
	cat <<'JSON'
{"status":"needs_input","summary":"Aider adapter setup is missing. Set CEO_AIDER_ADAPTER_COMMAND to a wrapper that supports CEO_HARNESS_ADAPTER_PROBE=version and dry-run.","confidence":0.4,"evidence":["adapter: aider","setup: docs/adapters/aider.md"]}
JSON
	exit 0
fi

printf '%s' "$prompt" | sh -c "$CEO_AIDER_ADAPTER_COMMAND"
