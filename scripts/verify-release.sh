#!/bin/sh
set -eu

root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
dist=${DIST:-"$root/dist"}
signing_public_key=${RELEASE_SIGNING_PUBLIC_KEY:-${SIGNING_PUBLIC_KEY:-}}

usage() {
  cat <<'USAGE'
Usage: sh scripts/verify-release.sh [--dist dist]
       sh scripts/verify-release.sh [dist]

Verifies release checksums, manifest integrity, and optional signatures.
USAGE
}

while [ "$#" -gt 0 ]; do
  case "$1" in
    --dist)
      [ "$#" -ge 2 ] || {
        printf '%s\n' "verify-release: --dist requires a value" >&2
        exit 2
      }
      dist="$2"
      shift 2
      ;;
    --help|-h)
      usage
      exit 0
      ;;
    --*)
      printf '%s\n' "verify-release: unknown argument: $1" >&2
      usage >&2
      exit 2
      ;;
    *)
      dist="$1"
      shift
      ;;
  esac
done

case "$dist" in
  /*) ;;
  *) dist="$(pwd)/$dist" ;;
esac

if [ ! -d "$dist" ]; then
  printf '%s\n' "verify-release: missing dist directory: $dist" >&2
  exit 2
fi

if [ ! -f "$dist/checksums.txt" ]; then
  printf '%s\n' "verify-release: missing checksums.txt" >&2
  exit 2
fi

if [ ! -f "$dist/release-manifest.json" ]; then
  printf '%s\n' "verify-release: missing release-manifest.json" >&2
  exit 2
fi

(cd "$dist" && shasum -a 256 -c checksums.txt)

python3 - "$dist" <<'PY'
import hashlib
import json
import pathlib
import sys

dist = pathlib.Path(sys.argv[1])
manifest = json.loads((dist / "release-manifest.json").read_text(encoding="utf-8"))
if manifest.get("schema_version") != 1:
    raise SystemExit("release manifest schema_version must be 1")
if not manifest.get("version") or not manifest.get("commit"):
    raise SystemExit("release manifest needs version and commit")

checksums = {}
for raw in (dist / "checksums.txt").read_text(encoding="utf-8").splitlines():
    parts = raw.split()
    if len(parts) >= 2:
        checksums[parts[1]] = parts[0]

artifacts = manifest.get("artifacts")
if not isinstance(artifacts, list) or not artifacts:
    raise SystemExit("release manifest needs artifacts")

for artifact in artifacts:
    name = artifact.get("name")
    if not name or "/" in name:
        raise SystemExit(f"invalid artifact name: {name!r}")
    path = dist / name
    if not path.exists():
        raise SystemExit(f"missing artifact: {name}")
    digest = hashlib.sha256(path.read_bytes()).hexdigest()
    if artifact.get("sha256") != digest:
        raise SystemExit(f"sha256 mismatch for {name}")
    if checksums.get(name) != digest:
        raise SystemExit(f"checksums.txt mismatch for {name}")
    if artifact.get("size_bytes") != path.stat().st_size:
        raise SystemExit(f"size mismatch for {name}")

print("release manifest ok")
PY

if [ -n "$signing_public_key" ]; then
  sh "$root/scripts/release-signatures.sh" --dist "$dist" --verify --public-key "$signing_public_key"
fi

printf '%s\n' "release verify ok"
