# Production Setup Actions

Use this as the single checklist before claiming public production readiness.

## Release

- Follow `.omo/evidence/production-finalize-final/release-readiness-final/setup-actions.md`.

## Providers

- openrouter: follow `.omo/evidence/production-finalize-final/provider-proof-openrouter/setup-checklist.md` and rerun `.omo/evidence/production-finalize-final/provider-proof-openrouter/commands.sh` after the required env var is set.
- kimi-code: follow `.omo/evidence/production-finalize-final/provider-proof-kimi-code/setup-checklist.md` and rerun `.omo/evidence/production-finalize-final/provider-proof-kimi-code/commands.sh` after the required env var is set.
- minimax: follow `.omo/evidence/production-finalize-final/provider-proof-minimax/setup-checklist.md` and rerun `.omo/evidence/production-finalize-final/provider-proof-minimax/commands.sh` after the required env var is set.

## Competitors

- Follow `.omo/evidence/production-finalize-final/competitor-smoke/setup-actions.md`.

## Final Rerun

```sh
go run ./cmd/ceo-packet production-finalize --workspace . --dry-run
go run ./cmd/ceo-packet production-finalize --workspace . --run-comparison
sh scripts/production-readiness.sh --dist dist --output-dir .omo/evidence/production-readiness-final
```
