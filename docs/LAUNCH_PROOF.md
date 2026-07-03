# CEO Harness Launch Proof

Date: 2026-07-02

## Go / No-Go

No-go for a public "10/10 beats competitors" launch claim.

The prior full-benchmark proof gap is closed for the eval scorer: all 21 benchmark tasks now have saved deterministic fixture reports, score JSON, per-task logs, and a summary with 21 pass / 0 partial / 0 fail / 0 skipped.

The remaining limit is competitor proof. This run upgraded comparison evidence from plan-only to local version/dry-run smoke where installed: Codex CLI and OpenCode ran successfully; Claude Code, Aider, and Goose were recorded as `skipped_missing_binary` with setup guidance. That is useful setup evidence, but it is not a live head-to-head task comparison.

## Evidence Root

- Task evidence: `.omo/evidence/task-14-ceo-harness-10-out-of-10/`
- Dogfood evidence: `.omo/evidence/dogfood-real/index.md`
- User report: `outputs/ceo-harness-10-out-of-10-report.md`

## Required Gates

| Gate | Invocation | Observable | Artifact |
| --- | --- | --- | --- |
| Repo CI | `make ci` | Exit 0; tests, vet, smoke, dogfood, build passed | `.omo/evidence/task-14-ceo-harness-10-out-of-10/make-ci-task14-fix.log` |
| Task 14 code review / slop | Programming Go + remove-ai-slops review over gate blocker scope | Oversized eval file split, dead test helper removed, no test slop found in scope | `.omo/evidence/task-14-ceo-harness-10-out-of-10-code-review-slop.md` |
| Go LOC ceiling | `awk` pure-LOC scan over `internal/eval/*.go` | Every Go file is <=250 pure LOC; `benchmark.go` is 124 | `.omo/evidence/task-14-ceo-harness-10-out-of-10/pure-loc-task14-fix.log` |
| Race/shuffle | `go test -race -shuffle=on -count=1 ./...` | Exit 0 on fresh 2026-07-02 rerun after adapter timeout-proof stabilization | `.omo/evidence/final-adapter-version-timeout-fix/go-test-race-shuffle-all.log` |
| Release + checksums | `VERSION=0.1.0-dev sh scripts/release-local.sh && cd dist && shasum -a 256 -c checksums.txt` | All three archives `OK` | `.omo/evidence/task-14-ceo-harness-10-out-of-10/release-local-checksums.log`, `dist/checksums.txt` |
| Eval catalog | `go run ./cmd/ceo-eval --list` | 21 task IDs listed | `.omo/evidence/task-14-ceo-harness-10-out-of-10/eval-list.txt` |
| Eval rubric | `go run ./cmd/ceo-eval --rubric` | `rubric_valid=true` | `.omo/evidence/task-14-ceo-harness-10-out-of-10/eval-rubric.txt` |
| Benchmark scoring pass | `go run ./cmd/ceo-eval --task bugfix-cli-timeout --report internal/eval/testdata/dirty-worktree/happy/report.json --workspace <temp>` | Verdict `pass`, 8/8 checks | `.omo/evidence/task-14-ceo-harness-10-out-of-10/eval-score-dirty-happy.json` |
| Full benchmark fixture suite | `go run ./cmd/ceo-eval --benchmark-fixtures --tasks evals/tasks --output-dir .omo/evidence/task-14-ceo-harness-10-out-of-10/benchmark-fixtures` | 21 tasks scored; 21 pass, 0 partial, 0 fail, 0 skipped | `.omo/evidence/task-14-ceo-harness-10-out-of-10/benchmark-fixtures/summary.json`, per-task `report.json`, `score.json`, `score.log` |
| Failure injection | `go run ./cmd/ceo-eval --task bugfix-cli-timeout --report internal/eval/testdata/corrupt/report.json --workspace .` | Expected non-zero; corrupt JSON rejected | `.omo/evidence/task-14-ceo-harness-10-out-of-10/eval-score-forced-failure.stderr` |
| Competitor comparison | `go run ./cmd/ceo-eval --validate-competitors --competitors evals/competitors.json`, `go run ./cmd/ceo-eval --comparison-plan --competitors evals/competitors.json`, and `go run ./cmd/ceo-eval --comparison-smoke --competitors evals/competitors.json --output-dir .omo/evidence/task-14-ceo-harness-10-out-of-10/competitor-smoke --timeout-seconds 15` | Config valid; plan exists; local smoke ran installed binaries only: 2 pass, 3 skipped missing binary | `.omo/evidence/task-14-ceo-harness-10-out-of-10/competitor-smoke/summary.json` |
| Real repo dogfood | `sh scripts/dogfood-real.sh --repo temp-real:<temp-git-repo> --timeout-ms 250` | Temp external git repo row `pass`; 5 scenarios captured | `.omo/evidence/dogfood-real/index.md` |
| Manual binary QA | `bin/ceo-packet` driven through start/config/provider/demo/write/TUI/history paths, then `cd .omo/evidence/task-14-ceo-harness-10-out-of-10/manual-qa && shasum -a 256 -c SHA256SUMS.txt` | Exit 0; per-surface artifacts hashed and every listed file verifies `OK` | `.omo/evidence/task-14-ceo-harness-10-out-of-10/manual-binary-qa.log`, `.omo/evidence/task-14-ceo-harness-10-out-of-10/manual-qa/SHA256SUMS.txt`, `.omo/evidence/task-14-ceo-harness-10-out-of-10/manual-qa-sha256-verify-task14-fix.log` |
| Final timeout/doc recheck | `go test -race -shuffle=on -count=20 ./internal/adapter`, `go test ./internal/checkrunner -run Test_Runner_Run_cancels_shell_process_group_when_timeout_expires -count=50 -v`, `go test ./internal/model -run Test_CommandClient_Complete_kills_shell_process_group_when_timeout_expires -count=20 -v`, and `go run ./cmd/ceo-packet --model-command-timeout-ms 50 ... sleep 5` | Adapter version retry proof stable; checkrunner/model process-tree proofs still pass; CLI timeout returns `provider_error_kind: command_timeout` in about 0.18s; no leftover sleep/ceo timeout processes | `.omo/evidence/final-adapter-version-timeout-fix/` |

