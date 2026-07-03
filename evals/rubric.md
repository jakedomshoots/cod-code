# CEO Harness Evaluation Rubric

## Artifact-first scoring

Scores come from saved reports, changed files, patch diffs, command results, and real evidence files. A run earns credit only when the evaluator can inspect those artifacts directly.

## Self-report exclusion

Model self-report, summaries, and claimed success are never enough. The report `verdict` may be recorded, but it is not used as proof of success.

## Scoring dimensions

- Task fit: required files changed and forbidden files avoided.
- Verification: required commands ran and passed with exit code 0.
- Diff evidence: required fix terms appear in patch diffs or previews.
- Artifact evidence: required evidence paths are listed and exist on disk.
- Report shape: required structured fields are present.

## Verdicts

- `pass`: every check passes.
- `partial`: at least one check passes and at least one check fails.
- `fail`: no checks pass, the report is malformed, or the task cannot be matched.

## Evidence paths

Evidence paths must be relative paths. The evaluator checks paths under the report directory and optional workspace root.

