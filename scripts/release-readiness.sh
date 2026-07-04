#!/bin/sh
set -eu

root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
dist=${DIST:-"$root/dist"}
output_dir=${OUTPUT_DIR:-"$root/.omo/evidence/release-readiness"}

usage() {
  cat <<'USAGE'
Usage: sh scripts/release-readiness.sh [--dist dist] [--output-dir dir]

Writes a release-readiness evidence packet without publishing anything.

The command verifies local artifacts, runs public release preflight, captures
git/GitHub state, and writes index.md, summary.json, and raw command logs.
It exits non-zero while public release blockers remain.
USAGE
}

while [ "$#" -gt 0 ]; do
  case "$1" in
    --dist)
      [ "$#" -ge 2 ] || {
        printf '%s\n' "release-readiness: --dist requires a value" >&2
        exit 2
      }
      dist="$2"
      shift 2
      ;;
    --output-dir)
      [ "$#" -ge 2 ] || {
        printf '%s\n' "release-readiness: --output-dir requires a value" >&2
        exit 2
      }
      output_dir="$2"
      shift 2
      ;;
    --help|-h)
      usage
      exit 0
      ;;
    *)
      printf '%s\n' "release-readiness: unknown argument: $1" >&2
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

mkdir -p "$output_dir"

if command -v ceo-packet >/dev/null 2>&1; then
  ceo_packet_cmd="ceo-packet"
else
  ceo_packet_cmd="go run ./cmd/ceo-packet"
fi

verify_status=blocked
if sh "$root/scripts/verify-release.sh" "$dist" >"$output_dir/verify-release.txt" 2>&1; then
  verify_status=pass
fi

preflight_exit=0
if sh "$root/scripts/release-preflight.sh" "$dist" >"$output_dir/preflight.md" 2>&1; then
  preflight_status=pass
else
  preflight_exit=$?
  preflight_status=blocked
fi

remote_url=$(git -C "$root" remote get-url origin 2>/dev/null || true)
if [ -n "$remote_url" ]; then
  printf '%s\n' "$remote_url" >"$output_dir/git-remote.txt"
else
  printf '%s\n' "no origin remote configured" >"$output_dir/git-remote.txt"
fi

github_auth_status=skipped
if command -v gh >/dev/null 2>&1; then
  if gh auth status >"$output_dir/github-auth.txt" 2>&1; then
    github_auth_status=pass
  else
    github_auth_status=blocked
  fi
else
  printf '%s\n' "gh CLI not installed" >"$output_dir/github-auth.txt"
fi

blocked_checks_file="$output_dir/blocked-checks.txt"
awk -F '|' '
  $0 ~ /^\|/ && $3 ~ /blocked/ {
    name=$2
    gsub(/^ +| +$/, "", name)
    if (name != "Check") {
      print name
    }
  }
' "$output_dir/preflight.md" >"$blocked_checks_file"

blocked_count=$(wc -l <"$blocked_checks_file" | tr -d ' ')
if [ "$preflight_status" = "pass" ]; then
  public_release_ready=true
  overall_status=pass
else
  public_release_ready=false
  overall_status=blocked
fi

json_array() {
  awk '
    BEGIN { first=1; printf "[" }
    {
      gsub(/\\/,"\\\\")
      gsub(/"/,"\\\"")
      if (!first) {
        printf ", "
      }
      printf "\"%s\"", $0
      first=0
    }
    END { printf "]" }
  ' "$1"
}

blocked_checks_json=$(json_array "$blocked_checks_file")

setup_actions_file="$output_dir/setup-actions.md"
write_setup_action() {
  check="$1"
  case "$check" in
    git_remote)
      printf '%s\n' "- git_remote: configure an origin remote for the public repo, for example \`git remote add origin git@github.com:<owner>/<repo>.git\`."
      ;;
    remote_release_url)
      printf '%s\n' "- remote_release_url: set \`RELEASE_URL\` or \`PUBLIC_RELEASE_URL\` to the public HTTPS release page."
      ;;
    github_release_assets)
      printf '%s\n' "- github_release_assets: push a \`v*\` tag, let the release workflow upload archives, \`checksums.txt\`, and \`release-manifest.json\`, then set \`GH_RELEASE_TAG\` and \`GH_REPO\` if no GitHub origin is configured."
      ;;
    homebrew_formula_url)
      printf '%s\n' "- homebrew_formula_url: publish the release archives to a stable HTTPS URL and update \`dist/homebrew/ceo-packet.rb\` or the tap formula so it uses that remote archive URL."
      ;;
    artifact_signatures)
      printf '%s\n' "- artifact_signatures: add \`.sig\`, \`.minisig\`, or \`.asc\` signatures for every archive, or set \`ALLOW_CHECKSUM_ONLY_RELEASE=1\` with a public \`CHECKSUM_ONLY_RELEASE_NOTES_URL\`."
      ;;
    *)
      printf '%s\n' "- $check: inspect \`preflight.md\` and resolve this blocked release preflight check."
      ;;
  esac
}

