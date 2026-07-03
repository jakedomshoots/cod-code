# Verification Record

Status date: 2026-07-03

## Passed

- Latest production gate:
  - `go test ./... -count=1`
  - `go vet ./...`
  - `sh scripts/smoke.sh`
  - `sh scripts/dogfood.sh`
  - `sh scripts/release-local.sh`
  - `task ci`
- Latest strict gate:
  - `gofumpt -l cmd internal`
  - `golangci-lint run ./...`
  - `nilaway ./...`
  - `sh -n` over shell scripts when ShellCheck is unavailable.
  - `sh scripts/strict-checks.sh`
- Latest live external-agent comparison:
  - `go run ./cmd/ceo-eval --local-agent-benchmark --local-agents ceo_harness,codex_cli,opencode,pi --local-agent-benchmark-task production-core --local-agent-benchmark-repeat 1 --tasks evals/tasks --output-dir .omo/evidence/external-agent-production-core-25-r1 --timeout-seconds 240 --ceo-benchmark-mode model-command --ceo-benchmark-model-command-json '["sh","/Users/jakedom/Documents/Codex/2026-06-30/new-chat/work/ceo-harness/scripts/benchmark-model-command.sh"]'`
  - Result: 100 runs / 99 pass / 0 partial / 0 fail / 1 timed out / 0 skipped / 1 incomplete evidence.
  - CEO Harness result: 25 pass / 0 partial / 0 fail / 0 timed out / 0 incomplete evidence.
  - Agent totals: Codex CLI 25 pass; OpenCode 25 pass; Pi 24 pass and 1 timed out.
- Latest focused external-agent comparison for the newest multi-file task:
  - `go run ./cmd/ceo-eval --local-agent-benchmark --local-agents ceo_harness,codex_cli,opencode,pi --local-agent-benchmark-task multi-file-operator-safety-flow --local-agent-benchmark-repeat 1 --local-agent-benchmark-concurrency 4 --ceo-binary ./bin/ceo-packet --tasks evals/tasks --output-dir .omo/evidence/external-agent-operator-safety-flow-r1 --timeout-seconds 240 --ceo-benchmark-mode model-command --ceo-benchmark-model-command-json '["sh","/Users/jakedom/Documents/Codex/2026-06-30/new-chat/work/ceo-harness/scripts/benchmark-model-command.sh"]'`
  - Result: 4 runs / 4 pass / 0 partial / 0 fail / 0 timed out / 0 skipped / 0 incomplete evidence.
- Latest market-parity-core CEO comparison:
  - `go run ./cmd/ceo-eval --local-agent-benchmark --local-agents ceo_harness --local-agent-benchmark-task market-parity-core --local-agent-benchmark-repeat 1 --ceo-binary ./bin/ceo-packet --tasks evals/tasks --output-dir .omo/evidence/market-parity-core-ceo-r2 --timeout-seconds 180 --ceo-benchmark-mode model-command --ceo-benchmark-model-command-json '["sh","/Users/jakedom/Documents/Codex/2026-06-30/new-chat/work/ceo-harness/scripts/benchmark-model-command.sh"]'`
  - Result: 10 tasks / 10 pass / 0 partial / 0 fail / 0 timed out / 0 skipped / 0 incomplete evidence.
- Latest expanded production-core CEO comparison:
  - `go run ./cmd/ceo-eval --local-agent-benchmark --local-agents ceo_harness --local-agent-benchmark-task production-core --local-agent-benchmark-repeat 1 --tasks evals/tasks --output-dir .omo/evidence/production-core-26-ceo-r1 --timeout-seconds 180 --ceo-benchmark-mode model-command --ceo-benchmark-model-command-json '["sh","/Users/jakedom/Documents/Codex/2026-06-30/new-chat/work/ceo-harness/scripts/benchmark-model-command.sh"]'`
  - Result: 26 tasks / 26 pass / 0 partial / 0 fail / 0 timed out / 0 incomplete evidence.
