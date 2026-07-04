# CEO Harness Launch Proof

Date: 2026-07-03

## Go / No-Go

No-go for a public "10/10 beats competitors" launch claim.

Go for a narrower claim: CEO Harness completed the controlled production-core coding harness suite cleanly, including the expanded 25-task all-agent comparison and newer 29-task CEO-only suite.

The current full-benchmark proof gap is closed for the eval scorer: all 31 benchmark tasks now have saved deterministic fixture reports, score JSON, per-task logs, and a summary with 31 pass / 0 partial / 0 fail / 0 skipped.

The live competitor proof gap is now materially stronger. CEO Harness, Codex CLI, OpenCode, and Pi completed a full 25-task production-core head-to-head run with 100 live runs. CEO Harness, Codex CLI, and OpenCode each passed 25/25. Pi passed 24/25 with 1 timeout. This supports controlled-suite parity with Codex CLI and OpenCode, not a broad market-win claim.

## Evidence Root

- Task evidence: `.omo/evidence/task-14-ceo-harness-10-out-of-10/`
- Dogfood evidence: `.omo/evidence/dogfood-real/index.md`
- Four-task external-agent evidence: `.omo/evidence/external-agent-4task-r2/summary.json`
- Previous 24-task external-agent evidence: `.omo/evidence/external-agent-production-core-r5/summary.json`
- Expanded 25-task external-agent evidence: `.omo/evidence/external-agent-production-core-25-r1/summary.json`
- Current 29-task external-agent evidence: `.omo/evidence/external-agent-production-core-29-r1/summary.json`
- Current 29-task external-agent rerun evidence: `.omo/evidence/external-agent-production-core-29-r2/summary.json`
- Focused path-safety rubric repair evidence: `.omo/evidence/external-agent-path-escape-rubric-r1/summary.json`
- Focused timeout-retry evidence: `.omo/evidence/external-agent-timeout-retry-r1/summary.json`
- Market-parity-core CEO evidence: `.omo/evidence/market-parity-core-ceo-r2/summary.json`
- Expanded 25-task CEO evidence: `.omo/evidence/production-core-25-ceo-r1/summary.json`
- Expanded 26-task CEO evidence: `.omo/evidence/production-core-26-ceo-r1/summary.json`
- Expanded 29-task CEO evidence: `.omo/evidence/production-core-29-ceo-r1/summary.json`
- Full 31-task fixture evidence: `.omo/evidence/benchmark-fixtures-31-r1/summary.json`
- Concurrent 25-task CEO evidence: `.omo/evidence/production-core-25-ceo-concurrency-r1/summary.json`
- Cross-language CEO evidence: `.omo/evidence/cross-language-core-ceo-r1/summary.json`
- Repeated real-repo dogfood evidence: `.omo/evidence/dogfood-real-repeat-self-r1/index.md`
- Copied-workspace dogfood evidence: `.omo/evidence/dogfood-real-copy-self-r1/index.md`
- Independent copied-workspace dogfood evidence: `.omo/evidence/dogfood-real-independent-r1/index.md`
- Expanded independent copied-workspace dogfood evidence: `.omo/evidence/dogfood-real-independent-r2/index.md`
- Copied-workspace write-probe dogfood evidence: `.omo/evidence/dogfood-real-write-probe-r1/index.md`
- Copied-workspace feature-edit dogfood evidence: `.omo/evidence/dogfood-real-feature-edit-r2/index.md`
- Copied-workspace app-code dogfood evidence: `.omo/evidence/dogfood-real-app-code-r1/index.md`
- Copied-workspace integrated app-code dogfood evidence: `.omo/evidence/dogfood-real-integrated-app-code-r1/index.md`
- Copied-workspace multi-file app-code dogfood evidence: `.omo/evidence/dogfood-real-multi-file-app-code-r1/index.md`
- Broadened Janus multi-file app-code dogfood evidence: `.omo/evidence/dogfood-real-multi-file-janus-r1/index.md`
- Broader app-shaped copied-workspace dogfood evidence: `.omo/evidence/dogfood-real-broader-apps-r1/index.md`
- Repeated real Kimi provider evidence: `.omo/evidence/provider-kimi-path-safety-repeat-r7/summary.json`
- Real Kimi JS app-code evidence: `.omo/evidence/provider-kimi-js-state-reducer-r2/summary.json`
- First-class Kimi provider proof evidence: `.omo/evidence/provider-proof-kimi-r2/index.md`
- First-class Codex provider proof evidence: `.omo/evidence/provider-proof-codex-r1/index.md`
- HTTP provider blocked-key evidence: `.omo/evidence/provider-proof-openrouter-blocked-r1/index.md`
- Public release readiness evidence: `.omo/evidence/release-readiness-r1/index.md`
- Production-readiness aggregate evidence: `.omo/evidence/production-readiness-r1/index.md`
- Endurance eval evidence: `.omo/evidence/endurance-local-r1/index.md`
- Longer endurance eval evidence: `.omo/evidence/endurance-local-r2/index.md`
- Extended endurance eval evidence: `.omo/evidence/endurance-local-r3/index.md`
- Nightly eval workflow: `.github/workflows/nightly-evals.yml`
- Focused multi-file CEO evidence: `.omo/evidence/multi-file-provider-fallback-ceo-r2/summary.json`, `.omo/evidence/multi-file-operator-safety-flow-ceo-r1/summary.json`
- User report: `outputs/ceo-harness-10-out-of-10-report.md`

