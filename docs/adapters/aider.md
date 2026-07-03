# Aider Adapter

Env var: `CEO_AIDER_ADAPTER_COMMAND`

Create a wrapper around your local Aider command. The wrapper must use the probe contract in `docs/adapters/README.md` and avoid changing files during dry-run.

The normal run path should translate Aider output into the harness JSON envelope before writing stdout.
