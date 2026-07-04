#!/bin/sh
set -eu

root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
dist=${DIST:-"$root/dist"}
private_key=${RELEASE_SIGNING_KEY:-}
public_key=${RELEASE_SIGNING_PUBLIC_KEY:-${SIGNING_PUBLIC_KEY:-}}
mode=sign

usage() {
  cat <<'USAGE'
Usage:
  sh scripts/release-signatures.sh [--dist dist] --private-key key.pem
  sh scripts/release-signatures.sh [--dist dist] --verify --public-key public.pem

Creates or verifies OpenSSL SHA-256 detached .sig files for every release
archive in dist. This command does not publish, upload, tag, or push anything.

Environment:
  RELEASE_SIGNING_KEY          Private key path used for signing.
  RELEASE_SIGNING_PUBLIC_KEY   Public key path used for verification.
  SIGNING_PUBLIC_KEY           Alternate public key env name.
USAGE
}

while [ "$#" -gt 0 ]; do
  case "$1" in
    --dist)
      [ "$#" -ge 2 ] || {
        printf '%s\n' "release-signatures: --dist requires a value" >&2
        exit 2
      }
      dist="$2"
      shift 2
      ;;
    --private-key)
      [ "$#" -ge 2 ] || {
        printf '%s\n' "release-signatures: --private-key requires a value" >&2
        exit 2
      }
      private_key="$2"
      shift 2
      ;;
    --public-key)
      [ "$#" -ge 2 ] || {
        printf '%s\n' "release-signatures: --public-key requires a value" >&2
        exit 2
      }
      public_key="$2"
      shift 2
      ;;
    --verify)
      mode=verify
      shift
      ;;
    --help|-h)
      usage
      exit 0
      ;;
    *)
      printf '%s\n' "release-signatures: unknown argument: $1" >&2
      usage >&2
      exit 2
      ;;
  esac
done

case "$dist" in
  /*) ;;
  *) dist="$(pwd)/$dist" ;;
esac

if [ ! -d "$dist" ]; then
  printf '%s\n' "release-signatures: missing dist directory: $dist" >&2
  exit 2
fi

if ! command -v openssl >/dev/null 2>&1; then
  printf '%s\n' "release-signatures: openssl is required" >&2
  exit 2
fi

archive_count=0
for archive in "$dist"/*.tar.gz; do
  [ -e "$archive" ] || continue
  archive_count=$((archive_count + 1))
done

if [ "$archive_count" -eq 0 ]; then
  printf '%s\n' "release-signatures: no .tar.gz archives found in $dist" >&2
  exit 2
fi

case "$mode" in
  sign)
    if [ -z "$private_key" ]; then
      printf '%s\n' "release-signatures: --private-key or RELEASE_SIGNING_KEY is required" >&2
      exit 2
    fi
    if [ ! -f "$private_key" ]; then
      printf '%s\n' "release-signatures: private key not found: $private_key" >&2
      exit 2
    fi
    for archive in "$dist"/*.tar.gz; do
      [ -e "$archive" ] || continue
      openssl dgst -sha256 -sign "$private_key" -out "$archive.sig" "$archive"
      printf '%s\n' "signed $(basename "$archive").sig"
    done
    ;;
  verify)
    if [ -z "$public_key" ]; then
      printf '%s\n' "release-signatures: --public-key, RELEASE_SIGNING_PUBLIC_KEY, or SIGNING_PUBLIC_KEY is required" >&2
      exit 2
    fi
    if [ ! -f "$public_key" ]; then
      printf '%s\n' "release-signatures: public key not found: $public_key" >&2
      exit 2
    fi
    for archive in "$dist"/*.tar.gz; do
      [ -e "$archive" ] || continue
      if [ ! -f "$archive.sig" ]; then
        printf '%s\n' "release-signatures: missing signature: $archive.sig" >&2
        exit 1
      fi
      openssl dgst -sha256 -verify "$public_key" -signature "$archive.sig" "$archive" >/dev/null
      printf '%s\n' "verified $(basename "$archive").sig"
    done
    ;;
  *)
    printf '%s\n' "release-signatures: invalid mode: $mode" >&2
    exit 2
    ;;
esac
