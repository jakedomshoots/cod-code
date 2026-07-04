# Release Readiness Evidence

Status: release readiness: blocked

| Check | Status | Evidence |
| --- | --- | --- |
| local_release_artifacts | pass | `verify-release.txt` |
| public_release_preflight | blocked | `preflight.md` |
| git_remote | blocked | `git-remote.txt` |
| github_auth | pass | `github-auth.txt` |

## Blocked Checks

- `git_remote`
- `remote_release_url`
- `github_release_assets`
- `homebrew_formula_url`
- `artifact_signatures`

Setup actions: `setup-actions.md`

## Publish Boundary

This command does not tag, push, upload artifacts, publish a tap, or create a GitHub release.
A public release claim is blocked until `preflight.md` reports pass.
