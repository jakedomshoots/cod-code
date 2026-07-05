# Comparison

Cod Code is compared as a local coding orchestration layer, not as an editor. The comparison set is Codex CLI, Claude Code, Aider, OpenCode, Goose, Pi CLI, and Oh My Pi.

## No Result Without Evidence Rule

No competitor gets `pass`, `partial`, or `fail` unless the exact command, raw logs, elapsed time, changed files, safety prompts, provider/cost notes, git status, and evidence paths are saved. A missing binary is `skipped_missing_binary`, not `fail`. A timed-out or hung command is only a failed run when the timeout command and logs are saved.

The gauntlet can produce partial/incomplete evidence. That is not a pass. It means the run needs missing setup or missing artifacts fixed before it can support a market claim.

## Metrics

Every comparison row must cover:

- `task_success`: task scored from saved artifacts, not model self-report.
- `time_to_complete`: wall-clock duration from command start to stop.
- `files_changed`: changed-file list plus git status evidence.
- `safety_prompts`: approvals, permission prompts, or write confirmations shown.
- `cost_provider_used`: provider/model/cost data when the tool exposes it, or `unavailable_with_logs`.
- `evidence_quality`: whether logs, diffs, reports, and artifacts are enough to audit the result.

## Commands

Validate the configured competitors:

```sh
go run ./cmd/ceo-eval --validate-competitors --competitors evals/competitors.json
```

Print plan-only placeholders without running external tools:

```sh
go run ./cmd/ceo-eval --comparison-plan --competitors evals/competitors.json
```

The plan-only output is not a benchmark result. It only says whether the configured binary is discoverable and records an empty placeholder until evidence exists.

Run installed competitor smoke checks before any serious comparison:

```sh
go run ./cmd/ceo-eval \
  --comparison-smoke \
  --competitors evals/competitors.json \
  --output-dir .omo/evidence/competitor-smoke-setup-r1 \
  --timeout-seconds 25
```

Smoke checks are setup proof, not benchmark proof. OpenCode uses a non-editing dry run with logs enabled so provider quota, auth, or credential failures become `setup_blocked` with saved stdout/stderr instead of a vague timeout.

Run installed local agents against the same safe task:

```sh
go run ./cmd/ceo-eval \
  --local-agent-suite \
  --local-agents ceo_harness,codex_cli,claude_code,aider,opencode,goose,pi,oh_my_pi \
  --ceo-binary ./bin/ceo-packet \
  --output-dir .omo/evidence/local-agent-suite-2026-07-02-r4 \
  --timeout-seconds 45
```

Run installed local agents against the same tiny edit task:

```sh
go run ./cmd/ceo-eval \
  --local-agent-suite \
  --local-agent-task edit-file \
  --local-agents ceo_harness,codex_cli,claude_code,aider,opencode,goose,pi,oh_my_pi \
  --ceo-binary ./bin/ceo-packet \
  --output-dir .omo/evidence/local-agent-edit-file-2026-07-02-postsplit-r2 \
  --timeout-seconds 180
```

Latest saved starter-suite results:

| Suite | Claude Code | Aider | Goose | Oh My Pi | Evidence |
| --- | --- | --- | --- | --- | --- |
| Readiness ping | setup-blocked: Claude auth works, but provider reports `Credit balance is too low` | pass, 4549ms | pass, 4739ms | pass, 3888ms | `.omo/evidence/expanded-runners-20260704T231533Z/local-agent-readiness-r2/summary.md`; Claude-only rerun: `.omo/evidence/expanded-runners-20260704T231533Z/claude-readiness-r3/summary.md` |
| Edit file | setup-blocked: Claude auth works, but provider reports `Credit balance is too low` | pass, 5296ms | previous pass, 17548ms; latest 4-agent smoke timed out at 240s | pass, 27435ms | `.omo/evidence/expanded-runners-20260704T231533Z/local-agent-edit-file-r2/summary.md`; Claude-only rerun: `.omo/evidence/expanded-runners-20260704T231533Z/claude-edit-file-r3/summary.md` |

The local-agent suite is still a starter comparison, not a full product benchmark. Exact-content scoring stays strict because previous runs caught newline mismatches and Aider path drift before clean reruns passed.

Run installed local agents against one scored benchmark task:

```sh
go run ./cmd/ceo-eval \
  --local-agent-benchmark \
  --local-agent-benchmark-task docs-roadmap-cli-first \
  --local-agents ceo_harness,codex_cli,claude_code,aider,opencode,goose,pi,oh_my_pi \
  --ceo-binary ./bin/ceo-packet \
  --tasks evals/tasks \
  --output-dir .omo/evidence/local-agent-benchmark-docs-roadmap-2026-07-02-r3 \
  --timeout-seconds 240
```

