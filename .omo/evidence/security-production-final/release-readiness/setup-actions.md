# Release Setup Actions

Run these before making a public production release claim.

- git_remote: configure an origin remote for the public repo, for example `git remote add origin git@github.com:<owner>/<repo>.git`, or set `GH_REPO=owner/name` for release-only verification.
- remote_release_url: set `RELEASE_PUBLIC_URL` (or `RELEASE_URL`) to the public HTTPS release page.
- github_release_assets: push a `v*` tag, let the release workflow upload archives, `checksums.txt`, and `release-manifest.json`, then set `RELEASE_VERSION` plus `GITHUB_REPOSITORY` (or `GH_RELEASE_TAG` plus `GH_REPO`) if no GitHub origin is configured.
- homebrew_formula_url: after release archives are public, run `GITHUB_REPOSITORY=<owner>/<repo> HOMEBREW_ARCHIVE_URL=<archive-url> sh scripts/release-homebrew-formula.sh --dist dist` or update the tap formula so it uses that remote archive URL.
- artifact_signatures: run `sh scripts/release-signatures.sh --dist dist --private-key <key.pem>` and rerun preflight with `RELEASE_SIGNING_PUBLIC_KEY=<public.pem>`, or set `ALLOW_CHECKSUM_ONLY_RELEASE=1` with a public `CHECKSUM_ONLY_RELEASE_NOTES_URL`.

After setup is complete, rerun:

```sh
sh scripts/release-readiness.sh --dist dist --output-dir .omo/evidence/release-readiness-final
go run ./cmd/ceo-packet production-finalize --workspace . --dry-run
```
