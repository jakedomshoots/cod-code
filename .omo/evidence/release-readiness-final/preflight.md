# Public Release Preflight

| Check | Status | Detail |
| --- | --- | --- |
| local_release_artifacts | pass | checksums and release-manifest verified |
| git_remote | blocked | no origin remote configured; set GITHUB_REPOSITORY=owner/name for release-only verification |
| remote_release_url | blocked | set RELEASE_PUBLIC_URL to the public HTTPS release page |
| github_release_assets | blocked | set RELEASE_VERSION and GITHUB_REPOSITORY after pushing a v* tag so gh can verify release assets |
| homebrew_formula_url | blocked | dist/homebrew/ceo-packet.rb still uses a local or placeholder URL |
| artifact_signatures | blocked | add archive signatures or set ALLOW_CHECKSUM_ONLY_RELEASE=1 with CHECKSUM_ONLY_RELEASE_NOTES_URL |

public release preflight: blocked (5)
