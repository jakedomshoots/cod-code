# External Adapters

CEO Harness treats external coding tools as adapter wrapper commands. The harness does not call vendor SDKs and does not require the real tool to be installed for tests.

`ceo-packet --config-check` reports all supported adapters:

- `pass`: wrapper command is configured, version probe ran, dry-run output parsed.
- `skip`: wrapper command is missing; read the setup doc and export the env var.
- `fail`: wrapper command ran but timed out, exited non-zero, or emitted invalid output.

## Wrapper Contract

Set the tool env var to an executable command:

```sh
export CEO_CODEX_ADAPTER_COMMAND=/path/to/codex-adapter
```

The wrapper reads the harness prompt on stdin and writes one structured JSON object to stdout:

```json
{"status":"pass","summary":"patch ready","evidence":["adapter dry-run"],"patches":[{"path":"app.txt","old":"old","new":"new"}]}
```

For doctor/config-check probes, wrappers must support:

- `CEO_HARNESS_ADAPTER_PROBE=version`: print a short version string and exit 0.
- `CEO_HARNESS_ADAPTER_PROBE=dry-run`: do not modify files or call paid services unless the wrapper owner explicitly allows it; print valid structured JSON.

Missing wrappers are setup work, not test failure.