## Manual QA Surfaces

| Scenario | Invocation | Observable | Artifact |
| --- | --- | --- | --- |
| First run guidance | `bin/ceo-packet start <temp> --format text` | `Start: pass`, next commands printed | `.omo/evidence/task-14-ceo-harness-10-out-of-10/manual-qa/first-run-start.txt` |
| Config doctor | `bin/ceo-packet config doctor --workspace <temp> --format text` | `Config doctor: pass` | `.omo/evidence/task-14-ceo-harness-10-out-of-10/manual-qa/config-doctor.txt` |
| Provider wizard | `bin/ceo-packet --workspace <temp> --provider-wizard openai --http-model gpt-5 --format text` | Provider config written | `.omo/evidence/task-14-ceo-harness-10-out-of-10/manual-qa/provider-wizard-openai.txt` |
| Provider doctor guidance | `bin/ceo-packet --workspace <temp> --doctor-provider main --format text` | Expected fail: missing `OPENAI_API_KEY`, no secret printed | `.omo/evidence/task-14-ceo-harness-10-out-of-10/manual-qa/provider-doctor-main.txt` |
| Adapter doctor/setup | `bin/ceo-packet config check --workspace <temp> --format text` | All five adapters reported as missing setup, not false failure | `.omo/evidence/task-14-ceo-harness-10-out-of-10/manual-qa/adapter-config-check.txt` |
| Demo repo default preview | `bin/ceo-packet --init-demo-repo <temp>/demo`, then `bin/ceo-packet --workspace <temp>/demo --replace app.txt old new --format json "Patch demo app"` | Patch approval status `previewed`; `patch_results` stayed null; `app.txt` stayed `hello old` because default writes do not mutate | `.omo/evidence/final-checkrunner-flake-fix/manual-default-preview-with-model.json`, `.omo/evidence/final-checkrunner-flake-fix/manual-default-preview-with-model-app.txt` |
| Approved write reject | `bin/ceo-packet --workspace <temp> --write-policy approved-write --replace app.txt old new --format json "Reject missing digest"` | Expected non-zero; missing digest rejected | `.omo/evidence/task-14-ceo-harness-10-out-of-10/manual-qa/approved-write-missing-digest.stderr` |
| Approved write apply | `bin/ceo-packet --workspace <temp> --write-policy approved-write --approve-preview <digest> --replace app.txt old new --format json "Apply approved patch"` | File became `hello new` | `.omo/evidence/task-14-ceo-harness-10-out-of-10/manual-qa/approved-write-apply.json` |
| Rollback | `bin/ceo-packet --workspace <temp> --rollback-report <apply-report> --format json` | File restored to `hello old` | `.omo/evidence/task-14-ceo-harness-10-out-of-10/manual-qa/rollback.json` |
| TUI snapshot | `bin/ceo-packet tui --workspace <demo> --snapshot` | Text dashboard rendered with job and next actions | `.omo/evidence/task-14-ceo-harness-10-out-of-10/manual-qa/tui-snapshot.txt` |
| History/report inspection | `status`, `--job latest`, `--job-report latest`, `--job-events latest`, `context latest` | Saved job, report, events, context trace readable | `.omo/evidence/task-14-ceo-harness-10-out-of-10/manual-qa/job-report-latest.json` |