- Latest concurrent production-core CEO comparison:
  - `go run ./cmd/ceo-packet gauntlet --suite production-core --agents ceo_harness --ceo-binary ./bin/ceo-packet --tasks evals/tasks --output-dir .omo/evidence/production-core-25-ceo-concurrency-r1 --timeout-seconds 120 --concurrency 4 --ceo-benchmark-mode model-command --ceo-benchmark-model-command-json '["sh","/Users/jakedom/Documents/Codex/2026-06-30/new-chat/work/ceo-harness/scripts/benchmark-model-command.sh"]'`
  - Result: 25 tasks / concurrency 4 / 25 pass / 0 partial / 0 fail / 0 timed out / 0 incomplete evidence.
- Latest cross-language CEO comparison:
  - `go run ./cmd/ceo-packet gauntlet --suite cross-language-core --agents ceo_harness --ceo-binary ./bin/ceo-packet --tasks evals/tasks --output-dir .omo/evidence/cross-language-core-ceo-r1 --timeout-seconds 120 --concurrency 2 --ceo-benchmark-mode model-command --ceo-benchmark-model-command-json '["sh","/Users/jakedom/Documents/Codex/2026-06-30/new-chat/work/ceo-harness/scripts/benchmark-model-command.sh"]'`
  - Result: 2 tasks / concurrency 2 / 2 pass / 0 partial / 0 fail / 0 timed out / 0 incomplete evidence.
- Latest focused multi-file task proof:
  - `go run ./cmd/ceo-eval --local-agent-benchmark --local-agents ceo_harness --local-agent-benchmark-task multi-file-provider-fallback-reporting --local-agent-benchmark-repeat 1 --tasks evals/tasks --output-dir .omo/evidence/multi-file-provider-fallback-ceo-r2 --timeout-seconds 120 --ceo-benchmark-mode model-command --ceo-benchmark-model-command-json '["sh","/Users/jakedom/Documents/Codex/2026-06-30/new-chat/work/ceo-harness/scripts/benchmark-model-command.sh"]'`
  - Result: 1 run / 1 pass / 9 scored checks / 0 incomplete evidence.
- Latest larger multi-file task proof:
  - `go run ./cmd/ceo-eval --local-agent-benchmark --local-agents ceo_harness --local-agent-benchmark-task multi-file-operator-safety-flow --local-agent-benchmark-repeat 1 --tasks evals/tasks --output-dir .omo/evidence/multi-file-operator-safety-flow-ceo-r1 --timeout-seconds 120 --ceo-benchmark-mode model-command --ceo-benchmark-model-command-json '["sh","/Users/jakedom/Documents/Codex/2026-06-30/new-chat/work/ceo-harness/scripts/benchmark-model-command.sh"]'`
  - Result: 1 run / 1 pass / 13 scored checks / 0 incomplete evidence.
- Latest full benchmark fixture scoring:
  - `go run ./cmd/ceo-eval --benchmark-fixtures --tasks evals/tasks --output-dir .omo/evidence/benchmark-fixtures-28-r1`
  - Result: 28 tasks / 28 pass / 0 partial / 0 fail / 0 skipped.
- Latest repeated real Kimi provider proof:
  - `go run ./cmd/ceo-eval --local-agent-benchmark --local-agents ceo_harness --local-agent-benchmark-task safety-policy-path-escape --local-agent-benchmark-repeat 3 --ceo-binary ./bin/ceo-packet --tasks evals/tasks --output-dir .omo/evidence/provider-kimi-path-safety-repeat-r7 --timeout-seconds 600 --ceo-benchmark-mode model-command --ceo-benchmark-model-command-json '["sh","/Users/jakedom/Documents/Codex/2026-06-30/new-chat/work/ceo-harness/scripts/kimi-model-command.sh"]'`
  - Result: 3 runs / 3 pass / 18 scored checks / 0 partial / 0 fail / 0 timed out / 0 incomplete evidence.
