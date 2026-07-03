#!/bin/sh

prompt=$(cat)

if [ -z "${CEO_CLAUDE_ADAPTER_COMMAND:-}" ]; then
	cat <<'JSON'
{"status":"needs_input","summary":"Claude Code adapter setup is missing. Set CEO_CLAUDE_ADAPTER_COMMAND to a wrapper that supports CEO_HARNESS_ADAPTER_PROBE=version and dry-run.","confidence":0.4,"evidence":["adapter: claude","setup: docs/adapters/claude.md"]}
JSON
	exit 0
fi

printf '%s' "$prompt" | sh -c "$CEO_CLAUDE_ADAPTER_COMMAND"
