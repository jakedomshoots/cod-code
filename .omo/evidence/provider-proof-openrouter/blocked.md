# Provider Proof Blocked

Provider `openrouter` has `OPENROUTER_API_KEY` set, but it is empty.
Set the environment variable, then rerun this command. The key value is not printed or saved.

## Next Command

```sh
# Export OPENROUTER_API_KEY in your shell or local secret manager first.
sh scripts/provider-setup-preflight.sh --providers openrouter --output-dir .omo/evidence/provider-setup-preflight-openrouter
sh scripts/provider-proof.sh --provider openrouter --output-dir .omo/evidence/provider-proof-openrouter --timeout-seconds 600
```
