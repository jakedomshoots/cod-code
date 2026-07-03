#!/bin/sh

prompt=$(cat)

if [ -z "${CEO_GOOSE_ADAPTER_COMMAND:-}" ]; then
	cat <<'JSON'
{"status":"needs_input","summary":"Goose adapter setup is missing. Set CEO_GOOSE_ADAPTER_COMMAND to a wrapper that supports CEO_HARNESS_ADAPTER_PROBE=version and dry-run.","confidence":0.4,"evidence":["adapter: goose","setup: docs/adapters/goose.md"]}
JSON
	exit 0
fi

printf '%s' "$prompt" | sh -c "$CEO_GOOSE_ADAPTER_COMMAND"
