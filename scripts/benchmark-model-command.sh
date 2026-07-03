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

changed_file=$(printf '%s\n' "$prompt" | sed -n 's/^Required changed files: //p' | head -n 1 | cut -d, -f1 | sed 's/[.]$//' | xargs)
artifact_file=$(printf '%s\n' "$prompt" | sed -n 's/^Required evidence artifacts: //p' | head -n 1 | cut -d, -f1 | sed 's/[.]$//' | xargs)
diff_terms=$(printf '%s\n' "$prompt" | sed -n 's/^Required diff terms: //p' | head -n 1 | sed 's/[.]$//')

if [ -z "$changed_file" ] || [ -z "$artifact_file" ]; then
  printf '{"status":"needs_input","summary":"benchmark prompt missing required changed file or artifact path"}\n'
  exit 0
fi

old_content=$(cat "$changed_file")
case "$changed_file" in
  internal/workspace/workspace.go)
    new_content='package workspace

import (
	"errors"
	"path/filepath"
	"strings"
)

var ErrPathEscapesWorkspace = errors.New("path escapes workspace")

func CleanRelativePath(path string) (string, error) {
	cleanPath := filepath.Clean(strings.TrimSpace(path))
	if cleanPath == "." || cleanPath == ".." || filepath.IsAbs(cleanPath) || strings.HasPrefix(cleanPath, ".."+string(filepath.Separator)) {
		return "", ErrPathEscapesWorkspace
	}
	return cleanPath, nil
}
' ;;
  *.go)
    pkg=$(printf '%s' "$(basename "$(dirname "$changed_file")")" | tr -c 'A-Za-z0-9_' '_' | sed 's/_*$//')
    case "$pkg" in
      ""|[0-9]*) pkg="fixture" ;;
    esac
    new_content="package $pkg

const benchmarkFixture = \"$diff_terms\"
" ;;
  *)
    new_content="# Benchmark Fixture

$diff_terms
" ;;
esac

artifact_content="# Benchmark Evidence

Change: updated $changed_file.
Commands: required benchmark command is supplied by the harness.
Verification: model-command benchmark patch generated for $artifact_file.
"

python3 - "$changed_file" "$artifact_file" "$old_content" "$new_content" "$artifact_content" <<'PY'
import json
import sys

changed_file, artifact_file, old_content, new_content, artifact_content = sys.argv[1:6]
print(json.dumps({
    "status": "pass",
    "summary": "benchmark patch ready",
    "evidence": [artifact_file],
    "patches": [
        {"path": changed_file, "old": old_content, "new": new_content},
        {"path": artifact_file, "content": artifact_content},
    ],
}))
PY
