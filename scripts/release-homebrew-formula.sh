#!/bin/sh
set -eu

root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
dist=${DIST:-"$root/dist"}
repo_url=${REPO_URL:-}
archive_base_url=${HOMEBREW_ARCHIVE_BASE_URL:-${HOMEBREW_ARCHIVE_URL:-}}
version=${VERSION:-}
output=${OUTPUT:-}

usage() {
  cat <<'USAGE'
Usage: sh scripts/release-homebrew-formula.sh [options]

Writes a Homebrew formula that points at the public HTTPS Darwin archive.
This command does not publish a tap or upload release files.

Options:
  --dist DIR                         Local release dist directory.
  --repo-url URL                     Public repo HTTPS URL.
  --homebrew-archive-base-url URL    Public HTTPS directory containing archives.
  --version VERSION                  Release version. Defaults to manifest version.
  --output PATH                      Formula path. Defaults to dist/homebrew/ceo-packet.rb.

Environment aliases:
  GITHUB_REPOSITORY                 owner/name fallback for --repo-url.
  HOMEBREW_ARCHIVE_URL              Full public archive URL; base path is inferred.
USAGE
}

while [ "$#" -gt 0 ]; do
  case "$1" in
    --dist)
      [ "$#" -ge 2 ] || { printf '%s\n' "release-homebrew-formula: --dist requires a value" >&2; exit 2; }
      dist="$2"
      shift 2
      ;;
    --repo-url)
      [ "$#" -ge 2 ] || { printf '%s\n' "release-homebrew-formula: --repo-url requires a value" >&2; exit 2; }
      repo_url="$2"
      shift 2
      ;;
    --homebrew-archive-base-url)
      [ "$#" -ge 2 ] || { printf '%s\n' "release-homebrew-formula: --homebrew-archive-base-url requires a value" >&2; exit 2; }
      archive_base_url="$2"
      shift 2
      ;;
    --version)
      [ "$#" -ge 2 ] || { printf '%s\n' "release-homebrew-formula: --version requires a value" >&2; exit 2; }
      version="$2"
      shift 2
      ;;
    --output)
      [ "$#" -ge 2 ] || { printf '%s\n' "release-homebrew-formula: --output requires a value" >&2; exit 2; }
      output="$2"
      shift 2
      ;;
    --help|-h)
      usage
      exit 0
      ;;
    *)
      printf '%s\n' "release-homebrew-formula: unknown argument: $1" >&2
      usage >&2
      exit 2
      ;;
  esac
done

case "$dist" in
  /*) ;;
  *) dist="$(pwd)/$dist" ;;
esac

if [ -z "$output" ]; then
  output="$dist/homebrew/ceo-packet.rb"
fi
case "$output" in
  /*) ;;
  *) output="$(pwd)/$output" ;;
esac

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

slash_trim() {
  printf '%s' "$1" | sed 's:/*$::'
}

if [ -z "$repo_url" ] && [ -n "${GITHUB_REPOSITORY:-}" ]; then
  repo_url="https://github.com/$GITHUB_REPOSITORY"
fi

case "$archive_base_url" in
  *.tar.gz)
    archive_base_url=${archive_base_url%/*}
    ;;
esac

if ! is_public_https_url "$repo_url"; then
  printf '%s\n' "release-homebrew-formula: --repo-url must be a real public HTTPS URL" >&2
  exit 2
fi

if ! is_public_https_url "$archive_base_url"; then
  printf '%s\n' "release-homebrew-formula: --homebrew-archive-base-url must be a real public HTTPS URL" >&2
  exit 2
fi

if [ ! -f "$dist/release-manifest.json" ]; then
  printf '%s\n' "release-homebrew-formula: missing release-manifest.json in $dist" >&2
  exit 2
fi

if [ ! -f "$dist/checksums.txt" ]; then
  printf '%s\n' "release-homebrew-formula: missing checksums.txt in $dist" >&2
  exit 2
fi

if [ -z "$version" ]; then
  version=$(python3 - "$dist/release-manifest.json" <<'PY'
import json
import sys
with open(sys.argv[1], "r", encoding="utf-8") as handle:
    print(json.load(handle).get("version", ""))
PY
)
fi

case "$version" in
  ""|*[!A-Za-z0-9._-]*)
    printf '%s\n' "release-homebrew-formula: invalid or missing version" >&2
    exit 2
    ;;
esac

archive_name="ceo-packet_${version}_darwin_arm64.tar.gz"
archive_sha=$(awk -v archive="$archive_name" '$2 == archive {print $1}' "$dist/checksums.txt")
if [ -z "$archive_sha" ]; then
  printf '%s\n' "release-homebrew-formula: missing checksum for $archive_name" >&2
  exit 2
fi

archive_base_url=$(slash_trim "$archive_base_url")
mkdir -p "$(dirname "$output")"
cat >"$output" <<EOF
class CeoPacket < Formula
  desc "Local Alpha Cod/swimmer coding harness"
  homepage "$repo_url"
  url "$archive_base_url/$archive_name"
  sha256 "$archive_sha"
  version "$version"

  def install
    bin.install "ceo-packet"
  end

  test do
    assert_match "ceo-packet $version", shell_output("#{bin}/ceo-packet --version")
  end
end
EOF

printf '%s\n' "release-homebrew-formula: wrote $output"