## Dogfood Summary

The dogfood runner used a temp external git repo, not a mocked internal fixture. It captured five scenarios:

- `scenario-01-doctor`: pass
- `scenario-02-plan-only`: pass
- `scenario-03-observe-run`: pass
- `scenario-04-patch-preview`: pass
- `scenario-05-timeout-guard`: `pass_expected_failure`

Artifact: `.omo/evidence/dogfood-real/repos/temp-real/summary.md`

## Benchmark Summary

The full benchmark suite was scored with deterministic fixture reports generated from `evals/tasks/benchmark_tasks.json`. This proves the scorer can evaluate every task definition and save per-task evidence. It does not prove an autonomous agent completed all 21 tasks live.

- Mode: `deterministic_fixture_scoring`
- Tasks: 21
- Passed: 21
- Partial: 0
- Failed: 0
- Skipped: 0

Artifacts:

- `.omo/evidence/task-14-ceo-harness-10-out-of-10/benchmark-fixtures/summary.json`
- `.omo/evidence/task-14-ceo-harness-10-out-of-10/benchmark-fixtures/*/report.json`
- `.omo/evidence/task-14-ceo-harness-10-out-of-10/benchmark-fixtures/*/score.json`
- `.omo/evidence/task-14-ceo-harness-10-out-of-10/benchmark-fixtures/*/score.log`

## Competitor Summary

The comparison harness now has local smoke evidence, not only a plan. It ran safe version/dry-run commands for installed competitor binaries and recorded missing tools as skips, not failures.

- `codex_cli`: `smoke_pass`, `/Users/jakedom/.local/bin/codex`, `codex-cli 0.142.4`
- `opencode`: `smoke_pass`, `/Users/jakedom/.opencode/bin/opencode`, `1.17.13`
- `claude_code`: `skipped_missing_binary`, install/auth setup required
- `aider`: `skipped_missing_binary`, install/provider setup required
- `goose`: `skipped_missing_binary`, install/provider setup required

Artifact: `.omo/evidence/task-14-ceo-harness-10-out-of-10/competitor-smoke/summary.json`

## Blockers / Risks

- Public "10/10 beats competitors" claim is still unsupported because no live head-to-head benchmark tasks were run against competitor tools.
- The 21/21 benchmark evidence is deterministic fixture-based scorer evidence, not proof that CEO Harness autonomously solved all 21 benchmark tasks.
- Provider doctor correctly fails without `OPENAI_API_KEY`; this is setup guidance, not a product pass against a real provider.
- Rollback QA passed for the supported simple replace path. Multiline/trailing-newline rollback remains a limitation.

## Cleanup

All temp QA workspaces recorded by task 14 were removed.

Artifact: `.omo/evidence/task-14-ceo-harness-10-out-of-10/cleanup.log`