- Latest real Kimi provider JS app-code proof:
  - `go run ./cmd/ceo-eval --local-agent-benchmark --local-agents ceo_harness --local-agent-benchmark-task cross-language-js-state-reducer --local-agent-benchmark-repeat 1 --ceo-binary ./bin/ceo-packet --tasks evals/tasks --output-dir .omo/evidence/provider-kimi-js-state-reducer-r2 --timeout-seconds 600 --ceo-benchmark-mode model-command --ceo-benchmark-model-command-json '["sh","/Users/jakedom/Documents/Codex/2026-06-30/new-chat/work/ceo-harness/scripts/kimi-model-command.sh"]'`
  - Result: 1 run / 1 pass / 6 scored checks / 0 incomplete evidence. Kimi changed `frontend/state.js`, created `.omo/evidence/cross-language-js-state-reducer.md`, and passed `node frontend/state.test.js`.
- Latest real-repo dogfood:
  - `sh scripts/dogfood-real.sh --repo ceo-harness-repeat:/Users/jakedom/Documents/Codex/2026-06-30/new-chat/work/ceo-harness --repeat 3 --timeout-ms 250 --output-dir .omo/evidence/dogfood-real-repeat-self-r1`
  - Result: 3 live passes / 0 fails, including expected timeout failure evidence in each run.
- Latest copied-workspace dogfood:
  - `sh scripts/dogfood-real.sh --copy-workspace --repo ceo-harness-copy:/Users/jakedom/Documents/Codex/2026-06-30/new-chat/work/ceo-harness --timeout-ms 250 --output-dir .omo/evidence/dogfood-real-copy-self-r1`
  - Result: pass; all five scenarios ran against `workspace-mode=copied`, with source and workspace paths recorded separately.
- Latest independent copied-workspace dogfood:
  - `sh scripts/dogfood-real.sh --copy-workspace --repo chemcheck:/Users/jakedom/Documents/chemcheck-main --repo axis-health:'/Users/jakedom/Documents/Axis health' --timeout-ms 250 --output-dir .omo/evidence/dogfood-real-independent-r1`
  - Result: ChemCheck pass and Axis Health pass; each ran doctor, plan-only, observe, patch-preview, and timeout-guard scenarios in copied workspaces.
- Latest expanded independent copied-workspace dogfood:
  - `sh scripts/dogfood-real.sh --copy-workspace --repo clicky:/Users/jakedom/Documents/clicky-main --repo dps:/Users/jakedom/Documents/DPS-internal-coms-main --repo janus:/Users/jakedom/Documents/janus-code --repo radian:'/Users/jakedom/Documents/Radian notes app ' --timeout-ms 250 --output-dir .omo/evidence/dogfood-real-independent-r2`
  - Result: Clicky pass, DPS Internal Comms pass, Janus pass, and Radian pass; each ran doctor, plan-only, observe, patch-preview, and timeout-guard scenarios in copied workspaces.
- Latest task-specific copied-workspace dogfood:
  - `sh scripts/dogfood-real.sh --copy-workspace --repo chemcheck:/Users/jakedom/Documents/chemcheck-main --repo axis-health:'/Users/jakedom/Documents/Axis health' --task 'Plan a repo-specific onboarding docs cleanup and inspect the safest first patch without writing source files' --timeout-ms 250 --output-dir .omo/evidence/dogfood-real-task-specific-r1`
  - Result: ChemCheck pass and Axis Health pass; both saved the custom task text, copied workspace path, git status evidence, plan-only output, observe-mode output, patch-preview digest, and expected timeout failure evidence.
- Latest copied-workspace write-probe dogfood:
  - `sh scripts/dogfood-real.sh --copy-workspace --write-probe --repo chemcheck:/Users/jakedom/Documents/chemcheck-main --repo axis-health:'/Users/jakedom/Documents/Axis health' --task 'Apply and prove a copied-workspace write probe without touching source checkouts' --timeout-ms 250 --output-dir .omo/evidence/dogfood-real-write-probe-r1`
  - Result: ChemCheck pass and Axis Health pass; both previewed, approved, and applied `ceo-dogfood-write-probe.txt` inside copied workspaces, saved after-state git status, and left the source checkouts without the marker file.
