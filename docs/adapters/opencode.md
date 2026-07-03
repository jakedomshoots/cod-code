# OpenCode Adapter

Env var: `CEO_OPENCODE_ADAPTER_COMMAND`

Create a wrapper around your local OpenCode command. The wrapper must read stdin, honor `CEO_HARNESS_ADAPTER_PROBE=version` and `CEO_HARNESS_ADAPTER_PROBE=dry-run`, and print structured JSON.

If OpenCode is not installed, leave the env var unset. Missing setup is reported as `skip`, not failure.
