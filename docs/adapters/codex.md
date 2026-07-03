# Codex CLI Adapter

Env var: `CEO_CODEX_ADAPTER_COMMAND`

Create a small wrapper around your local Codex CLI command. The wrapper must honor `CEO_HARNESS_ADAPTER_PROBE=version` and `CEO_HARNESS_ADAPTER_PROBE=dry-run`, then emit the structured JSON shape from `docs/adapters/README.md`.

Run:

```sh
ceo-packet --config-check --format text
```

Expected setup result: `codex: pass` when the wrapper is configured and returns valid dry-run JSON. If Codex CLI is not installed, leave the env var unset; config-check will report `skip` with setup guidance.
