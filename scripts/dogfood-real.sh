#!/bin/sh
set -eu

invocation_dir=$(pwd)
root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
cd "$root"
evidence_dir="$root/.omo/evidence/dogfood-real"
repos_file=$(mktemp)
dry_run=0
timeout_ms=250
build_tmp=""

trap 'rm -f "$repos_file"; if [ -n "$build_tmp" ]; then rm -rf "$build_tmp"; fi' EXIT

usage() {
  cat <<'USAGE'
Usage: sh scripts/dogfood-real.sh [--dry-run] [--repo name:/path/to/repo] [--timeout-ms n]

Creates durable dogfood evidence under .omo/evidence/dogfood-real.

Options:
  --dry-run          List scenarios and write evidence without running commands or touching repos.
  --repo value       Repo to include. Use name:/path/to/repo or just /path/to/repo.
  --timeout-ms n     Timeout used by the hung-command probe in live mode. Default: 250.
  --help             Show this help.
USAGE
}

add_repo() {
  if [ -z "$1" ]; then
    printf '%s\n' "dogfood-real: --repo requires a value" >&2
    exit 2
  fi
  printf '%s\n' "$1" >>"$repos_file"
}

while [ "$#" -gt 0 ]; do
  case "$1" in
    --dry-run)
      dry_run=1
      shift
      ;;
    --repo)
      shift
      if [ "$#" -eq 0 ]; then
        add_repo ""
      fi
      add_repo "${1:-}"
      shift
      ;;
    --repo=*)
      add_repo "${1#--repo=}"
      shift
      ;;
    --timeout-ms)
      shift
      if [ "$#" -eq 0 ]; then
        printf '%s\n' "dogfood-real: --timeout-ms requires a value" >&2
        exit 2
      fi
      timeout_ms="${1:-}"
      shift
      ;;
    --timeout-ms=*)
      timeout_ms="${1#--timeout-ms=}"
      shift
      ;;
    --help|-h)
      usage
      exit 0
      ;;
    *)
      printf '%s\n' "dogfood-real: unknown argument: $1" >&2
      usage >&2
      exit 2
      ;;
  esac
done

case "$timeout_ms" in
  ''|*[!0-9]*)
    printf '%s\n' "dogfood-real: --timeout-ms must be a non-negative integer" >&2
    exit 2
    ;;
esac

if [ ! -s "$repos_file" ]; then
  add_repo "self:$root"
fi

mode="live"
if [ "$dry_run" -eq 1 ]; then
  mode="dry-run"
fi

mkdir -p "$evidence_dir"

archive_previous_run() {
  if [ ! -f "$evidence_dir/index.md" ]; then
    return
  fi
  archive_name=$(date -u '+%Y%m%dT%H%M%SZ')-$$
  archive_dir="$evidence_dir/_archive/$archive_name"
  mkdir -p "$archive_dir"
  for item in index.md repos build; do
    if [ -e "$evidence_dir/$item" ]; then
      cp -R "$evidence_dir/$item" "$archive_dir/$item"
    fi
  done
}

archive_previous_run
rm -rf "$evidence_dir/index.md" "$evidence_dir/repos" "$evidence_dir/build"
mkdir -p "$evidence_dir/repos"
index="$evidence_dir/index.md"
generated_at=$(date -u '+%Y-%m-%dT%H:%M:%SZ')

slugify() {
  slug=$(printf '%s' "$1" | tr -c 'A-Za-z0-9._-' '-' | sed 's/^-*//; s/-*$//')
  if [ -z "$slug" ]; then
    slug="repo"
  fi
  printf '%s' "$slug"
}

repo_name_from_spec() {
  case "$1" in
    *:*) printf '%s' "${1%%:*}" ;;
    *) basename "$1" ;;
  esac
}

