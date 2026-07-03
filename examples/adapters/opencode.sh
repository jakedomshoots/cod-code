#!/bin/sh

prompt=$(cat)

if [ -z "${CEO_OPENCODE_ADAPTER_COMMAND:-}" ]; then
	cat <<'JSON'
{"status":"needs_input","summary":"OpenCode adapter setup is missing. Set CEO_OPENCODE_ADAPTER_COMMAND to a wrapper that supports CEO_HARNESS_ADAPTER_PROBE=version and dry-run.","confidence":0.4,"evidence":["adapter: opencode","setup: docs/adapters/opencode.md"]}
JSON
	exit 0
fi

printf '%s' "$prompt" | sh -c "$CEO_OPENCODE_ADAPTER_COMMAND"
