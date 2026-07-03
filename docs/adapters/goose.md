# Goose Adapter

Env var: `CEO_GOOSE_ADAPTER_COMMAND`

Create a wrapper around your local Goose command. The wrapper must support version and dry-run probes, then emit the structured JSON envelope described in `docs/adapters/README.md`.

If the wrapper hangs, `ceo-packet --config-check` records a typed `timeout` health issue.
