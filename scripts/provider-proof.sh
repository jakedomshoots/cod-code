#!/bin/sh
set -eu

root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
invocation_dir=$(pwd)
cd "$root"

dry_run=0
provider="kimi"
timeout_seconds=600
output_dir=""
http_model=""
api_key_env=""

usage() {
  cat <<'USAGE'
Usage: sh scripts/provider-proof.sh [--dry-run] [--provider kimi|codex|openai|openrouter|moonshot] [--timeout-seconds n] [--output-dir path]

Runs real-provider benchmark proofs and writes durable evidence.

Options:
  --dry-run            Write the provider proof plan without running commands.
  --provider name      Provider bridge to use. Supported: kimi, codex, openai,
                       openrouter, moonshot.
  --http-model name    Override the default HTTP provider model.
  --api-key-env name   Override the default HTTP provider API key env var.
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
    --http-model)
      shift
      if [ "$#" -eq 0 ]; then
        printf '%s\n' "provider-proof: --http-model requires a value" >&2
        exit 2
      fi
      http_model="${1:-}"
      shift
      ;;
    --http-model=*)
      http_model="${1#--http-model=}"
      shift
      ;;
    --api-key-env)
      shift
      if [ "$#" -eq 0 ]; then
        printf '%s\n' "provider-proof: --api-key-env requires a value" >&2
        exit 2
      fi
      api_key_env="${1:-}"
      shift
      ;;
    --api-key-env=*)
      api_key_env="${1#--api-key-env=}"
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

case "$provider" in
  kimi)
    provider_mode="cli-model-command"
    provider_cli="kimi"
    model_command_script="$root/scripts/kimi-model-command.sh"
    model_command_display="scripts/kimi-model-command.sh"
    ;;
  codex)
    provider_mode="cli-model-command"
    provider_cli="codex"
    model_command_script="$root/scripts/codex-model-command.sh"
    model_command_display="scripts/codex-model-command.sh"
    ;;
  openai)
    provider_mode="http-provider"
    http_preset="openai"
    default_http_model="${CEO_PROVIDER_PROOF_OPENAI_MODEL:-gpt-5}"
    default_api_key_env="OPENAI_API_KEY"
    ;;
  openrouter)
    provider_mode="http-provider"
    http_preset="openrouter"
    default_http_model="${CEO_PROVIDER_PROOF_OPENROUTER_MODEL:-openai/gpt-5-mini}"
    default_api_key_env="OPENROUTER_API_KEY"
    ;;
  moonshot)
    provider_mode="http-provider"
    http_preset="moonshot"
    default_http_model="${CEO_PROVIDER_PROOF_MOONSHOT_MODEL:-moonshot-v1-128k}"
    default_api_key_env="MOONSHOT_API_KEY"
    ;;
  *)
    printf '%s\n' "provider-proof: unsupported provider: $provider" >&2
    exit 2
    ;;
esac

if [ "$provider_mode" = "http-provider" ]; then
  if [ -z "$http_model" ]; then
    http_model="$default_http_model"
  fi
  if [ -z "$api_key_env" ]; then
    api_key_env="$default_api_key_env"
  fi
  case "$api_key_env" in
    ''|*[!A-Za-z0-9_]*|[0-9]*)
      printf '%s\n' "provider-proof: --api-key-env must be a valid environment variable name" >&2
      exit 2
      ;;
  esac
fi

if [ -z "$output_dir" ]; then
  output_dir="$root/.omo/evidence/provider-proof-$provider"
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
rm -rf "$output_dir"/index.md "$output_dir"/blocked.md "$output_dir"/build "$output_dir"/cross-language-js-state-reducer "$output_dir"/cross-language-python-retry-policy
index="$output_dir/index.md"
generated_at=$(date -u '+%Y-%m-%dT%H:%M:%SZ')
if [ "$provider_mode" = "cli-model-command" ]; then
  model_command_json=$(printf '["sh","%s"]' "$model_command_script")