repo_path_from_spec() {
  case "$1" in
    *:*) raw="${1#*:}" ;;
    *) raw="$1" ;;
  esac
  case "$raw" in
    /*) printf '%s' "$raw" ;;
    *) printf '%s/%s' "$invocation_dir" "$raw" ;;
  esac
}

write_hash() {
  file="$1"
  target="$2"
  if command -v shasum >/dev/null 2>&1; then
    shasum -a 256 "$file" | awk '{print $1}' >"$target"
  elif command -v sha256sum >/dev/null 2>&1; then
    sha256sum "$file" | awk '{print $1}' >"$target"
  else
    cksum "$file" >"$target"
  fi
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
  set +e
  "$@" >"$command_dir/stdout.txt" 2>"$command_dir/stderr.txt"
  status=$?
  set -e
  printf '%s\n' "$status" >"$command_dir/exit-code.txt"
  write_hash "$command_dir/stdout.txt" "$command_dir/stdout.sha256"
  return "$status"
}

preview_digest_from_stdout() {
  sed -n 's/.*"preview_digest": "\([^"]*\)".*/\1/p' "$1" | head -n 1
}

write_index_header() {
  {
    printf '%s\n' "# Real Repo Dogfood Evidence"
    printf '\n'
    printf '%s\n' "- Generated: $generated_at"
    printf '%s\n' "- Mode: $mode"
    printf '%s\n' "- Runner: scripts/dogfood-real.sh"
    printf '%s\n' "- Evidence root: .omo/evidence/dogfood-real"
    printf '%s\n' "- Secret API keys: not required for smoke path"
    printf '%s\n' "- Real-provider path: skipped by default; this runner uses local command/dry-run surfaces unless a repo config routes providers itself"
    printf '\n'
    printf '%s\n' "## Scenario Catalog"
    printf '\n'
    printf '%s\n' "| Scenario | Purpose | Dry-run behavior | Live evidence |"
    printf '%s\n' "| --- | --- | --- | --- |"
    printf '%s\n' "| scenario-01-doctor | Build and run the no-key doctor smoke | listed only | command output, report hash |"
    printf '%s\n' "| scenario-02-plan-only | Preview a bounded real-repo task packet | listed only | plan report, route metadata |"
    printf '%s\n' "| scenario-03-observe-run | Run CEO Harness with a local deterministic model in observe mode | listed only | JSON report, pass/fail note |"
    printf '%s\n' "| scenario-04-patch-preview | Capture a patch approval digest on a controlled fixture | listed only | preview report and digest |"
    printf '%s\n' "| scenario-05-timeout-guard | Prove hung model commands do not look successful | listed only | expected-failure transcript |"
    printf '\n'
    printf '%s\n' "## Repo Results"
    printf '\n'
    printf '%s\n' "| Repo | Status | Path | Notes |"
    printf '%s\n' "| --- | --- | --- | --- |"
  } >"$index"
}

append_repo_row() {
  printf '| %s | %s | `%s` | %s |\n' "$1" "$2" "$3" "$4" >>"$index"
}

write_dry_run_plan() {
  repo_dir="$1"
  repo_name="$2"
  repo_path="$3"
  mkdir -p "$repo_dir"
  {
    printf '%s\n' "# Dry-run Plan: $repo_name"
    printf '\n'
    printf '%s\n' "- Repo path: $repo_path"
    printf '%s\n' "- Status: planned"
    printf '%s\n' "- External repo writes: none"
    printf '\n'
    printf '%s\n' "## Planned Commands"
    printf '\n'
    printf '%s\n' "1. ceo-packet --doctor --format json"
    printf '%s\n' "2. ceo-packet --workspace \"$repo_path\" --plan-only --format json \"Plan a bounded real-repo fix\""
    printf '%s\n' "3. ceo-packet --workspace \"$repo_path\" --write-policy observe --format json --model-command sh examples/command-model.sh -- \"Inspect repo with local model\""
    printf '%s\n' "4. ceo-packet --workspace <controlled-fixture> --dry-run --replace app.txt old new --format json \"Preview patch approval digest\""
    printf '%s\n' "5. ceo-packet --workspace \"$repo_path\" --write-policy observe --model-command-timeout-ms $timeout_ms --model-command sh -c 'sleep 5' -- \"Probe timeout guard\""
  } >"$repo_dir/plan.md"
}

