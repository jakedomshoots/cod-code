# release-readiness [release_proof] state: setup_blocked reason: release setup blocked: git_remote, remote_release_url, github_release_assets, homebrew_formula_url, artifact_signatures
# setup actions:
# - git_remote: configure an origin remote for the public repo, for example `git remote add origin git@github.com:<owner>/<repo>.git`, or set `GH_REPO=owner/name` for release-only verification.
# - remote_release_url: set `RELEASE_URL` or `PUBLIC_RELEASE_URL` to the public HTTPS release page.
# - github_release_assets: push a `v*` tag, let the release workflow upload archives, `checksums.txt`, and `release-manifest.json`, then set `GH_RELEASE_TAG` and `GH_REPO` if no GitHub origin is configured.
# - homebrew_formula_url: after release archives are public, run `sh scripts/release-homebrew-formula.sh --dist dist --repo-url <repo-url> --homebrew-archive-base-url <archive-base-url>` or update the tap formula so it uses that remote archive URL.
# - artifact_signatures: run `sh scripts/release-signatures.sh --dist dist --private-key <key.pem>` and rerun preflight with `RELEASE_SIGNING_PUBLIC_KEY=<public.pem>`, or set `ALLOW_CHECKSUM_ONLY_RELEASE=1` with a public `CHECKSUM_ONLY_RELEASE_NOTES_URL`.
# blocked command: sh scripts/release-readiness.sh --dist dist --output-dir .omo/evidence/release-readiness-final
# provider-openai [provider_proof] missing env: OPENAI_API_KEY state: missing_env reason: missing required env: OPENAI_API_KEY
# setup checklist:
# 1. Export `OPENAI_API_KEY` in the shell or local secret manager.
# 2. Keep the key out of git, logs, reports, and evidence folders.
# 3. Run `sh scripts/provider-setup-preflight.sh --providers openai --output-dir .omo/evidence/provider-setup-preflight-openai`.
# 4. Run `commands.sh` from the repo root.
# 5. Confirm `index.md` says `- Overall: pass`.
# 6. Re-run production readiness.
# blocked command: sh scripts/provider-proof.sh --provider openai --output-dir .omo/evidence/provider-proof-openai --timeout-seconds 600
# provider-openrouter [provider_proof] requires env: OPENROUTER_API_KEY state: ready reason: provider proof already passed
sh scripts/provider-proof.sh --provider openrouter --output-dir .omo/evidence/provider-proof-openrouter --timeout-seconds 600
# provider-moonshot [provider_proof] empty env: MOONSHOT_API_KEY state: empty_env reason: required env is set but empty: MOONSHOT_API_KEY
# blocked command: sh scripts/provider-proof.sh --provider moonshot --output-dir .omo/evidence/provider-proof-moonshot --timeout-seconds 600
# competitor-smoke-command [competitor_setup] state: ready reason: status is planned
go run ./cmd/ceo-packet production-finalize --workspace . --dry-run
# production-readiness [final_readiness] state: waiting reason: waiting on: release-readiness, provider-openai, provider-moonshot, competitor-smoke-command waiting on: release-readiness, provider-openai, provider-moonshot, competitor-smoke-command
# blocked command: sh scripts/production-readiness.sh --dist dist --output-dir .omo/evidence/production-readiness-final
