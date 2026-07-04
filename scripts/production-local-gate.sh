#!/bin/sh
set -eu

root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
dist=${DIST:-"$root/dist"}
output_dir=${OUTPUT_DIR:-"$root/.omo/evidence/production-local-gate"}
evidence_root=${EVIDENCE_ROOT:-"$root/.omo/evidence"}

usage() {
  cat <<'USAGE'
Usage: sh scripts/production-local-gate.sh [--dist dir] [--output-dir dir] [--evidence-root dir]

Runs production-readiness and fails only when local production readiness is not
green. Public release/provider/comparison blockers are allowed here because CI
does not have release credentials or paid provider keys.
USAGE
}

while [ "$#" -gt 0 ]; do
  case "$1" in
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

mkdir -p "$output_dir"

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
go run "$root/cmd/ceo-packet" production-actions --workspace "$root" --format json >"$output_dir/production-actions.json" 2>"$output_dir/production-actions.stderr.txt"
actions_status=$?
go run "$root/cmd/ceo-packet" production-actions --workspace "$root" --commands-only >"$output_dir/production-actions.commands.sh" 2>>"$output_dir/production-actions.stderr.txt"
commands_status=$?
go run "$root/cmd/ceo-packet" production-status --workspace "$root" --format json >"$output_dir/production-status.json" 2>"$output_dir/production-status.stderr.txt"
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
    if not bool(readiness.get("public_production_ready")):
        print("production-local-gate: fail production actions missing while public blockers remain")
        raise SystemExit(1)
    print("production-local-gate: production_actions=missing")
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
if "OPENAI_API_KEY=" in commands_text or "OPENROUTER_API_KEY=" in commands_text or "MOONSHOT_API_KEY=" in commands_text:
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
print(f"production-local-gate: finalizer_setup_actions={finalizer.get('setup_required_action_count', 0)}")
PY
