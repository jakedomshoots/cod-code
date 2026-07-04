# Trust Surface

CEO Harness should be installable, reviewable, and honest about what has actually shipped.

## Verified Today

- Local install from source with `scripts/install-local.sh`.
- Local release archives from `scripts/release-local.sh`.
- SHA-256 checksums in `dist/checksums.txt`.
- Machine-readable release manifest in `dist/release-manifest.json`.
- Public release preflight through `scripts/release-preflight.sh`.
- Durable release-readiness evidence through `scripts/release-readiness.sh`.
- Secret leak scan through `scripts/secret-scan.sh`.
- Checksum verification from inside `dist/`.
- Smoke and dogfood scripts that drive the CLI surface.
- Local production readiness through `production-status`: local ready is true while public ready remains blocked, with `External setup required: true` when only release/provider setup remains.
- Guarded production finalizer evidence: release/provider/readiness commands stay commented while setup is blocked, with no publish/tag/upload or secret-saving behavior.
- Final 29-task all-agent comparison evidence at `.omo/evidence/external-agent-production-core-29-final-result-retry-r1/summary.json`: 116 runs / 116 pass / 0 partial / 0 fail / 0 timeout / 0 incomplete evidence.

## Not Claimed Yet

- No remote `curl | sh` installer is published.
- No Homebrew tap is published.
- No signed release artifacts are published.
- No public production claim is made until release readiness and OpenAI/OpenRouter/Moonshot HTTP provider proofs pass with real external evidence.

## Release Integrity

Build archives:

```sh
VERSION=0.1.0-dev sh scripts/release-local.sh
```

Verify checksums from the release directory:

```sh
cd dist
shasum -a 256 -c checksums.txt
```

Or verify checksums and manifest together:

```sh
sh scripts/verify-release.sh dist
```

Check public-release readiness without publishing:

```sh
sh scripts/release-preflight.sh dist
sh scripts/release-readiness.sh --dist dist --output-dir .omo/evidence/release-readiness
```

If the checkout has no `origin` remote, release verification can use explicit GitHub metadata instead:

```sh
GH_REPO=owner/name GH_RELEASE_TAG=v0.1.0 RELEASE_URL=https://github.com/owner/name/releases/tag/v0.1.0 sh scripts/release-preflight.sh dist
```

Unsigned checksum-only releases must opt in explicitly:

```sh
ALLOW_CHECKSUM_ONLY_RELEASE=1 CHECKSUM_ONLY_RELEASE_NOTES_URL=https://<release-notes-url> sh scripts/release-preflight.sh dist
```

Current releases are checksum-only. Signing is planned after a real release identity is chosen.

## Provider Proof

Paid HTTP provider proof is setup-blocked until these env vars are present and non-empty:

```sh
OPENAI_API_KEY
OPENROUTER_API_KEY
MOONSHOT_API_KEY
```

Check setup without printing or saving secret values:

```sh
sh scripts/provider-setup-preflight.sh --output-dir .omo/evidence/provider-setup-preflight
```

Then run the guarded finalizer:

```sh
go run ./cmd/ceo-packet production-finalize --workspace . --run-comparison
go run ./cmd/ceo-packet production-status --workspace . --format text
```

## Secret Scan

Before release or proof publication, scan source, docs, scripts, and workflow files for committed secret values:

```sh
sh scripts/secret-scan.sh
```

The scan allows placeholders like `OPENAI_API_KEY=...` and skips test fixtures, but blocks real-looking provider keys and GitHub tokens.

## Local Docs Link Check

This grep-based check verifies local markdown links in `README.md`, `SECURITY.md`, `CONTRIBUTING.md`, `CHANGELOG.md`, and `docs/*.md` resolve to existing files:

```sh
python3 - <<'PY'
import pathlib, re, sys
roots = [pathlib.Path("README.md"), pathlib.Path("SECURITY.md"), pathlib.Path("CONTRIBUTING.md"), pathlib.Path("CHANGELOG.md")]
roots += sorted(pathlib.Path("docs").rglob("*.md"))
missing = []
for path in roots:
    text = path.read_text()
    for target in re.findall(r"\[[^\]]+\]\(([^)]+)\)", text):
        if "://" in target or target.startswith("#") or target.startswith("mailto:"):
            continue
        clean = target.split("#", 1)[0]
        if not clean:
            continue
        resolved = (path.parent / clean).resolve()
        if not resolved.exists():
            missing.append(f"{path}: {target}")
if missing:
    print("\n".join(missing))
    sys.exit(1)
print("local markdown links ok")
PY
```