Run the full 30-task production suite against CEO Harness:

```sh
go run ./cmd/ceo-packet gauntlet \
  --suite production-core \
  --agents ceo_harness \
  --ceo-binary ./bin/ceo-packet \
  --tasks evals/tasks \
  --output-dir .omo/evidence/production-gauntlet \
  --concurrency 4 \
  --timeout-seconds 240
```

Latest production-core all-agent result:

| Benchmark | CEO Harness | Codex CLI | OpenCode | Pi | Evidence |
| --- | --- | --- | --- | --- | --- |
| 25-task `production-core` | 25/25 pass | 25/25 pass | 25/25 pass | 24/25 pass, 1 timeout | `.omo/evidence/external-agent-production-core-25-r1/summary.json` |
| 29-task `production-core` | 29/29 pass | 29/29 pass | 29/29 pass | 29/29 pass | `.omo/evidence/external-agent-production-core-29-final-result-retry-r1/summary.json` |

Latest expanded-runner scored slice:

| Benchmark | Claude Code | Aider | Goose | Oh My Pi | Evidence |
| --- | --- | --- | --- | --- | --- |
| `docs-roadmap-cli-first` | setup-blocked: Claude auth works, but provider reports `Credit balance is too low` | pass 5/5 | pass 5/5 | pass 5/5 | `.omo/evidence/expanded-runners-20260704T231533Z/benchmark-docs-roadmap-r3/summary.json` |

The latest stable 29-task run is a clean comparison pass after enabling one timeout retry and one result retry. Expanded full-runner support is wired, but the expanded matrix is not clean until the Claude Code account has usable provider credit/quota and the full production-core suite reruns with `claude_code,aider,goose,oh_my_pi` included.

Latest CEO-only production-core result:

| Benchmark | CEO Harness | Evidence |
| --- | --- | --- |
| 30-task `production-core` | 30/30 pass, 0 incomplete evidence | `.omo/evidence/production-core-final/summary.json` |

Use `--concurrency N` on long comparisons to shard independent task/agent runs while keeping the final summary in planned task order.

Use bounded timeout retries when comparing flaky external CLIs:

```sh
go run ./cmd/ceo-eval \
  --local-agent-benchmark \
  --local-agent-benchmark-task production-core \
  --local-agent-benchmark-timeout-retries 1 \
  --local-agent-benchmark-result-retries 1 \
  --local-agents ceo_harness,codex_cli,claude_code,aider,opencode,goose,pi,oh_my_pi \
  --ceo-binary ./bin/ceo-packet \
  --tasks evals/tasks \
  --output-dir .omo/evidence/external-agent-production-core-29-retry \
  --concurrency 4 \
  --timeout-seconds 240
```

Timeout retries apply to timed-out runs. Result retries apply to partial or failed runs. Each retry gets its own `attempt-XX` evidence folder, and the summary keeps prior attempts under `prior_attempts`.

Use per-agent timeout overrides when one tool is known to need a longer ceiling:

```sh
go run ./cmd/ceo-eval \
  --local-agent-benchmark \
  --local-agent-benchmark-task docs-product-status-weak-spots \
  --local-agent-benchmark-agent-timeouts opencode=600 \
  --local-agents opencode \
  --tasks evals/tasks \
  --output-dir .omo/evidence/opencode-agent-timeout-r1 \
  --timeout-seconds 240
```

The override is recorded in `summary.json` as `agent_timeouts`.

Use per-agent model overrides when a competitor CLI has multiple configured providers and the default provider is stale, quota-blocked, or the wrong model family:

```sh
go run ./cmd/ceo-eval \
  --local-agent-benchmark \
  --local-agent-benchmark-task docs-product-status-weak-spots \
  --local-agent-benchmark-agent-models opencode=openai/gpt-5.4-mini \
  --local-agents opencode \
  --tasks evals/tasks \
  --output-dir .omo/evidence/opencode-model-override \
  --timeout-seconds 240
```

The selected model is recorded in `summary.json` as `agent_models`. OpenCode benchmark runs also include `--print-logs --log-level INFO` so provider quota/auth failures are visible in `stderr.log`. When those logs show quota, token refresh, or invalid-key errors, the benchmark records `setup_blocked` instead of treating the run as a silent timeout.

Run the focused cross-language suite against CEO Harness:

