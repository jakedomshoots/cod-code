# Provider Setup Checklist

1. Export `KIMI_CODE_API_KEY` in the shell or local secret manager.
2. Keep the key out of git, logs, reports, and evidence folders.
3. Run `sh scripts/provider-setup-preflight.sh --providers kimi-code --output-dir .omo/evidence/provider-setup-preflight-kimi-code`.
4. Run `commands.sh` from the repo root.
5. Confirm `index.md` says `- Overall: pass`.
6. Re-run production readiness.
