# Claude Code Adapter

Env var: `CEO_CLAUDE_ADAPTER_COMMAND`

Create a wrapper around your local Claude Code command. The wrapper must support the version and dry-run probes from `docs/adapters/README.md`.

The dry-run path should return structured JSON only. If it prints prose such as `done` or `success`, CEO Harness records `invalid_output` in provider health.

Run:

```sh
ceo-packet --config-check --format text
```
