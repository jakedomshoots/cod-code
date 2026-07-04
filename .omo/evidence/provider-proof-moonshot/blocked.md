# Provider Proof Blocked

Provider `moonshot` has `MOONSHOT_API_KEY` set, but it is empty.
Set the environment variable, then rerun this command. The key value is not printed or saved.

## Next Command

```sh
# Export MOONSHOT_API_KEY in your shell or local secret manager first.
sh scripts/provider-setup-preflight.sh --providers moonshot --output-dir .omo/evidence/provider-setup-preflight-moonshot
sh scripts/provider-proof.sh --provider moonshot --output-dir .omo/evidence/provider-proof-moonshot --timeout-seconds 600
```
