#!/bin/sh
set -eu

root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
invocation_dir=$(pwd)
output_dir="$root/.omo/evidence/production-finalize"
evidence_root="$root/.omo/evidence"
dist="$root/dist"
provider_timeout_seconds=600
comparison_timeout_seconds=240
comparison_timeout_retries=1
comparison_result_retries=1
dry_run=0
run_comparison=0
skip_release_readiness=0
skip_provider_proofs=0
skip_competitor_smoke=0
skip_production_readiness=0

usage() {
  cat <<'USAGE'
Usage: sh scripts/production-finalize.sh [options]

Runs the guarded final production evidence sequence. This command does not
publish, push, tag, upload, create releases, or print provider secret values.

Options:
  --dry-run                         Write commands and summary without running them.
  --run-comparison                  Run the expensive 29-task all-agent comparison.
  --output-dir dir                  Evidence directory. Default: .omo/evidence/production-finalize
  --evidence-root dir               Canonical evidence root. Default: .omo/evidence
  --dist dir                        Release dist directory. Default: dist
  --provider-timeout-seconds n      Provider proof timeout. Default: 600
  --comparison-timeout-seconds n    All-agent comparison timeout. Default: 240
  --comparison-timeout-retries n    Retry timed-out all-agent comparison runs. Default: 1
  --comparison-result-retries n     Retry partial/failed all-agent comparison runs. Default: 1
  --skip-release-readiness          Skip release-readiness step.
  --skip-provider-proofs            Skip HTTP provider proof steps.
  --skip-competitor-smoke           Skip competitor smoke preflight.
  --skip-production-readiness       Skip final production-readiness aggregate.
  --help                            Show this help.
USAGE
}

while [ "$#" -gt 0 ]; do
  case "$1" in
    --dry-run)
      dry_run=1
      shift
      ;;
    --run-comparison)
      run_comparison=1
      shift
      ;;
    --output-dir)
      [ "$#" -ge 2 ] || {
        printf '%s\n' "production-finalize: --output-dir requires a value" >&2
        exit 2
      }
      output_dir="$2"
      shift 2
      ;;
    --evidence-root)
      [ "$#" -ge 2 ] || {
        printf '%s\n' "production-finalize: --evidence-root requires a value" >&2
        exit 2
      }
      evidence_root="$2"
      shift 2
      ;;
    --dist)
      [ "$#" -ge 2 ] || {
        printf '%s\n' "production-finalize: --dist requires a value" >&2
        exit 2
      }
      dist="$2"
      shift 2
      ;;
    --provider-timeout-seconds)
      [ "$#" -ge 2 ] || {
        printf '%s\n' "production-finalize: --provider-timeout-seconds requires a value" >&2
        exit 2
      }
      provider_timeout_seconds="$2"
      shift 2
      ;;
    --comparison-timeout-seconds)
      [ "$#" -ge 2 ] || {
        printf '%s\n' "production-finalize: --comparison-timeout-seconds requires a value" >&2
        exit 2
      }
      comparison_timeout_seconds="$2"
      shift 2
      ;;
    --comparison-timeout-retries)
      [ "$#" -ge 2 ] || {
        printf '%s\n' "production-finalize: --comparison-timeout-retries requires a value" >&2
        exit 2
      }
      comparison_timeout_retries="$2"
      shift 2
      ;;
    --comparison-result-retries)
      [ "$#" -ge 2 ] || {
        printf '%s\n' "production-finalize: --comparison-result-retries requires a value" >&2
        exit 2
      }
      comparison_result_retries="$2"
      shift 2
      ;;
    --skip-release-readiness)
      skip_release_readiness=1
      shift
      ;;
    --skip-provider-proofs)
      skip_provider_proofs=1
      shift
      ;;
    --skip-competitor-smoke)
      skip_competitor_smoke=1
      shift
      ;;
    --skip-production-readiness)
      skip_production_readiness=1
      shift
      ;;
    --help|-h)
      usage
      exit 0
      ;;
    *)
      printf '%s\n' "production-finalize: unknown argument: $1" >&2
      usage >&2
      exit 2
      ;;
  esac
done