## Required Gates

| Gate | Invocation | Observable | Artifact |
| --- | --- | --- | --- |
| Repo CI | `make ci` | Exit 0; tests, vet, smoke, dogfood, build passed | `.omo/evidence/task-14-ceo-harness-10-out-of-10/make-ci-task14-fix.log` |
| Task 14 code review / slop | Programming Go + remove-ai-slops review over gate blocker scope | Oversized eval file split, dead test helper removed, no test slop found in scope | `.omo/evidence/task-14-ceo-harness-10-out-of-10-code-review-slop.md` |
| Go LOC ceiling | `awk` pure-LOC scan over `internal/eval/*.go` | Every Go file is <=250 pure LOC; `benchmark.go` is 124 | `.omo/evidence/task-14-ceo-harness-10-out-of-10/pure-loc-task14-fix.log` |
| Race/shuffle | `go test -race -shuffle=on -count=1 ./...` | Exit 0 on fresh 2026-07-02 rerun after adapter timeout-proof stabilization | `.omo/evidence/final-adapter-version-timeout-fix/go-test-race-shuffle-all.log` |
| Release + manifest verification | `VERSION=0.1.0-dev sh scripts/release-local.sh && sh scripts/verify-release.sh dist` | All three archives verified against `checksums.txt` and `release-manifest.json` | `.omo/evidence/task-14-ceo-harness-10-out-of-10/release-local-checksums.log`, `dist/checksums.txt`, `dist/release-manifest.json` |
| Public release readiness packet | `sh scripts/release-readiness.sh --dist dist --output-dir .omo/evidence/release-readiness-r1` | Local artifacts verified and GitHub auth passed; public release remains blocked by missing origin remote, release URL, remote Homebrew URL, and signatures/checksum-only notes | `.omo/evidence/release-readiness-r1/index.md`, `.omo/evidence/release-readiness-r1/summary.json` |
| Eval catalog | `go run ./cmd/ceo-eval --list` | 31 task IDs listed | `.omo/evidence/benchmark-fixtures-31-r1/summary.json` |
| Eval rubric | `go run ./cmd/ceo-eval --rubric` | `rubric_valid=true` | `.omo/evidence/task-14-ceo-harness-10-out-of-10/eval-rubric.txt` |
| Benchmark scoring pass | `go run ./cmd/ceo-eval --task bugfix-cli-timeout --report internal/eval/testdata/dirty-worktree/happy/report.json --workspace <temp>` | Verdict `pass`, 8/8 checks | `.omo/evidence/task-14-ceo-harness-10-out-of-10/eval-score-dirty-happy.json` |
| Full benchmark fixture suite | `go run ./cmd/ceo-eval --benchmark-fixtures --tasks evals/tasks --output-dir .omo/evidence/benchmark-fixtures-31-r1` | 31 tasks scored; 31 pass, 0 partial, 0 fail, 0 skipped | `.omo/evidence/benchmark-fixtures-31-r1/summary.json`, per-task `report.json`, `score.json`, `score.log` |
| Failure injection | `go run ./cmd/ceo-eval --task bugfix-cli-timeout --report internal/eval/testdata/corrupt/report.json --workspace .` | Expected non-zero; corrupt JSON rejected | `.omo/evidence/task-14-ceo-harness-10-out-of-10/eval-score-forced-failure.stderr` |
| Competitor comparison | `go run ./cmd/ceo-eval --validate-competitors --competitors evals/competitors.json`, `go run ./cmd/ceo-eval --comparison-plan --competitors evals/competitors.json`, and `go run ./cmd/ceo-eval --comparison-smoke --competitors evals/competitors.json --output-dir .omo/evidence/task-14-ceo-harness-10-out-of-10/competitor-smoke --timeout-seconds 15` | Config valid; plan exists; local smoke ran installed binaries only: 2 pass, 3 skipped missing binary | `.omo/evidence/task-14-ceo-harness-10-out-of-10/competitor-smoke/summary.json` |
| Four-task external-agent comparison | `go run ./cmd/ceo-packet gauntlet --suite docs-roadmap-cli-first,bugfix-cli-timeout,safety-policy-path-escape,recovery-resume-retry --agents ceo_harness,codex_cli,opencode,pi ...` | 16 runs; 16 pass, 0 partial, 0 fail, 0 timed out, 0 skipped, 0 incomplete evidence | `.omo/evidence/external-agent-4task-r2/summary.json` |
| Market-parity-core CEO comparison | `go run ./cmd/ceo-eval --local-agent-benchmark --local-agents ceo_harness --local-agent-benchmark-task market-parity-core --timeout-seconds 180 ...` | 10 runs; CEO Harness 10/10 pass; 0 partial; 0 timeout; 0 incomplete evidence | `.omo/evidence/market-parity-core-ceo-r2/summary.json` |
| Full expanded production-core external-agent comparison | `go run ./cmd/ceo-eval --local-agent-benchmark --local-agents ceo_harness,codex_cli,opencode,pi --local-agent-benchmark-task production-core --timeout-seconds 240 ...` | 100 runs; CEO Harness 25/25 pass; Codex CLI 25/25 pass; OpenCode 25/25 pass; Pi 24 pass and 1 timeout | `.omo/evidence/external-agent-production-core-25-r1/summary.json` |
| Current 29-task external-agent comparison | `go run ./cmd/ceo-eval --local-agent-benchmark --local-agents ceo_harness,codex_cli,opencode,pi --local-agent-benchmark-task production-core --local-agent-benchmark-concurrency 4 --timeout-seconds 240 ...` | 116 runs; first run 115 pass / 1 partial; focused rubric repair 4/4 pass; rerun 110 pass / 6 external-tool timeouts. CEO Harness stayed clean. | `.omo/evidence/external-agent-production-core-29-r1/summary.json`, `.omo/evidence/external-agent-path-escape-rubric-r1/summary.json`, `.omo/evidence/external-agent-production-core-29-r2/summary.json` |
| Focused timeout-retry comparison | `go run ./cmd/ceo-eval --local-agent-benchmark --local-agent-benchmark-timeout-retries 1 --local-agent-benchmark-task <5 timeout-heavy tasks> ...` | 20 runs; CEO Harness 5/5 pass, Codex CLI 5/5 pass, Pi 5/5 pass, OpenCode 5 exhausted timeouts with prior attempts preserved | `.omo/evidence/external-agent-timeout-retry-r1/summary.json` |
| Expanded production-core CEO comparison | `go run ./cmd/ceo-eval --local-agent-benchmark --local-agents ceo_harness --local-agent-benchmark-task production-core --timeout-seconds 120 ...` | 25 runs; CEO Harness 25/25 pass; 0 partial; 0 timeout; 0 incomplete evidence | `.omo/evidence/production-core-25-ceo-r1/summary.json` |
| Expanded 26-task production-core CEO comparison | `go run ./cmd/ceo-eval --local-agent-benchmark --local-agents ceo_harness --local-agent-benchmark-task production-core --timeout-seconds 180 ...` | 26 runs; CEO Harness 26/26 pass; 0 partial; 0 timeout; 0 incomplete evidence | `.omo/evidence/production-core-26-ceo-r1/summary.json` |
| Expanded 29-task production-core CEO comparison | `go run ./cmd/ceo-packet gauntlet --suite production-core --agents ceo_harness --concurrency 4 --timeout-seconds 180 ...` | 29 runs; CEO Harness 29/29 pass; 0 partial; 0 timeout; 0 incomplete evidence | `.omo/evidence/production-core-29-ceo-r1/summary.json` |
| Concurrent production-core CEO comparison | `go run ./cmd/ceo-packet gauntlet --suite production-core --agents ceo_harness --concurrency 4 --timeout-seconds 120 ...` | 25 runs; concurrency 4; CEO Harness 25/25 pass; planned result order preserved | `.omo/evidence/production-core-25-ceo-concurrency-r1/summary.json` |
| Cross-language CEO comparison | `go run ./cmd/ceo-packet gauntlet --suite cross-language-core --agents ceo_harness --concurrency 2 --timeout-seconds 120 ...` | 2 runs across JavaScript and Python fixtures; CEO Harness 2/2 pass; 0 incomplete evidence | `.omo/evidence/cross-language-core-ceo-r1/summary.json` |
| Repeated real Kimi provider path-safety proof | `go run ./cmd/ceo-eval --local-agent-benchmark --local-agents ceo_harness --local-agent-benchmark-task safety-policy-path-escape --local-agent-benchmark-repeat 3 --timeout-seconds 600 ... scripts/kimi-model-command.sh` | 3 runs; CEO Harness 3/3 pass; 18/18 scored checks; 0 partial; 0 fail; 0 incomplete evidence | `.omo/evidence/provider-kimi-path-safety-repeat-r7/summary.json` |
| Real Kimi JS app-code proof | `go run ./cmd/ceo-eval --local-agent-benchmark --local-agents ceo_harness --local-agent-benchmark-task cross-language-js-state-reducer --timeout-seconds 600 ... scripts/kimi-model-command.sh` | 1 run; CEO Harness 1/1 pass; 6/6 scored checks; Kimi changed `frontend/state.js`, created required evidence, and passed `node frontend/state.test.js` | `.omo/evidence/provider-kimi-js-state-reducer-r2/summary.json` |
| First-class Kimi provider proof gate | `sh scripts/provider-proof.sh --provider kimi --output-dir .omo/evidence/provider-proof-kimi-r2` | JS reducer passed 6/6; Python retry policy passed 7/7; both source edits and evidence artifacts were produced through Kimi-backed CEO Harness | `.omo/evidence/provider-proof-kimi-r2/index.md` |
| First-class Codex provider proof gate | `sh scripts/provider-proof.sh --provider codex --output-dir .omo/evidence/provider-proof-codex-r1` | JS reducer passed 6/6; Python retry policy passed 7/7; both source edits and evidence artifacts were produced through Codex-backed CEO Harness | `.omo/evidence/provider-proof-codex-r1/index.md` |
| HTTP provider proof blocked-key gate | `sh scripts/provider-proof.sh --provider openrouter --output-dir .omo/evidence/provider-proof-openrouter-blocked-r1` | OpenRouter HTTP mode is wired but blocked because `OPENROUTER_API_KEY` is missing; evidence is setup-blocked, not benchmark-failed | `.omo/evidence/provider-proof-openrouter-blocked-r1/index.md` |
| Multi-file provider/config benchmark | `go run ./cmd/ceo-eval --local-agent-benchmark --local-agents ceo_harness --local-agent-benchmark-task multi-file-provider-fallback-reporting --timeout-seconds 120 ...` | Required two files across `internal/cli` and `internal/config`; 9/9 scored checks passed | `.omo/evidence/multi-file-provider-fallback-ceo-r2/summary.json` |
| Four-file operator safety benchmark | `go run ./cmd/ceo-eval --local-agent-benchmark --local-agents ceo_harness --local-agent-benchmark-task multi-file-operator-safety-flow --timeout-seconds 120 ...` | Required four files across `internal/cli`, `internal/workspace`, `internal/config`, and docs; 13/13 scored checks passed | `.omo/evidence/multi-file-operator-safety-flow-ceo-r1/summary.json` |
| Real repo dogfood | `sh scripts/dogfood-real.sh --repo temp-real:<temp-git-repo> --timeout-ms 250` | Temp external git repo row `pass`; 5 scenarios captured | `.omo/evidence/dogfood-real/index.md` |
| Repeated real-repo dogfood | `sh scripts/dogfood-real.sh --repo ceo-harness-repeat:<repo> --repeat 3 --timeout-ms 250 --output-dir .omo/evidence/dogfood-real-repeat-self-r1` | 3 live passes; 0 fails; each run captured doctor, plan-only, observe, patch-preview, and timeout-guard evidence | `.omo/evidence/dogfood-real-repeat-self-r1/index.md` |
| Copied-workspace dogfood | `sh scripts/dogfood-real.sh --copy-workspace --repo ceo-harness-copy:<repo> --timeout-ms 250 --output-dir .omo/evidence/dogfood-real-copy-self-r1` | Source path and copied workspace path recorded separately; all five scenarios passed against the copy | `.omo/evidence/dogfood-real-copy-self-r1/index.md` |
| Copied-workspace write-probe dogfood | `sh scripts/dogfood-real.sh --copy-workspace --write-probe --repo chemcheck:<repo> --repo axis-health:<repo> --timeout-ms 250 --output-dir .omo/evidence/dogfood-real-write-probe-r1` | Both copied workspaces passed six scenarios, including preview plus approved write; source checkouts stayed without `ceo-dogfood-write-probe.txt` | `.omo/evidence/dogfood-real-write-probe-r1/index.md` |
| Copied-workspace feature-edit dogfood | `sh scripts/dogfood-real.sh --copy-workspace --feature-edit-probe --repo chemcheck:<repo> --repo axis-health:<repo> --timeout-ms 250 --output-dir .omo/evidence/dogfood-real-feature-edit-r2` | Both copied workspaces passed feature-note edits through preview plus approved write; source checkouts stayed without `ceo-dogfood-feature.md` | `.omo/evidence/dogfood-real-feature-edit-r2/index.md` |
| Copied-workspace app-code dogfood | `sh scripts/dogfood-real.sh --copy-workspace --app-code-probe --repo chemcheck:<repo> --repo axis-health:<repo> --timeout-ms 250 --output-dir .omo/evidence/dogfood-real-app-code-r1` | Both copied workspaces passed source-module edits through preview plus approved write; source checkouts stayed without `src/ceoDogfoodProbe.mjs` | `.omo/evidence/dogfood-real-app-code-r1/index.md` |
| Copied-workspace integrated app-code dogfood | `sh scripts/dogfood-real.sh --copy-workspace --integrated-app-code-probe --repo chemcheck:<repo> --repo axis-health:<repo> --timeout-ms 250 --output-dir .omo/evidence/dogfood-real-integrated-app-code-r1` | Both copied workspaces passed existing `src/App.jsx` edits through preview plus approved write; source checkouts stayed without `ceoDogfoodIntegratedProbe` | `.omo/evidence/dogfood-real-integrated-app-code-r1/index.md` |
| Copied-workspace multi-file app-code dogfood | `sh scripts/dogfood-real.sh --copy-workspace --multi-file-app-code-probe --repo chemcheck:<repo> --repo axis-health:<repo> --timeout-ms 250 --output-dir .omo/evidence/dogfood-real-multi-file-app-code-r1` | Both copied workspaces passed existing `src/App.jsx` and `src/main.jsx` edits through preview plus approved write; source checkouts stayed without `ceoDogfoodMultiFileProbe` | `.omo/evidence/dogfood-real-multi-file-app-code-r1/index.md` |
| Broadened Janus multi-file app-code dogfood | `sh scripts/dogfood-real.sh --copy-workspace --multi-file-app-code-probe --repo janus:<repo> --timeout-ms 250 --output-dir .omo/evidence/dogfood-real-multi-file-janus-r1` | Janus passed existing `src/cli/args.ts` and `src/cli/base64-payload-byte-count.ts` edits through preview plus approved write; source checkout stayed without `ceoDogfoodMultiFileProbe` | `.omo/evidence/dogfood-real-multi-file-janus-r1/index.md` |
| Broader app-shaped copied-workspace dogfood | `sh scripts/dogfood-real.sh --copy-workspace --multi-file-app-code-probe --repo family-os:<repo> --repo pools:<repo> --repo dps-frontend:<repo> --repo janus-mobile:<repo> --timeout-ms 250 --output-dir .omo/evidence/dogfood-real-broader-apps-r1` | Family OS, Pools, DPS frontend, and Janus Mobile passed two existing-source-file edits through preview plus approved write; source checkouts stayed without `ceoDogfoodMultiFileProbe` | `.omo/evidence/dogfood-real-broader-apps-r1/index.md` |
| Short endurance smoke | `sh scripts/endurance.sh --iterations 3 --output-dir .omo/evidence/endurance-local-r1` | 3 iterations; 3 pass; 0 fail; each iteration ran fixture scoring, cross-language gauntlet, and real-repo dogfood | `.omo/evidence/endurance-local-r1/index.md` |
| Longer local endurance run | `sh scripts/endurance.sh --iterations 10 --output-dir .omo/evidence/endurance-local-r2` | 10 iterations; 10 pass; 0 fail; elapsed 30 seconds; each iteration ran build, 28-task fixture scoring, cross-language gauntlet, and real-repo dogfood | `.omo/evidence/endurance-local-r2/index.md` |
| Extended local endurance run | `sh scripts/endurance.sh --iterations 30 --output-dir .omo/evidence/endurance-local-r3` | 30 iterations; 30 pass; 0 fail; elapsed 102 seconds | `.omo/evidence/endurance-local-r3/index.md` |
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

