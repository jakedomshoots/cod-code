# Launch Checklist

Local production ready: true
Public production ready: false

## Required Before Public Production Claim

- Publish release proof: push an explicit `v*` tag so the GitHub release workflow publishes verified tarballs, `checksums.txt`, and `release-manifest.json`; then rerun `sh scripts/release-readiness.sh --dist dist --output-dir .omo/evidence/release-readiness-final` with the public release/Homebrew/signing inputs in place.
- Prove OpenRouter provider: export `OPENROUTER_API_KEY`, run `sh scripts/provider-proof.sh --provider openrouter --output-dir .omo/evidence/provider-proof-openrouter`, and keep the key out of evidence.
- Prove Kimi Code provider: export `KIMI_CODE_API_KEY`, run `sh scripts/provider-proof.sh --provider kimi-code --output-dir .omo/evidence/provider-proof-kimi-code`, and keep the key out of evidence.
- Prove MiniMax provider: export `MINIMAX_API_KEY`, run `sh scripts/provider-proof.sh --provider minimax --output-dir .omo/evidence/provider-proof-minimax`, and keep the key out of evidence.

## Final Gate

After every item above is complete, run:

```sh
sh scripts/production-readiness.sh --dist dist --output-dir .omo/evidence/production-readiness
```
