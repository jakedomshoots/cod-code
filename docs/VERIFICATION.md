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
  - `sh scripts/strict-checks.sh`
- Latest live external-agent comparison:
  - `go run ./cmd/ceo-eval --local-agent-benchmark --local-agents ceo_harness,codex_cli,opencode,pi --local-agent-benchmark-task production-core --local-agent-benchmark-repeat 1 --tasks evals/tasks --output-dir .omo/evidence/external-agent-production-core-25-r1 --timeout-seconds 240 --ceo-benchmark-mode model-command --ceo-benchmark-model-command-json '["sh","/Users/jakedom/Documents/Codex/2026-06-30/new-chat/work/ceo-harness/scripts/benchmark-model-command.sh"]'`
  - Result: 100 runs / 99 pass / 0 partial / 0 fail / 1 timed out / 0 skipped / 1 incomplete evidence.
  - CEO Harness result: 25 pass / 0 partial / 0 fail / 0 timed out / 0 incomplete evidence.
  - Agent totals: Codex CLI 25 pass; OpenCode 25 pass; Pi 24 pass and 1 timed out.
- Latest expanded production-core CEO comparison:
  - `go run ./cmd/ceo-eval --local-agent-benchmark --local-agents ceo_harness --local-agent-benchmark-task production-core --local-agent-benchmark-repeat 1 --tasks evals/tasks --output-dir .omo/evidence/production-core-25-ceo-r1 --timeout-seconds 120 --ceo-benchmark-mode model-command --ceo-benchmark-model-command-json '["sh","/Users/jakedom/Documents/Codex/2026-06-30/new-chat/work/ceo-harness/scripts/benchmark-model-command.sh"]'`
  - Result: 25 tasks / 25 pass / 0 partial / 0 fail / 0 timed out / 0 incomplete evidence.
- Latest concurrent production-core CEO comparison:
  - `go run ./cmd/ceo-packet gauntlet --suite production-core --agents ceo_harness --ceo-binary ./bin/ceo-packet --tasks evals/tasks --output-dir .omo/evidence/production-core-25-ceo-concurrency-r1 --timeout-seconds 120 --concurrency 4 --ceo-benchmark-mode model-command --ceo-benchmark-model-command-json '["sh","/Users/jakedom/Documents/Codex/2026-06-30/new-chat/work/ceo-harness/scripts/benchmark-model-command.sh"]'`
  - Result: 25 tasks / concurrency 4 / 25 pass / 0 partial / 0 fail / 0 timed out / 0 incomplete evidence.
- Latest focused multi-file task proof:
  - `go run ./cmd/ceo-eval --local-agent-benchmark --local-agents ceo_harness --local-agent-benchmark-task multi-file-provider-fallback-reporting --local-agent-benchmark-repeat 1 --tasks evals/tasks --output-dir .omo/evidence/multi-file-provider-fallback-ceo-r2 --timeout-seconds 120 --ceo-benchmark-mode model-command --ceo-benchmark-model-command-json '["sh","/Users/jakedom/Documents/Codex/2026-06-30/new-chat/work/ceo-harness/scripts/benchmark-model-command.sh"]'`
  - Result: 1 run / 1 pass / 9 scored checks / 0 incomplete evidence.
- Latest real-repo dogfood:
  - `sh scripts/dogfood-real.sh --repo ceo-harness:<repo> --timeout-ms 250`
  - Result: pass, including expected timeout failure evidence.

- Focused additions test:
  - `go test ./internal/cli -run 'Test_Run_(start|inbox|provider_wizard|init_demo_repo|tui|write_policy|init_config_uses_external_adapter|prints_help)' -count=1`
- `make ci`
  - `gofmt -w ./cmd ./internal`
  - `go test ./... -count=1`
  - `go vet ./...`
  - `sh scripts/smoke.sh`
  - `sh scripts/dogfood.sh`
  - `go build ... ./cmd/ceo-packet`
- `go test -race -shuffle=on -count=1 ./...`
- `VERSION=0.1.0-dev sh scripts/release-local.sh`
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

## Tooling Available Locally

These optional strict tools are installed under the local Go bin and passed during the latest gate:

- `gofumpt`
- `golangci-lint`
- `nilaway`
- `task`

## Tooling Not Available Locally

- `shellcheck`

ShellCheck is still optional for a source install. Missing optional tools should not block `scripts/install-local.sh`.