The latest repeated dogfood runner used this real repo in live mode and captured five scenarios across three attempts:

- `scenario-01-doctor`: pass
- `scenario-02-plan-only`: pass
- `scenario-03-observe-run`: pass
- `scenario-04-patch-preview`: pass
- `scenario-05-timeout-guard`: `pass_expected_failure`

Artifact: `.omo/evidence/dogfood-real-repeat-self-r1/repos/ceo-harness-repeat/summary.md`

The latest independent copied-workspace dogfood run used two additional local repos and avoided writing to the source checkouts:

- ChemCheck: pass, copied workspace, five scenarios.
- Axis Health: pass, copied workspace, five scenarios.

Artifact: `.omo/evidence/dogfood-real-independent-r1/index.md`

The expanded independent copied-workspace dogfood run added four more local repos:

- Clicky: pass, copied workspace, five scenarios.
- DPS Internal Comms: pass, copied workspace, five scenarios.
- Janus: pass, copied workspace, five scenarios.
- Radian: pass, copied workspace, five scenarios.

Artifact: `.omo/evidence/dogfood-real-independent-r2/index.md`

The latest copied-workspace write-probe dogfood run added real approved-write proof without touching the source checkouts:

- ChemCheck: pass, copied workspace, six scenarios, write probe applied only inside the copy.
- Axis Health: pass, copied workspace, six scenarios, write probe applied only inside the copy.