```sh
go run ./cmd/ceo-packet gauntlet \
  --suite cross-language-core \
  --agents ceo_harness \
  --ceo-binary ./bin/ceo-packet \
  --tasks evals/tasks \
  --output-dir .omo/evidence/cross-language-core-ceo-r1 \
  --concurrency 2 \
  --timeout-seconds 120
```

Latest cross-language result:

| Benchmark | CEO Harness | Evidence |
| --- | --- | --- |
| 2-task `cross-language-core` | 2/2 pass, 0 incomplete evidence | `.omo/evidence/cross-language-core-ceo-r1/summary.json` |

Latest real-provider result:

| Benchmark | Provider Path | CEO Harness | Evidence |
| --- | --- | --- | --- |
| `safety-policy-path-escape` | Kimi CLI OAuth via `scripts/kimi-model-command.sh` | 3/3 pass, 18/18 scored checks, 0 partial, 0 fail, 0 incomplete evidence | `.omo/evidence/provider-kimi-path-safety-repeat-r7/summary.json` |
| `cross-language-js-state-reducer` + `cross-language-python-retry-policy` | Kimi CLI via `scripts/provider-proof.sh --provider kimi` | JS 6/6 pass; Python 7/7 pass; 0 incomplete evidence | `.omo/evidence/provider-proof-kimi-r2/index.md` |
| `cross-language-js-state-reducer` + `cross-language-python-retry-policy` | Codex CLI via `scripts/provider-proof.sh --provider codex` | JS 6/6 pass; Python 7/7 pass; 0 incomplete evidence | `.omo/evidence/provider-proof-codex-r1/index.md` |

Latest focused multi-file external result:

| Benchmark | CEO Harness | Codex CLI | OpenCode | Pi | Evidence |
| --- | --- | --- | --- | --- | --- |
| `multi-file-operator-safety-flow` | pass 13/13 | pass 13/13 | pass 13/13 | pass 13/13 | `.omo/evidence/external-agent-operator-safety-flow-r1/summary.json` |

Latest timeout-retry evidence:

| Benchmark | CEO Harness | Codex CLI | OpenCode | Pi | Evidence |
| --- | --- | --- | --- | --- | --- |
| 5 timeout-heavy current-suite tasks with 1 retry | 5/5 pass | 5/5 pass | 0/5 pass, 5 exhausted timeouts | 5/5 pass | `.omo/evidence/external-agent-timeout-retry-r1/summary.json` |
| OpenCode focused docs task with 600s timeout | n/a | n/a | 0/1 pass, timed out at 600s | n/a | `.omo/evidence/opencode-agent-timeout-r1/summary.json` |

Latest saved benchmark result:

| Benchmark | CEO Harness | Codex CLI | OpenCode | Pi | Evidence |
| --- | --- | --- | --- | --- | --- |
| `docs-roadmap-cli-first` | pass 5/5, 275ms, 0 extra files | pass 5/5, 23726ms, 0 extra files | pass 5/5, 27447ms, 0 extra files | pass 5/5, 94567ms, 0 extra files | `.omo/evidence/local-agent-benchmark-docs-roadmap-2026-07-02-r3/summary.md` |

The first benchmark loop exposed a concrete CEO Harness gap: runtime artifacts were polluting the scored workspace. That is now fixed by running CEO Harness benchmarks with an external `--artifact-root`, so `ceo-artifacts` remain auditable without counting as task changes.

Run a repeated multi-task benchmark:

```sh
go run ./cmd/ceo-eval \
  --local-agent-benchmark \
  --local-agent-benchmark-task docs-roadmap-cli-first,docs-product-status-weak-spots \
  --local-agent-benchmark-repeat 2 \
  --local-agents ceo_harness,codex_cli,claude_code,aider,opencode,goose,pi,oh_my_pi \
  --ceo-binary ./bin/ceo-packet \
  --tasks evals/tasks \
  --output-dir .omo/evidence/local-agent-benchmark-expanded-2026-07-02-r1 \
  --timeout-seconds 240
```

Latest expanded benchmark aggregate:

| Benchmark | CEO Harness | Codex CLI | OpenCode | Pi | Evidence |
| --- | --- | --- | --- | --- | --- |
| 2 docs tasks x 2 repeats | 4/4 pass, avg 81ms, 0 extra files | 4/4 pass, avg 50444ms, 0 extra files | 4/4 pass, avg 19939ms, 0 extra files | 4/4 pass, avg 78689ms, 0 extra files | `.omo/evidence/local-agent-benchmark-expanded-2026-07-02-r1/summary.md` |

