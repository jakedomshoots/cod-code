#!/bin/sh
set -eu

invocation_dir=$(pwd)
root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
cd "$root"
evidence_dir="$root/.omo/evidence/dogfood-real"
repos_file=$(mktemp)
dry_run=0
timeout_ms=250
repeat_count=1
copy_workspace=0
write_probe=0
feature_edit_probe=0
app_code_probe=0
integrated_app_code_probe=0
multi_file_app_code_probe=0
build_tmp=""
task_text="Plan a bounded real-repo fix without writing files"

trap 'rm -f "$repos_file"; if [ -n "$build_tmp" ]; then rm -rf "$build_tmp"; fi' EXIT

usage() {
  cat <<'USAGE'
Usage: sh scripts/dogfood-real.sh [--dry-run] [--repo name:/path/to/repo] [--task text] [--timeout-ms n] [--repeat n] [--output-dir path] [--copy-workspace] [--write-probe] [--feature-edit-probe] [--app-code-probe] [--integrated-app-code-probe] [--multi-file-app-code-probe]

Creates durable dogfood evidence under .omo/evidence/dogfood-real.

Options:
  --dry-run          List scenarios and write evidence without running commands or touching repos.
  --repo value       Repo to include. Use name:/path/to/repo or just /path/to/repo.
  --task text        Task text used by plan/model scenarios. Default is a bounded fix plan.
  --timeout-ms n     Timeout used by the hung-command probe in live mode. Default: 250.
  --repeat n         Repeat each repo scenario set n times. Default: 1.
  --output-dir path  Evidence directory. Default: .omo/evidence/dogfood-real.
  --copy-workspace   Run live scenarios against a copied workspace instead of the source repo.
  --write-probe      In live copied-workspace mode, preview and approve one real write against the copy.
  --feature-edit-probe
                     In live copied-workspace mode, preview and approve a repo-specific feature note.
  --app-code-probe   In live copied-workspace mode, preview and approve a source-code module edit.
  --integrated-app-code-probe
                     In live copied-workspace mode, preview and approve an edit to an existing source file.
  --multi-file-app-code-probe
                     In live copied-workspace mode, preview and approve edits to two existing source files.
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
    --task)
      shift
      if [ "$#" -eq 0 ] || [ -z "${1:-}" ]; then
        printf '%s\n' "dogfood-real: --task requires a non-empty value" >&2
        exit 2
      fi
      task_text="${1:-}"
      shift
      ;;
    --task=*)
      task_text="${1#--task=}"
      if [ -z "$task_text" ]; then
        printf '%s\n' "dogfood-real: --task requires a non-empty value" >&2
        exit 2
      fi
      shift
      ;;
    --repeat)
      shift
      if [ "$#" -eq 0 ]; then
        printf '%s\n' "dogfood-real: --repeat requires a value" >&2
        exit 2
      fi
      repeat_count="${1:-}"
      shift
      ;;
    --repeat=*)
      repeat_count="${1#--repeat=}"
      shift
      ;;
    --output-dir)
      shift
      if [ "$#" -eq 0 ]; then
        printf '%s\n' "dogfood-real: --output-dir requires a value" >&2
        exit 2
      fi
      evidence_dir="${1:-}"
      shift
      ;;
    --output-dir=*)
      evidence_dir="${1#--output-dir=}"
      shift
      ;;
    --copy-workspace)
      copy_workspace=1
      shift
      ;;
    --write-probe)
      write_probe=1
      shift
      ;;
    --feature-edit-probe)
      feature_edit_probe=1
      shift
      ;;
    --app-code-probe)
      app_code_probe=1
      shift
      ;;
    --integrated-app-code-probe)
      integrated_app_code_probe=1
      shift
      ;;
    --multi-file-app-code-probe)
      multi_file_app_code_probe=1
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

case "$repeat_count" in
  ''|*[!0-9]*|0)
    printf '%s\n' "dogfood-real: --repeat must be a positive integer" >&2
    exit 2
    ;;
esac

