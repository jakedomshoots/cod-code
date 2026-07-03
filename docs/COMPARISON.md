# Competitor Comparison

CEO Harness is compared as a local coding orchestration layer, not as an editor. The comparison set is Codex CLI, Claude Code, Aider, OpenCode, and Goose.

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

Run installed local agents against the same safe task:

```sh
go run ./cmd/ceo-eval \
  --local-agent-suite \
  --local-agents ceo_harness,codex_cli,opencode,pi \
  --ceo-binary ./bin/ceo-packet \
  --output-dir .omo/evidence/local-agent-suite-2026-07-02-r4 \
  --timeout-seconds 45
```

Run installed local agents against the same tiny edit task:

```sh
go run ./cmd/ceo-eval \
  --local-agent-suite \
  --local-agent-task edit-file \
  --local-agents ceo_harness,codex_cli,opencode,pi \
  --ceo-binary ./bin/ceo-packet \
  --output-dir .omo/evidence/local-agent-edit-file-2026-07-02-postsplit-r2 \
  --timeout-seconds 180
```

Latest saved live results:

| Suite | CEO Harness | Codex CLI | OpenCode | Pi | Evidence |
| --- | --- | --- | --- | --- | --- |
| Readiness ping | pass, 19ms | pass, 5028ms | pass, 3320ms | pass, 7588ms | `.omo/evidence/local-agent-suite-2026-07-02-r4/summary.md` |
| Edit file | pass, 6ms | pass, 19815ms | pass, 11728ms | pass, 25796ms | `.omo/evidence/local-agent-edit-file-2026-07-02-postsplit-r2/summary.md` |

The local-agent suite is still a starter comparison, not a full product benchmark. The first improvement-loop item is `benchmark-task-runner`: run real benchmark tasks with repo reset, scoring, changed-file checks, and per-agent artifacts instead of only a one-file mutation. Exact-content scoring stays strict because one post-split run caught a real newline mismatch before a clean rerun passed.

Run installed local agents against one scored benchmark task:

```sh
go run ./cmd/ceo-eval \
  --local-agent-benchmark \
  --local-agent-benchmark-task docs-roadmap-cli-first \
  --local-agents ceo_harness,codex_cli,opencode,pi \
  --ceo-binary ./bin/ceo-packet \
  --tasks evals/tasks \
  --output-dir .omo/evidence/local-agent-benchmark-docs-roadmap-2026-07-02-r3 \
  --timeout-seconds 240
```

Run the full 24-task production suite against CEO Harness:

```sh
go run ./cmd/ceo-packet gauntlet \
  --suite production-core \
  --agents ceo_harness \
  --ceo-binary ./bin/ceo-packet \
  --tasks evals/tasks \
  --output-dir .omo/evidence/production-gauntlet \
  --timeout-seconds 240
```

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
  --local-agents ceo_harness,codex_cli,opencode,pi \
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
  --local-agents ceo_harness,codex_cli,opencode,pi \
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
  --output-dir .omo/evidence/local-agent-benchmark-safety-path-escape-kimi-real-2026-07-02-r3 \
  --timeout-seconds 600
```

Latest Kimi CLI real-model result:

| Benchmark | CEO Harness | Evidence |
| --- | --- | --- |
| `safety-policy-path-escape` via Kimi CLI OAuth + CEO model command | pass 6/6, 97616ms, 0 extra files | `.omo/evidence/local-agent-benchmark-safety-path-escape-kimi-real-2026-07-02-r3/summary.md` |

The Kimi wrapper uses `kimi -p ... --output-format stream-json`, extracts the assistant JSON, and feeds that into CEO Harness. The final passing run wired Kimi into both subagent patching and CEO delegation/review.

HTTP provider mode is also wired for OpenAI-compatible providers, but the local shell did not have a real API key available during the first OpenRouter attempt. The saved blocked run is `.omo/evidence/local-agent-benchmark-safety-path-escape-openrouter-gpt5mini-2026-07-02-r1/summary.md`, with stderr reporting that `OPENROUTER_API_KEY` was required.

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