write_skipped_repo() {
  repo_dir="$1"
  repo_name="$2"
  repo_path="$3"
  mkdir -p "$repo_dir"
  {
    printf '%s\n' "# Skipped Repo: $repo_name"
    printf '\n'
    printf '%s\n' "- Repo path: $repo_path"
    printf '%s\n' "- Status: skipped_missing_repo"
    printf '%s\n' "- Reason: path does not exist or is not a directory"
    printf '%s\n' "- False-success guard: this is recorded as skipped, not pass"
  } >"$repo_dir/skipped.md"
}

capture_git_state() {
  repo_path="$1"
  repo_dir="$2"
  if [ -d "$repo_path/.git" ] || git -C "$repo_path" rev-parse --is-inside-work-tree >/dev/null 2>&1; then
    set +e
    git -C "$repo_path" rev-parse HEAD >"$repo_dir/git-head.txt" 2>"$repo_dir/git-head.stderr"
    git -C "$repo_path" status --short >"$repo_dir/git-status.txt" 2>"$repo_dir/git-status.stderr"
    set -e
    write_hash "$repo_dir/git-status.txt" "$repo_dir/git-status.sha256"
  else
    printf '%s\n' "not a git worktree" >"$repo_dir/git-status.txt"
    write_hash "$repo_dir/git-status.txt" "$repo_dir/git-status.sha256"
  fi
}

run_live_repo() {
  repo_name="$1"
  repo_path="$2"
  repo_dir="$3"
  bin="$4"
  mkdir -p "$repo_dir"
  capture_git_state "$repo_path" "$repo_dir"

  overall="pass"

  if run_capture "$repo_dir/scenario-01-doctor" "$bin" --doctor --format json; then
    scenario_01="pass"
  else
    scenario_01="fail"
    overall="fail"
  fi

  if run_capture "$repo_dir/scenario-02-plan-only" "$bin" --workspace "$repo_path" --plan-only --format json "Plan a bounded real-repo fix without writing files"; then
    scenario_02="pass"
  else
    scenario_02="fail"
    overall="fail"
  fi

  if run_capture "$repo_dir/scenario-03-observe-run" "$bin" --workspace "$repo_path" --write-policy observe --format json --model-command sh "$root/examples/command-model.sh" -- "Inspect the repo with a local deterministic model and no writes"; then
    scenario_03="pass"
  else
    scenario_03="fail"
    overall="fail"
  fi

  fixture="$repo_dir/patch-preview-workspace"
  mkdir -p "$fixture"
  printf '%s\n' "old" >"$fixture/app.txt"
  if run_capture "$repo_dir/scenario-04-patch-preview" "$bin" --workspace "$fixture" --dry-run --replace app.txt old new --format json "Preview controlled patch approval digest"; then
    digest=$(preview_digest_from_stdout "$repo_dir/scenario-04-patch-preview/stdout.txt")
    if [ -n "$digest" ]; then
      printf '%s\n' "$digest" >"$repo_dir/scenario-04-patch-preview/preview-digest.txt"
      scenario_04="pass"
    else
      printf '%s\n' "missing preview digest" >"$repo_dir/scenario-04-patch-preview/pass-fail-note.txt"
      scenario_04="fail"
      overall="fail"
    fi
  else
    scenario_04="fail"
    overall="fail"
  fi

  if run_capture "$repo_dir/scenario-05-timeout-guard" "$bin" --workspace "$repo_path" --write-policy observe --format json --model-command-timeout-ms "$timeout_ms" --model-command sh -c 'sleep 5' -- "Probe hung model command timeout guard"; then
    scenario_05="fail"
    overall="fail"
    printf '%s\n' "timeout probe unexpectedly exited zero" >"$repo_dir/scenario-05-timeout-guard/pass-fail-note.txt"
  else
    scenario_05="pass_expected_failure"
    printf '%s\n' "timeout probe exited non-zero as expected" >"$repo_dir/scenario-05-timeout-guard/pass-fail-note.txt"
  fi

  {
    printf '%s\n' "# Live Dogfood Summary: $repo_name"
    printf '\n'
    printf '%s\n' "- Repo path: $repo_path"
    printf '%s\n' "- Overall: $overall"
    printf '%s\n' "- Git status evidence: git-status.txt"
    printf '\n'
    printf '%s\n' "| Scenario | Status | Evidence |"
    printf '%s\n' "| --- | --- | --- |"
    printf '%s\n' "| scenario-01-doctor | $scenario_01 | scenario-01-doctor/stdout.txt |"
    printf '%s\n' "| scenario-02-plan-only | $scenario_02 | scenario-02-plan-only/stdout.txt |"
    printf '%s\n' "| scenario-03-observe-run | $scenario_03 | scenario-03-observe-run/stdout.txt |"
    printf '%s\n' "| scenario-04-patch-preview | $scenario_04 | scenario-04-patch-preview/preview-digest.txt |"
    printf '%s\n' "| scenario-05-timeout-guard | $scenario_05 | scenario-05-timeout-guard/pass-fail-note.txt |"
  } >"$repo_dir/summary.md"

  append_repo_row "$repo_name" "$overall" "$repo_path" "see repos/$(basename "$repo_dir")/summary.md"
}

