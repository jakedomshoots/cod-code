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

## GitHub Release Publish

Pushing a `v*` tag runs `.github/workflows/release.yml`.

The workflow:

- Derives the CLI version from the tag, for example `v0.1.0` builds `0.1.0`.
- Runs `scripts/release-local.sh`.
- Runs `scripts/verify-release.sh dist`.
- Creates a GitHub Release for the existing tag.
- Uploads `dist/*.tar.gz`, `dist/checksums.txt`, and `dist/release-manifest.json`.
- Writes release notes with the checksum verification command.

This is the explicit public-publish path. Local release commands still do not tag, push, or create a remote release.

## Release Gate

Before cutting a tag, run:

```sh
make ci
make test-race
VERSION=0.1.0 make release-local
sh scripts/verify-release.sh dist
sh scripts/release-bootstrap.sh --dist dist --output-dir .omo/evidence/release-bootstrap
sh scripts/release-preflight.sh dist
sh scripts/release-readiness.sh --dist dist --output-dir .omo/evidence/release-readiness
```

If `task` is installed, the equivalent commands are:

```sh
task ci
task test:race
VERSION=0.1.0 task release:local
sh scripts/verify-release.sh dist
task release:bootstrap
task release:preflight
```

Optional local formula inspection:

```sh
sed -n '1,80p' dist/homebrew/ceo-packet.rb
```

`scripts/verify-release.sh` checks `checksums.txt`, verifies every archive hash and size against `release-manifest.json`, and fails if any artifact is missing or mismatched.

`scripts/release-bootstrap.sh` prepares the public release packet without publishing anything. It writes `index.md`, `summary.json`, `commands.sh`, `env.template`, `release-checklist.md`, `remote-homebrew-formula.rb`, and `verify-release.txt`. It exits non-zero until public repo, release, Homebrew archive, and signing/checksum policy inputs are explicit. `summary.json` records the checklist item count and SHA-256 fingerprints for the bootstrap files.

`scripts/release-preflight.sh` checks whether a release can honestly be called public. It verifies local artifacts, then blocks until a git remote, public release URL, remote Homebrew archive URL, and archive signatures or explicit checksum-only release notes are handled. It does not tag, push, upload, or publish anything.

After the GitHub Release exists, preflight can verify the real release assets:

```sh
GH_RELEASE_TAG=v0.1.0 GH_REPO=<owner>/<repo> sh scripts/release-preflight.sh dist
```

When `GH_RELEASE_TAG` is set, preflight uses `gh release view` to prove the public release has every archive from `release-manifest.json` plus `checksums.txt` and `release-manifest.json`. `GH_REPO` is optional when `origin` is a GitHub remote.

`scripts/release-readiness.sh` writes the durable evidence packet for that decision: `index.md`, `summary.json`, `preflight.md`, `verify-release.txt`, `git-remote.txt`, and `github-auth.txt`. It exits non-zero while public release blockers remain, but still writes the evidence folder so the next action is obvious. Blocked setup evidence records `setup_command_policy: no_publish_no_secret_assignment`, `publish_actions_performed: false`, and `secret_value_saved: false`.

For an unsigned checksum-only first release, the preflight must be explicit:

```sh
ALLOW_CHECKSUM_ONLY_RELEASE=1 CHECKSUM_ONLY_RELEASE_NOTES_URL=https://<release-notes-url> sh scripts/release-preflight.sh dist
```

Bootstrap a first public release plan:

```sh
sh scripts/release-bootstrap.sh \
  --dist dist \
  --output-dir .omo/evidence/release-bootstrap \
  --repo-url https://github.com/<owner>/<repo> \
  --release-url https://github.com/<owner>/<repo>/releases/tag/v0.1.0 \
  --homebrew-archive-base-url https://github.com/<owner>/<repo>/releases/download/v0.1.0 \
  --checksum-notes-url https://github.com/<owner>/<repo>/releases/tag/v0.1.0
```

## Signing

Current local releases are checksum-only. Do not claim signed artifacts until a release signing identity is configured and the signature verification command is documented here.

Planned signing gate:

1. Choose the signing tool and release identity.
2. Sign every archive in `dist/`.
3. Publish the public verification key.
4. Add a copy-paste verification command next to the checksum command.

## Publish Boundary

Do not tag, push, publish a tap, or create a remote release from the local release command. Public publishing happens only through an explicit `v*` tag push to the GitHub release workflow.

A public release is not claimed until the tag, remote artifacts, checksum file, and install docs are visible and verified.

Blocked prerequisites before publishing:

- Public repository or release storage URL.
- Replaced Homebrew placeholder homepage and archive URL.
- Verified checksum from the remote artifact, not just the local file.
- Archive signatures, or `ALLOW_CHECKSUM_ONLY_RELEASE=1` with public checksum-only release notes.
