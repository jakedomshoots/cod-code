#!/bin/sh
set -eu

root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
invocation_dir=$(pwd)
cd "$root"

dry_run=0
provider="kimi"
timeout_seconds=600
output_dir="$root/.omo/evidence/provider-proof-kimi"

usage() {
  cat <<'USAGE'
Usage: sh scripts/provider-proof.sh [--dry-run] [--provider kimi] [--timeout-seconds n] [--output-dir path]

Runs real-provider benchmark proofs and writes durable evidence.

Options:
  --dry-run            Write the provider proof plan without running commands.
  --provider name      Provider bridge to use. Currently supported: kimi.
  --timeout-seconds n  Timeout for each benchmark command. Default: 600.
  --output-dir path    Evidence directory. Default: .omo/evidence/provider-proof-kimi.
  --help               Show this help.
USAGE
}

while [ "$#" -gt 0 ]; do
  case "$1" in
    --dry-run)
      dry_run=1
      shift
      ;;
    --provider)
      shift
      if [ "$#" -eq 0 ]; then
        printf '%s\n' "provider-proof: --provider requires a value" >&2
        exit 2
      fi
      provider="${1:-}"
      shift
      ;;
    --provider=*)
      provider="${1#--provider=}"
      shift
      ;;
    --timeout-seconds)
      shift
      if [ "$#" -eq 0 ]; then
        printf '%s\n' "provider-proof: --timeout-seconds requires a value" >&2
        exit 2
      fi
      timeout_seconds="${1:-}"
      shift
      ;;
    --timeout-seconds=*)
      timeout_seconds="${1#--timeout-seconds=}"
      shift
      ;;
    --output-dir)
      shift
      if [ "$#" -eq 0 ]; then
        printf '%s\n' "provider-proof: --output-dir requires a value" >&2
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
      printf '%s\n' "provider-proof: unknown argument: $1" >&2
      usage >&2
      exit 2
      ;;
  esac
done

case "$timeout_seconds" in
  ''|*[!0-9]*|0)
    printf '%s\n' "provider-proof: --timeout-seconds must be a positive integer" >&2
    exit 2
    ;;
esac

if [ "$provider" != "kimi" ]; then
  printf '%s\n' "provider-proof: unsupported provider: $provider" >&2
  exit 2
fi

case "$output_dir" in
  /*) ;;
  *) output_dir="$invocation_dir/$output_dir" ;;
esac

mode="live"
if [ "$dry_run" -eq 1 ]; then
  mode="dry-run"
fi

mkdir -p "$output_dir"
rm -rf "$output_dir"/index.md "$output_dir"/build "$output_dir"/cross-language-js-state-reducer "$output_dir"/cross-language-python-retry-policy
index="$output_dir/index.md"
generated_at=$(date -u '+%Y-%m-%dT%H:%M:%SZ')
model_command_json=$(printf '["sh","%s/scripts/kimi-model-command.sh"]' "$root")

display_path() {
  case "$output_dir" in
    "$root"/*) printf '%s' "${output_dir#"$root"/}" ;;
    *) printf '%s' "$output_dir" ;;
  esac
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
    printf '%s\n' "# Provider Proof Evidence"
    printf '\n'
    printf '%s\n' "- Generated: $generated_at"
    printf '%s\n' "- Mode: $mode"
    printf '%s\n' "- Provider: $provider"
    printf '%s\n' "- Evidence root: $(display_path)"
    printf '%s\n' "- Model command: scripts/kimi-model-command.sh"
    printf '%s\n' "- Timeout seconds: $timeout_seconds"
    printf '\n'
    printf '%s\n' "## Task Results"
    printf '\n'
    printf '%s\n' "| Task | Status | Evidence |"
    printf '%s\n' "| --- | --- | --- |"
  } >"$index"
}

append_result() {
  printf '| %s | %s | %s |\n' "$1" "$2" "$3" >>"$index"
}

write_plan() {
  task_id="$1"
  task_dir="$output_dir/$task_id"
  mkdir -p "$task_dir"
  {
    printf '%s\n' "# Provider Proof Plan: $task_id"
    printf '\n'
    printf '%s\n' "1. Build `bin/ceo-packet`."
    printf '%s\n' "2. Run `ceo-eval --local-agent-benchmark` for `$task_id`."
    printf '%s\n' "3. Route CEO Harness subagent and CEO review through `scripts/kimi-model-command.sh`."
    printf '%s\n' "4. Save command output, score JSON, report JSON, diff, and changed-files evidence."
  } >"$task_dir/plan.md"
  append_result "$task_id" "planned" "$task_id/plan.md"
}

run_task() {
  task_id="$1"
  task_dir="$output_dir/$task_id"
  if run_capture "$task_dir/command" go run ./cmd/ceo-eval \
    --local-agent-benchmark \
    --local-agents ceo_harness \
    --local-agent-benchmark-task "$task_id" \
    --local-agent-benchmark-repeat 1 \
    --ceo-binary ./bin/ceo-packet \
    --tasks evals/tasks \
    --output-dir "$task_dir/run" \
    --timeout-seconds "$timeout_seconds" \
    --ceo-benchmark-mode model-command \
    --ceo-benchmark-model-command-json "$model_command_json"; then
    if python3 - "$task_dir/run/summary.json" <<'PY'
import json
import sys

with open(sys.argv[1], "r", encoding="utf-8") as handle:
    summary = json.load(handle)

passed = int(summary.get("passed", 0) or 0)
partial = int(summary.get("partial", 0) or 0)
failed = int(summary.get("failed", 0) or 0)
timed_out = int(summary.get("timed_out", 0) or 0)
incomplete = int(summary.get("incomplete_evidence", 0) or 0)

if passed > 0 and partial == 0 and failed == 0 and timed_out == 0 and incomplete == 0:
    raise SystemExit(0)
raise SystemExit(1)
PY
    then
      append_result "$task_id" "pass" "$task_id/run/summary.json"
      return 0
    fi
  fi
  append_result "$task_id" "fail" "$task_id/run/summary.json"
  return 1
}

write_index_header

if [ "$dry_run" -eq 1 ]; then
  write_plan "cross-language-js-state-reducer"
  write_plan "cross-language-python-retry-policy"
  overall="planned"
else
  if ! command -v kimi >/dev/null 2>&1; then
    printf '%s\n' "provider-proof: kimi CLI not found on PATH" >&2
    exit 1
  fi
  overall="pass"
  if ! run_capture "$output_dir/build" go build -trimpath -o bin/ceo-packet ./cmd/ceo-packet; then
    append_result "build" "fail" "build/stderr.txt"
    overall="fail"
  fi
  if [ "$overall" = "pass" ] && ! run_task "cross-language-js-state-reducer"; then
    overall="fail"
  fi
  if [ "$overall" = "pass" ] && ! run_task "cross-language-python-retry-policy"; then
    overall="fail"
  fi
fi

{
  printf '\n'
  printf '%s\n' "## Summary"
  printf '\n'
  printf '%s\n' "- Overall: $overall"
} >>"$index"

printf '%s\n' "provider-proof: mode=$mode"
printf '%s\n' "provider-proof: evidence=$index"

if [ "$overall" = "fail" ]; then
  exit 1
fi
