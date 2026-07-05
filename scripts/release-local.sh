#!/bin/sh
set -eu

root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
version=${VERSION:-0.1.0-dev}
commit=${COMMIT:-$(git -C "$root" rev-parse --short HEAD 2>/dev/null || printf local)}
dist=${DIST:-"$root/dist"}

usage() {
  cat <<'USAGE'
Usage: sh scripts/release-local.sh [--dist dist]

Builds local release archives, checksums, manifest, and Homebrew formula draft.
USAGE
}

while [ "$#" -gt 0 ]; do
  case "$1" in
    --dist)
      [ "$#" -ge 2 ] || {
        printf '%s\n' "release-local: --dist requires a value" >&2
        exit 2
      }
      dist="$2"
      shift 2
      ;;
    --help|-h)
      usage
      exit 0
      ;;
    *)
      printf '%s\n' "release-local: unknown argument: $1" >&2
      usage >&2
      exit 2
      ;;
  esac
done

case "$dist" in
  /*) ;;
  *) dist="$(pwd)/$dist" ;;
esac

formula_dir="$dist/homebrew"

case "$version" in
*[!A-Za-z0-9._-]*)
  printf '%s\n' "VERSION may only contain letters, numbers, dots, underscores, or dashes" >&2
  exit 2
  ;;
esac

case "$commit" in
*[!A-Za-z0-9._-]*)
  printf '%s\n' "COMMIT may only contain letters, numbers, dots, underscores, or dashes" >&2
  exit 2
  ;;
esac

rm -rf "$dist"
mkdir -p "$dist"

build_target() {
  goos=$1
  goarch=$2
  name="ceo-packet_${version}_${goos}_${goarch}"
  outdir="$dist/$name"
  mkdir -p "$outdir"
  GOOS=$goos GOARCH=$goarch go build \
    -trimpath \
    -ldflags "-s -w -X ceoharness/internal/cli.Version=$version -X ceoharness/internal/cli.Commit=$commit" \
    -o "$outdir/ceo-packet" \
    ./cmd/ceo-packet
  tar -C "$dist" -czf "$dist/$name.tar.gz" "$name"
  rm -rf "$outdir"
}

cd "$root"
build_target darwin arm64
build_target linux amd64
build_target linux arm64

(cd "$dist" && shasum -a 256 *.tar.gz >checksums.txt)
python3 - "$dist" "$version" "$commit" <<'PY'
import datetime
import hashlib
import json
import pathlib
import sys

dist = pathlib.Path(sys.argv[1])
version = sys.argv[2]
commit = sys.argv[3]
artifacts = []
for archive in sorted(dist.glob("*.tar.gz")):
    payload = archive.read_bytes()
    artifacts.append({
        "name": archive.name,
        "sha256": hashlib.sha256(payload).hexdigest(),
        "size_bytes": len(payload),
    })
manifest = {
    "schema_version": 1,
    "name": "ceo-packet",
    "version": version,
    "commit": commit,
    "generated_at": datetime.datetime.now(datetime.timezone.utc).replace(microsecond=0).isoformat().replace("+00:00", "Z"),
    "checksum_file": "checksums.txt",
    "artifacts": artifacts,
}
(dist / "release-manifest.json").write_text(json.dumps(manifest, indent=2) + "\n", encoding="utf-8")
PY
mkdir -p "$formula_dir"

darwin_archive="ceo-packet_${version}_darwin_arm64.tar.gz"
darwin_sha=$(awk -v archive="$darwin_archive" '$2 == archive {print $1}' "$dist/checksums.txt")
cat >"$formula_dir/ceo-packet.rb" <<EOF
class CeoPacket < Formula
  desc "Local Alpha Cod/swimmer coding harness"
  homepage "https://example.invalid/cod-code"
  url "file://$dist/$darwin_archive"
  sha256 "$darwin_sha"
  version "$version"

  def install
    bin.install "ceo-packet"
  end

  test do
    assert_match "ceo-packet $version", shell_output("#{bin}/ceo-packet --version")
  end
end
EOF

printf 'release artifacts written to %s\n' "$dist"
printf 'homebrew formula draft written to %s\n' "$formula_dir/ceo-packet.rb"
