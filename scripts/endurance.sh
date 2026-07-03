#!/bin/sh
set -eu

root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
invocation_dir=$(pwd)
cd "$root"

dry_run=0
iterations=3
min_seconds=0
output_dir="$root/.omo/evidence/endurance"

usage() {
  cat <<'USAGE'
Usage: sh scripts/endurance.sh [--dry-run] [--iterations n] [--min-seconds n] [--output-dir path]

Runs repeated no-key eval loops and writes durable evidence.

Options:
  --dry-run          Write the endurance plan without running commands.
  --iterations n     Minimum number of eval loops to run. Default: 3.
  --min-seconds n    Keep running until this many seconds have elapsed. Default: 0.
  --output-dir path  Evidence directory. Default: .omo/evidence/endurance.
  --help             Show this help.
USAGE
}

while [ "$#" -gt 0 ]; do
  case "$1" in
    --dry-run)
      dry_run=1
      shift
      ;;
    --iterations)
      shift
      if [ "$#" -eq 0 ]; then
        printf '%s\n' "endurance: --iterations requires a value" >&2
        exit 2
      fi
      iterations="${1:-}"
      shift
      ;;
    --iterations=*)
      iterations="${1#--iterations=}"
      shift
      ;;
    --min-seconds)
      shift
      if [ "$#" -eq 0 ]; then
        printf '%s\n' "endurance: --min-seconds requires a value" >&2
        exit 2
      fi
      min_seconds="${1:-}"
      shift
      ;;
    --min-seconds=*)
      min_seconds="${1#--min-seconds=}"
      shift
      ;;
    --output-dir)
      shift
      if [ "$#" -eq 0 ]; then
        printf '%s\n' "endurance: --output-dir requires a value" >&2
        exit 2
      fi
      output_dir="${1:-}"
      shift
      ;;
    --output-dir=*)
      output_dir="${1#--output-dir=}"
      shift
      ;;
    --help|-h)
      usage
      exit 0
      ;;
    *)
      printf '%s\n' "endurance: unknown argument: $1" >&2
      usage >&2
      exit 2
      ;;
  esac
done

case "$iterations" in
  ''|*[!0-9]*|0)
    printf '%s\n' "endurance: --iterations must be a positive integer" >&2
    exit 2
    ;;
esac

case "$min_seconds" in
  ''|*[!0-9]*)
    printf '%s\n' "endurance: --min-seconds must be a non-negative integer" >&2
    exit 2
    ;;
esac

