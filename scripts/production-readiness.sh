#!/bin/sh
set -eu

root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
invocation_dir=$(pwd)
evidence_root="$root/.omo/evidence"
output_dir="$root/.omo/evidence/production-readiness"
dist="$root/dist"
skip_secret_scan=0
skip_release_readiness=0

usage() {
  cat <<'USAGE'
Usage: sh scripts/production-readiness.sh [--evidence-root dir] [--output-dir dir] [--dist dir] [--skip-secret-scan] [--skip-release-readiness]

Writes one production-readiness packet from the release, provider, eval,
security, endurance, and competitor evidence gates.

This command does not publish, push, tag, upload, or call paid providers.
It exits non-zero while public-production blockers remain.
USAGE
}

while [ "$#" -gt 0 ]; do
  case "$1" in
    --evidence-root)
      [ "$#" -ge 2 ] || {
        printf '%s\n' "production-readiness: --evidence-root requires a value" >&2
        exit 2
      }
      evidence_root="$2"
      shift 2
      ;;
    --output-dir)
      [ "$#" -ge 2 ] || {
        printf '%s\n' "production-readiness: --output-dir requires a value" >&2
        exit 2
      }
      output_dir="$2"
      shift 2
      ;;
    --dist)
      [ "$#" -ge 2 ] || {
        printf '%s\n' "production-readiness: --dist requires a value" >&2
        exit 2
      }
      dist="$2"
      shift 2
      ;;
    --skip-secret-scan)
      skip_secret_scan=1
      shift
      ;;
    --skip-release-readiness)
      skip_release_readiness=1
      shift
      ;;
    --help|-h)
      usage
      exit 0
      ;;
    *)
      printf '%s\n' "production-readiness: unknown argument: $1" >&2
      usage >&2
      exit 2
      ;;
  esac
done

abspath() {
  case "$1" in
    /*) printf '%s\n' "$1" ;;
    *) printf '%s\n' "$invocation_dir/$1" ;;
  esac
}

evidence_root=$(abspath "$evidence_root")
output_dir=$(abspath "$output_dir")
dist=$(abspath "$dist")

mkdir -p "$output_dir"
: >"$output_dir/checks.tsv"
: >"$output_dir/blockers.txt"

status_for_json() {
  case "$1" in
    pass|blocked|fail|missing|skipped) printf '%s\n' "$1" ;;
    *) printf '%s\n' "fail" ;;
  esac
}

add_check() {
  category="$1"
  name="$2"
  status=$(status_for_json "$3")
  evidence="$4"
  detail="$5"
  printf '%s\t%s\t%s\t%s\t%s\n' "$category" "$name" "$status" "$evidence" "$detail" >>"$output_dir/checks.tsv"
  case "$status" in
    pass|skipped) ;;
    *) printf '%s: %s (%s)\n' "$category" "$name" "$status" >>"$output_dir/blockers.txt" ;;
  esac
}

json_check() {
  file="$1"
  expression="$2"
  [ -f "$file" ] || return 2
  python3 - "$file" "$expression" <<'PY'
import json
import sys

with open(sys.argv[1], "r", encoding="utf-8") as handle:
    data = json.load(handle)

allowed_names = {"int": int, "str": str, "len": len, "bool": bool}
ok = bool(eval(sys.argv[2], {"__builtins__": {}}, {"data": data, **allowed_names}))
raise SystemExit(0 if ok else 1)
PY
}

provider_index_has_pass() {
  file="$1"
  [ -f "$file" ] || return 2
  grep -q -- "- Overall: pass" "$file"
}

provider_index_has_blocked() {
  file="$1"
  [ -f "$file" ] || return 2
  grep -q -- "- Overall: blocked" "$file"
}

endurance_index_has_pass() {
  file="$1"
  [ -f "$file" ] || return 2
  grep -q -- "- Overall: pass" "$file" && grep -q -- "- Completed iterations: 30" "$file"
}

latest_evidence_summary() {
  prefix="$1"
  python3 - "$evidence_root" "$prefix" <<'PY'
import glob
import os
import sys

root, prefix = sys.argv[1], sys.argv[2]
candidates = glob.glob(os.path.join(root, prefix + "*/summary.json"))
if not candidates:
    print("")
    raise SystemExit(0)

latest = max(candidates, key=lambda path: (os.path.getmtime(path), path))
print(latest)
PY
}