case "$provider_timeout_seconds" in
  ''|*[!0-9]*|0)
    printf '%s\n' "production-finalize: --provider-timeout-seconds must be a positive integer" >&2
    exit 2
    ;;
esac

case "$comparison_timeout_seconds" in
  ''|*[!0-9]*|0)
    printf '%s\n' "production-finalize: --comparison-timeout-seconds must be a positive integer" >&2
    exit 2
    ;;
esac

case "$comparison_timeout_retries" in
  ''|*[!0-9]*)
    printf '%s\n' "production-finalize: --comparison-timeout-retries must be a non-negative integer" >&2
    exit 2
    ;;
esac

case "$comparison_result_retries" in
  ''|*[!0-9]*)
    printf '%s\n' "production-finalize: --comparison-result-retries must be a non-negative integer" >&2
    exit 2
    ;;
esac

abspath() {
  case "$1" in
    /*) printf '%s\n' "$1" ;;
    *) printf '%s\n' "$invocation_dir/$1" ;;
  esac
}

output_dir=$(abspath "$output_dir")
evidence_root=$(abspath "$evidence_root")
dist=$(abspath "$dist")

mkdir -p "$output_dir"
: >"$output_dir/steps.tsv"
: >"$output_dir/commands.sh"
chmod +x "$output_dir/commands.sh"

quote_command_arg() {
  case "$1" in
    '')
      printf "%s" "''"
      ;;
    *[!abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_./:=,+@%-]*)
      printf "'"
      printf "%s" "$1" | sed "s/'/'\\\\''/g"
      printf "'"
      ;;
    *)
      printf "%s" "$1"
      ;;
  esac
}

display_path() {
  case "$1" in
    "$root"/*) printf '%s\n' "${1#"$root"/}" ;;
    "$root") printf '%s\n' "." ;;
    *) printf '%s\n' "$1" ;;
  esac
}

if command -v ceo-packet >/dev/null 2>&1; then
  ceo_packet_cmd="ceo-packet"
  ceo_packet_words="ceo-packet"
else
  ceo_packet_cmd="go run ./cmd/ceo-packet"
  ceo_packet_words="go	run	./cmd/ceo-packet"
fi

quote_display_path() {
  path=$(display_path "$1")
  quote_command_arg "$path"
}

write_command_with_prefix() {
  prefix="$1"
  shift
  printf '%s' "$prefix" >>"$output_dir/commands.sh"
  first=1
  for arg in "$@"; do
    if [ "$first" -eq 1 ]; then
      first=0
    else
      printf ' ' >>"$output_dir/commands.sh"
    fi
    quote_command_arg "$arg" >>"$output_dir/commands.sh"
  done
  printf '\n' >>"$output_dir/commands.sh"
}

write_command() {
  write_command_with_prefix "" "$@"
}

write_blocked_command() {
  reason="$1"
  shift
  write_command_with_prefix "# blocked command: " "$@"
  printf '%s\n' "# reason: $reason" >>"$output_dir/commands.sh"
  write_command_with_prefix "# " "$@"
}

add_step() {
  name="$1"
  status="$2"
  evidence="$3"
  detail="$4"
  printf '%s\t%s\t%s\t%s\n' "$name" "$status" "$evidence" "$detail" >>"$output_dir/steps.tsv"
}

run_step() {
  name="$1"
  evidence_rel="$2"
  shift 2
  step_dir="$output_dir/$name"
  mkdir -p "$step_dir"
  if [ "$dry_run" -eq 1 ]; then
    write_command "$@"
    add_step "$name" "planned" "$evidence_rel" "Dry-run only"
    return 0
  fi
  set +e
  "$@" >"$step_dir/stdout.txt" 2>"$step_dir/stderr.txt"
  code=$?
  set -e
  printf '%s\n' "$code" >"$step_dir/exit-code.txt"
  if [ "$code" -eq 0 ]; then
    write_command "$@"
    add_step "$name" "pass" "$evidence_rel" "Command exited 0"
    return 0
  fi
  write_blocked_command "step $name exited $code; inspect $name/stdout.txt and $name/stderr.txt before rerunning." "$@"
  add_step "$name" "blocked" "$evidence_rel" "Command exited $code; see stdout/stderr"
  return 1
}

competitor_smoke_clean() {
  summary="$1"
  [ -f "$summary" ] || return 1
  python3 - "$summary" <<'PY'
import json
import sys

with open(sys.argv[1], "r", encoding="utf-8") as handle:
    data = json.load(handle)

failed = int(data.get("smoke_failed", 0) or 0)
setup_blocked = int(data.get("setup_blocked", 0) or 0)
passed = int(data.get("smoke_passed", 0) or 0)
competitors = int(data.get("competitors", 0) or 0)

if competitors > 0 and passed > 0 and failed == 0 and setup_blocked == 0:
    raise SystemExit(0)
raise SystemExit(1)
PY
}

comparison_summary_clean() {
  summary="$1"
  [ -f "$summary" ] || return 1
  python3 - "$summary" <<'PY'
import json
import sys

with open(sys.argv[1], "r", encoding="utf-8") as handle:
    data = json.load(handle)

run_count = int(data.get("run_count", 0) or 0)
ok = (
    run_count > 0
    and int(data.get("task_count", 0) or 0) >= 29
    and int(data.get("agent_count", 0) or 0) >= 4
    and int(data.get("passed", 0) or 0) >= run_count
    and int(data.get("partial", 0) or 0) == 0
    and int(data.get("failed", 0) or 0) == 0
    and int(data.get("timed_out", 0) or 0) == 0
    and int(data.get("setup_blocked", 0) or 0) == 0
    and int(data.get("skipped", 0) or 0) == 0
    and int(data.get("incomplete_evidence", 0) or 0) == 0
)
raise SystemExit(0 if ok else 1)
PY
}

{
  printf '%s\n' "#!/bin/sh"
  printf '%s\n' "set -eu"
  printf '\n'
  printf '%s\n' "# Generated by scripts/production-finalize.sh."
  printf '%s\n' "# Fill provider API key environment variables before running provider proof commands."
} >"$output_dir/commands.sh"

overall="pass"

release_readiness_output="$evidence_root/release-readiness-final"
if [ "$skip_release_readiness" -eq 1 ]; then
  add_step "release-readiness" "skipped" "not-run" "Skipped by flag"
else
  if ! run_step "release-readiness" "$release_readiness_output/index.md" sh "$root/scripts/release-readiness.sh" --dist "$dist" --output-dir "$release_readiness_output"; then
    overall="blocked"
  fi
fi

if [ "$skip_provider_proofs" -eq 1 ]; then
  add_step "provider-openrouter" "skipped" "not-run" "Skipped by flag"
  add_step "provider-kimi-code" "skipped" "not-run" "Skipped by flag"
  add_step "provider-minimax" "skipped" "not-run" "Skipped by flag"
else
  for provider in openrouter kimi-code minimax; do
    provider_output="$evidence_root/provider-proof-$provider"
    if ! run_step "provider-$provider" "$provider_output/index.md" sh "$root/scripts/provider-proof.sh" --provider "$provider" --output-dir "$provider_output" --timeout-seconds "$provider_timeout_seconds"; then
      overall="blocked"
    fi
  done
fi

competitor_setup_blocked=0
if [ "$skip_competitor_smoke" -eq 1 ]; then
  add_step "competitor-smoke" "skipped" "not-run" "Skipped by flag"
else
  competitor_smoke_summary="$output_dir/competitor-smoke/summary.json"
  if ! run_step "competitor-smoke-command" "competitor-smoke/summary.json" go run "$root/cmd/ceo-eval" --comparison-smoke --competitors "$root/evals/competitors.json" --output-dir "$output_dir/competitor-smoke" --timeout-seconds 25; then
    competitor_setup_blocked=1
    overall="blocked"
  elif competitor_smoke_clean "$competitor_smoke_summary"; then
    add_step "competitor-smoke" "pass" "competitor-smoke/summary.json" "Smoke summary has no failed or setup-blocked competitors"
  else
    competitor_setup_blocked=1
    add_step "competitor-smoke" "blocked" "competitor-smoke/summary.json" "Smoke summary has failed or setup-blocked competitors"
    overall="blocked"
  fi
fi

comparison_output="$evidence_root/external-agent-production-core-29-final-result-retry-r1"
if [ "$run_comparison" -eq 1 ]; then
  if [ "$competitor_setup_blocked" -eq 1 ]; then
    add_step "all-agent-29-comparison" "blocked" "commands.sh" "Waiting on clean competitor smoke before running all-agent comparison"
    overall="blocked"
  elif ! run_step "all-agent-29-comparison" "$comparison_output/summary.json" go run "$root/cmd/ceo-eval" \
    --local-agent-benchmark \
    --local-agents ceo_harness,codex_cli,claude_code,aider,opencode,goose,pi,oh_my_pi \
    --local-agent-benchmark-task production-core \
    --local-agent-benchmark-repeat 1 \
    --local-agent-benchmark-concurrency 4 \
    --local-agent-benchmark-timeout-retries "$comparison_timeout_retries" \
    --local-agent-benchmark-result-retries "$comparison_result_retries" \
    --ceo-binary "$root/bin/ceo-packet" \
    --tasks "$root/evals/tasks" \
    --output-dir "$comparison_output" \
    --timeout-seconds "$comparison_timeout_seconds" \
    --ceo-benchmark-mode model-command \
    --ceo-benchmark-model-command-json "[\"sh\",\"$root/scripts/benchmark-model-command.sh\"]"; then
    overall="blocked"
  fi
else
  write_command go run ./cmd/ceo-eval \
    --local-agent-benchmark \
    --local-agents ceo_harness,codex_cli,claude_code,aider,opencode,goose,pi,oh_my_pi \
    --local-agent-benchmark-task production-core \
    --local-agent-benchmark-repeat 1 \
    --local-agent-benchmark-concurrency 4 \
    --local-agent-benchmark-timeout-retries "$comparison_timeout_retries" \
    --local-agent-benchmark-result-retries "$comparison_result_retries" \
    --ceo-binary ./bin/ceo-packet \
    --tasks evals/tasks \
    --output-dir .omo/evidence/external-agent-production-core-29-final-result-retry-r1 \
    --timeout-seconds "$comparison_timeout_seconds" \
    --ceo-benchmark-mode model-command \
    --ceo-benchmark-model-command-json "[\"sh\",\"$root/scripts/benchmark-model-command.sh\"]"
  if comparison_summary_clean "$comparison_output/summary.json"; then
    add_step "all-agent-29-comparison" "pass" "$comparison_output/summary.json" "Existing clean all-agent comparison evidence found"
  else
    add_step "all-agent-29-comparison" "planned" "commands.sh" "Use --run-comparison to execute the expensive all-agent suite"
  fi
  if [ "$dry_run" -eq 0 ] && ! comparison_summary_clean "$comparison_output/summary.json"; then
    overall="blocked"
  fi
fi

production_readiness_output="$evidence_root/production-readiness-final"
if [ "$skip_production_readiness" -eq 1 ]; then
  add_step "production-readiness" "skipped" "not-run" "Skipped by flag"
else
  if ! run_step "production-readiness" "$production_readiness_output/index.md" sh "$root/scripts/production-readiness.sh" --dist "$dist" --output-dir "$production_readiness_output"; then
    overall="blocked"
  fi
fi

if [ "$dry_run" -eq 1 ]; then
  overall="planned"
fi

{
  printf '%s\n' "# Production Finalize Next Actions"
  printf '\n'
  printf '%s\n' "Status: $overall"
  printf '\n'
  action_count=0
  competitor_action_written=0
  while IFS='	' read -r name status evidence detail; do
    case "$status" in
      pass|skipped) continue ;;
    esac
    case "$name" in
      competitor-smoke|competitor-smoke-command)
        if [ "$competitor_action_written" -eq 1 ]; then
          continue
        fi
        competitor_action_written=1
        ;;
    esac
    action_count=$((action_count + 1))
    case "$name" in
      release-readiness)
        printf '%s\n' "- Publish and verify release evidence: set public release metadata, then rerun \`sh scripts/release-readiness.sh --dist $(quote_display_path "$dist") --output-dir $(quote_display_path "$evidence_root/release-readiness-final")\`. Evidence: \`$(display_path "$evidence")\`."
        ;;
      provider-openrouter)
        printf '%s\n' "- Prove OpenRouter HTTP provider: export \`OPENROUTER_API_KEY\`, run \`sh scripts/provider-setup-preflight.sh --providers openrouter --output-dir .omo/evidence/provider-setup-preflight-openrouter\`, then rerun \`sh scripts/provider-proof.sh --provider openrouter --output-dir $(quote_display_path "$evidence_root/provider-proof-openrouter") --timeout-seconds $provider_timeout_seconds\`. Evidence: \`$(display_path "$evidence")\`."
        ;;
      provider-kimi-code)
        printf '%s\n' "- Prove Kimi Code HTTP provider: export \`KIMI_CODE_API_KEY\`, run \`sh scripts/provider-setup-preflight.sh --providers kimi-code --output-dir .omo/evidence/provider-setup-preflight-kimi-code\`, then rerun \`sh scripts/provider-proof.sh --provider kimi-code --output-dir $(quote_display_path "$evidence_root/provider-proof-kimi-code") --timeout-seconds $provider_timeout_seconds\`. Evidence: \`$(display_path "$evidence")\`."
        ;;
      provider-minimax)
        printf '%s\n' "- Prove MiniMax HTTP provider: export \`MINIMAX_API_KEY\`, run \`sh scripts/provider-setup-preflight.sh --providers minimax --output-dir .omo/evidence/provider-setup-preflight-minimax\`, then rerun \`sh scripts/provider-proof.sh --provider minimax --output-dir $(quote_display_path "$evidence_root/provider-proof-minimax") --timeout-seconds $provider_timeout_seconds\`. Evidence: \`$(display_path "$evidence")\`."
        ;;
      competitor-smoke|competitor-smoke-command)
        printf '%s\n' "- Fix competitor setup before final comparison: inspect \`$(display_path "$output_dir/competitor-smoke/summary.json")\`, install missing binaries or fix provider auth/quota, then rerun \`$ceo_packet_cmd production-finalize --workspace . --dry-run\` or the full finalizer."
        ;;
      all-agent-29-comparison)
        printf '%s\n' "- Run the final all-agent 29-task comparison after setup is clean: \`$ceo_packet_cmd production-finalize --workspace . --run-comparison\`."
        ;;
      production-readiness)
        printf '%s\n' "- Re-run the final readiness aggregate after release, provider, smoke, and comparison proof are clean: \`sh scripts/production-readiness.sh --dist $(quote_display_path "$dist") --output-dir $(quote_display_path "$evidence_root/production-readiness-final")\`. Evidence: \`$(display_path "$evidence")\`."
        ;;
      *)
        printf '%s\n' "- Resolve \`$name\`: $detail. Evidence: \`$evidence\`."
        ;;
    esac
  done <"$output_dir/steps.tsv"
  if [ "$action_count" -eq 0 ]; then
    printf '%s\n' "- No next actions remain."
  fi
} >"$output_dir/next-actions.md"

{
  printf '%s\n' "# Production Setup Actions"
  printf '\n'
  printf '%s\n' "Use this as the single checklist before claiming public production readiness."
  printf '\n'
  if [ -f "$evidence_root/release-readiness-final/setup-actions.md" ]; then
    printf '%s\n' "## Release"
    printf '\n'
    printf '%s\n' "- Follow \`$(display_path "$evidence_root/release-readiness-final/setup-actions.md")\`."
    printf '\n'
  fi
  printf '%s\n' "## Providers"
  printf '\n'
  for provider in openrouter kimi-code minimax; do
    checklist="$evidence_root/provider-proof-$provider/setup-checklist.md"
    commands_file="$evidence_root/provider-proof-$provider/commands.sh"
    if [ -f "$checklist" ]; then
      printf '%s\n' "- $provider: follow \`$(display_path "$checklist")\` and rerun \`$(display_path "$commands_file")\` after the required env var is set."
    else
      printf '%s\n' "- $provider: run \`sh scripts/provider-proof.sh --provider $provider --output-dir $(quote_display_path "$evidence_root/provider-proof-$provider") --timeout-seconds $provider_timeout_seconds\`."
    fi
  done
  printf '\n'
  if [ -f "$output_dir/competitor-smoke/setup-actions.md" ]; then
    printf '%s\n' "## Competitors"
    printf '\n'
    printf '%s\n' "- Follow \`$(display_path "$output_dir/competitor-smoke/setup-actions.md")\`."
    printf '\n'
  fi
  printf '%s\n' "## Final Rerun"
  printf '\n'
  printf '%s\n' '```sh'
  printf '%s\n' "$ceo_packet_cmd production-finalize --workspace . --dry-run"
  printf '%s\n' "$ceo_packet_cmd production-finalize --workspace . --run-comparison"
  printf '%s\n' 'sh scripts/production-readiness.sh --dist dist --output-dir .omo/evidence/production-readiness-final'
  printf '%s\n' '```'
} >"$output_dir/setup-actions.md"

python3 - "$output_dir/steps.tsv" "$output_dir/summary.json" "$overall" "$output_dir/next-actions.md" "$output_dir/next-actions.json" "$provider_timeout_seconds" "$output_dir/setup-actions.md" "$ceo_packet_words" <<'PY'
import json
import sys
import pathlib
import hashlib

steps = []
with open(sys.argv[1], "r", encoding="utf-8") as handle:
    for line in handle:
        name, status, evidence, detail = line.rstrip("\n").split("\t", 3)
        steps.append({
            "name": name,
            "status": status,
            "evidence": evidence,
            "detail": detail,
        })

action_lines = [
    line[2:] for line in pathlib.Path(sys.argv[4]).read_text(encoding="utf-8").splitlines()
    if line.startswith("- ") and line != "- No next actions remain."
]
actions = []
action_text_by_step = {}
nonpass_steps = [step for step in steps if step["status"] not in {"pass", "skipped"}]
competitor_step_present = any(step["name"] == "competitor-smoke" for step in nonpass_steps)
action_steps = []
for step in nonpass_steps:
    if step["name"] == "competitor-smoke-command" and competitor_step_present:
        continue
    action_steps.append(step)
for step, line in zip(action_steps, action_lines):
    action_text_by_step[step["name"]] = line

provider_timeout = sys.argv[6]
next_actions_dir = pathlib.Path(sys.argv[5]).resolve().parent
ceo_packet_prefix = sys.argv[8].split("\t")

def action_output_dir(step):
    path = pathlib.Path(step["evidence"])
    if not path.is_absolute():
        path = next_actions_dir / path
    return str(path.parent)

def evidence_metadata(action):
    files = []
    for field in ("evidence", "inspect"):
        value = action.get(field)
        if not value:
            continue
        path = pathlib.Path(value)
        if not path.is_absolute():
            path = next_actions_dir / path
        entry = {
            "field": field,
            "path": str(path),
        }
        try:
            content = path.read_bytes()
        except OSError as error:
            entry["exists"] = False
            entry["error"] = str(error)
        else:
            entry["exists"] = True
            entry["size_bytes"] = len(content)
            entry["sha256"] = hashlib.sha256(content).hexdigest()
        files.append(entry)
    return files

def action_for_step(step):
    name = step["name"]
    action = {
        "id": name,
        "status": step["status"],
        "text": action_text_by_step.get(name, step["detail"]),
        "evidence": step["evidence"],
    }
    if name == "release-readiness":
        action["kind"] = "release_proof"
        action["command"] = ["sh", "scripts/release-readiness.sh", "--dist", "dist", "--output-dir", action_output_dir(step)]
    elif name == "provider-openrouter":
        action["kind"] = "provider_proof"
        action["provider"] = "openrouter"
        action["required_env"] = "OPENROUTER_API_KEY"
        action["command"] = ["sh", "scripts/provider-proof.sh", "--provider", "openrouter", "--output-dir", action_output_dir(step), "--timeout-seconds", provider_timeout]
    elif name == "provider-kimi-code":
        action["kind"] = "provider_proof"
        action["provider"] = "kimi-code"
        action["required_env"] = "KIMI_CODE_API_KEY"
        action["command"] = ["sh", "scripts/provider-proof.sh", "--provider", "kimi-code", "--output-dir", action_output_dir(step), "--timeout-seconds", provider_timeout]
    elif name == "provider-minimax":
        action["kind"] = "provider_proof"
        action["provider"] = "minimax"
        action["required_env"] = "MINIMAX_API_KEY"
        action["command"] = ["sh", "scripts/provider-proof.sh", "--provider", "minimax", "--output-dir", action_output_dir(step), "--timeout-seconds", provider_timeout]
    elif name in {"competitor-smoke", "competitor-smoke-command"}:
        action["kind"] = "competitor_setup"
        action["inspect"] = "competitor-smoke/summary.json"
        action["command"] = ceo_packet_prefix + ["production-finalize", "--workspace", ".", "--dry-run"]
    elif name == "all-agent-29-comparison":
        action["kind"] = "comparison"
        action["command"] = ceo_packet_prefix + ["production-finalize", "--workspace", ".", "--run-comparison"]
    elif name == "production-readiness":
        action["kind"] = "final_readiness"
        action["command"] = ["sh", "scripts/production-readiness.sh", "--dist", "dist", "--output-dir", action_output_dir(step)]
    else:
        action["kind"] = "manual"
    files = evidence_metadata(action)
    if files:
        action["declared_evidence_files"] = files
    return action

for step in action_steps:
    actions.append(action_for_step(step))

with open(sys.argv[5], "w", encoding="utf-8") as handle:
    json.dump({
        "schema_version": 1,
        "status": sys.argv[3],
        "required_action_count": len(actions),
        "actions": actions,
    }, handle, indent=2)
    handle.write("\n")

summary = {
    "schema_version": 1,
    "status": sys.argv[3],
    "step_count": len(steps),
    "blocked_steps": [step["name"] for step in steps if step["status"] == "blocked"],
    "planned_steps": [step["name"] for step in steps if step["status"] == "planned"],
    "skipped_steps": [step["name"] for step in steps if step["status"] == "skipped"],
    "secret_value_saved": False,
    "publish_actions_performed": False,
    "next_actions": {
        "path": "next-actions.md",
        "json_path": "next-actions.json",
        "required_action_count": len(actions),
    },
    "setup_actions": {
        "path": "setup-actions.md",
        "required_action_count": sum(
            1 for line in pathlib.Path(sys.argv[7]).read_text(encoding="utf-8").splitlines()
            if line.startswith("- ")
        ),
        "sha256": hashlib.sha256(pathlib.Path(sys.argv[7]).read_bytes()).hexdigest(),
    },
    "steps": steps,
}
with open(sys.argv[2], "w", encoding="utf-8") as handle:
    json.dump(summary, handle, indent=2)
    handle.write("\n")
PY

{
  printf '%s\n' "# Production Finalize Evidence"
  printf '\n'
  printf '%s\n' "Status: $overall"
  printf '%s\n' "Publishes or tags: false"
  printf '%s\n' "Secret values saved: false"
  printf '\n'
  printf '%s\n' "| Step | Status | Evidence | Detail |"
  printf '%s\n' "| --- | --- | --- | --- |"
  while IFS='	' read -r name status evidence detail; do
    printf '| %s | %s | `%s` | %s |\n' "$name" "$status" "$evidence" "$detail"
  done <"$output_dir/steps.tsv"
  printf '\n'
  printf '%s\n' "## Next Actions"
  printf '\n'
  printf '%s\n' "Open \`next-actions.md\` for the exact remaining commands and setup steps."
  printf '%s\n' "Open \`setup-actions.md\` for one consolidated public-readiness checklist."
  printf '\n'
  printf '%s\n' "## Commands"
  printf '\n'
  printf '%s\n' "Replay or inspect the generated command list in \`commands.sh\`."
} >"$output_dir/index.md"

printf '%s\n' "production-finalize: wrote $output_dir/index.md"
printf '%s\n' "production-finalize: $overall"

case "$overall" in
  pass|planned) exit 0 ;;
  *) exit 1 ;;
esac