Artifact: `.omo/evidence/dogfood-real-write-probe-r1/index.md`

The latest copied-workspace feature-edit dogfood run added a repo-specific approved-write feature note without touching the source checkouts:

- ChemCheck: pass, copied workspace, feature note applied only inside the copy.
- Axis Health: pass, copied workspace, feature note applied only inside the copy.

Artifact: `.omo/evidence/dogfood-real-feature-edit-r2/index.md`

The latest copied-workspace app-code dogfood run added a source module through approved write without touching the source checkouts:

- ChemCheck: pass, copied workspace, `src/ceoDogfoodProbe.mjs` applied only inside the copy.
- Axis Health: pass, copied workspace, `src/ceoDogfoodProbe.mjs` applied only inside the copy.

Artifact: `.omo/evidence/dogfood-real-app-code-r1/index.md`

## Benchmark Summary

The full benchmark suite was scored with deterministic fixture reports generated from `evals/tasks/benchmark_tasks.json`. This proves the scorer can evaluate every task definition and save per-task evidence. It does not prove an autonomous agent completed all 31 tasks live.

- Mode: `deterministic_fixture_scoring`
- Tasks: 31
- Passed: 31
- Partial: 0
- Failed: 0
- Skipped: 0

Artifacts:

- `.omo/evidence/benchmark-fixtures-31-r1/summary.json`
- `.omo/evidence/benchmark-fixtures-31-r1/*/report.json`
- `.omo/evidence/benchmark-fixtures-31-r1/*/score.json`
- `.omo/evidence/benchmark-fixtures-31-r1/*/score.log`