case "$output_dir" in
  /*) ;;
  *) output_dir="$invocation_dir/$output_dir" ;;
esac

mode="live"
if [ "$dry_run" -eq 1 ]; then
  mode="dry-run"
fi

mkdir -p "$output_dir"
rm -rf "$output_dir"/index.md "$output_dir"/run-*
index="$output_dir/index.md"
generated_at=$(date -u '+%Y-%m-%dT%H:%M:%SZ')

display_path() {
  case "$output_dir" in
    "$root"/*) printf '%s' "${output_dir#"$root"/}" ;;
    *) printf '%s' "$output_dir" ;;
  esac
}

run_slug() {
  printf 'run-%02d' "$1"
}

now_seconds() {
  date -u '+%s'
}

write_command_files() {
  command_dir="$1"
  shift
  : >"$command_dir/command.argv"
  : >"$command_dir/command.txt"
  for arg in "$@"; do
    printf '%s\n' "$arg" >>"$command_dir/command.argv"
    printf " '%s'" "$arg" >>"$command_dir/command.txt"
  done
  printf '\n' >>"$command_dir/command.txt"
}

run_capture() {
  command_dir="$1"
  shift
  mkdir -p "$command_dir"
  write_command_files "$command_dir" "$@"
  started=$(now_seconds)
  set +e
  "$@" >"$command_dir/stdout.txt" 2>"$command_dir/stderr.txt"
  command_status=$?
  set -e
  ended=$(now_seconds)
  printf '%s\n' "$command_status" >"$command_dir/exit-code.txt"
  printf '%s\n' "$((ended - started))" >"$command_dir/duration-seconds.txt"
  return "$command_status"
}

write_index_header() {
  {
    printf '%s\n' "# Endurance Evidence"
    printf '\n'
    printf '%s\n' "- Generated: $generated_at"
    printf '%s\n' "- Mode: $mode"
    printf '%s\n' "- Evidence root: $(display_path)"
    printf '%s\n' "- Planned iterations: $iterations"
    printf '%s\n' "- Minimum seconds: $min_seconds"
    printf '\n'
    printf '%s\n' "## Run Results"
    printf '\n'
    printf '%s\n' "| Run | Status | Duration seconds | Evidence |"
    printf '%s\n' "| --- | --- | ---: | --- |"
  } >"$index"
}

append_result() {
  printf '| %s | %s | %s | %s |\n' "$1" "$2" "$3" "$4" >>"$index"
}

write_plan() {
  run_dir="$1"
  mkdir -p "$run_dir"
  {
    printf '%s\n' "# Endurance Iteration Plan"
    printf '\n'
    printf '%s\n' '1. Build `bin/ceo-packet`.'
    printf '%s\n' "2. Run deterministic benchmark fixture scoring."
    printf '%s\n' "3. Run the cross-language CEO gauntlet."
    printf '%s\n' "4. Run one real-repo dogfood pass."
  } >"$run_dir/plan.md"
}

run_iteration() {
  attempt="$1"
  run_name=$(run_slug "$attempt")
  run_dir="$output_dir/$run_name"
  mkdir -p "$run_dir"
  started=$(now_seconds)

  if [ "$dry_run" -eq 1 ]; then
    write_plan "$run_dir"
    append_result "$run_name" "planned" "0" "$run_name/plan.md"
    return 0
  fi

  status="pass"
  if ! run_capture "$run_dir/build" go build -trimpath -o bin/ceo-packet ./cmd/ceo-packet; then
    status="fail"
  fi
  if [ "$status" = "pass" ] && ! run_capture "$run_dir/benchmark-fixtures-command" go run ./cmd/ceo-eval --benchmark-fixtures --tasks evals/tasks --output-dir "$run_dir/benchmark-fixtures"; then
    status="fail"
  fi
  if [ "$status" = "pass" ] && ! run_capture "$run_dir/cross-language-command" go run ./cmd/ceo-packet gauntlet --suite cross-language-core --agents ceo_harness --ceo-binary ./bin/ceo-packet --tasks evals/tasks --output-dir "$run_dir/cross-language-core" --timeout-seconds 120 --concurrency 2 --ceo-benchmark-mode model-command --ceo-benchmark-model-command-json "[\"sh\",\"$root/scripts/benchmark-model-command.sh\"]"; then
    status="fail"
  fi
  if [ "$status" = "pass" ] && ! run_capture "$run_dir/dogfood-real-command" sh scripts/dogfood-real.sh --repo "ceo-harness:$root" --repeat 1 --timeout-ms 250 --output-dir "$run_dir/dogfood-real"; then
    status="fail"
  fi

  ended=$(now_seconds)
  duration=$((ended - started))
  {
    printf '%s\n' "# Endurance Iteration $run_name"
    printf '\n'
    printf '%s\n' "- Status: $status"
    printf '%s\n' "- Duration seconds: $duration"
    printf '%s\n' "- Benchmark fixtures: benchmark-fixtures/summary.json"
    printf '%s\n' "- Cross-language: cross-language-core/summary.json"
    printf '%s\n' "- Dogfood: dogfood-real/index.md"
  } >"$run_dir/summary.md"
  append_result "$run_name" "$status" "$duration" "$run_name/summary.md"
  [ "$status" = "pass" ]
}

write_index_header
started_all=$(now_seconds)
attempt=1
overall="pass"
while :; do
  if ! run_iteration "$attempt"; then
    overall="fail"
    break
  fi
  elapsed=$(( $(now_seconds) - started_all ))
  if [ "$attempt" -ge "$iterations" ] && [ "$elapsed" -ge "$min_seconds" ]; then
    break
  fi
  attempt=$((attempt + 1))
done

elapsed_all=$(( $(now_seconds) - started_all ))
{
  printf '\n'
  printf '%s\n' "## Summary"
  printf '\n'
  printf '%s\n' "- Overall: $overall"
  printf '%s\n' "- Completed iterations: $attempt"
  printf '%s\n' "- Elapsed seconds: $elapsed_all"
} >>"$index"

printf '%s\n' "endurance: mode=$mode"
printf '%s\n' "endurance: evidence=$index"

if [ "$overall" != "pass" ]; then
  exit 1
fi
