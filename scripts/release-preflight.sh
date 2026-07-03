#!/bin/sh
set -eu

root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
dist=${1:-${DIST:-"$root/dist"}}
release_url=${RELEASE_URL:-${PUBLIC_RELEASE_URL:-}}
checksum_only_notes_url=${CHECKSUM_ONLY_RELEASE_NOTES_URL:-}
allow_checksum_only=${ALLOW_CHECKSUM_ONLY_RELEASE:-}
blockers=0

case "$dist" in
  /*) ;;
  *) dist="$(pwd)/$dist" ;;
esac

usage() {
  cat <<'USAGE'
Usage: sh scripts/release-preflight.sh [dist]

Checks whether a local release is ready to be called public. This command does
not tag, push, upload, publish a tap, or create a release.

Environment:
  RELEASE_URL or PUBLIC_RELEASE_URL   Public HTTPS URL for the release page.
  ALLOW_CHECKSUM_ONLY_RELEASE=1       Allow unsigned archives only when paired
                                      with CHECKSUM_ONLY_RELEASE_NOTES_URL.
  CHECKSUM_ONLY_RELEASE_NOTES_URL     Public HTTPS notes explaining checksum-only
                                      verification.
USAGE
}

if [ "${1:-}" = "--help" ] || [ "${1:-}" = "-h" ]; then
  usage
  exit 0
fi

row() {
  name="$1"
  status="$2"
  detail="$3"
  printf '| %s | %s | %s |\n' "$name" "$status" "$detail"
  if [ "$status" = "blocked" ]; then
    blockers=$((blockers + 1))
  fi
}

is_public_https_url() {
  candidate="$1"
  case "$candidate" in
    https://*example.invalid*|https://example.com*|https://example.org*|"")
      return 1
      ;;
    https://*)
      return 0
      ;;
    *)
      return 1
      ;;
  esac
}

has_public_release_url() {
  is_public_https_url "$release_url"
}

formula_is_remote() {
  formula="$dist/homebrew/ceo-packet.rb"
  [ -f "$formula" ] || return 1
  if grep -q 'example.invalid\|url "file://' "$formula"; then
    return 1
  fi
  grep -q 'url "https://' "$formula"
}

signatures_exist() {
  found=0
  for archive in "$dist"/*.tar.gz; do
    [ -e "$archive" ] || continue
    found=1
    if [ ! -f "$archive.sig" ] && [ ! -f "$archive.minisig" ] && [ ! -f "$archive.asc" ]; then
      return 1
    fi
  done
  [ "$found" -eq 1 ]
}

printf '%s\n' '# Public Release Preflight'
printf '\n'
printf '| Check | Status | Detail |\n'
printf '| --- | --- | --- |\n'

if [ -d "$dist" ] && sh "$root/scripts/verify-release.sh" "$dist" >/tmp/ceo-release-preflight-verify.$$ 2>&1; then
  row "local_release_artifacts" "pass" "checksums and release-manifest verified"
else
  detail="run VERSION=<version> sh scripts/release-local.sh, then sh scripts/verify-release.sh dist"
  row "local_release_artifacts" "blocked" "$detail"
fi
rm -f /tmp/ceo-release-preflight-verify.$$

remote_url=$(git -C "$root" remote get-url origin 2>/dev/null || true)
if [ -n "$remote_url" ]; then
  row "git_remote" "pass" "$remote_url"
else
  row "git_remote" "blocked" "no origin remote configured"
fi

if has_public_release_url; then
  row "remote_release_url" "pass" "$release_url"
else
  row "remote_release_url" "blocked" "set RELEASE_URL to the public HTTPS release page"
fi

if formula_is_remote; then
  row "homebrew_formula_url" "pass" "formula uses a remote HTTPS archive URL"
else
  row "homebrew_formula_url" "blocked" "dist/homebrew/ceo-packet.rb still uses a local or placeholder URL"
fi

if signatures_exist; then
  row "artifact_signatures" "pass" "every archive has .sig, .minisig, or .asc"
elif [ "$allow_checksum_only" = "1" ] && is_public_https_url "$checksum_only_notes_url"; then
  row "artifact_signatures" "pass" "checksum-only release explicitly documented at $checksum_only_notes_url"
else
  row "artifact_signatures" "blocked" "add archive signatures or set ALLOW_CHECKSUM_ONLY_RELEASE=1 with CHECKSUM_ONLY_RELEASE_NOTES_URL"
fi

if [ "$blockers" -eq 0 ]; then
  printf '\n%s\n' "public release preflight: pass"
  exit 0
fi

printf '\n%s\n' "public release preflight: blocked ($blockers)"
exit 1