fi

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
    printf '%s\n' "- Provider mode: $provider_mode"
    printf '%s\n' "- Evidence root: $(display_path)"
    if [ "$provider_mode" = "cli-model-command" ]; then
      printf '%s\n' "- Model command: $model_command_display"
    else
      printf '%s\n' "- HTTP preset: $http_preset"
      printf '%s\n' "- HTTP model: $http_model"
      printf '%s\n' "- API key env: $api_key_env"
    fi
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
    printf '%s\n' '1. Build `bin/ceo-packet`.'
    printf '%s\n' "2. Run \`ceo-eval --local-agent-benchmark\` for \`$task_id\`."
    if [ "$provider_mode" = "cli-model-command" ]; then
      printf '%s\n' "3. Route CEO Harness subagent and CEO review through \`$model_command_display\`."
    else
      printf '%s\n' "3. Route CEO Harness subagent and CEO review through \`$http_preset\` HTTP provider model \`$http_model\`."
      printf '%s\n' "4. Require \`$api_key_env\` to be present without printing the key value."
    fi
    printf '%s\n' "5. Save command output, score JSON, report JSON, diff, and changed-files evidence."
  } >"$task_dir/plan.md"
  append_result "$task_id" "planned" "$task_id/plan.md"
}

write_http_setup_blocked() {
  blocked_reason="${1:-missing_api_key_env}"
  if [ "$blocked_reason" = "empty_api_key_env" ]; then
    blocked_message="Provider \`$provider\` has \`$api_key_env\` set, but it is empty."
    setup_action="Fill \`$api_key_env\` with a non-empty value in the shell or local secret manager."
    result_status="blocked_empty_key"
  else
    blocked_message="Provider \`$provider\` requires \`$api_key_env\` for HTTP benchmark mode."
    setup_action="Export \`$api_key_env\` in the shell or local secret manager."
    result_status="blocked_missing_key"
  fi
  {
    printf '%s\n' "# Provider Proof Blocked"
    printf '\n'
    printf '%s\n' "$blocked_message"
    printf '%s\n' "Set the environment variable, then rerun this command. The key value is not printed or saved."
    printf '\n'
    printf '%s\n' "## Next Command"
    printf '\n'
    printf '%s\n' "\`\`\`sh"
    printf '%s\n' "# Export $api_key_env in your shell or local secret manager first."
    printf '%s\n' "sh scripts/provider-proof.sh --provider $provider --output-dir .omo/evidence/provider-proof-$provider --timeout-seconds $timeout_seconds"
    printf '%s\n' "\`\`\`"
  } >"$output_dir/blocked.md"

  {
    printf '%s\n' "$api_key_env="
    printf '%s\n' "CEO_PROVIDER_PROOF_$(printf '%s' "$provider" | tr '[:lower:]' '[:upper:]')_MODEL=$http_model"
  } >"$output_dir/env.template"

  {
    printf '%s\n' "#!/bin/sh"
    printf '%s\n' "set -eu"
    printf '\n'
    printf '%s\n' "# Export $api_key_env in your shell or local secret manager before running."
    printf '%s\n' "# Do not paste secret values into this file or any evidence artifact."
    printf '%s\n' "if [ -z \"\${$api_key_env+x}\" ]; then"
    printf '%s\n' "  printf '%s\\n' 'provider setup: $api_key_env is not set' >&2"
    printf '%s\n' "  exit 2"
    printf '%s\n' "fi"
    printf '%s\n' "if [ -z \"\${$api_key_env}\" ]; then"
    printf '%s\n' "  printf '%s\\n' 'provider setup: $api_key_env is empty' >&2"
    printf '%s\n' "  exit 2"
    printf '%s\n' "fi"
    printf '%s\n' "sh scripts/provider-proof.sh --provider $provider --output-dir .omo/evidence/provider-proof-$provider --timeout-seconds $timeout_seconds"
    printf '%s\n' "sh scripts/production-readiness.sh --dist dist --output-dir .omo/evidence/production-readiness"
  } >"$output_dir/commands.sh"
  chmod +x "$output_dir/commands.sh"

  {
    printf '%s\n' "# Provider Setup Checklist"
    printf '\n'
    printf '%s\n' "1. $setup_action"
    printf '%s\n' "2. Keep the key out of git, logs, reports, and evidence folders."
    printf '%s\n' "3. Run \`commands.sh\` from the repo root."
    printf '%s\n' "4. Confirm \`index.md\` says \`- Overall: pass\`."
    printf '%s\n' "5. Re-run production readiness."
  } >"$output_dir/setup-checklist.md"

  setup_checklist_item_count=$(awk '/^[0-9]+[.]/ { count += 1 } END { print count + 0 }' "$output_dir/setup-checklist.md")
  setup_artifacts_sha256=$(python3 - "$output_dir/blocked.md" "$output_dir/env.template" "$output_dir/commands.sh" "$output_dir/setup-checklist.md" <<'PY'
import hashlib
import json
import pathlib
import sys

result = {}
for raw_path in sys.argv[1:]:
    path = pathlib.Path(raw_path)
    result[path.name] = hashlib.sha256(path.read_bytes()).hexdigest()
print(json.dumps(result, sort_keys=True))
PY
)

  cat >"$output_dir/summary.json" <<JSON
{
  "schema_version": 1,
  "status": "blocked",
  "provider": "$provider",
  "provider_mode": "$provider_mode",
  "http_preset": "$http_preset",
  "http_model": "$http_model",
  "api_key_env": "$api_key_env",
  "blocked_reason": "$blocked_reason",
  "setup_result_status": "$result_status",
  "setup_checklist_item_count": $setup_checklist_item_count,
  "setup_artifacts_sha256": $setup_artifacts_sha256,
  "command_script_secret_policy": "no_secret_assignment",
  "secret_value_saved": false,
  "artifacts": {
    "index": "index.md",
    "blocked": "blocked.md",
    "env_template": "env.template",
    "commands": "commands.sh",
    "checklist": "setup-checklist.md"
  }
}
JSON
}

