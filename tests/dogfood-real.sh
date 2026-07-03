#!/bin/sh
set -eu

root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
cd "$root"

evidence_dir=".omo/evidence/dogfood-real"
rm -rf "$evidence_dir"

dry_run_output=$(mktemp)
missing_output=$(mktemp)
tmp_repo=$(mktemp -d)
missing_parent=$(mktemp -d)
trap 'rm -f "$dry_run_output" "$missing_output"; rm -rf "$tmp_repo" "$missing_parent"' EXIT

printf '%s\n' "fixture" >"$tmp_repo/fixture.txt"
before=$(find "$tmp_repo" -type f | sort)

sh scripts/dogfood-real.sh --dry-run --repo "fixture:$tmp_repo" >"$dry_run_output"

index="$evidence_dir/index.md"
if [ ! -f "$index" ]; then
  printf '%s\n' "missing dogfood-real index: $index" >&2
  exit 1
fi

scenario_count=$(grep -c '^| scenario-' "$index")
if [ "$scenario_count" -lt 5 ]; then
  printf '%s\n' "expected at least 5 scenarios in $index, found $scenario_count" >&2
  exit 1
fi

after=$(find "$tmp_repo" -type f | sort)
if [ "$before" != "$after" ]; then
  printf '%s\n' "dry-run touched fixture repo" >&2
  exit 1
fi

missing_repo="$missing_parent/does-not-exist"
sh scripts/dogfood-real.sh --dry-run --repo "missing:$missing_repo" >"$missing_output"

if [ -e "$missing_repo" ]; then
  printf '%s\n' "missing repo path was created: $missing_repo" >&2
  exit 1
fi

if ! grep -q '| missing | skipped_missing_repo |' "$index"; then
  printf '%s\n' "missing repo was not recorded as skipped" >&2
  exit 1
fi

if grep -q '| missing | pass |' "$index"; then
  printf '%s\n' "missing repo was falsely recorded as pass" >&2
  exit 1
fi

printf '%s\n' "dogfood-real test ok"
