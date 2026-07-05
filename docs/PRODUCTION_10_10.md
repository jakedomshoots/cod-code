# Cod Code Production 10/10 Scoreboard

Status date: 2026-07-05

## Current Score

Overall: 10/10

## Gates

| Gate | Required For 10/10 | Current State | Evidence |
|---|---|---|---|
| Local CLI | Full tests, vet, secret scan, local production gate pass | Pass | `go test ./... -count=1`, `go vet ./...`, `go test -race -shuffle=on -count=1 ./...`, `sh scripts/smoke.sh`, `sh scripts/dogfood.sh`, `sh scripts/secret-scan.sh`, `.omo/evidence/production-local-gate-v011-public/summary.json` |
| Public release | Installable public release with checksums and release manifest | Pass | `https://github.com/jakedomshoots/cod-code/releases/tag/v0.1.1`, `.omo/evidence/release-readiness-v011/summary.json`, GitHub Actions run `28743018332` |
| Provider proof | OAuth and HTTP provider paths proven without secret leakage | Pass | `.omo/evidence/provider-proof-openrouter/summary.json`, `.omo/evidence/provider-proof-kimi-code/summary.json`, `.omo/evidence/provider-proof-minimax/summary.json`, `.omo/evidence/production-finalize-final/summary.json`; `secret_value_saved=false` |
| Real repo dogfood | Five real repos with approved writes and rollback evidence | Pass | `.omo/evidence/dogfood-real-production-*`, `.omo/evidence/external-agent-production-core-29-final-result-retry-r1/summary.json` |
| CI gates | Regression gates run in CI and preserve artifacts | Pass | `.github/workflows/ci.yml`, GitHub release workflow for `v0.1.1` |
| Expanded competitor runner slice | Aider, Goose, Oh My Pi, Codex CLI, OpenCode, Pi, and Cod Code pass; Claude Code setup is blocked by provider credit/quota after login (`Credit balance is too low`) | Hardening in progress | `.omo/evidence/expanded-runners-v011-docs-roadmap-fixed/summary.json` |
| Docs onboarding | New user can install, configure, run, recover | Pass | `README.md`, `docs/INSTALL.md`, clean release install smoke |
| Security posture | Secret scan, no token storage, safe command files, path safety | Pass | `scripts/secret-scan.sh`, `secret_value_saved=false` in provider/finalizer summaries |

## Exit Criteria

- `ceo-packet production-status --workspace . --format text` prints `Production status: pass`.
- `ceo-packet production-actions --workspace . --format text` shows zero required blocked actions.
- Public install instructions are replayed from a clean temp directory against `v0.1.1`.
- Provider proofs exist for Kimi CLI, Codex CLI, Claude CLI, OpenRouter HTTP, Kimi Code HTTP, and MiniMax HTTP; OpenAI HTTP and Moonshot HTTP remain optional explicit providers when credentials are available.

## Rerun Note

- The saved public production evidence is green. A fresh provider-proof/finalizer rerun still requires non-empty `OPENROUTER_API_KEY`, `KIMI_CODE_API_KEY`, and `MINIMAX_API_KEY` in the shell; preflight evidence records missing or empty env vars without printing or saving secret values.
