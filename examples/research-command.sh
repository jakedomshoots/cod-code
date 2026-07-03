#!/bin/sh

query=$(cat)

if [ -z "${CEO_RESEARCH_QUERY:-}" ]; then
	echo "missing CEO_RESEARCH_QUERY" >&2
	exit 1
fi

if [ "$query" != "$CEO_RESEARCH_QUERY" ]; then
	echo "stdin query did not match CEO_RESEARCH_QUERY" >&2
	exit 1
fi

printf 'research example handled %s\n' "$CEO_RESEARCH_QUERY"