if [ "$skip_secret_scan" -eq 1 ]; then
  add_check "security" "secret_scan" "skipped" "not-run" "Skipped by flag"
else
  if sh "$root/scripts/secret-scan.sh" >"$output_dir/secret-scan.txt" 2>&1; then
    add_check "security" "secret_scan" "pass" "secret-scan.txt" "No committed secret values found"
  else
    add_check "security" "secret_scan" "fail" "secret-scan.txt" "Secret scan failed"
  fi
fi

release_summary=$(latest_evidence_summary "release-readiness-")
if [ -z "$release_summary" ]; then
  release_summary="$evidence_root/release-readiness-r1/summary.json"
fi
if [ "$skip_release_readiness" -eq 1 ]; then
  add_check "release" "public_release_readiness_run" "skipped" "not-run" "Skipped by flag"
else
  release_output="$output_dir/release-readiness"
  if sh "$root/scripts/release-readiness.sh" --dist "$dist" --output-dir "$release_output" >"$output_dir/release-readiness.stdout.txt" 2>"$output_dir/release-readiness.stderr.txt"; then
    add_check "release" "public_release_readiness_run" "pass" "release-readiness/index.md" "Public release preflight passed"
  else
    add_check "release" "public_release_readiness_run" "blocked" "release-readiness/index.md" "Public release preflight still has blockers"
  fi
  release_summary="$release_output/summary.json"
fi

if json_check "$release_summary" "data.get('status') == 'pass' and data.get('public_release_ready') is True"; then
  add_check "release" "public_release_ready" "pass" "$release_summary" "Public release evidence is green"
elif json_check "$release_summary" "data.get('status') == 'blocked' and data.get('public_release_ready') is False and data.get('publish_actions_performed') is False and data.get('secret_value_saved') is False and data.get('setup_command_policy') == 'no_publish_no_secret_assignment'"; then
  add_check "release" "public_release_ready" "blocked" "$release_summary" "Remote/release/Homebrew/signing evidence is not complete"
else
  add_check "release" "public_release_ready" "blocked" "$release_summary" "Release evidence missing setup safety policy"
fi

if json_check "$evidence_root/production-core-29-ceo-r1/summary.json" "int(data.get('passed', 0)) >= 29 and int(data.get('failed', 0)) == 0 and int(data.get('partial', 0)) == 0 and int(data.get('timed_out', 0)) == 0 and int(data.get('incomplete_evidence', 0)) == 0"; then
  add_check "eval" "ceo_29_task_production_core" "pass" "$evidence_root/production-core-29-ceo-r1/summary.json" "CEO Harness passed the current 29-task suite"
else
  add_check "eval" "ceo_29_task_production_core" "blocked" "$evidence_root/production-core-29-ceo-r1/summary.json" "Current CEO 29-task proof is missing or not clean"
fi

if json_check "$evidence_root/benchmark-fixtures-31-r1/summary.json" "int(data.get('task_count', 0)) >= 31 and int(data.get('passed', 0)) >= 31 and int(data.get('failed', 0)) == 0 and int(data.get('partial', 0)) == 0"; then
  add_check "eval" "full_fixture_catalog" "pass" "$evidence_root/benchmark-fixtures-31-r1/summary.json" "Full deterministic task catalog scores cleanly"
else
  add_check "eval" "full_fixture_catalog" "blocked" "$evidence_root/benchmark-fixtures-31-r1/summary.json" "Full fixture catalog proof is missing or not clean"
fi

comparison_summary=$(latest_evidence_summary "external-agent-production-core-29-")
if json_check "$comparison_summary" "int(data.get('task_count', 0)) >= 29 and int(data.get('agent_count', 0)) >= 4 and int(data.get('failed', 0)) == 0 and int(data.get('partial', 0)) == 0 and int(data.get('timed_out', 0)) == 0 and int(data.get('setup_blocked', 0)) == 0 and int(data.get('incomplete_evidence', 0)) == 0"; then
  add_check "comparison" "all_agent_29_task_comparison" "pass" "$comparison_summary" "All configured agents have clean current-suite evidence"
else
  add_check "comparison" "all_agent_29_task_comparison" "blocked" "$comparison_summary" "Latest all-agent current-suite evidence is missing, partial, timed out, or incomplete"
fi

if provider_index_has_pass "$evidence_root/provider-proof-kimi-r2/index.md"; then
  add_check "provider" "kimi_real_provider" "pass" "$evidence_root/provider-proof-kimi-r2/index.md" "Kimi provider proof passed"