Run a Go-file dirty-worktree-sensitive benchmark:

```sh
go run ./cmd/ceo-eval \
  --local-agent-benchmark \
  --local-agent-benchmark-task bugfix-cli-timeout \
  --local-agents ceo_harness,codex_cli,claude_code,aider,opencode,goose,pi,oh_my_pi \
  --ceo-binary ./bin/ceo-packet \
  --tasks evals/tasks \
  --output-dir .omo/evidence/local-agent-benchmark-go-bugfix-2026-07-02-r2 \
  --timeout-seconds 240
```

Latest Go-file benchmark result:

| Benchmark | CEO Harness | Codex CLI | OpenCode | Pi | Evidence |
| --- | --- | --- | --- | --- | --- |
| `bugfix-cli-timeout` | pass 8/8, 255ms, 0 extra files | pass 8/8, 97392ms, 0 extra files | pass 8/8, 36754ms, 0 extra files | pass 8/8, 142473ms, 0 extra files | `.omo/evidence/local-agent-benchmark-go-bugfix-2026-07-02-r2/summary.md` |

The second benchmark loop exposed two runner gaps and closed both: repeated task coverage now writes separate evidence folders, and Go-file fixtures now compile before agents edit them. Dirty-worktree evidence is also included in local-agent reports, so dirty-sensitive tasks score from saved git status rather than partial reports.

Run the strengthened path-safety benchmark against CEO Harness only:

```sh
go run ./cmd/ceo-eval \
  --local-agent-benchmark \
  --local-agent-benchmark-task safety-policy-path-escape \
  --local-agents ceo_harness \
  --ceo-binary ./bin/ceo-packet \
  --tasks evals/tasks \
  --output-dir .omo/evidence/local-agent-benchmark-safety-path-escape-2026-07-02-r1 \
  --timeout-seconds 240
```

Latest path-safety result:

| Benchmark | CEO Harness | Evidence |
| --- | --- | --- |
| `safety-policy-path-escape` | pass 6/6, 141ms, 0 extra files | `.omo/evidence/local-agent-benchmark-safety-path-escape-2026-07-02-r1/summary.md` |

This is a CEO Harness proof run, not a cross-agent comparison. The fixture now includes real path-escape tests, and the score requires the forbidden `../outside.txt` path to stay absent.

Run the same path-safety task through the real CEO model-command patch path:

```sh
./bin/ceo-eval \
  --local-agent-benchmark \
  --local-agent-benchmark-task safety-policy-path-escape \
  --local-agents ceo_harness \
  --ceo-binary ./bin/ceo-packet \
  --ceo-benchmark-mode model-command \
  --ceo-benchmark-model-command-json '["go","run","/Users/jakedom/Documents/Codex/2026-06-30/new-chat/work/ceo-harness/scripts/eval-path-escape-model.go"]' \
  --tasks evals/tasks \
  --output-dir .omo/evidence/local-agent-benchmark-safety-path-escape-model-command-verified-2026-07-02-r1 \
  --timeout-seconds 240
```

Latest model-command path result:

| Benchmark | CEO Harness | Evidence |
| --- | --- | --- |
| `safety-policy-path-escape` via `--model-command` + CEO check | pass 6/6, 1194ms, 0 extra files | `.omo/evidence/local-agent-benchmark-safety-path-escape-model-command-verified-2026-07-02-r1/summary.md` |

This proves the full CEO patch route for this task: scanner/coder/reviewer delegation, lean context packets, model-command output, model-sourced patch audit, patch application, CEO-side check execution, CEO verdict, and saved scorer evidence. The CEO report records `verification_contract.status=pass` with one required check run and passed. It uses a deterministic local model helper, not an external frontier model, so it proves runtime plumbing rather than broad autonomous reasoning quality.

Run the same path-safety task through a real Codex CLI model command:

```sh
./bin/ceo-eval \
  --local-agent-benchmark \
  --local-agent-benchmark-task safety-policy-path-escape \
  --local-agents ceo_harness \
  --ceo-binary ./bin/ceo-packet \
  --ceo-benchmark-mode model-command \
  --ceo-benchmark-model-command-json '["sh","/Users/jakedom/Documents/Codex/2026-06-30/new-chat/work/ceo-harness/scripts/codex-model-command.sh"]' \
  --tasks evals/tasks \
  --output-dir .omo/evidence/local-agent-benchmark-safety-path-escape-codex-real-2026-07-02-r2 \
  --timeout-seconds 600
```

Latest real-model path result:

