#!/bin/sh
set -eu

root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
scan_root="$root"

usage() {
  cat <<'USAGE'
Usage: sh scripts/secret-scan.sh [--root path]

Scans source, docs, scripts, workflows, and provider command evidence files for committed secret values.
Placeholders such as OPENAI_API_KEY=... are allowed in docs, but real key assignments, bearer values, OAuth token filenames, and copied token files are blocked.
USAGE
}

while [ "$#" -gt 0 ]; do
  case "$1" in
    --root)
      [ "$#" -ge 2 ] || {
        printf '%s\n' "secret-scan: --root requires a value" >&2
        exit 2
      }
      scan_root="$2"
      shift 2
      ;;
    --help|-h)
      usage
      exit 0
      ;;
    *)
      printf '%s\n' "secret-scan: unknown argument: $1" >&2
      usage >&2
      exit 2
      ;;
  esac
done

case "$scan_root" in
  /*) ;;
  *) scan_root="$(pwd)/$scan_root" ;;
esac

if [ ! -d "$scan_root" ]; then
  printf '%s\n' "secret-scan: missing root: $scan_root" >&2
  exit 2
fi

tmp=$(mktemp -t ceo-secret-scan.XXXXXX)
trap 'rm -f "$tmp"' EXIT

find "$scan_root" \
  \( -path "$scan_root/.git" -o -path "$scan_root/.git/*" \
    -o -path "$scan_root/.omo" -o -path "$scan_root/.omo/*" \
    -o -path "$scan_root/bin" -o -path "$scan_root/bin/*" \
    -o -path "$scan_root/dist" -o -path "$scan_root/dist/*" \
    -o -path "$scan_root/vendor" -o -path "$scan_root/vendor/*" \) -prune \
  -o -type f \
  \( -name 'README.md' -o -name 'SECURITY.md' -o -name 'CONTRIBUTING.md' -o -name 'CHANGELOG.md' \
    -o -path "$scan_root/docs/*.md" \
    -o -path "$scan_root/scripts/*.sh" \
    -o -path "$scan_root/.github/workflows/*.yml" \
    -o -path "$scan_root/.github/workflows/*.yaml" \
    -o -path "$scan_root/cmd/*.go" -o -path "$scan_root/cmd/*/*.go" \
    -o -path "$scan_root/internal/*.go" -o -path "$scan_root/internal/*/*.go" -o -path "$scan_root/internal/*/*/*.go" \) \
  ! -name '*_test.go' \
  -print >"$tmp"

if [ -d "$scan_root/.omo/evidence" ]; then
  find "$scan_root/.omo/evidence" \
    -type f \
    \( -name 'commands.sh' -o -name 'command.txt' -o -name 'command.argv' -o -name 'command.json' \
      -o -name 'env.template' -o -name 'setup-checklist.md' -o -name 'blocked.md' \) \
    -print >>"$tmp"
fi

secret_pattern='(OPENAI_API_KEY|OPENROUTER_API_KEY|KIMI_CODE_API_KEY|MOONSHOT_API_KEY|MINIMAX_API_KEY|ANTHROPIC_API_KEY)[[:space:]]*=[[:space:]]*['\''"]?(sk-[A-Za-z0-9_-]{16,}|sk-proj-[A-Za-z0-9_-]{16,}|sk-or-v1-[A-Za-z0-9_-]{16,}|[A-Za-z0-9_.-]{40,})|Authorization:[[:space:]]*Bearer[[:space:]]+[A-Za-z0-9._~+/=-]{16,}|gh[opsu]_[A-Za-z0-9_]{20,}'
oauth_token_filename_pattern='(^|/)(oauth|access|refresh)[A-Za-z0-9_.-]*(token|credential)[A-Za-z0-9_.-]*$|(^|/)[A-Za-z0-9_.-]*(token|credential)[A-Za-z0-9_.-]*(oauth|refresh)[A-Za-z0-9_.-]*$'

failed=0
while IFS= read -r file; do
  [ -n "$file" ] || continue
  rel="$file"
  case "$file" in
    "$scan_root"/*) rel="${file#"$scan_root"/}" ;;
  esac
  if printf '%s\n' "$rel" | LC_ALL=C grep -Eq "$oauth_token_filename_pattern"; then
    failed=1
    printf '%s\n' "secret-scan: possible OAuth token file committed at $rel" >&2
    continue
  fi
  if LC_ALL=C grep -nE "$secret_pattern" "$file" >/tmp/ceo-secret-scan-match.$$ 2>/dev/null; then
    failed=1
    rel="$file"
    case "$file" in
      "$scan_root"/*) rel="${file#"$scan_root"/}" ;;
    esac
    while IFS= read -r match; do
      line=${match%%:*}
      printf '%s\n' "secret-scan: possible secret in $rel:$line" >&2
    done </tmp/ceo-secret-scan-match.$$
  fi
  rm -f /tmp/ceo-secret-scan-match.$$
done <"$tmp"
rm -f /tmp/ceo-secret-scan-match.$$

if [ "$failed" -ne 0 ]; then
  printf '%s\n' "secret-scan: failed" >&2
  exit 1
fi

printf '%s\n' "secret-scan ok"
