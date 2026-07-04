# Provider Proof Blocked

Provider `minimax` has `MINIMAX_API_KEY` set, but it is empty.
Set the environment variable, then rerun this command. The key value is not printed or saved.

## Next Command

```sh
# Export MINIMAX_API_KEY in your shell or local secret manager first.
sh scripts/provider-setup-preflight.sh --providers minimax --output-dir .omo/evidence/provider-setup-preflight-minimax
sh scripts/provider-proof.sh --provider minimax --output-dir .omo/evidence/provider-proof-minimax --timeout-seconds 600
```