case "$evidence_dir" in
  /*) ;;
  *) evidence_dir="$invocation_dir/$evidence_dir" ;;
esac

if [ "$write_probe" -eq 1 ] && [ "$copy_workspace" -eq 0 ]; then
  printf '%s\n' "dogfood-real: --write-probe requires --copy-workspace" >&2
  exit 2
fi

if [ "$feature_edit_probe" -eq 1 ] && [ "$copy_workspace" -eq 0 ]; then
  printf '%s\n' "dogfood-real: --feature-edit-probe requires --copy-workspace" >&2
  exit 2
fi

if [ "$app_code_probe" -eq 1 ] && [ "$copy_workspace" -eq 0 ]; then
  printf '%s\n' "dogfood-real: --app-code-probe requires --copy-workspace" >&2
  exit 2
fi

if [ "$integrated_app_code_probe" -eq 1 ] && [ "$copy_workspace" -eq 0 ]; then
  printf '%s\n' "dogfood-real: --integrated-app-code-probe requires --copy-workspace" >&2
  exit 2
fi

if [ "$multi_file_app_code_probe" -eq 1 ] && [ "$copy_workspace" -eq 0 ]; then
  printf '%s\n' "dogfood-real: --multi-file-app-code-probe requires --copy-workspace" >&2
  exit 2
fi

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

attempt_slug() {
  printf 'run-%02d' "$1"
}

evidence_rel_path() {
  case "$1" in
    "$evidence_dir"/*) printf '%s' "${1#"$evidence_dir"/}" ;;
    *) basename "$1" ;;
  esac
}

evidence_display_path() {
  case "$evidence_dir" in
    "$root"/*) printf '%s' "${evidence_dir#"$root"/}" ;;
    *) printf '%s' "$evidence_dir" ;;
  esac
}

workspace_mode() {
  if [ "$copy_workspace" -eq 1 ]; then
    printf '%s' "copied"
  else
    printf '%s' "source"
  fi
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

prepare_repo_workspace() {
  source_path="$1"
  evidence_repo_dir="$2"
  mkdir -p "$evidence_repo_dir"
  printf '%s\n' "$source_path" >"$evidence_repo_dir/source-path.txt"
  if [ "$copy_workspace" -eq 0 ]; then
    printf '%s\n' "source" >"$evidence_repo_dir/workspace-mode.txt"
    printf '%s\n' "$source_path" >"$evidence_repo_dir/workspace-path.txt"
    printf '%s\n' "$source_path"
    return 0
  fi

  copy_path="$evidence_repo_dir/workspace-copy"
  rm -rf "$copy_path"
  if git -C "$source_path" rev-parse --is-inside-work-tree >/dev/null 2>&1; then
    source_top=$(git -C "$source_path" rev-parse --show-toplevel)
    source_prefix=$(git -C "$source_path" rev-parse --show-prefix)
    git clone --quiet --no-hardlinks "$source_top" "$copy_path"
    if [ -n "$source_prefix" ]; then
      copy_path="$copy_path/${source_prefix%/}"
    fi
  else
    case "$evidence_repo_dir" in
      "$source_path"/*)
        printf '%s\n' "dogfood-real: --copy-workspace for non-git repos needs --output-dir outside the source path" >&2
        return 1
        ;;
    esac
    mkdir -p "$copy_path"
    cp -R "$source_path/." "$copy_path"
  fi
  printf '%s\n' "copied" >"$evidence_repo_dir/workspace-mode.txt"
  printf '%s\n' "$copy_path" >"$evidence_repo_dir/workspace-path.txt"
  printf '%s\n' "$copy_path"
}

write_index_header() {
  {
    printf '%s\n' "# Real Repo Dogfood Evidence"
    printf '\n'
    printf '%s\n' "- Generated: $generated_at"
    printf '%s\n' "- Mode: $mode"
    printf '%s\n' "- Repeat count: $repeat_count"
    printf '%s\n' "- Workspace mode: $(workspace_mode)"
    printf '%s\n' "- Task: $task_text"
    if [ "$write_probe" -eq 1 ]; then
      printf '%s\n' "- Write probe: enabled"
    else
      printf '%s\n' "- Write probe: disabled"
    fi
    if [ "$feature_edit_probe" -eq 1 ]; then
      printf '%s\n' "- Feature edit probe: enabled"
    else
      printf '%s\n' "- Feature edit probe: disabled"
    fi
    if [ "$app_code_probe" -eq 1 ]; then
      printf '%s\n' "- App-code probe: enabled"
    else
      printf '%s\n' "- App-code probe: disabled"
    fi
    if [ "$integrated_app_code_probe" -eq 1 ]; then
      printf '%s\n' "- Integrated app-code probe: enabled"
    else
      printf '%s\n' "- Integrated app-code probe: disabled"
    fi
    if [ "$multi_file_app_code_probe" -eq 1 ]; then
      printf '%s\n' "- Multi-file app-code probe: enabled"
    else
      printf '%s\n' "- Multi-file app-code probe: disabled"
    fi
    printf '%s\n' "- Runner: scripts/dogfood-real.sh"
    printf '%s\n' "- Evidence root: $(evidence_display_path)"
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
    printf '%s\n' "| scenario-06-write-probe | Prove preview plus approved write mutates only the copied workspace | listed only | preview digest, apply report, after-state git status |"
    printf '%s\n' "| scenario-07-feature-edit-probe | Prove a repo-specific feature note can be previewed and approved in a copied workspace | listed only | feature file, preview digest, after-state git status |"
    printf '%s\n' "| scenario-08-app-code-probe | Prove a source-code module can be previewed and approved in a copied workspace | listed only | app-code file, preview digest, after-state git status |"
    printf '%s\n' "| scenario-09-integrated-app-code-probe | Prove an existing source file can be previewed and approved in a copied workspace | listed only | target path, modified source file, preview digest, after-state git status |"
    printf '%s\n' "| scenario-10-multi-file-app-code-probe | Prove two existing source files can be previewed and approved in a copied workspace | listed only | target paths, modified source files, preview digests, after-state git status |"
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
  plan_repo_dir="$1"
  plan_repo_name="$2"
  plan_repo_path="$3"
  mkdir -p "$plan_repo_dir"
  {
    printf '%s\n' "# Dry-run Plan: $plan_repo_name"
    printf '\n'
    printf '%s\n' "- Repo path: $plan_repo_path"
    printf '%s\n' "- Status: planned"
    printf '%s\n' "- Workspace mode: $(workspace_mode)"
    printf '%s\n' "- External repo writes: none"
    printf '\n'
    printf '%s\n' "## Planned Commands"
    printf '\n'
    printf '%s\n' "1. ceo-packet --doctor --format json"
    printf '%s\n' "2. ceo-packet --workspace \"$plan_repo_path\" --plan-only --format json \"$task_text\""
    printf '%s\n' "3. ceo-packet --workspace \"$plan_repo_path\" --write-policy observe --format json --model-command sh examples/command-model.sh -- \"$task_text\""
    printf '%s\n' "4. ceo-packet --workspace <controlled-fixture> --dry-run --replace app.txt old new --format json \"Preview patch approval digest\""
    printf '%s\n' "5. ceo-packet --workspace \"$plan_repo_path\" --write-policy observe --model-command-timeout-ms $timeout_ms --model-command sh -c 'sleep 5' -- \"Probe timeout guard\""
    if [ "$write_probe" -eq 1 ]; then
      printf '%s\n' "6. ceo-packet --workspace <copied-workspace> --write-policy approved-write --approve-preview <digest> --replace ceo-dogfood-write-probe.txt old new --format json \"Apply copied workspace write probe\""
    fi
    if [ "$feature_edit_probe" -eq 1 ]; then
      printf '%s\n' "7. ceo-packet --workspace <copied-workspace> --write-policy approved-write --approve-preview <digest> --replace ceo-dogfood-feature.md <seed> <repo-specific-note> --format json \"Apply copied workspace feature edit probe\""
    fi
    if [ "$app_code_probe" -eq 1 ]; then
      printf '%s\n' "8. ceo-packet --workspace <copied-workspace> --write-policy approved-write --approve-preview <digest> --replace src/ceoDogfoodProbe.mjs <seed> <source-module> --format json \"Apply copied workspace app-code probe\""
    fi
    if [ "$integrated_app_code_probe" -eq 1 ]; then
      printf '%s\n' "9. ceo-packet --workspace <copied-workspace> --write-policy approved-write --approve-preview <digest> --replace <existing-source-file> <old> <old-plus-export> --format json \"Apply copied workspace integrated app-code probe\""
    fi
    if [ "$multi_file_app_code_probe" -eq 1 ]; then
      printf '%s\n' "10. ceo-packet --workspace <copied-workspace> --write-policy approved-write --approve-preview <digest> --replace <existing-source-file-a> <old> <old-plus-export> --format json \"Apply copied workspace multi-file app-code probe\""
      printf '%s\n' "11. ceo-packet --workspace <copied-workspace> --write-policy approved-write --approve-preview <digest> --replace <existing-source-file-b> <old> <old-plus-export> --format json \"Apply copied workspace multi-file app-code probe\""
    fi
  } >"$plan_repo_dir/plan.md"
  printf '%s\n' "planned" >"$plan_repo_dir/status.txt"
}

write_skipped_repo() {
  skipped_repo_dir="$1"
  skipped_repo_name="$2"
  skipped_repo_path="$3"
  mkdir -p "$skipped_repo_dir"
  {
    printf '%s\n' "# Skipped Repo: $skipped_repo_name"
    printf '\n'
    printf '%s\n' "- Repo path: $skipped_repo_path"
    printf '%s\n' "- Status: skipped_missing_repo"
    printf '%s\n' "- Reason: path does not exist or is not a directory"
    printf '%s\n' "- False-success guard: this is recorded as skipped, not pass"
  } >"$skipped_repo_dir/skipped.md"
  printf '%s\n' "skipped_missing_repo" >"$skipped_repo_dir/status.txt"
}

write_repeat_summary() {
  summary_repo_dir="$1"
  summary_repo_name="$2"
  summary_repo_path="$3"
  mkdir -p "$summary_repo_dir"
  {
    printf '%s\n' "# Repeated Dogfood Summary: $summary_repo_name"
    printf '\n'
    printf '%s\n' "- Repo path: $summary_repo_path"
    printf '%s\n' "- Repeat count: $repeat_count"
    printf '\n'
    printf '%s\n' "| Run | Status | Evidence |"
    printf '%s\n' "| --- | --- | --- |"
    attempt=1
    while [ "$attempt" -le "$repeat_count" ]; do
      run_slug=$(attempt_slug "$attempt")
      summary_run_dir="$summary_repo_dir/$run_slug"
      status="missing"
      if [ -f "$summary_run_dir/status.txt" ]; then
        status=$(sed -n '1p' "$summary_run_dir/status.txt")
      fi
      evidence="plan.md"
      if [ -f "$summary_run_dir/summary.md" ]; then
        evidence="summary.md"
      fi
      printf '| %s | %s | %s/%s |\n' "$run_slug" "$status" "$run_slug" "$evidence"
      attempt=$((attempt + 1))
    done
  } >"$summary_repo_dir/summary.md"
}

capture_git_state() {
  git_repo_path="$1"
  git_repo_dir="$2"
  if [ -d "$git_repo_path/.git" ] || git -C "$git_repo_path" rev-parse --is-inside-work-tree >/dev/null 2>&1; then
    set +e
    git -C "$git_repo_path" rev-parse HEAD >"$git_repo_dir/git-head.txt" 2>"$git_repo_dir/git-head.stderr"
    git -C "$git_repo_path" status --short >"$git_repo_dir/git-status.txt" 2>"$git_repo_dir/git-status.stderr"
    set -e
    write_hash "$git_repo_dir/git-status.txt" "$git_repo_dir/git-status.sha256"
  else
    printf '%s\n' "not a git worktree" >"$git_repo_dir/git-status.txt"
    write_hash "$git_repo_dir/git-status.txt" "$git_repo_dir/git-status.sha256"
  fi
}

capture_git_status_only() {
  git_repo_path="$1"
  target_file="$2"
  if [ -d "$git_repo_path/.git" ] || git -C "$git_repo_path" rev-parse --is-inside-work-tree >/dev/null 2>&1; then
    git -C "$git_repo_path" status --short >"$target_file" 2>"$target_file.stderr"
  else
    printf '%s\n' "not a git worktree" >"$target_file"
  fi
  write_hash "$target_file" "$target_file.sha256"
}

run_write_probe() {
  probe_repo_path="$1"
  probe_dir="$2"
  probe_bin="$3"
  mkdir -p "$probe_dir"
  probe_file="$probe_repo_path/ceo-dogfood-write-probe.txt"
  printf '%s\n' "old" >"$probe_file"

  if ! run_capture "$probe_dir/preview" "$probe_bin" --workspace "$probe_repo_path" --dry-run --replace ceo-dogfood-write-probe.txt old new --format json "Preview copied workspace write probe"; then
    printf '%s\n' "write probe preview failed" >"$probe_dir/pass-fail-note.txt"
    return 1
  fi

  digest=$(preview_digest_from_stdout "$probe_dir/preview/stdout.txt")
  if [ -z "$digest" ]; then
    printf '%s\n' "missing preview digest" >"$probe_dir/pass-fail-note.txt"
    return 1
  fi
  printf '%s\n' "$digest" >"$probe_dir/preview-digest.txt"

  if ! run_capture "$probe_dir/apply" "$probe_bin" --workspace "$probe_repo_path" --write-policy approved-write --approve-preview "$digest" --replace ceo-dogfood-write-probe.txt old new --format json "Apply copied workspace write probe"; then
    printf '%s\n' "write probe approved apply failed" >"$probe_dir/pass-fail-note.txt"
    return 1
  fi

  if [ "$(cat "$probe_file")" != "new" ]; then
    printf '%s\n' "write probe file content did not change to expected value" >"$probe_dir/pass-fail-note.txt"
    return 1
  fi

  capture_git_status_only "$probe_repo_path" "$probe_dir/git-status-after.txt"
  printf '%s\n' "approved write changed copied workspace only" >"$probe_dir/pass-fail-note.txt"
  return 0
}

run_feature_edit_probe() {
  feature_repo_name="$1"
  feature_task_text="$2"
  feature_repo_path="$3"
  feature_dir="$4"
  feature_bin="$5"
  mkdir -p "$feature_dir"
  feature_file="$feature_repo_path/ceo-dogfood-feature.md"
  old_content="pending feature probe"
  new_content=$(printf '%s\n\n%s\n%s\n%s' "# CEO Dogfood Feature Probe" "- Repo: $feature_repo_name" "- Task: $feature_task_text" "- Result: approved copied-workspace feature edit")
  printf '%s\n' "$old_content" >"$feature_file"

  if ! run_capture "$feature_dir/preview" "$feature_bin" --workspace "$feature_repo_path" --dry-run --replace ceo-dogfood-feature.md "$old_content" "$new_content" --format json "Preview copied workspace feature edit probe"; then
    printf '%s\n' "feature edit probe preview failed" >"$feature_dir/pass-fail-note.txt"
    return 1
  fi

  digest=$(preview_digest_from_stdout "$feature_dir/preview/stdout.txt")
  if [ -z "$digest" ]; then
    printf '%s\n' "missing preview digest" >"$feature_dir/pass-fail-note.txt"
    return 1
  fi
  printf '%s\n' "$digest" >"$feature_dir/preview-digest.txt"

  if ! run_capture "$feature_dir/apply" "$feature_bin" --workspace "$feature_repo_path" --write-policy approved-write --approve-preview "$digest" --replace ceo-dogfood-feature.md "$old_content" "$new_content" --format json "Apply copied workspace feature edit probe"; then
    printf '%s\n' "feature edit probe approved apply failed" >"$feature_dir/pass-fail-note.txt"
    return 1
  fi

  if ! grep -F -q "# CEO Dogfood Feature Probe" "$feature_file" || ! grep -F -q -- "- Task: $feature_task_text" "$feature_file"; then
    printf '%s\n' "feature edit probe file content did not include expected task note" >"$feature_dir/pass-fail-note.txt"
    return 1
  fi

  capture_git_status_only "$feature_repo_path" "$feature_dir/git-status-after.txt"
  cp "$feature_file" "$feature_dir/feature-file.md"
  printf '%s\n' "approved feature edit changed copied workspace only" >"$feature_dir/pass-fail-note.txt"
  return 0
}

js_string_escape() {
  printf '%s' "$1" | sed 's/\\/\\\\/g; s/"/\\"/g'
}

run_app_code_probe() {
  app_repo_name="$1"
  app_task_text="$2"
  app_repo_path="$3"
  app_dir="$4"
  app_bin="$5"
  mkdir -p "$app_dir"
  src_dir="$app_repo_path/src"
  mkdir -p "$src_dir"
  app_file="$src_dir/ceoDogfoodProbe.mjs"
  old_content="export const ceoDogfoodProbeStatus = \"pending\";"
  repo_js=$(js_string_escape "$app_repo_name")
  task_js=$(js_string_escape "$app_task_text")
  new_content=$(printf '%s\n' \
    "export function ceoDogfoodProbe() {" \
    "  return {" \
    "    repo: \"$repo_js\"," \
    "    task: \"$task_js\"," \
    "    result: \"approved copied-workspace app-code edit\"" \
    "  };" \
    "}")
  printf '%s\n' "$old_content" >"$app_file"

  if ! run_capture "$app_dir/preview" "$app_bin" --workspace "$app_repo_path" --dry-run --replace src/ceoDogfoodProbe.mjs "$old_content" "$new_content" --format json "Preview copied workspace app-code probe"; then
    printf '%s\n' "app-code probe preview failed" >"$app_dir/pass-fail-note.txt"
    return 1
  fi

  digest=$(preview_digest_from_stdout "$app_dir/preview/stdout.txt")
  if [ -z "$digest" ]; then
    printf '%s\n' "missing preview digest" >"$app_dir/pass-fail-note.txt"
    return 1
  fi
  printf '%s\n' "$digest" >"$app_dir/preview-digest.txt"

  if ! run_capture "$app_dir/apply" "$app_bin" --workspace "$app_repo_path" --write-policy approved-write --approve-preview "$digest" --replace src/ceoDogfoodProbe.mjs "$old_content" "$new_content" --format json "Apply copied workspace app-code probe"; then
    printf '%s\n' "app-code probe approved apply failed" >"$app_dir/pass-fail-note.txt"
    return 1
  fi

  if ! grep -F -q "export function ceoDogfoodProbe()" "$app_file" || ! grep -F -q "approved copied-workspace app-code edit" "$app_file"; then
    printf '%s\n' "app-code probe file content did not include expected module" >"$app_dir/pass-fail-note.txt"
    return 1
  fi

  if command -v node >/dev/null 2>&1; then
    if ! node --check "$app_file" >"$app_dir/syntax-check.txt" 2>"$app_dir/syntax-check.stderr"; then
      printf '%s\n' "app-code probe syntax check failed" >"$app_dir/pass-fail-note.txt"
      return 1
    fi
  else
    printf '%s\n' "node unavailable; syntax check skipped" >"$app_dir/syntax-check.txt"
  fi

  capture_git_status_only "$app_repo_path" "$app_dir/git-status-after.txt"
  cp "$app_file" "$app_dir/app-code-file.mjs"
  printf '%s\n' "approved app-code edit changed copied workspace only" >"$app_dir/pass-fail-note.txt"
  return 0
}

select_integrated_source_file() {
  integrated_repo_path="$1"
  for candidate in \
    "$integrated_repo_path/src/App.jsx" \
    "$integrated_repo_path/src/App.tsx" \
    "$integrated_repo_path/src/main.jsx" \
    "$integrated_repo_path/src/main.tsx" \
    "$integrated_repo_path/src/App.js" \
    "$integrated_repo_path/src/main.js"; do
    if [ -f "$candidate" ]; then
      printf '%s\n' "$candidate"
      return 0
    fi
  done
  if [ -d "$integrated_repo_path/src" ]; then
    find "$integrated_repo_path/src" -type f \( -name '*.js' -o -name '*.jsx' -o -name '*.ts' -o -name '*.tsx' \) \
      | grep -Ev '/(__tests__|test)/|[._](test|spec)\.' \
      | sort \
      | sed -n '1p'
  fi
}

run_integrated_app_code_probe() {
  integrated_repo_name="$1"
  integrated_task_text="$2"
  integrated_repo_path="$3"
  integrated_dir="$4"
  integrated_bin="$5"
  mkdir -p "$integrated_dir"
  target_file=$(select_integrated_source_file "$integrated_repo_path")
  if [ -z "$target_file" ] || [ ! -f "$target_file" ]; then
    printf '%s\n' "no source file found under src for integrated app-code probe" >"$integrated_dir/pass-fail-note.txt"
    return 1
  fi
  target_rel=${target_file#"$integrated_repo_path"/}
  printf '%s\n' "$target_rel" >"$integrated_dir/target-file.txt"

  old_content=$(cat "$target_file")
  repo_js=$(js_string_escape "$integrated_repo_name")
  task_js=$(js_string_escape "$integrated_task_text")
  marker=$(printf '%s\n' \
    "export const ceoDogfoodIntegratedProbe = {" \
    "  repo: \"$repo_js\"," \
    "  task: \"$task_js\"," \
    "  result: \"approved integrated copied-workspace app-code edit\"" \
    "};")
  new_content=$(printf '%s\n\n%s' "$old_content" "$marker")

  if ! run_capture "$integrated_dir/preview" "$integrated_bin" --workspace "$integrated_repo_path" --dry-run --replace "$target_rel" "$old_content" "$new_content" --format json "Preview copied workspace integrated app-code probe"; then
    printf '%s\n' "integrated app-code probe preview failed" >"$integrated_dir/pass-fail-note.txt"
    return 1
  fi

  digest=$(preview_digest_from_stdout "$integrated_dir/preview/stdout.txt")
  if [ -z "$digest" ]; then
    printf '%s\n' "missing preview digest" >"$integrated_dir/pass-fail-note.txt"
    return 1
  fi
  printf '%s\n' "$digest" >"$integrated_dir/preview-digest.txt"

  if ! run_capture "$integrated_dir/apply" "$integrated_bin" --workspace "$integrated_repo_path" --write-policy approved-write --approve-preview "$digest" --replace "$target_rel" "$old_content" "$new_content" --format json "Apply copied workspace integrated app-code probe"; then
    printf '%s\n' "integrated app-code probe approved apply failed" >"$integrated_dir/pass-fail-note.txt"
    return 1
  fi

  if ! grep -F -q "ceoDogfoodIntegratedProbe" "$target_file" || ! grep -F -q "approved integrated copied-workspace app-code edit" "$target_file"; then
    printf '%s\n' "integrated app-code probe target file did not include expected marker" >"$integrated_dir/pass-fail-note.txt"
    return 1
  fi

  capture_git_status_only "$integrated_repo_path" "$integrated_dir/git-status-after.txt"
  cp "$target_file" "$integrated_dir/integrated-source-file.txt"
  printf '%s\n' "approved integrated app-code edit changed copied workspace only" >"$integrated_dir/pass-fail-note.txt"
  return 0
}

select_multi_file_source_files() {
  multi_repo_path="$1"
  {
    for candidate in \
      "$multi_repo_path/src/App.jsx" \
      "$multi_repo_path/src/main.jsx" \
      "$multi_repo_path/src/App.tsx" \
      "$multi_repo_path/src/main.tsx" \
      "$multi_repo_path/src/App.js" \
      "$multi_repo_path/src/main.js" \
      "$multi_repo_path/src/App.ts" \
      "$multi_repo_path/src/main.ts"; do
      if [ -f "$candidate" ]; then
        printf '%s\n' "$candidate"
      fi
    done
    if [ -d "$multi_repo_path/src" ]; then
      find "$multi_repo_path/src" -type f \( -name '*.js' -o -name '*.jsx' -o -name '*.ts' -o -name '*.tsx' \) \
        | grep -Ev '/(__tests__|test)/|[._](test|spec)\.' \
        | sort
    fi
  } | awk '!seen[$0]++' | sed -n '1,2p'
}

artifact_name_for_relpath() {
  printf '%s' "$1" | tr '/ ' '--' | tr -c 'A-Za-z0-9._-' '-'
}

run_multi_file_single_app_code_edit() {
  multi_repo_name="$1"
  multi_task_text="$2"
  multi_repo_path="$3"
  multi_dir="$4"
  multi_bin="$5"
  target_file="$6"
  variable_name="$7"
  label="$8"
  target_rel=${target_file#"$multi_repo_path"/}
  old_content=$(cat "$target_file")
  repo_js=$(js_string_escape "$multi_repo_name")
  task_js=$(js_string_escape "$multi_task_text")
  marker=$(printf '%s\n' \
    "export const $variable_name = {" \
    "  repo: \"$repo_js\"," \
    "  task: \"$task_js\"," \
    "  result: \"approved multi-file copied-workspace app-code edit\"" \
    "};")
  new_content=$(printf '%s\n\n%s' "$old_content" "$marker")

  if ! run_capture "$multi_dir/preview-$label" "$multi_bin" --workspace "$multi_repo_path" --dry-run --replace "$target_rel" "$old_content" "$new_content" --format json "Preview copied workspace multi-file app-code probe"; then
    printf '%s\n' "multi-file app-code probe preview failed for $target_rel" >"$multi_dir/pass-fail-note.txt"
    return 1
  fi

  digest=$(preview_digest_from_stdout "$multi_dir/preview-$label/stdout.txt")
  if [ -z "$digest" ]; then
    printf '%s\n' "missing preview digest for $target_rel" >"$multi_dir/pass-fail-note.txt"
    return 1
  fi
  printf '%s %s\n' "$target_rel" "$digest" >>"$multi_dir/preview-digests.txt"

  if ! run_capture "$multi_dir/apply-$label" "$multi_bin" --workspace "$multi_repo_path" --write-policy approved-write --approve-preview "$digest" --replace "$target_rel" "$old_content" "$new_content" --format json "Apply copied workspace multi-file app-code probe"; then
    printf '%s\n' "multi-file app-code probe approved apply failed for $target_rel" >"$multi_dir/pass-fail-note.txt"
    return 1
  fi

  if ! grep -F -q "$variable_name" "$target_file" || ! grep -F -q "approved multi-file copied-workspace app-code edit" "$target_file"; then
    printf '%s\n' "multi-file app-code target file did not include expected marker for $target_rel" >"$multi_dir/pass-fail-note.txt"
    return 1
  fi

  artifact_name=$(artifact_name_for_relpath "$target_rel")
  cp "$target_file" "$multi_dir/modified-source-files/$artifact_name.txt"
  return 0
}

run_multi_file_app_code_probe() {
  multi_repo_name="$1"
  multi_task_text="$2"
  multi_repo_path="$3"
  multi_dir="$4"
  multi_bin="$5"
  mkdir -p "$multi_dir/modified-source-files"
  targets_tmp="$multi_dir/selected-targets.txt"
  select_multi_file_source_files "$multi_repo_path" >"$targets_tmp"
  target_count=$(wc -l <"$targets_tmp" | tr -d ' ')
  if [ "$target_count" -lt 2 ]; then
    printf '%s\n' "fewer than two source files found under src for multi-file app-code probe" >"$multi_dir/pass-fail-note.txt"
    return 1
  fi

  target_one=$(sed -n '1p' "$targets_tmp")
  target_two=$(sed -n '2p' "$targets_tmp")
  target_one_rel=${target_one#"$multi_repo_path"/}
  target_two_rel=${target_two#"$multi_repo_path"/}
  {
    printf '%s\n' "$target_one_rel"
    printf '%s\n' "$target_two_rel"
  } >"$multi_dir/target-files.txt"
  : >"$multi_dir/preview-digests.txt"

  if ! run_multi_file_single_app_code_edit "$multi_repo_name" "$multi_task_text" "$multi_repo_path" "$multi_dir" "$multi_bin" "$target_one" "ceoDogfoodMultiFileProbeApp" "one"; then
    return 1
  fi
  if ! run_multi_file_single_app_code_edit "$multi_repo_name" "$multi_task_text" "$multi_repo_path" "$multi_dir" "$multi_bin" "$target_two" "ceoDogfoodMultiFileProbeEntry" "two"; then
    return 1
  fi

  capture_git_status_only "$multi_repo_path" "$multi_dir/git-status-after.txt"
  printf '%s\n' "approved multi-file app-code edit changed copied workspace only" >"$multi_dir/pass-fail-note.txt"
  return 0
}

run_live_repo() {
  live_repo_name="$1"
  live_repo_path="$2"
  live_repo_dir="$3"
  live_bin="$4"
  mkdir -p "$live_repo_dir"
  capture_git_state "$live_repo_path" "$live_repo_dir"

  overall="pass"

  if run_capture "$live_repo_dir/scenario-01-doctor" "$live_bin" --doctor --format json; then
    scenario_01="pass"
  else
    scenario_01="fail"
    overall="fail"
  fi

  if run_capture "$live_repo_dir/scenario-02-plan-only" "$live_bin" --workspace "$live_repo_path" --plan-only --format json "$task_text"; then
    scenario_02="pass"
  else
    scenario_02="fail"
    overall="fail"
  fi

  if run_capture "$live_repo_dir/scenario-03-observe-run" "$live_bin" --workspace "$live_repo_path" --write-policy observe --format json --model-command sh "$root/examples/command-model.sh" -- "$task_text"; then
    scenario_03="pass"
  else
    scenario_03="fail"
    overall="fail"
  fi

  fixture="$live_repo_dir/patch-preview-workspace"
  mkdir -p "$fixture"
  printf '%s\n' "old" >"$fixture/app.txt"
  if run_capture "$live_repo_dir/scenario-04-patch-preview" "$live_bin" --workspace "$fixture" --dry-run --replace app.txt old new --format json "Preview controlled patch approval digest"; then
    digest=$(preview_digest_from_stdout "$live_repo_dir/scenario-04-patch-preview/stdout.txt")
    if [ -n "$digest" ]; then
      printf '%s\n' "$digest" >"$live_repo_dir/scenario-04-patch-preview/preview-digest.txt"
      scenario_04="pass"
    else
      printf '%s\n' "missing preview digest" >"$live_repo_dir/scenario-04-patch-preview/pass-fail-note.txt"
      scenario_04="fail"
      overall="fail"
    fi
  else
    scenario_04="fail"
    overall="fail"
  fi

  if run_capture "$live_repo_dir/scenario-05-timeout-guard" "$live_bin" --workspace "$live_repo_path" --write-policy observe --format json --model-command-timeout-ms "$timeout_ms" --model-command sh -c 'sleep 5' -- "Probe hung model command timeout guard"; then
    scenario_05="fail"
    overall="fail"
    printf '%s\n' "timeout probe unexpectedly exited zero" >"$live_repo_dir/scenario-05-timeout-guard/pass-fail-note.txt"
  else
    scenario_05="pass_expected_failure"
    printf '%s\n' "timeout probe exited non-zero as expected" >"$live_repo_dir/scenario-05-timeout-guard/pass-fail-note.txt"
  fi

  scenario_06="skipped_disabled"
  if [ "$write_probe" -eq 1 ]; then
    if run_write_probe "$live_repo_path" "$live_repo_dir/scenario-06-write-probe" "$live_bin"; then
      scenario_06="pass"
    else
      scenario_06="fail"
      overall="fail"
    fi
  fi

  scenario_07="skipped_disabled"
  if [ "$feature_edit_probe" -eq 1 ]; then
    if run_feature_edit_probe "$live_repo_name" "$task_text" "$live_repo_path" "$live_repo_dir/scenario-07-feature-edit-probe" "$live_bin"; then
      scenario_07="pass"
    else
      scenario_07="fail"
      overall="fail"
    fi
  fi

  scenario_08="skipped_disabled"
  if [ "$app_code_probe" -eq 1 ]; then
    if run_app_code_probe "$live_repo_name" "$task_text" "$live_repo_path" "$live_repo_dir/scenario-08-app-code-probe" "$live_bin"; then
      scenario_08="pass"
    else
      scenario_08="fail"
      overall="fail"
    fi
  fi

  scenario_09="skipped_disabled"
  if [ "$integrated_app_code_probe" -eq 1 ]; then
    if run_integrated_app_code_probe "$live_repo_name" "$task_text" "$live_repo_path" "$live_repo_dir/scenario-09-integrated-app-code-probe" "$live_bin"; then
      scenario_09="pass"
    else
      scenario_09="fail"
      overall="fail"
    fi
  fi

  scenario_10="skipped_disabled"
  if [ "$multi_file_app_code_probe" -eq 1 ]; then
    if run_multi_file_app_code_probe "$live_repo_name" "$task_text" "$live_repo_path" "$live_repo_dir/scenario-10-multi-file-app-code-probe" "$live_bin"; then
      scenario_10="pass"
    else
      scenario_10="fail"
      overall="fail"
    fi
  fi

  {
    printf '%s\n' "# Live Dogfood Summary: $live_repo_name"
    printf '\n'
    printf '%s\n' "- Repo path: $live_repo_path"
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
    if [ "$write_probe" -eq 1 ]; then
      printf '%s\n' "| scenario-06-write-probe | $scenario_06 | scenario-06-write-probe/pass-fail-note.txt |"
    fi
    if [ "$feature_edit_probe" -eq 1 ]; then
      printf '%s\n' "| scenario-07-feature-edit-probe | $scenario_07 | scenario-07-feature-edit-probe/pass-fail-note.txt |"
    fi
    if [ "$app_code_probe" -eq 1 ]; then
      printf '%s\n' "| scenario-08-app-code-probe | $scenario_08 | scenario-08-app-code-probe/pass-fail-note.txt |"
    fi
    if [ "$integrated_app_code_probe" -eq 1 ]; then
      printf '%s\n' "| scenario-09-integrated-app-code-probe | $scenario_09 | scenario-09-integrated-app-code-probe/pass-fail-note.txt |"
    fi
    if [ "$multi_file_app_code_probe" -eq 1 ]; then
      printf '%s\n' "| scenario-10-multi-file-app-code-probe | $scenario_10 | scenario-10-multi-file-app-code-probe/pass-fail-note.txt |"
    fi
  } >"$live_repo_dir/summary.md"
  printf '%s\n' "$overall" >"$live_repo_dir/status.txt"

  append_repo_row "$live_repo_name" "$overall" "$live_repo_path" "see $(evidence_rel_path "$live_repo_dir")/summary.md"
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

  if [ "$repeat_count" -eq 1 ]; then
    if [ "$dry_run" -eq 1 ]; then
      write_dry_run_plan "$repo_dir" "$repo_name" "$repo_path"
      append_repo_row "$repo_name" "planned" "$repo_path" "dry-run only; no commands run"
    else
      workspace_path=$(prepare_repo_workspace "$repo_path" "$repo_dir")
      run_live_repo "$repo_name" "$workspace_path" "$repo_dir" "$bin"
    fi
  else
    attempt=1
    while [ "$attempt" -le "$repeat_count" ]; do
      run_slug=$(attempt_slug "$attempt")
      run_dir="$repo_dir/$run_slug"
      run_name="$repo_name $run_slug"
      if [ "$dry_run" -eq 1 ]; then
        write_dry_run_plan "$run_dir" "$run_name" "$repo_path"
        append_repo_row "$run_name" "planned" "$repo_path" "see $(evidence_rel_path "$run_dir")/plan.md"
      else
        workspace_path=$(prepare_repo_workspace "$repo_path" "$run_dir")
        run_live_repo "$run_name" "$workspace_path" "$run_dir" "$bin"
      fi
      attempt=$((attempt + 1))
    done
    write_repeat_summary "$repo_dir" "$repo_name" "$repo_path"
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
