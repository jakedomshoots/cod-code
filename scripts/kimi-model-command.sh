#!/bin/sh
set -eu

tmp=$(mktemp -t ceo-kimi-model.XXXXXX)
log=$(mktemp -t ceo-kimi-model-log.XXXXXX)
run_dir=$(mktemp -d -t ceo-kimi-model-run.XXXXXX)
skills_dir="$run_dir/skills"
mkdir -p "$skills_dir"

cleanup() {
  rm -f "$tmp" "$log"
  rm -rf "$run_dir"
}
trap cleanup EXIT

prompt=$(cat)
kind="${CEO_MODEL_REQUEST_KIND:-unknown}"
agent="${CEO_AGENT_NAME:-unknown}"
role="${CEO_AGENT_ROLE:-unknown}"
context="${CEO_CONTEXT_MODE:-unknown}"

workspace_context=$(python3 - "$prompt" <<'PY'
import os
import pathlib
import sys

prompt = sys.argv[1]

def prompt_line(prefix):
    for line in prompt.splitlines():
        if line.startswith(prefix):
            return line[len(prefix):].strip().rstrip(".")
    return ""

def prompt_list(prefix):
    value = prompt_line(prefix)
    if not value:
        return []
    return [part.strip().rstrip(".") for part in value.split(",") if part.strip()]

def safe_relative(path):
    clean = pathlib.PurePosixPath(path.replace(os.sep, "/"))
    if clean.is_absolute() or ".." in clean.parts:
        return None
    return pathlib.Path(*clean.parts)

changed_files = prompt_list("Required changed files: ")
artifact_files = prompt_list("Required evidence artifacts: ")

def sibling_context_files(paths):
    extras = []
    seen = set(paths)
    for raw_path in paths:
        rel = safe_relative(raw_path)
        if rel is None:
            continue
        candidates = []
        if rel.suffix == ".go" and not rel.name.endswith("_test.go"):
            candidates.append(rel.with_name(rel.stem + "_test.go"))
        if rel.suffix in {".js", ".jsx", ".ts", ".tsx", ".mjs", ".cjs"} and not any(part in rel.stem for part in [".test", ".spec"]):
            candidates.append(rel.with_name(rel.stem + ".test" + rel.suffix))
            candidates.append(rel.with_name(rel.stem + ".spec" + rel.suffix))
        for test_name in candidates:
            test_raw = test_name.as_posix()
            if test_raw not in seen and test_name.exists():
                extras.append(test_raw)
                seen.add(test_raw)
    return extras

context_files = changed_files + sibling_context_files(changed_files)
lines = []
if context_files:
    lines.append("Workspace snippets:")
    for raw_path in context_files:
        rel = safe_relative(raw_path)
        if rel is None:
            lines.append(f"File: {raw_path}")
            lines.append("Status: skipped unsafe path")
            continue
        lines.append(f"File: {raw_path}")
        try:
            content = rel.read_text(encoding="utf-8")
        except FileNotFoundError:
            lines.append("Status: missing")
            continue
        if len(content) > 12000:
            content = content[:12000] + "\n[truncated]\n"
        lines.append("Content:")
        lines.append(content)
diff_terms = prompt_list("Required diff terms: ")
if diff_terms:
    lines.append("Required diff terms parsed from harness prompt:")
    for term in diff_terms:
        lines.append(f"- {term}")
if artifact_files:
    lines.append("Required evidence artifacts parsed from harness prompt:")
    for artifact in artifact_files:
        lines.append(f"- {artifact}")

print("\n".join(lines))
PY
)

run_prompt=$(printf '%s\n' \
  "You are a real model backend for CEO Harness, not the outer coding agent." \
  "You are running in an isolated temporary directory. The workspace snippets below are the authoritative file contents." \
  "Return one JSON object only. Do not include markdown unless you cannot avoid it." \
  "Do not edit files directly. If an edit is needed, propose it in the JSON patches array." \
  "Do not use shell, filesystem, or agentic actions. Only return the JSON object for the requested contract." \
  "For existing files, every patch must include path, exact old text, and new text. Use content only when creating a new file." \
  "If the harness prompt lists Required diff terms, the changed file must include those exact terms." \
  "If you need more workspace content, request it with tool_requests instead of guessing." \
  "Return needs_input only for a missing user decision, never just because you want to inspect files." \
  "" \
  "Request metadata:" \
  "kind: $kind" \
  "agent: $agent" \
  "role: $role" \
  "context: $context" \
  "" \
  "JSON contracts:" \
  "ceo_delegation -> {\"selected_subagents\":[\"coder\"],\"summary\":\"short reason\"}; choose the smallest useful set." \
  "ceo_delegation must include a non-empty selected_subagents array using candidate names from the harness prompt." \
  "ceo_review -> {\"recommended_verdict\":\"pass|fail\",\"summary\":\"short reason\"}" \
  "For ceo_review, guard_verdict, checks, changed_files, patch_results, and workspace snippets are observed facts." \
  "Do not contradict guard_verdict, checks, patch_results, or workspace snippets. Recommend fail only for a concrete unmet requirement visible in those facts." \
  "subagent work -> {\"status\":\"pass|fail|needs_input\",\"summary\":\"short result\",\"confidence\":0.0,\"evidence\":[\"item\"],\"tool_requests\":[],\"patches\":[{\"path\":\"existing.txt\",\"old\":\"exact old text\",\"new\":\"replacement text\"}]}" \
  "" \
  "$workspace_context" \
  "" \
  "Harness prompt:" \
  "$prompt")

