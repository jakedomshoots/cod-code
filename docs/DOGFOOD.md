# Dogfood Gate

CEO Harness is not considered product-ready unless the real CLI passes the dogfood gate.

Run it with:

```sh
sh scripts/dogfood.sh
```

Run the market-facing gauntlet with:

```sh
ceo-packet gauntlet --agents ceo_harness --output-dir .omo/evidence/gauntlet
```

Run the production-readiness gauntlet with:

```sh
ceo-packet gauntlet --suite production-core --agents ceo_harness --output-dir .omo/evidence/production-gauntlet
```

The gate builds a fresh local binary and drives these user-visible paths:

1. Built-in demo renders a compact text report.
2. Dry-run patch preview emits a patch approval digest.
3. Approved patch digest applies the exact preview and passes a verification command.
4. A failing verification command produces a non-zero job outcome.
5. A model-backed subagent asks for input.
6. Saved job context prints a copy-pasteable resume command.

This is intentionally not a unit test. It exercises the product surface a user would actually touch.

## Evidence Honesty

The current gauntlet can report partial/incomplete evidence. That is expected when a provider key, CLI login, timeout log, git status snapshot, or scorer artifact is missing. Do not turn that into a false pass; keep the result partial until the raw logs and artifacts prove the outcome.

For recovery drills, use the real commands:

```sh
ceo-packet explain-failure latest --workspace .
ceo-packet retry latest --workspace .
ceo-packet rollback .ceo-harness/history/job-000001.json --workspace .
```