## Competitor Summary

The comparison harness now has expanded production-core live task evidence for installed external agents, not only a plan, version smoke, or four-task subset.

- Scope: 25 tasks x 4 agents = 100 live runs.
- Agents: CEO Harness, Codex CLI, OpenCode, Pi.
- Result: 99 pass, 0 partial, 0 fail, 1 timed out, 0 skipped, 1 incomplete evidence.
- CEO Harness: 25 pass, 0 partial, 0 fail, 0 timed out, 0 incomplete evidence.
- Codex CLI: 25 pass, 0 partial, 0 fail, 0 timed out, 0 incomplete evidence.
- OpenCode: 25 pass, 0 partial, 0 fail, 0 timed out, 0 incomplete evidence.
- Pi: 24 pass, 0 partial, 1 timed out, 1 incomplete evidence.

Artifact: `.omo/evidence/external-agent-production-core-25-r1/summary.json`

Expanded suite follow-up:

- Scope: 25 tasks x CEO Harness = 25 live runs.
- Added task: `multi-file-provider-fallback-reporting`, requiring edits in both `internal/cli/provider_fallback_report.go` and `internal/config/provider_fallback_policy.go`.
- Result: CEO Harness 25 pass, 0 partial, 0 fail, 0 timed out, 0 incomplete evidence.
- Artifact: `.omo/evidence/production-core-25-ceo-r1/summary.json`