write_index_header

if [ "$dry_run" -eq 0 ]; then
  mkdir -p "$evidence_dir/build"
  build_tmp=$(mktemp -d)
  bin="$build_tmp/ceo-packet"
  if ! run_capture "$evidence_dir/build" go build -trimpath -o "$bin" ./cmd/ceo-packet; then
    append_repo_row "build" "fail" "$root" "go build failed; see build/stderr.txt"
    printf '%s\n' "dogfood-real: build failed; evidence: $index" >&2
    exit 1
  fi
else
  bin=""
fi

while IFS= read -r repo_spec; do
  repo_name=$(repo_name_from_spec "$repo_spec")
  repo_path=$(repo_path_from_spec "$repo_spec")
  repo_slug=$(slugify "$repo_name")
  repo_dir="$evidence_dir/repos/$repo_slug"

  if [ ! -d "$repo_path" ]; then
    write_skipped_repo "$repo_dir" "$repo_name" "$repo_path"
    append_repo_row "$repo_name" "skipped_missing_repo" "$repo_path" "path missing; no commands run"
    continue
  fi

  if [ "$dry_run" -eq 1 ]; then
    write_dry_run_plan "$repo_dir" "$repo_name" "$repo_path"
    append_repo_row "$repo_name" "planned" "$repo_path" "dry-run only; no commands run"
  else
    run_live_repo "$repo_name" "$repo_path" "$repo_dir" "$bin"
  fi
done <"$repos_file"

{
  printf '\n'
  printf '%s\n' "## Adversarial Coverage"
  printf '\n'
  printf '%s\n' "- stale_state: live mode captures git HEAD and git status hashes before repo scenarios; dry-run records this as planned only."
  printf '%s\n' "- misleading_success_output: missing repos are recorded as skipped_missing_repo, and timeout probes must exit non-zero to pass."
  printf '%s\n' "- dirty_worktree: live mode saves git-status.txt and git-status.sha256 for review; dirty status is evidence, not an automatic pass."
  printf '%s\n' "- hung/long commands: live mode runs scenario-05-timeout-guard with --model-command-timeout-ms $timeout_ms."
} >>"$index"

printf '%s\n' "dogfood-real: mode=$mode"
printf '%s\n' "dogfood-real: evidence=$index"