run_task() {
  task_id="$1"
  task_dir="$output_dir/$task_id"
  if [ "$provider_mode" = "cli-model-command" ]; then
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
      task_command_passed=1
    else
      task_command_passed=0
    fi
  else
    if run_capture "$task_dir/command" go run ./cmd/ceo-eval \
      --local-agent-benchmark \
      --local-agents ceo_harness \
      --local-agent-benchmark-task "$task_id" \
      --local-agent-benchmark-repeat 1 \
      --ceo-binary ./bin/ceo-packet \
      --tasks evals/tasks \
      --output-dir "$task_dir/run" \
      --timeout-seconds "$timeout_seconds" \
      --ceo-benchmark-mode http-provider \
      --ceo-benchmark-provider-preset "$http_preset" \
      --ceo-benchmark-provider-model "$http_model" \
      --ceo-benchmark-provider-api-key-env "$api_key_env"; then
      task_command_passed=1
    else
      task_command_passed=0
    fi
  fi
  if [ "$task_command_passed" -eq 1 ]; then
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
  if [ "$provider_mode" = "cli-model-command" ]; then
    if ! command -v "$provider_cli" >/dev/null 2>&1; then
      printf '%s\n' "provider-proof: $provider_cli CLI not found on PATH" >&2
      exit 1
    fi
  else
    eval "api_key_present=\${$api_key_env+x}"
    eval "configured_api_key=\${$api_key_env:-}"
    if [ -z "$configured_api_key" ]; then
      if [ "$api_key_present" = "x" ]; then
        write_http_setup_blocked "empty_api_key_env"
      else
        write_http_setup_blocked "missing_api_key_env"
      fi
      append_result "provider_setup" "$result_status" "blocked.md"
      overall="blocked"
    fi
  fi
  if [ "${overall:-}" != "blocked" ]; then
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
fi

{
  printf '\n'
  printf '%s\n' "## Summary"
  printf '\n'
  printf '%s\n' "- Overall: $overall"
} >>"$index"

printf '%s\n' "provider-proof: mode=$mode"
printf '%s\n' "provider-proof: evidence=$index"

if [ "$overall" = "fail" ] || [ "$overall" = "blocked" ]; then
  exit 1
fi