Larger multi-file follow-up:

- Scope: 26 tasks x CEO Harness = 26 live runs.
- Added task: `multi-file-operator-safety-flow`, requiring edits across `internal/cli`, `internal/workspace`, `internal/config`, and docs.
- Result: CEO Harness 26 pass, 0 partial, 0 fail, 0 timed out, 0 incomplete evidence.
- Artifact: `.omo/evidence/production-core-26-ceo-r1/summary.json`

Harder production-core follow-up:

- Scope: 29 tasks x CEO Harness = 29 live runs with `--concurrency 4`.
- Added tasks: `multi-file-release-readiness-publish-boundary`, `multi-file-lean-context-autonomy`, and `multi-file-secret-safe-provider-proof`.
- Result: CEO Harness 29 pass, 0 partial, 0 fail, 0 timed out, 0 incomplete evidence.
- Artifact: `.omo/evidence/production-core-29-ceo-r1/summary.json`

Focused external-agent follow-up for the newest multi-file task:

- Scope: `multi-file-operator-safety-flow` x 4 agents = 4 live runs with concurrency 4.
- Agents: CEO Harness, Codex CLI, OpenCode, Pi.
- Result: 4 pass, 0 partial, 0 fail, 0 timed out, 0 skipped, 0 incomplete evidence.
- Artifact: `.omo/evidence/external-agent-operator-safety-flow-r1/summary.json`