else
  add_check "provider" "kimi_real_provider" "blocked" "$evidence_root/provider-proof-kimi-r2/index.md" "Kimi provider proof missing or not passing"
fi

if provider_index_has_pass "$evidence_root/provider-proof-codex-r1/index.md"; then
  add_check "provider" "codex_real_provider" "pass" "$evidence_root/provider-proof-codex-r1/index.md" "Codex provider proof passed"
else
  add_check "provider" "codex_real_provider" "blocked" "$evidence_root/provider-proof-codex-r1/index.md" "Codex provider proof missing or not passing"
fi

http_blocked=0
for provider in openai openrouter moonshot; do
  index="$evidence_root/provider-proof-$provider/index.md"
  summary="$evidence_root/provider-proof-$provider/summary.json"
  if provider_index_has_pass "$index"; then
    add_check "provider" "${provider}_http_provider" "pass" "$index" "$provider HTTP provider proof passed"
  elif provider_index_has_blocked "$index" && json_check "$summary" "data.get('status') == 'blocked' and data.get('blocked_reason') in ['missing_api_key_env', 'empty_api_key_env'] and data.get('secret_value_saved') is False and data.get('command_script_secret_policy') == 'no_secret_assignment'"; then
    add_check "provider" "${provider}_http_provider" "blocked" "$index" "$provider HTTP provider proof is blocked by setup"
    http_blocked=1
  else
    add_check "provider" "${provider}_http_provider" "blocked" "$index" "$provider HTTP provider proof missing"
    http_blocked=1
  fi
done

if endurance_index_has_pass "$evidence_root/endurance-local-r3/index.md"; then
  add_check "endurance" "thirty_loop_local_endurance" "pass" "$evidence_root/endurance-local-r3/index.md" "30 local loops passed"
else
  add_check "endurance" "thirty_loop_local_endurance" "blocked" "$evidence_root/endurance-local-r3/index.md" "30-loop endurance proof missing or not clean"
fi

local_blockers=$(awk -F '\t' '$1 != "release" && $1 != "comparison" && !($1 == "provider" && $2 ~ /_http_provider$/) && $3 != "pass" && $3 != "skipped" { count++ } END { print count+0 }' "$output_dir/checks.tsv")
public_blockers=$(awk -F '\t' '$3 != "pass" && $3 != "skipped" { count++ } END { print count+0 }' "$output_dir/checks.tsv")

if [ "$local_blockers" -eq 0 ]; then
  local_ready=true
else
  local_ready=false
fi

if [ "$public_blockers" -eq 0 ] && [ "$http_blocked" -eq 0 ]; then
  public_ready=true
  overall_status=pass
else
  public_ready=false
  overall_status=blocked
fi

{
  printf '%s\n' "# Launch Checklist"
  printf '\n'
  printf '%s\n' "Local production ready: $local_ready"
  printf '%s\n' "Public production ready: $public_ready"
  printf '\n'
  printf '%s\n' "## Required Before Public Production Claim"
  printf '\n'
  if [ "$public_blockers" -eq 0 ]; then
    printf '%s\n' "- No public-production blockers remain."
  else
    while IFS='	' read -r category name status evidence detail; do
      case "$status" in
        pass|skipped) continue ;;
      esac
      case "$category.$name" in
        release.public_release_readiness_run|release.public_release_ready)
          printf '%s\n' "- Publish release proof: push an explicit \`v*\` tag so the GitHub release workflow publishes verified tarballs, \`checksums.txt\`, and \`release-manifest.json\`; then rerun \`sh scripts/release-readiness.sh --dist dist --output-dir .omo/evidence/release-readiness-final\` with the public release/Homebrew/signing inputs in place."
          ;;
        comparison.all_agent_29_task_comparison)
          printf '%s\n' "- Refresh market comparison: rerun the 29-task all-agent gauntlet after external CLI auth/timeout setup, then confirm \`comparison-report.md\` says \`Overall comparison: pass\`."
          ;;
        provider.openai_http_provider)
          printf '%s\n' "- Prove OpenAI provider: export \`OPENAI_API_KEY\`, run \`sh scripts/provider-proof.sh --provider openai --output-dir .omo/evidence/provider-proof-openai\`, and keep the key out of evidence."
          ;;
        provider.openrouter_http_provider)
          printf '%s\n' "- Prove OpenRouter provider: export \`OPENROUTER_API_KEY\`, run \`sh scripts/provider-proof.sh --provider openrouter --output-dir .omo/evidence/provider-proof-openrouter\`, and keep the key out of evidence."
          ;;
        provider.moonshot_http_provider)
          printf '%s\n' "- Prove Moonshot provider: export \`MOONSHOT_API_KEY\`, run \`sh scripts/provider-proof.sh --provider moonshot --output-dir .omo/evidence/provider-proof-moonshot\`, and keep the key out of evidence."
          ;;
        *)
          printf -- '- Resolve `%s.%s`: %s Evidence: `%s`.\n' "$category" "$name" "$detail" "$evidence"
          ;;
      esac
    done <"$output_dir/checks.tsv" | awk '!seen[$0]++'
  fi
  printf '\n'
  printf '%s\n' "## Final Gate"
  printf '\n'
  printf '%s\n' "After every item above is complete, run:"
  printf '\n'
  printf '%s\n' "\`\`\`sh"
  printf '%s\n' "sh scripts/production-readiness.sh --dist dist --output-dir .omo/evidence/production-readiness"
  printf '%s\n' "\`\`\`"
} >"$output_dir/launch-checklist.md"

