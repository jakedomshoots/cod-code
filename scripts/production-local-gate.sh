#!/bin/sh
set -eu

root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
dist=${DIST:-"$root/dist"}
output_dir=${OUTPUT_DIR:-"$root/.omo/evidence/production-local-gate"}
evidence_root=${EVIDENCE_ROOT:-"$root/.omo/evidence"}
workspace=${WORKSPACE:-"$root"}

usage() {
  cat <<'USAGE'
Usage: sh scripts/production-local-gate.sh [--workspace dir] [--dist dir] [--output-dir dir] [--evidence-root dir]

Runs production-readiness and fails only when local production readiness is not
green. Public release/provider/comparison blockers are allowed here because CI
does not have release credentials or paid provider keys.
USAGE
}

while [ "$#" -gt 0 ]; do
  case "$1" in
    --workspace)
      [ "$#" -ge 2 ] || {
        printf '%s\n' "production-local-gate: --workspace requires a value" >&2
        exit 2
      }
      workspace="$2"
      shift 2
      ;;
    --dist)
      [ "$#" -ge 2 ] || {
        printf '%s\n' "production-local-gate: --dist requires a value" >&2
        exit 2
      }
      dist="$2"
      shift 2
      ;;
    --output-dir)
      [ "$#" -ge 2 ] || {
        printf '%s\n' "production-local-gate: --output-dir requires a value" >&2
        exit 2
      }
      output_dir="$2"
      shift 2
      ;;
    --evidence-root)
      [ "$#" -ge 2 ] || {
        printf '%s\n' "production-local-gate: --evidence-root requires a value" >&2
        exit 2
      }
      evidence_root="$2"
      shift 2
      ;;
    --help|-h)
      usage
      exit 0
      ;;
    *)
      printf '%s\n' "production-local-gate: unknown argument: $1" >&2
      usage >&2
      exit 2
      ;;
  esac
done

