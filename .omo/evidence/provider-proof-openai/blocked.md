# Provider Proof Blocked

Provider `openai` requires `OPENAI_API_KEY` for HTTP benchmark mode.
Set the environment variable, then rerun this command. The key value is not printed or saved.

## Next Command

```sh
# Export OPENAI_API_KEY in your shell or local secret manager first.
sh scripts/provider-setup-preflight.sh --providers openai --output-dir .omo/evidence/provider-setup-preflight-openai
sh scripts/provider-proof.sh --provider openai --output-dir .omo/evidence/provider-proof-openai --timeout-seconds 600
```