if ! (cd "$run_dir" && kimi -p "$run_prompt" --skills-dir "$skills_dir" --output-format stream-json >"$tmp" 2>"$log"); then
  cat "$log" >&2
  exit 1
fi

python3 - "$tmp" "$prompt" <<'PY'
import json
import os
import pathlib
import re
import sys

prompt = sys.argv[2]

def prompt_line(prefix):
    for line in prompt.splitlines():
        if line.startswith(prefix):
            return line[len(prefix):].strip().rstrip(".")
    return ""

def prompt_list(prefix):
    value = prompt_line(prefix)
    if not value:
        return []
    return [part.strip().rstrip(".") for part in value.split(",") if part.strip()]

def safe_relative(path):
    clean = pathlib.PurePosixPath(str(path).replace(os.sep, "/"))
    if clean.is_absolute() or ".." in clean.parts:
        return None
    return pathlib.Path(*clean.parts)

def normalize_patch(patch):
    if not isinstance(patch, dict):
        return
    raw_path = patch.get("path")
    if not isinstance(raw_path, str) or not raw_path.strip():
        return
    rel = safe_relative(raw_path)
    if rel is None:
        return
    content = patch.get("content")
    old = patch.get("old")
    new = patch.get("new")
    has_content = isinstance(content, str) and content != ""
    has_old = isinstance(old, str) and old != ""
    has_new = isinstance(new, str) and new != ""
    if has_content:
        try:
            current = rel.read_text(encoding="utf-8")
        except FileNotFoundError:
            return
        patch["old"] = current
        patch["new"] = content
        patch.pop("content", None)
        return
    if has_new and not has_old:
        try:
            current = rel.read_text(encoding="utf-8")
        except FileNotFoundError:
            patch["content"] = new
            patch.pop("old", None)
            patch.pop("new", None)
            return
        patch["old"] = current

def normalize_payload(payload):
    if not isinstance(payload, dict):
        return payload
    patches = payload.get("patches")
    if isinstance(patches, list):
        for patch in patches:
            normalize_patch(patch)
    validate_required_diff_terms(payload)
    return payload

def validate_required_diff_terms(payload):
    required_files = set(prompt_list("Required changed files: "))
    terms = prompt_list("Required diff terms: ")
    if not required_files or not terms:
        return
    patches = payload.get("patches")
    if not isinstance(patches, list):
        return
    text_by_path = {}
    for patch in patches:
        if not isinstance(patch, dict):
            continue
        path = patch.get("path")
        if path not in required_files:
            continue
        text = patch.get("new")
        if not isinstance(text, str) or text == "":
            text = patch.get("content")
        if isinstance(text, str) and text != "":
            text_by_path[path] = text_by_path.get(path, "") + "\n" + text
    for path, text in text_by_path.items():
        missing = [term for term in terms if term not in text]
        if missing:
            raise SystemExit(f"required diff terms missing from patch for {path}: {', '.join(missing)}")

def emit_payload(payload):
    print(json.dumps(normalize_payload(payload), separators=(",", ":")))
    raise SystemExit(0)

content = None
with open(sys.argv[1], "r", encoding="utf-8") as handle:
    for raw_line in handle:
        line = raw_line.strip()
        if not line:
            continue
        try:
            event = json.loads(line)
        except json.JSONDecodeError:
            continue
        if event.get("role") == "assistant" and isinstance(event.get("content"), str):
            content = event["content"].strip()

if not content:
    raise SystemExit("kimi model command returned no assistant content")

if content.startswith("{"):
    try:
        emit_payload(json.loads(content))
    except json.JSONDecodeError:
        pass

fenced = re.search(r"```(?:json)?\s*(\{.*?\})\s*```", content, re.DOTALL)
if fenced:
    try:
        emit_payload(json.loads(fenced.group(1).strip()))
    except json.JSONDecodeError:
        pass

decoder = json.JSONDecoder()
for index, char in enumerate(content):
    if char != "{":
        continue
    try:
        payload, _ = decoder.raw_decode(content[index:])
    except json.JSONDecodeError:
        continue
    emit_payload(payload)

raise SystemExit("kimi model command returned no JSON object")
PY
