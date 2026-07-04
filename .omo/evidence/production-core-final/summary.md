# Local Agent Benchmark

Mode: `local_agent_benchmark`

Tasks: 30
Repeats: 1
Concurrency: 1
Timeout retries: 0
Result retries: 0
Runs: 30

| Task | Run | Retry | Agent | Status | Score | Exit | Duration ms | Extra files | Changed files | Evidence |
| --- | ---: | ---: | --- | --- | ---: | ---: | ---: | ---: | --- | --- |
| `bugfix-cli-timeout` | 1 | 1 | CEO Harness | `pass` | 8/8 | 0 | 887 | 0 | `internal/cli/run.go, .omo/evidence/bugfix-cli-timeout.md` | `.omo/evidence/production-core-final/bugfix-cli-timeout/run-01/ceo_harness/score.json` |
| `bugfix-history-latest` | 1 | 1 | CEO Harness | `pass` | 5/5 | 0 | 516 | 0 | `internal/history/query.go, .omo/evidence/bugfix-history-latest.md` | `.omo/evidence/production-core-final/bugfix-history-latest/run-01/ceo_harness/score.json` |
| `bugfix-provider-health-rollup` | 1 | 1 | CEO Harness | `pass` | 5/5 | 0 | 585 | 0 | `internal/cli/provider_health_rollup.go, .omo/evidence/bugfix-provider-health-rollup.md` | `.omo/evidence/production-core-final/bugfix-provider-health-rollup/run-01/ceo_harness/score.json` |
| `bugfix-report-context-truncation` | 1 | 1 | CEO Harness | `pass` | 5/5 | 0 | 512 | 0 | `internal/ceo/context_budget.go, .omo/evidence/bugfix-report-context-truncation.md` | `.omo/evidence/production-core-final/bugfix-report-context-truncation/run-01/ceo_harness/score.json` |
| `refactor-model-selection-split` | 1 | 1 | CEO Harness | `pass` | 5/5 | 0 | 518 | 0 | `internal/cli/model_selection.go, .omo/evidence/refactor-model-selection-split.md` | `.omo/evidence/production-core-final/refactor-model-selection-split/run-01/ceo_harness/score.json` |
| `refactor-text-report-sections` | 1 | 1 | CEO Harness | `pass` | 5/5 | 0 | 498 | 0 | `internal/cli/text_report_sections.go, .omo/evidence/refactor-text-report-sections.md` | `.omo/evidence/production-core-final/refactor-text-report-sections/run-01/ceo_harness/score.json` |
| `refactor-check-fix-prompt` | 1 | 1 | CEO Harness | `pass` | 5/5 | 0 | 559 | 0 | `internal/ceo/check_fix_prompt.go, .omo/evidence/refactor-check-fix-prompt.md` | `.omo/evidence/production-core-final/refactor-check-fix-prompt/run-01/ceo_harness/score.json` |
| `refactor-workspace-brief-excludes` | 1 | 1 | CEO Harness | `pass` | 5/5 | 0 | 499 | 0 | `internal/workspace/brief.go, .omo/evidence/refactor-workspace-brief-excludes.md` | `.omo/evidence/production-core-final/refactor-workspace-brief-excludes/run-01/ceo_harness/score.json` |
| `multi-file-provider-fallback-reporting` | 1 | 1 | CEO Harness | `pass` | 9/9 | 0 | 933 | 0 | `internal/cli/provider_fallback_report.go, internal/config/provider_fallback_policy.go, .omo/evidence/multi-file-provider-fallback-reporting.md` | `.omo/evidence/production-core-final/multi-file-provider-fallback-reporting/run-01/ceo_harness/score.json` |
| `test-repair-require-checks` | 1 | 1 | CEO Harness | `pass` | 5/5 | 0 | 850 | 0 | `internal/cli/require_checks_test.go, .omo/evidence/test-repair-require-checks.md` | `.omo/evidence/production-core-final/test-repair-require-checks/run-01/ceo_harness/score.json` |
| `test-repair-provider-policy` | 1 | 1 | CEO Harness | `pass` | 5/5 | 0 | 837 | 0 | `internal/config/provider_policy_test.go, .omo/evidence/test-repair-provider-policy.md` | `.omo/evidence/production-core-final/test-repair-provider-policy/run-01/ceo_harness/score.json` |
| `test-repair-run-events` | 1 | 1 | CEO Harness | `pass` | 5/5 | 0 | 679 | 0 | `internal/ceo/run_events_test.go, .omo/evidence/test-repair-run-events.md` | `.omo/evidence/production-core-final/test-repair-run-events/run-01/ceo_harness/score.json` |
| `test-repair-smoke-script` | 1 | 1 | CEO Harness | `pass` | 5/5 | 0 | 727 | 0 | `internal/cli/smoke_script_test.go, .omo/evidence/test-repair-smoke-script.md` | `.omo/evidence/production-core-final/test-repair-smoke-script/run-01/ceo_harness/score.json` |
| `docs-verification-record` | 1 | 1 | CEO Harness | `pass` | 5/5 | 0 | 94 | 0 | `docs/VERIFICATION.md, .omo/evidence/docs-verification-record.md` | `.omo/evidence/production-core-final/docs-verification-record/run-01/ceo_harness/score.json` |
| `docs-product-status-weak-spots` | 1 | 1 | CEO Harness | `pass` | 5/5 | 0 | 113 | 0 | `docs/PRODUCT_STATUS.md, .omo/evidence/docs-product-status-weak-spots.md` | `.omo/evidence/production-core-final/docs-product-status-weak-spots/run-01/ceo_harness/score.json` |
| `docs-roadmap-cli-first` | 1 | 1 | CEO Harness | `pass` | 5/5 | 0 | 84 | 0 | `docs/ROADMAP.md, .omo/evidence/docs-roadmap-cli-first.md` | `.omo/evidence/production-core-final/docs-roadmap-cli-first/run-01/ceo_harness/score.json` |
| `provider-config-openai-compatible` | 1 | 1 | CEO Harness | `pass` | 5/5 | 0 | 705 | 0 | `internal/config/http_provider.go, .omo/evidence/provider-config-openai-compatible.md` | `.omo/evidence/production-core-final/provider-config-openai-compatible/run-01/ceo_harness/score.json` |
| `provider-config-budget-metadata` | 1 | 1 | CEO Harness | `pass` | 5/5 | 0 | 511 | 0 | `internal/cli/provider_budget_test.go, .omo/evidence/provider-config-budget-metadata.md` | `.omo/evidence/production-core-final/provider-config-budget-metadata/run-01/ceo_harness/score.json` |
| `provider-config-health-policy` | 1 | 1 | CEO Harness | `pass` | 5/5 | 0 | 510 | 0 | `internal/config/provider_health_policy.go, .omo/evidence/provider-config-health-policy.md` | `.omo/evidence/production-core-final/provider-config-health-policy/run-01/ceo_harness/score.json` |
| `safety-policy-observe-no-write` | 1 | 1 | CEO Harness | `pass` | 6/6 | 0 | 503 | 0 | `internal/cli/write_policy.go, .omo/evidence/safety-policy-observe-no-write.md` | `.omo/evidence/production-core-final/safety-policy-observe-no-write/run-01/ceo_harness/score.json` |
| `safety-policy-approved-digest` | 1 | 1 | CEO Harness | `pass` | 5/5 | 0 | 511 | 0 | `internal/ceo/patch_approval.go, .omo/evidence/safety-policy-approved-digest.md` | `.omo/evidence/production-core-final/safety-policy-approved-digest/run-01/ceo_harness/score.json` |
| `safety-policy-path-escape` | 1 | 1 | CEO Harness | `pass` | 6/6 | 0 | 500 | 0 | `internal/workspace/workspace.go, .omo/evidence/safety-policy-path-escape.md` | `.omo/evidence/production-core-final/safety-policy-path-escape/run-01/ceo_harness/score.json` |
| `recovery-resume-retry` | 1 | 1 | CEO Harness | `pass` | 7/7 | 0 | 511 | 0 | `internal/cli/resume.go, .omo/evidence/recovery-resume-retry.md` | `.omo/evidence/production-core-final/recovery-resume-retry/run-01/ceo_harness/score.json` |
| `safety-policy-rollback-report` | 1 | 1 | CEO Harness | `pass` | 5/5 | 0 | 501 | 0 | `internal/cli/rollback.go, .omo/evidence/safety-policy-rollback-report.md` | `.omo/evidence/production-core-final/safety-policy-rollback-report/run-01/ceo_harness/score.json` |
| `multi-file-operator-safety-flow` | 1 | 1 | CEO Harness | `pass` | 13/13 | 0 | 1038 | 0 | `docs/OPERATOR_SAFETY_FLOW.md, internal/cli/operator_safety_flow.go, internal/config/operator_safety_flow.go, internal/workspace/operator_safety_flow.go, .omo/evidence/multi-file-operator-safety-flow.md` | `.omo/evidence/production-core-final/multi-file-operator-safety-flow/run-01/ceo_harness/score.json` |
| `multi-file-release-readiness-publish-boundary` | 1 | 1 | CEO Harness | `pass` | 12/12 | 0 | 774 | 0 | `docs/RELEASE_READINESS.md, internal/cli/release_readiness_gate.go, internal/config/release_publish_policy.go, .omo/evidence/multi-file-release-readiness-publish-boundary.md` | `.omo/evidence/production-core-final/multi-file-release-readiness-publish-boundary/run-01/ceo_harness/score.json` |
| `multi-file-lean-context-autonomy` | 1 | 1 | CEO Harness | `pass` | 12/12 | 0 | 773 | 0 | `docs/LEAN_CONTEXT_AUTONOMY.md, internal/ceo/lean_context_autonomy.go, internal/subagent/lean_context_packet.go, .omo/evidence/multi-file-lean-context-autonomy.md` | `.omo/evidence/production-core-final/multi-file-lean-context-autonomy/run-01/ceo_harness/score.json` |
| `multi-file-secret-safe-provider-proof` | 1 | 1 | CEO Harness | `pass` | 12/12 | 0 | 883 | 0 | `docs/PROVIDER_PROOF.md, internal/cli/provider_proof_gate.go, internal/model/provider_secret_redaction.go, .omo/evidence/multi-file-secret-safe-provider-proof.md` | `.omo/evidence/production-core-final/multi-file-secret-safe-provider-proof/run-01/ceo_harness/score.json` |
| `multi-file-finalizer-check-fix` | 1 | 1 | CEO Harness | `pass` | 13/13 | 0 | 873 | 0 | `docs/FINALIZER_CHECK_FIX.md, internal/ceo/finalizer_check_fix.go, internal/cli/finalizer_check_fix.go, .omo/evidence/multi-file-finalizer-check-fix.md` | `.omo/evidence/production-core-final/multi-file-finalizer-check-fix/run-01/ceo_harness/score.json` |
| `report-quality-evidence-summary` | 1 | 1 | CEO Harness | `pass` | 7/7 | 0 | 111 | 0 | `docs/REPORT_SCHEMA.md, .omo/evidence/report-quality-evidence-summary.md` | `.omo/evidence/production-core-final/report-quality-evidence-summary/run-01/ceo_harness/score.json` |

Passed: 30
Partial: 0
Failed: 0
Timed out: 0
Setup blocked: 0
Skipped: 0
Incomplete evidence: 0
