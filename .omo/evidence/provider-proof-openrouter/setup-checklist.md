# Provider Setup Checklist

1. Fill `OPENROUTER_API_KEY` with a non-empty value in the shell or local secret manager.
2. Keep the key out of git, logs, reports, and evidence folders.
3. Run `sh scripts/provider-setup-preflight.sh --providers openrouter --output-dir .omo/evidence/provider-setup-preflight-openrouter`.
4. Run `commands.sh` from the repo root.
5. Confirm `index.md` says `- Overall: pass`.
6. Re-run production readiness.
