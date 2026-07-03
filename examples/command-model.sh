#!/bin/sh

prompt=$(cat)

if ! printf '%s' "$prompt" | grep -q '^role:'; then
	cat <<'JSON'
{"status":"fail","summary":"example command model did not receive a role line","confidence":0.4,"evidence":["prompt contract missing role"]}
JSON
	exit 0
fi

printf '{"status":"pass","summary":"example command model handled %s","confidence":0.8,"evidence":["request kind: %s","role: %s","context: %s"]}\n' \
	"${CEO_AGENT_NAME:-unknown}" \
	"${CEO_MODEL_REQUEST_KIND:-unknown}" \
	"${CEO_AGENT_ROLE:-unknown}" \
	"${CEO_CONTEXT_MODE:-unknown}"