case "$dist" in
  /*) ;;
  *) dist="$(pwd)/$dist" ;;
esac

case "$output_dir" in
  /*) ;;
  *) output_dir="$(pwd)/$output_dir" ;;
esac

case "$evidence_root" in
  /*) ;;
  *) evidence_root="$(pwd)/$evidence_root" ;;
esac

case "$workspace" in
  /*) ;;
  *) workspace="$(pwd)/$workspace" ;;
esac

mkdir -p "$output_dir"

set +e
sh "$root/scripts/secret-scan.sh" --root "$workspace" >"$output_dir/secret-scan.stdout.txt" 2>"$output_dir/secret-scan.stderr.txt"
secret_scan_status=$?
set -e
if [ "$secret_scan_status" -ne 0 ]; then
  printf '%s\n' "production-local-gate: fail secret scan"
  exit 1
fi
printf '%s\n' "production-local-gate: secret_scan=pass"

set +e
sh "$root/scripts/production-readiness.sh" --dist "$dist" --evidence-root "$evidence_root" --output-dir "$output_dir" >"$output_dir/production-readiness.stdout.txt" 2>"$output_dir/production-readiness.stderr.txt"
readiness_status=$?
set -e

python3 - "$output_dir/summary.json" "$readiness_status" <<'PY'
import json
import sys

summary_path = sys.argv[1]
readiness_status = int(sys.argv[2])
with open(summary_path, "r", encoding="utf-8") as handle:
    summary = json.load(handle)

local_ready = bool(summary.get("local_production_ready"))
public_ready = bool(summary.get("public_production_ready"))
checklist = summary.get("launch_checklist") or {}
checklist_ok = checklist.get("status") == "pass" and int(checklist.get("required_action_count", 0) or 0) >= 0

if not local_ready:
    print("production-local-gate: fail local_production_ready=false")
    raise SystemExit(1)
if not checklist_ok:
    print("production-local-gate: fail launch checklist missing or invalid")
    raise SystemExit(1)
if public_ready and readiness_status != 0:
    print("production-local-gate: fail public ready but production-readiness exited non-zero")
    raise SystemExit(1)

print(f"production-local-gate: pass local_production_ready=true public_production_ready={str(public_ready).lower()}")
print(f"production-local-gate: blocked_count={summary.get('blocked_count', 0)}")
print(f"production-local-gate: checklist_actions={checklist.get('required_action_count', 0)}")
PY

set +e
go run "$root/cmd/ceo-packet" production-actions --workspace "$workspace" --format json >"$output_dir/production-actions.json" 2>"$output_dir/production-actions.stderr.txt"
actions_status=$?
go run "$root/cmd/ceo-packet" production-actions --workspace "$workspace" --commands-only >"$output_dir/production-actions.commands.sh" 2>>"$output_dir/production-actions.stderr.txt"
commands_status=$?
go run "$root/cmd/ceo-packet" production-status --workspace "$workspace" --format json >"$output_dir/production-status.json" 2>"$output_dir/production-status.stderr.txt"
status_status=$?
set -e

python3 - "$output_dir/summary.json" "$output_dir/production-actions.json" "$output_dir/production-actions.commands.sh" "$output_dir/production-status.json" "$actions_status" "$commands_status" "$status_status" <<'PY'
import json
import os
import sys

summary_path, actions_path, commands_path, status_path = sys.argv[1], sys.argv[2], sys.argv[3], sys.argv[4]
actions_status, commands_status, status_status = int(sys.argv[5]), int(sys.argv[6]), int(sys.argv[7])
if actions_status != 0 or commands_status != 0 or status_status != 0:
    print("production-local-gate: fail production-actions command failed")
    raise SystemExit(1)

with open(summary_path, "r", encoding="utf-8") as handle:
    readiness = json.load(handle)

with open(actions_path, "r", encoding="utf-8") as handle:
    actions = json.load(handle)

with open(status_path, "r", encoding="utf-8") as handle:
    production_status = json.load(handle)

status = actions.get("status")
if status == "missing":
    print("production-local-gate: production_actions=missing")
    raise SystemExit(0)
if status == "pass" and int(actions.get("required_action_count", 0) or 0) == 0:
    print("production-local-gate: production_actions=pass")
    raise SystemExit(0)

source_path = actions.get("path")
if not source_path or not os.path.exists(source_path):
    print("production-local-gate: fail production action source missing")
    raise SystemExit(1)

runnable = int(actions.get("runnable_command_count", 0) or 0)
blocked = int(actions.get("blocked_command_count", 0) or 0)
required = int(actions.get("required_action_count", 0) or 0)
if required > 0 and runnable + blocked <= 0:
    print("production-local-gate: fail production action command counts missing")
    raise SystemExit(1)

action_rows = actions.get("actions") or []
if required != len(action_rows):
    print(f"production-local-gate: fail action rows={len(action_rows)} expected={required}")
    raise SystemExit(1)
state_counts = actions.get("action_state_counts") or {}
state_total = sum(int(value or 0) for value in state_counts.values())
if required > 0 and state_total != len(action_rows):
    print(f"production-local-gate: fail action state counts={state_total} expected={len(action_rows)}")
    raise SystemExit(1)
missing_reasons = []
for action in action_rows:
    action_id = action.get("id") or "<unknown>"
    if not action.get("action_state"):
        print(f"production-local-gate: fail action_state missing for {action_id}")
        raise SystemExit(1)
    if not action.get("action_reason"):
        missing_reasons.append(action_id)
    if action.get("kind") == "release_proof":
        release_summary = action.get("release_summary") or {}
        if release_summary.get("setup_command_policy") != "no_publish_no_secret_assignment":
            print(f"production-local-gate: fail release setup policy missing for {action_id}")
            raise SystemExit(1)
        if release_summary.get("publish_actions_performed") is not False:
            print(f"production-local-gate: fail release publish action flag unsafe for {action_id}")
            raise SystemExit(1)
        if release_summary.get("secret_value_saved") is not False:
            print(f"production-local-gate: fail release secret flag unsafe for {action_id}")
            raise SystemExit(1)
        setup_actions_path = release_summary.get("setup_actions_path") or ""
        if not setup_actions_path or not os.path.exists(setup_actions_path):
            print(f"production-local-gate: fail release setup actions file missing for {action_id}")
            raise SystemExit(1)
        setup_actions_text = open(setup_actions_path, "r", encoding="utf-8").read()
        for blocked_check in release_summary.get("blocked_checks") or []:
            if f"- {blocked_check}:" not in setup_actions_text:
                print(f"production-local-gate: fail release setup action missing for {action_id}: {blocked_check}")
                raise SystemExit(1)
        if "sh scripts/release-readiness.sh --dist dist --output-dir .omo/evidence/release-readiness-final" not in setup_actions_text:
            print(f"production-local-gate: fail release setup actions missing readiness rerun for {action_id}")
            raise SystemExit(1)
        has_installed_finalizer = "ceo-packet production-finalize --workspace . --dry-run" in setup_actions_text
        has_source_finalizer = "go run ./cmd/ceo-packet production-finalize --workspace . --dry-run" in setup_actions_text
        if not (has_installed_finalizer or has_source_finalizer):
            print(f"production-local-gate: fail release setup actions missing finalizer rerun for {action_id}")
            raise SystemExit(1)
        setup_commands_path = release_summary.get("setup_commands_path") or ""
        if setup_commands_path or release_summary.get("setup_commands_sha256"):
            if not setup_commands_path or not os.path.exists(setup_commands_path):
                print(f"production-local-gate: fail release setup command file missing for {action_id}")
                raise SystemExit(1)
            setup_commands_text = open(setup_commands_path, "r", encoding="utf-8").read()
            if "RELEASE_URL=" in setup_commands_text and "# blocked remote_release_url:" not in setup_commands_text:
                print(f"production-local-gate: fail release setup commands assign release URL for {action_id}")
                raise SystemExit(1)
            if "git push" in setup_commands_text or "gh release create" in setup_commands_text:
                print(f"production-local-gate: fail release setup commands include publish command for {action_id}")
                raise SystemExit(1)
            for blocked_check in release_summary.get("blocked_checks") or []:
                if f"# blocked {blocked_check}:" not in setup_commands_text:
                    print(f"production-local-gate: fail release setup command missing for {action_id}: {blocked_check}")
                    raise SystemExit(1)
    if action.get("kind") == "provider_proof":
        provider_summary = action.get("provider_summary") or {}
        provider_name = provider_summary.get("provider") or action.get("provider") or ""
        api_key_env = provider_summary.get("api_key_env") or action.get("required_env") or ""
        provider_status = provider_summary.get("status") or ""
        if provider_summary.get("command_script_secret_policy") != "no_secret_assignment":
            print(f"production-local-gate: fail provider command secret policy missing for {action_id}")
            raise SystemExit(1)
        if provider_summary.get("secret_value_saved") is not False:
            print(f"production-local-gate: fail provider secret flag unsafe for {action_id}")
            raise SystemExit(1)
        if provider_status == "pass":
            continue
        setup_hashes = provider_summary.get("setup_artifacts_sha256") or {}
        if setup_hashes:
            for artifact in ("blocked.md", "commands.sh", "env.template", "setup-checklist.md"):
                if not setup_hashes.get(artifact):
                    print(f"production-local-gate: fail provider setup artifact hash missing for {action_id}: {artifact}")
                    raise SystemExit(1)
            provider_commands_path = provider_summary.get("commands_path") or ""
            if not provider_commands_path or not os.path.exists(provider_commands_path):
                print(f"production-local-gate: fail provider command file missing for {action_id}")
                raise SystemExit(1)
            provider_commands = open(provider_commands_path, "r", encoding="utf-8").read()
            if api_key_env and f"{api_key_env}=" in provider_commands:
                print(f"production-local-gate: fail provider command file assigns secret env for {action_id}")
                raise SystemExit(1)
            if "<redacted>" in provider_commands:
                print(f"production-local-gate: fail provider command file uses redacted secret placeholder for {action_id}")
                raise SystemExit(1)
            if "Do not paste secret values into this file or any evidence artifact." not in provider_commands:
                print(f"production-local-gate: fail provider command file missing secret-safety comment for {action_id}")
                raise SystemExit(1)
            if provider_name and f"scripts/provider-proof.sh --provider {provider_name}" not in provider_commands:
                print(f"production-local-gate: fail provider command file missing rerun command for {action_id}")
                raise SystemExit(1)
            if "provider setup:" in provider_commands:
                if api_key_env and f"${{{api_key_env}+x}}" not in provider_commands:
                    print(f"production-local-gate: fail provider command file missing env-present guard for {action_id}")
                    raise SystemExit(1)
                if api_key_env and f"${{{api_key_env}}}" not in provider_commands:
                    print(f"production-local-gate: fail provider command file missing env-empty guard for {action_id}")
                    raise SystemExit(1)
if missing_reasons:
    print("production-local-gate: fail action_reason missing for " + ", ".join(missing_reasons))
    raise SystemExit(1)

mismatches = int(actions.get("evidence_declared_mismatch_count", 0) or 0)
if mismatches > 0:
    print(f"production-local-gate: fail production action evidence mismatches={mismatches}")
    raise SystemExit(1)

finalizer = production_status.get("finalizer_next_actions") or {}
launch = production_status.get("launch_checklist") or {}
if launch.get("matches_declared") is False:
    print("production-local-gate: fail launch checklist fingerprint mismatch")
    raise SystemExit(1)
if not bool(readiness.get("public_production_ready")):
    if int(finalizer.get("evidence_declared_mismatch_count", 0) or 0) > 0:
        print("production-local-gate: fail finalizer evidence mismatch")
        raise SystemExit(1)
    if int(finalizer.get("setup_required_action_count", 0) or 0) <= 0 or not finalizer.get("setup_sha256"):
        print("production-local-gate: fail finalizer setup metadata missing")
        raise SystemExit(1)
    if finalizer.get("setup_matches_declared") is False:
        print("production-local-gate: fail finalizer setup fingerprint mismatch")
        raise SystemExit(1)

commands_text = open(commands_path, "r", encoding="utf-8").read()
secret_env_assignments = (
    "OPENAI_API_KEY=",
    "OPENROUTER_API_KEY=",
    "KIMI_CODE_API_KEY=",
    "MOONSHOT_API_KEY=",
    "MINIMAX_API_KEY=",
)
if any(marker in commands_text for marker in secret_env_assignments):
    print("production-local-gate: fail production action commands include secret assignment")
    raise SystemExit(1)
blocked_lines = sum(1 for line in commands_text.splitlines() if line.startswith("# blocked command:"))
if blocked_lines != blocked:
    print(f"production-local-gate: fail blocked command lines={blocked_lines} expected={blocked}")
    raise SystemExit(1)
reason_lines = sum(1 for line in commands_text.splitlines() if line.startswith("# ") and " reason: " in line)
if required > 0 and reason_lines <= 0:
    print("production-local-gate: fail production action commands missing blocker reasons")
    raise SystemExit(1)

print(f"production-local-gate: production_actions={status}")
print(f"production-local-gate: runnable_commands={runnable}")
print(f"production-local-gate: blocked_commands={blocked}")
print(f"production-local-gate: action_reasons={len(action_rows)}")
print("production-local-gate: release_setup_policy=verified")
print("production-local-gate: provider_setup_policy=verified")
print(f"production-local-gate: finalizer_setup_actions={finalizer.get('setup_required_action_count', 0)}")
PY
