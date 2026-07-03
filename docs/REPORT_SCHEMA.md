# CEO Harness Report Schema v1

This document defines the stable JSON report artifact saved under:

`ceo-artifacts/jobs/<job_id>.json`

## Versioning

New report artifacts include:

```json
{
  "schema_version": 1
}
```

`schema_version` is the top-level contract version for the saved report artifact. Readers must treat missing `schema_version` as a legacy report, not as a fatal error.

## Compatibility

Legacy reports without `schema_version` are compatible with v1 readers when they are valid JSON objects. The CLI keeps the original payload fields readable and adds this marker only when printing a legacy report through `--job-report`:

```json
{
  "schema_compatibility": {
    "status": "legacy",
    "warning": "missing schema_version; treating as legacy report compatible with schema v1",
    "assumed_schema_version": 0,
    "reader_schema_version": 1
  }
}
```

`--job-events` reads legacy reports through the same compatibility path, then emits the saved `run_events` JSONL stream.

## Required Top-Level Fields

New v1 reports should include these top-level fields:

- `schema_version`: integer, currently `1`.
- `job_packet`: compact task packet and selected subagents.
- `job_owner`: current owner role.
- `lifecycle_state`: final lifecycle state.
- `lifecycle_events`: ordered lifecycle events.
- `run_ledger`: compact operator ledger.
- `run_manifest`: run parameters and counts.
- `run_events`: compact event stream used by `--job-events`.
- `context_trace`: metadata-only per-agent packet trace used by `--context-trace`.
- `subagent_results`: bounded subagent outputs.
- `changed_files`: changed workspace artifact paths.
- `check_results`: verification command results.
- `verification_summary`: aggregate verification counts.
- `execution_plan`: final next action.
- `verdict`: final verdict.

Optional fields may appear when that feature is active, including `workspace_brief`, `resume`, `continuation`, `patch_results`, `patch_previews`, `patch_audit`, `patch_approval`, `ceo_delegation`, `ceo_review`, `history_path`, and `job_id`.

`context_trace` entries must stay metadata-only: agent identity, task/assignment summaries, byte budgets, actual context bytes, truncation fields, workspace brief counts/bytes, prior-finding counts/bytes, and excluded content metadata. They must not include raw prompts, repo file contents, environment values, or API keys.

## Reader Rules

- Valid v1 report: parse JSON object with `schema_version: 1`.
- Legacy report: parse JSON object missing `schema_version`, mark it as legacy-compatible, and continue.
- Malformed report: return an error; do not emit a success-looking report.
- Unknown future version: keep the payload readable, but do not claim v1 compatibility unless a migration exists.
