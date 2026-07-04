#!/bin/sh
set -eu

root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
dist=${1:-${DIST:-"$root/dist"}}
release_url=${RELEASE_URL:-${PUBLIC_RELEASE_URL:-}}
github_release_tag=${GH_RELEASE_TAG:-${RELEASE_TAG:-}}
github_repo=${GH_REPO:-}
checksum_only_notes_url=${CHECKSUM_ONLY_RELEASE_NOTES_URL:-}
allow_checksum_only=${ALLOW_CHECKSUM_ONLY_RELEASE:-}
signing_public_key=${RELEASE_SIGNING_PUBLIC_KEY:-${SIGNING_PUBLIC_KEY:-}}
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
  GH_RELEASE_TAG or RELEASE_TAG       Optional GitHub release tag to verify.
  GH_REPO                             Optional GitHub repo in owner/name form.
  ALLOW_CHECKSUM_ONLY_RELEASE=1       Allow unsigned archives only when paired
                                      with CHECKSUM_ONLY_RELEASE_NOTES_URL.
  CHECKSUM_ONLY_RELEASE_NOTES_URL     Public HTTPS notes explaining checksum-only
                                      verification.
  RELEASE_SIGNING_PUBLIC_KEY          Optional public key used to verify .sig
                                      files before accepting signatures.
  SIGNING_PUBLIC_KEY                  Alternate public key env name.
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

github_repo_from_remote() {
  remote="$1"
  case "$remote" in
    https://github.com/*)
      printf '%s\n' "$remote" | sed 's#^https://github.com/##; s#\.git$##'
      ;;
    git@github.com:*)
      printf '%s\n' "$remote" | sed 's#^git@github.com:##; s#\.git$##'
      ;;
    *)
      return 1
      ;;
  esac
}

is_github_repo_name() {
  case "$1" in
    */*)
      case "$1" in
        *://*|*' '*|/*|*/|*//*)
          return 1
          ;;
        *)
          return 0
          ;;
      esac
      ;;
    *)
      return 1
      ;;
  esac
}

verify_github_release_assets() {
  tag="$1"
  repo="$2"
  out="$3"
  [ -n "$tag" ] || return 1
  [ -n "$repo" ] || return 1
  command -v gh >/dev/null 2>&1 || return 1
  gh release view "$tag" --repo "$repo" --json url,assets >"$out" 2>/dev/null || return 1
  python3 - "$dist/release-manifest.json" "$out" <<'PY'
import json
import sys

manifest_path, release_path = sys.argv[1], sys.argv[2]
with open(manifest_path, "r", encoding="utf-8") as handle:
    manifest = json.load(handle)
with open(release_path, "r", encoding="utf-8") as handle:
    release = json.load(handle)

asset_names = {asset.get("name", "") for asset in release.get("assets", [])}
required = {artifact.get("name", "") for artifact in manifest.get("artifacts", [])}
required.update({"checksums.txt", "release-manifest.json"})
missing = sorted(name for name in required if name not in asset_names)
url = release.get("url", "")
if missing or not url.startswith("https://"):
    raise SystemExit(1)
print(url)
PY
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
  if [ -z "$github_repo" ]; then
    github_repo=$(github_repo_from_remote "$remote_url" || true)
  fi
elif is_github_repo_name "$github_repo"; then
  row "git_remote" "pass" "GH_REPO=$github_repo"
else
  row "git_remote" "blocked" "no origin remote configured; set GH_REPO=owner/name for release-only verification"
fi

github_release_json="/tmp/ceo-release-preflight-gh-release.$$"
github_release_url=""
github_release_assets_status=blocked
github_release_assets_detail="set GH_RELEASE_TAG after pushing a v* tag so gh can verify release assets"
if [ -n "$github_release_tag" ]; then
  if github_release_url=$(verify_github_release_assets "$github_release_tag" "$github_repo" "$github_release_json"); then
    github_release_assets_status=pass
    github_release_assets_detail="GitHub release $github_release_tag has all archives, checksums.txt, and release-manifest.json"
    if [ -z "$release_url" ]; then
      release_url="$github_release_url"
    fi
  elif ! command -v gh >/dev/null 2>&1; then
    github_release_assets_detail="gh CLI is required to verify GitHub release assets"
  elif [ -z "$github_repo" ]; then
    github_release_assets_detail="set GH_REPO=owner/name or configure a GitHub origin remote"
  else
    github_release_assets_detail="GitHub release $github_release_tag is missing required assets or is not readable"
  fi
fi
rm -f "$github_release_json"

if has_public_release_url; then
  row "remote_release_url" "pass" "$release_url"
else
  row "remote_release_url" "blocked" "set RELEASE_URL to the public HTTPS release page"
fi

row "github_release_assets" "$github_release_assets_status" "$github_release_assets_detail"

if formula_is_remote; then
  row "homebrew_formula_url" "pass" "formula uses a remote HTTPS archive URL"
else
  row "homebrew_formula_url" "blocked" "dist/homebrew/ceo-packet.rb still uses a local or placeholder URL"
fi

if signatures_exist; then
  if [ -n "$signing_public_key" ]; then
    if sh "$root/scripts/release-signatures.sh" --dist "$dist" --verify --public-key "$signing_public_key" >/tmp/ceo-release-preflight-signatures.$$ 2>&1; then
      row "artifact_signatures" "pass" "every archive signature verified with public key"
    else
      row "artifact_signatures" "blocked" "signature files exist but verification failed with RELEASE_SIGNING_PUBLIC_KEY"
    fi
    rm -f /tmp/ceo-release-preflight-signatures.$$
  else
    row "artifact_signatures" "pass" "every archive has .sig, .minisig, or .asc"
  fi
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
