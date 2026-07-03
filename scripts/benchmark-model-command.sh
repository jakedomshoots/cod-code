#!/bin/sh
set -eu

prompt=$(cat)
if [ -z "$prompt" ] && [ "$#" -gt 0 ]; then
  prompt="$*"
fi

case "$prompt" in
  *'"selected_subagents"'*)
    printf '{"selected_subagents":["coder"],"summary":"Use coder for benchmark patch ownership."}\n'
    exit 0
    ;;
  *'"recommended_verdict":"pass|fail"'*)
    printf '{"recommended_verdict":"pass","summary":"Benchmark artifacts and checks were reviewed."}\n'
    exit 0
    ;;
esac

python3 - "$prompt" <<'PY'
import json
import os
import re
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
    return [part.strip() for part in value.split(",") if part.strip()]

changed_files = prompt_list("Required changed files: ")
artifact_files = prompt_list("Required evidence artifacts: ")
diff_terms = prompt_line("Required diff terms: ")

if not changed_files or not artifact_files:
    print(json.dumps({
        "status": "needs_input",
        "summary": "benchmark prompt missing required changed files or artifact paths",
    }))
    raise SystemExit(0)

def go_package(path):
    name = os.path.basename(os.path.dirname(path))
    clean = re.sub(r"[^A-Za-z0-9_]", "_", name).rstrip("_")
    if not clean or clean[0].isdigit():
        return "fixture"
    return clean

def replacement_for(path):
    if path == "internal/workspace/workspace.go":
        return """package workspace

import (
\t"errors"
\t"path/filepath"
\t"strings"
)

var ErrPathEscapesWorkspace = errors.New("path escapes workspace")

func CleanRelativePath(path string) (string, error) {
\tcleanPath := filepath.Clean(strings.TrimSpace(path))
\tif cleanPath == "." || cleanPath == ".." || filepath.IsAbs(cleanPath) || strings.HasPrefix(cleanPath, ".."+string(filepath.Separator)) {
\t\treturn "", ErrPathEscapesWorkspace
\t}
\treturn cleanPath, nil
}
"""
    if path.endswith(".go"):
        return f"package {go_package(path)}\n\nconst benchmarkFixture = {json.dumps(diff_terms)}\n"
    return f"# Benchmark Fixture\n\n{diff_terms}\n"

patches = []
for changed_file in changed_files:
    with open(changed_file, "r", encoding="utf-8") as handle:
        old_content = handle.read()
    patches.append({
        "path": changed_file,
        "old": old_content,
        "new": replacement_for(changed_file),
    })

artifact_content = f"""# Benchmark Evidence

Change: updated {", ".join(changed_files)}.
Commands: required benchmark command is supplied by the harness.
Verification: model-command benchmark patch generated for {", ".join(artifact_files)}.
"""

for artifact_file in artifact_files:
    patches.append({
        "path": artifact_file,
        "content": artifact_content,
    })

print(json.dumps({
    "status": "pass",
    "summary": "benchmark patch ready",
    "evidence": artifact_files,
    "patches": patches,
}))
PY
