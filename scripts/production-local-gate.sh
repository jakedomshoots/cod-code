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