Concurrent runner follow-up:

- Scope: 25 tasks x CEO Harness = 25 live runs with `--concurrency 4`.
- Result: CEO Harness 25 pass, 0 partial, 0 fail, 0 timed out, 0 incomplete evidence.
- Artifact: `.omo/evidence/production-core-25-ceo-concurrency-r1/summary.json`

Cross-language follow-up:

- Scope: 2 tasks x CEO Harness = 2 live runs with `--concurrency 2`.
- Languages covered: JavaScript and Python fixture edits with native test commands.
- Result: CEO Harness 2 pass, 0 partial, 0 fail, 0 timed out, 0 incomplete evidence.
- Artifact: `.omo/evidence/cross-language-core-ceo-r1/summary.json`

## Blockers / Risks

- Public "10/10 beats competitors" claim is still unsupported because Codex CLI and OpenCode matched CEO Harness at 25/25 on the controlled suite.
- The expanded 29/29 CEO Harness live result, focused 4/4 external-agent multi-file result, 2/2 cross-language result, 3-pass real-repo dogfood result, six-repo copied-workspace dogfood result, two-repo task-specific dogfood result, 30-iteration local endurance run, and repeated 3/3 real Kimi provider proof are still not enough for a broad production-market claim. Deeper task-specific real-repo jobs with real writes, additional provider families, and overnight or truly long-duration tasks are still needed.
- Provider doctor correctly fails without `OPENAI_API_KEY`; this is setup guidance, not a product pass against a real provider.
- Rollback QA now covers normal replacements, trailing-newline replacements, and created-file model patches. Arbitrary hand-edited diff rollback is still not claimed.

## Cleanup

All temp QA workspaces recorded by task 14 were removed.

Artifact: `.omo/evidence/task-14-ceo-harness-10-out-of-10/cleanup.log`