| Benchmark | CEO Harness | Evidence |
| --- | --- | --- |
| `safety-policy-path-escape` via Codex CLI real model + CEO model command | pass 6/6, 27387ms, 0 extra files | `.omo/evidence/local-agent-benchmark-safety-path-escape-codex-real-2026-07-02-r2/summary.md` |

This is the first saved real-provider proof for the CEO Harness path. The benchmark command wired the real model into both `--model-command` and `--ceo-model-command`. The CEO model selected only `coder`, the coder proposed a patch to `internal/workspace/workspace.go`, the harness applied it, the required Go path test passed, the CEO model returned `recommended_verdict=pass`, and the external scorer passed 6/6.

Run the same path-safety task through the OAuth-backed Kimi CLI:

```sh
./bin/ceo-eval \
  --local-agent-benchmark \
  --local-agent-benchmark-task safety-policy-path-escape \
  --local-agents ceo_harness \
  --ceo-binary ./bin/ceo-packet \
  --ceo-benchmark-mode model-command \
  --ceo-benchmark-model-command-json '["sh","/Users/jakedom/Documents/Codex/2026-06-30/new-chat/work/ceo-harness/scripts/kimi-model-command.sh"]' \
  --tasks evals/tasks \
  --local-agent-benchmark-repeat 3 \
  --output-dir .omo/evidence/provider-kimi-path-safety-repeat-r7 \
  --timeout-seconds 600
```

Latest Kimi CLI real-model result:

| Benchmark | CEO Harness | Evidence |
| --- | --- | --- |
| `safety-policy-path-escape` via Kimi CLI OAuth + CEO model command | 3/3 pass, 18/18 scored checks, 0 partial, 0 fail, 0 incomplete evidence | `.omo/evidence/provider-kimi-path-safety-repeat-r7/summary.json` |

The Kimi wrapper uses `kimi -p ... --output-format stream-json`, extracts the assistant JSON, isolates Kimi in a temporary directory so it cannot mutate the scored workspace directly, normalizes common patch/tool-request shorthand, and feeds strict JSON into CEO Harness. The repeated passing run wired Kimi into both subagent patching and CEO delegation/review.

HTTP provider mode is also wired for OpenAI-compatible providers, but the local shell did not have a real API key available during the first OpenRouter attempt. The saved blocked run is `.omo/evidence/local-agent-benchmark-safety-path-escape-openrouter-gpt5mini-2026-07-02-r1/summary.md`, with stderr reporting that `OPENROUTER_API_KEY` was required.

Current comparison reports include a `Readiness Decision` section. This keeps the market claim honest: the report can show `CEO Harness result: clean` while still marking `Overall comparison: blocked` when external CLIs time out or leave incomplete evidence.

Operator setup shorthand:

- Codex CLI path: use the `codex` adapter preset or the saved Codex model-command benchmark path.
- Kimi CLI path: use the `kimi` adapter preset or `scripts/kimi-model-command.sh` when OAuth-backed Kimi CLI is available.
- OpenRouter path: use `ceo-packet --provider-wizard openrouter`; a missing `OPENROUTER_API_KEY` is blocked setup, not a failed product result.
- Codex/Kimi/OpenRouter missing key or missing login states must be recorded as setup blockers with command output.

## Result Statuses

- `planned_no_result`: binary was found, but no benchmark task has been run.
- `skipped_missing_binary`: binary was not found on `PATH`; setup is missing, not product failure.
- `pass`: saved evidence proves the task passed the rubric.
- `partial`: saved evidence proves some checks passed and some failed.
- `fail`: saved command/log evidence proves the task failed.
- `timeout`: saved timeout/log evidence proves the tool exceeded the limit.

## Evidence Layout

Real comparison runs should save artifacts under:

```text
.omo/evidence/comparison/<date>/<tool>/<task-id>/
```

Each task folder should include:

- `command.txt`
- `stdout.log`
- `stderr.log`
- `git-status-before.txt`
- `git-status-after.txt`
- `changed-files.txt`
- `timing.txt`
- `provider-cost.txt`
- `safety-prompts.txt`
- `result.json`

## Adversarial Checks

- `stale_state`: compare git status before and after, and reject stale reports whose evidence no longer matches the workspace.
- `misleading_success_output`: ignore self-reported success unless command exit status, diffs, artifacts, and rubric checks support it.
- `dirty_worktree`: record pre-existing dirty files before running any tool.
- `hung_long_commands`: wrap real tool runs with a timeout and save timeout output before assigning `timeout` or `fail`.
