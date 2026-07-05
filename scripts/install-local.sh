#!/bin/sh
set -eu

root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
prefix=${PREFIX:-"$HOME/.local"}
bindir=${BINDIR:-"$prefix/bin"}
version=${VERSION:-dev}
commit=${COMMIT:-local}

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

tmpdir=$(mktemp -d)
trap 'rm -rf "$tmpdir"' EXIT

cd "$root"
go build \
	-ldflags "-X ceoharness/internal/cli.Version=$version -X ceoharness/internal/cli.Commit=$commit" \
	-o "$tmpdir/cod" \
	./cmd/ceo-packet

mkdir -p "$bindir"
install -m 0755 "$tmpdir/cod" "$bindir/cod"
install -m 0755 "$tmpdir/cod" "$bindir/ceo-packet"
"$bindir/cod" --version
printf 'installed %s\n' "$bindir/cod"
