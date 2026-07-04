# Provider Proof Blocked

Provider `kimi-code` requires `KIMI_CODE_API_KEY` for HTTP benchmark mode.
Set the environment variable, then rerun this command. The key value is not printed or saved.

## Next Command

```sh
# Export KIMI_CODE_API_KEY in your shell or local secret manager first.
sh scripts/provider-setup-preflight.sh --providers kimi-code --output-dir .omo/evidence/provider-setup-preflight-kimi-code
sh scripts/provider-proof.sh --provider kimi-code --output-dir .omo/evidence/provider-proof-kimi-code --timeout-seconds 600
```
