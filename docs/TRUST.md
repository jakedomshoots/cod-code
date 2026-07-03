# Trust Surface

CEO Harness should be installable, reviewable, and honest about what has actually shipped.

## Verified Today

- Local install from source with `scripts/install-local.sh`.
- Local release archives from `scripts/release-local.sh`.
- SHA-256 checksums in `dist/checksums.txt`.
- Checksum verification from inside `dist/`.
- Smoke and dogfood scripts that drive the CLI surface.

## Not Claimed Yet

- No remote `curl | sh` installer is published.
- No Homebrew tap is published.
- No signed release artifacts are published.

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

Current releases are checksum-only. Signing is planned after a real release identity is chosen.

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