if [ "$blocked_count" -gt 0 ]; then
  {
    printf '%s\n' "# Release Setup Actions"
    printf '\n'
    printf '%s\n' "Run these before making a public production release claim."
    printf '\n'
    while IFS= read -r check; do
      [ -n "$check" ] || continue
      write_setup_action "$check"
    done <"$blocked_checks_file"
    printf '\n'
    printf '%s\n' "After setup is complete, rerun:"
    printf '\n'
    printf '%s\n' '```sh'
    printf '%s\n' 'sh scripts/release-readiness.sh --dist dist --output-dir .omo/evidence/release-readiness-final'
    printf '%s\n' "$ceo_packet_cmd production-finalize --workspace . --dry-run"
    printf '%s\n' '```'
  } >"$setup_actions_file"
fi

setup_action_count=0
setup_actions_sha256=""
if [ "$blocked_count" -gt 0 ] && [ -f "$setup_actions_file" ]; then
  setup_action_count=$(awk '/^- / { count += 1 } END { print count + 0 }' "$setup_actions_file")
  setup_actions_sha256=$(python3 - "$setup_actions_file" <<'PY'
import hashlib
import sys

with open(sys.argv[1], "rb") as handle:
    print(hashlib.sha256(handle.read()).hexdigest())
PY
)
fi

cat >"$output_dir/summary.json" <<JSON
{
  "schema_version": 1,
  "public_release_ready": $public_release_ready,
  "status": "$overall_status",
  "dist": "$dist",
  "release_artifacts_verified": $(if [ "$verify_status" = "pass" ]; then printf true; else printf false; fi),
  "preflight_status": "$preflight_status",
  "preflight_exit_code": $preflight_exit,
  "blocked_count": $blocked_count,
  "blocked_checks": $blocked_checks_json,
  "setup_actions": "$(if [ "$blocked_count" -gt 0 ]; then printf 'setup-actions.md'; fi)",
  "setup_action_count": $setup_action_count,
  "setup_actions_sha256": "$setup_actions_sha256",
  "setup_command_policy": "no_publish_no_secret_assignment",
  "publish_actions_performed": false,
  "secret_value_saved": false,
  "origin_remote_configured": $(if [ -n "$remote_url" ]; then printf true; else printf false; fi),
  "github_auth_status": "$github_auth_status",
  "artifacts": {
    "index": "index.md",
    "summary": "summary.json",
    "preflight": "preflight.md",
    "verify_release": "verify-release.txt",
    "git_remote": "git-remote.txt",
    "github_auth": "github-auth.txt",
    "setup_actions": "$(if [ "$blocked_count" -gt 0 ]; then printf 'setup-actions.md'; fi)"
  }
}
JSON

{
  printf '%s\n' "# Release Readiness Evidence"
  printf '\n'
  printf '%s\n' "Status: release readiness: $overall_status"
  printf '\n'
  printf '| Check | Status | Evidence |\n'
  printf '| --- | --- | --- |\n'
  printf '| local_release_artifacts | %s | `verify-release.txt` |\n' "$verify_status"
  printf '| public_release_preflight | %s | `preflight.md` |\n' "$preflight_status"
  printf '| git_remote | %s | `git-remote.txt` |\n' "$(if [ -n "$remote_url" ]; then printf pass; else printf blocked; fi)"
  printf '| github_auth | %s | `github-auth.txt` |\n' "$github_auth_status"
  printf '\n'
  if [ "$blocked_count" -gt 0 ]; then
    printf '%s\n' "## Blocked Checks"
    printf '\n'
    while IFS= read -r check; do
      [ -n "$check" ] || continue
      printf -- '- `%s`\n' "$check"
    done <"$blocked_checks_file"
    printf '\n'
    printf '%s\n' "Setup actions: \`setup-actions.md\`"
    printf '\n'
  fi
  printf '%s\n' "## Publish Boundary"
  printf '\n'
  printf '%s\n' "This command does not tag, push, upload artifacts, publish a tap, or create a GitHub release."
  printf '%s\n' 'A public release claim is blocked until `preflight.md` reports pass.'
} >"$output_dir/index.md"

printf '%s\n' "release-readiness: wrote $output_dir/index.md"
printf '%s\n' "release-readiness: $overall_status"

if [ "$overall_status" = "pass" ]; then
  exit 0
fi
exit 1