python3 - "$output_dir/checks.tsv" "$output_dir/summary.json" "$output_dir/launch-checklist.md" "$local_ready" "$public_ready" "$overall_status" <<'PY'
import hashlib
import json
import pathlib
import sys

checks = []
with open(sys.argv[1], "r", encoding="utf-8") as handle:
    for line in handle:
        category, name, status, evidence, detail = line.rstrip("\n").split("\t", 4)
        checks.append({
            "category": category,
            "name": name,
            "status": status,
            "evidence": evidence,
            "detail": detail,
        })

checklist_path = pathlib.Path(sys.argv[3])
checklist_text = checklist_path.read_text(encoding="utf-8")
actions = [line for line in checklist_text.splitlines() if line.startswith("- ")]
blocked = [check for check in checks if check["status"] not in {"pass", "skipped"}]
summary = {
    "schema_version": 1,
    "status": sys.argv[6],
    "local_production_ready": sys.argv[4] == "true",
    "public_production_ready": sys.argv[5] == "true",
    "check_count": len(checks),
    "blocked_count": len(blocked),
    "blocked_checks": [f"{check['category']}.{check['name']}" for check in blocked],
    "launch_checklist": {
        "path": "launch-checklist.md",
        "sha256": hashlib.sha256(checklist_text.encode("utf-8")).hexdigest(),
        "required_action_count": len(actions),
        "status": "pass" if "# Launch Checklist" in checklist_text and "## Final Gate" in checklist_text else "fail",
    },
    "checks": checks,
}
with open(sys.argv[2], "w", encoding="utf-8") as handle:
    json.dump(summary, handle, indent=2)
    handle.write("\n")
PY

{
  printf '%s\n' "# Production Readiness Evidence"
  printf '\n'
  printf '%s\n' "Status: $overall_status"
  printf '%s\n' "Local production ready: $local_ready"
  printf '%s\n' "Public production ready: $public_ready"
  printf '\n'
  printf '%s\n' "| Category | Check | Status | Evidence |"
  printf '%s\n' "| --- | --- | --- | --- |"
  while IFS='	' read -r category name status evidence detail; do
    printf '| %s | %s | %s | `%s` |\n' "$category" "$name" "$status" "$evidence"
  done <"$output_dir/checks.tsv"
  printf '\n'
  if [ "$public_blockers" -gt 0 ]; then
    printf '%s\n' "## Blockers"
    printf '\n'
    while IFS= read -r blocker; do
      [ -n "$blocker" ] || continue
      printf -- '- %s\n' "$blocker"
    done <"$output_dir/blockers.txt"
    printf '\n'
  fi
  printf '%s\n' "## Launch Checklist"
  printf '\n'
  printf '%s\n' "Next public-production actions are in \`launch-checklist.md\`."
  printf '\n'
  printf '%s\n' "## Publish Boundary"
  printf '\n'
  printf '%s\n' "This command does not publish, push, tag, upload, create releases, or call paid providers."
} >"$output_dir/index.md"

printf '%s\n' "production-readiness: wrote $output_dir/index.md"
printf '%s\n' "production-readiness: $overall_status"

if [ "$overall_status" = "pass" ]; then
  exit 0
fi
exit 1