- Latest copied-workspace feature-edit dogfood:
  - `sh scripts/dogfood-real.sh --copy-workspace --feature-edit-probe --repo chemcheck:/Users/jakedom/Documents/chemcheck-main --repo axis-health:'/Users/jakedom/Documents/Axis health' --task 'Add a copied-workspace onboarding note that proves approved feature edits stay isolated' --timeout-ms 250 --output-dir .omo/evidence/dogfood-real-feature-edit-r2`
  - Result: ChemCheck pass and Axis Health pass; both previewed, approved, and applied a repo-specific `ceo-dogfood-feature.md` note inside copied workspaces, saved the final feature file plus after-state git status, and left source checkouts without the marker file.
- Latest copied-workspace app-code dogfood:
  - `sh scripts/dogfood-real.sh --copy-workspace --app-code-probe --repo chemcheck:/Users/jakedom/Documents/chemcheck-main --repo axis-health:'/Users/jakedom/Documents/Axis health' --task 'Add a copied-workspace source module proving approved app-code edits stay isolated' --timeout-ms 250 --output-dir .omo/evidence/dogfood-real-app-code-r1`
  - Result: ChemCheck pass and Axis Health pass; both previewed, approved, and applied `src/ceoDogfoodProbe.mjs` inside copied workspaces, saved the final source file plus after-state git status, and left source checkouts without the marker file.
- Latest copied-workspace integrated app-code dogfood:
  - `sh scripts/dogfood-real.sh --copy-workspace --integrated-app-code-probe --repo chemcheck:/Users/jakedom/Documents/chemcheck-main --repo axis-health:'/Users/jakedom/Documents/Axis health' --task 'Wire a copied-workspace app-code marker into an existing source file without touching source checkouts' --timeout-ms 250 --output-dir .omo/evidence/dogfood-real-integrated-app-code-r1`
  - Result: ChemCheck pass and Axis Health pass; both previewed, approved, and applied `ceoDogfoodIntegratedProbe` to copied `src/App.jsx`, saved the target path plus modified source file and after-state git status, and left source checkouts without the marker.
- Latest copied-workspace multi-file app-code dogfood:
  - `sh scripts/dogfood-real.sh --copy-workspace --multi-file-app-code-probe --repo chemcheck:/Users/jakedom/Documents/chemcheck-main --repo axis-health:'/Users/jakedom/Documents/Axis health' --task 'Wire a copied-workspace app-code marker across two existing source files without touching source checkouts' --timeout-ms 250 --output-dir .omo/evidence/dogfood-real-multi-file-app-code-r1`
  - Result: ChemCheck pass and Axis Health pass; both previewed, approved, and applied `ceoDogfoodMultiFileProbe*` markers to copied `src/App.jsx` and `src/main.jsx`, saved target paths plus modified source files and after-state git status, and left source checkouts without the marker.
- Broadened copied-workspace multi-file app-code dogfood:
  - `sh scripts/dogfood-real.sh --copy-workspace --multi-file-app-code-probe --repo janus:/Users/jakedom/Documents/janus-code --task 'Wire a copied-workspace app-code marker across two existing source files without touching the Janus source checkout' --timeout-ms 250 --output-dir .omo/evidence/dogfood-real-multi-file-janus-r1`
  - Result: Janus pass; previewed, approved, and applied `ceoDogfoodMultiFileProbe*` markers to copied `src/cli/args.ts` and `src/cli/base64-payload-byte-count.ts`, saved target paths plus modified source files and after-state git status, and left the Janus source checkout without the marker.
- Nightly eval task:
  - `make eval-nightly`
  - `/Users/jakedom/go/bin/task eval:nightly`
  - Result: both passed locally; each ran 27/27 deterministic fixture scoring, 2/2 cross-language CEO gauntlet, and 2-pass real-repo dogfood under `.omo/evidence/nightly/`.
- Endurance eval task:
  - `sh scripts/endurance.sh --iterations 3 --output-dir .omo/evidence/endurance-local-r1`
  - Result: 3 iterations / 3 pass / 0 fail / elapsed 8 seconds, with per-iteration command logs and summary rows.
