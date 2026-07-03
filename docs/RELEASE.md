# Release Process

CEO Harness releases are built from the Go CLI directly. The first supported release shape is archive-based: tarballs plus SHA-256 checksums.

## Local Release Build

```sh
VERSION=0.1.0 sh scripts/release-local.sh
```

Artifacts are written to `dist/`:

- `ceo-packet_<version>_darwin_arm64.tar.gz`
- `ceo-packet_<version>_linux_amd64.tar.gz`
- `ceo-packet_<version>_linux_arm64.tar.gz`
- `checksums.txt`
- `release-manifest.json`
- `homebrew/ceo-packet.rb`

The Homebrew formula is a local draft that points at the generated Darwin archive with the matching checksum. It is for review or local formula testing only; it is not a published tap.

## Release Gate

Before cutting a tag, run:

```sh
make ci
make test-race
VERSION=0.1.0 make release-local
sh scripts/verify-release.sh dist
sh scripts/release-preflight.sh dist
sh scripts/release-readiness.sh --dist dist --output-dir .omo/evidence/release-readiness
```

If `task` is installed, the equivalent commands are:

```sh
task ci
task test:race
VERSION=0.1.0 task release:local
sh scripts/verify-release.sh dist
task release:preflight
```

Optional local formula inspection:

```sh
sed -n '1,80p' dist/homebrew/ceo-packet.rb
```

`scripts/verify-release.sh` checks `checksums.txt`, verifies every archive hash and size against `release-manifest.json`, and fails if any artifact is missing or mismatched.

`scripts/release-preflight.sh` checks whether a release can honestly be called public. It verifies local artifacts, then blocks until a git remote, public release URL, remote Homebrew archive URL, and archive signatures or explicit checksum-only release notes are handled. It does not tag, push, upload, or publish anything.

`scripts/release-readiness.sh` writes the durable evidence packet for that decision: `index.md`, `summary.json`, `preflight.md`, `verify-release.txt`, `git-remote.txt`, and `github-auth.txt`. It exits non-zero while public release blockers remain, but still writes the evidence folder so the next action is obvious.

For an unsigned checksum-only first release, the preflight must be explicit:

```sh
ALLOW_CHECKSUM_ONLY_RELEASE=1 CHECKSUM_ONLY_RELEASE_NOTES_URL=https://<release-notes-url> sh scripts/release-preflight.sh dist
```

## Signing

Current local releases are checksum-only. Do not claim signed artifacts until a release signing identity is configured and the signature verification command is documented here.

Planned signing gate:

1. Choose the signing tool and release identity.
2. Sign every archive in `dist/`.
3. Publish the public verification key.
4. Add a copy-paste verification command next to the checksum command.

## Publish Boundary

Do not tag, push, publish a tap, or create a remote release from the local release command. That needs explicit user approval and a real public release URL.

A public release is not claimed until the tag, remote artifacts, checksum file, and install docs are visible and verified.

Blocked prerequisites before publishing:

- Public repository or release storage URL.
- Replaced Homebrew placeholder homepage and archive URL.
- Verified checksum from the remote artifact, not just the local file.
- Archive signatures, or `ALLOW_CHECKSUM_ONLY_RELEASE=1` with public checksum-only release notes.
