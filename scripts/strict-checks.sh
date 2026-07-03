#!/bin/sh
set -eu

root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
cd "$root"

go_bin_dir() {
  gobin=$(go env GOBIN)
  if [ -n "$gobin" ]; then
    printf '%s\n' "$gobin"
    return
  fi
  printf '%s/bin\n' "$(go env GOPATH)"
}

tool_path() {
  name="$1"
  bin_dir=$(go_bin_dir)
  if command -v "$name" >/dev/null 2>&1; then
    command -v "$name"
    return
  fi
  if [ -x "$bin_dir/$name" ]; then
    printf '%s/%s\n' "$bin_dir" "$name"
    return
  fi
  return 1
}

require_tool() {
  name="$1"
  if ! path=$(tool_path "$name"); then
    printf '%s\n' "strict-checks: missing required tool: $name" >&2
    exit 1
  fi
  printf '%s\n' "$path"
}

run_shell_syntax_check() {
  failed=0
  found=0
  for script in scripts/*.sh examples/*.sh examples/adapters/*.sh tests/*.sh; do
    [ -f "$script" ] || continue
    found=1
    if ! sh -n "$script"; then
      failed=1
    fi
  done
  if [ "$found" -eq 0 ]; then
    printf '%s\n' "strict-checks: no shell scripts found" >&2
    exit 1
  fi
  if [ "$failed" -ne 0 ]; then
    printf '%s\n' "strict-checks: shell syntax failed" >&2
    exit 1
  fi
}

gofumpt=$(require_tool gofumpt)
golangci_lint=$(require_tool golangci-lint)
nilaway=$(require_tool nilaway)

unformatted=$("$gofumpt" -l cmd internal)
if [ -n "$unformatted" ]; then
  printf '%s\n' "$unformatted" >&2
  printf '%s\n' "strict-checks: gofumpt drift" >&2
  exit 1
fi

"$golangci_lint" run ./...
"$nilaway" ./...

run_shell_syntax_check

if shellcheck_path=$(tool_path shellcheck); then
  "$shellcheck_path" scripts/*.sh examples/*.sh examples/adapters/*.sh tests/*.sh
else
  printf '%s\n' "strict-checks: shellcheck unavailable; ran sh -n on shell scripts"
fi

printf '%s\n' "strict-checks ok"