- Longer endurance eval task:
  - `sh scripts/endurance.sh --iterations 10 --output-dir .omo/evidence/endurance-local-r2`
  - Result: 10 iterations / 10 pass / 0 fail / elapsed 30 seconds, with each iteration running build, 28-task fixture scoring, cross-language gauntlet, and real-repo dogfood.
- Extended endurance eval task:
  - `sh scripts/endurance.sh --iterations 30 --output-dir .omo/evidence/endurance-local-r3`
  - Result: 30 iterations / 30 pass / 0 fail / elapsed 102 seconds, with each iteration writing a run summary.

- Focused additions test:
  - `go test ./internal/cli -run 'Test_Run_(start|inbox|provider_wizard|init_demo_repo|tui|write_policy|init_config_uses_external_adapter|prints_help)' -count=1`
- Focused rollback test:
  - `go test ./internal/workspace -run Test_Workspace_RollbackReplaceText -count=1`
  - `go test ./internal/cli -run 'Test_Run_rollback_report|Test_HelperProcess_cli_model_create_file_patch' -count=1`
- `make ci`
  - `gofmt -w ./cmd ./internal`
  - `go test ./... -count=1`
  - `go vet ./...`
  - `sh scripts/smoke.sh`
  - `sh scripts/dogfood.sh`
  - `go build ... ./cmd/ceo-packet`
- `go test -race -shuffle=on -count=1 ./...`
- `VERSION=0.1.0-dev sh scripts/release-local.sh`
- `sh scripts/verify-release.sh dist`
- `sh scripts/release-preflight.sh dist` blocks public release claims when remote URL, public release URL, Homebrew URL, and signature or checksum-only notes are missing.
- `shasum -a 256 -c checksums.txt` from `dist/`
- Temporary install QA via `scripts/install-local.sh`
- Local markdown link check from [Trust Surface](TRUST.md)
- Shared output binary QA:
  - `ceo-packet --help`
  - `ceo-packet --demo --format text`
  - `ceo-packet --doctor` with bundled example model, CEO, and research adapters
  - `outputs/ceo-packet --version`
- New operator binary QA:
  - `bin/ceo-packet --start <temp> --format text`
  - `bin/ceo-packet --workspace <temp> --provider-wizard openai --http-model gpt-5 --format text`
  - `bin/ceo-packet --init-demo-repo <temp> --format text`
  - `bin/ceo-packet --workspace <demo> --write-policy dry-run --replace app.txt old new --format text Patch demo app`
  - `bin/ceo-packet --workspace <demo> --replace app.txt old new --format text Patch demo app`
  - `bin/ceo-packet --workspace <demo> --inbox`
  - `bin/ceo-packet --workspace <demo> --tui`
  - `bin/ceo-packet --workspace <demo> --write-policy preview --replace app.txt old new --format json Patch demo app`
  - `bin/ceo-packet --workspace <demo> --write-policy approved-write --approve-preview <preview_digest> --replace app.txt old new --format json Patch demo app`
  - `bin/ceo-packet --workspace <demo> --write-policy approved-write --replace app.txt old new --format json Patch demo app` failed as expected without `--approve-preview <preview_digest>`

## Release Artifacts Verified

- `dist/ceo-packet_0.1.0-dev_darwin_arm64.tar.gz`
- `dist/ceo-packet_0.1.0-dev_linux_amd64.tar.gz`
- `dist/ceo-packet_0.1.0-dev_linux_arm64.tar.gz`
- `dist/checksums.txt`
- `dist/release-manifest.json`

## Tooling Available Locally

These optional strict tools are installed under the local Go bin and passed during the latest gate:

- `gofumpt`
- `golangci-lint`
- `nilaway`
- `task`

## Tooling Not Available Locally

- `shellcheck`

ShellCheck is still optional for a source install. `scripts/strict-checks.sh` now runs `sh -n` over shell scripts when ShellCheck is unavailable, so shell syntax is still checked.
