#!/bin/sh

prompt=$(cat)

case "${CEO_MODEL_REQUEST_KIND:-}" in
ceo_delegation)
	if ! printf '%s' "$prompt" | grep -q 'candidate_subagents'; then
		cat <<'JSON'
{"selected_subagents":[],"summary":"missing candidate subagents"}
JSON
		exit 0
	fi
	cat <<'JSON'
{"selected_subagents":["coder"],"summary":"CEO example delegated to coder"}
JSON
	;;
ceo_review)
	if ! printf '%s' "$prompt" | grep -q 'guard_verdict:'; then
		cat <<'JSON'
{"recommended_verdict":"fail","summary":"missing guard verdict"}
JSON
		exit 0
	fi
	cat <<'JSON'
{"recommended_verdict":"pass","summary":"CEO example approved the run"}
JSON
	;;
*)
	cat <<'JSON'
{"recommended_verdict":"fail","summary":"unknown CEO model request kind"}
JSON
	;;
esac
