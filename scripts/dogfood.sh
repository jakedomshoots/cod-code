#!/bin/sh
set -eu

root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
tmpdir=$(mktemp -d)
trap 'rm -rf "$tmpdir"' EXIT

bin="$tmpdir/ceo-packet"
workspace="$tmpdir/workspace"
mkdir -p "$workspace"

cd "$root"
go build -trimpath -o "$bin" ./cmd/ceo-packet

"$bin" --demo --format text >/dev/null

printf '%s\n' old >"$workspace/app.txt"
preview=$("$bin" \
  --workspace "$workspace" \
  --dry-run \
  --replace app.txt old new \
  "Patch app text")

digest=$(printf '%s\n' "$preview" | sed -n 's/.*"preview_digest": "\([^"]*\)".*/\1/p' | head -n 1)
if [ -z "$digest" ]; then
  printf '%s\n' "dogfood: missing preview digest" >&2
  exit 1
fi

"$bin" \
  --workspace "$workspace" \
  --approve-preview "$digest" \
  --replace app.txt old new \
  --check sh -c 'grep -q new app.txt' -- \
  "Patch app text" >/dev/null

set +e
"$bin" \
  --workspace "$workspace" \
  --check sh -c 'exit 7' -- \
  "Exercise failed check handling" >/dev/null 2>/dev/null
failed_check_status=$?

"$bin" \
  --workspace "$workspace" \
  --model-command sh -c 'prompt=$(cat); case "$prompt" in *"agent: scanner"*) printf "%s\n" "{\"status\":\"needs_input\",\"summary\":\"Need target package\",\"questions\":[\"Which package should I change?\"]}" ;; *) printf "%s\n" "{\"summary\":\"ok\"}" ;; esac' -- \
  "Fix ambiguous package" >/dev/null 2>/dev/null
needs_input_status=$?
set -e

if [ "$failed_check_status" -eq 0 ]; then
  printf '%s\n' "dogfood: failed-check scenario unexpectedly passed" >&2
  exit 1
fi

if [ "$needs_input_status" -eq 0 ]; then
  printf '%s\n' "dogfood: needs-input scenario unexpectedly passed" >&2
  exit 1
fi

"$bin" --workspace "$workspace" --job-context latest --format text | grep -q 'Resume: ceo-packet'

printf '%s\n' "dogfood ok"
