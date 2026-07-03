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
  - `ceo-packet gauntlet --suite docs-roadmap-cli-first,bugfix-cli-timeout,safety-policy-path-escape,recovery-resume-retry --agents ceo_harness,codex_cli,opencode,pi`
  - Result: 16 runs / 16 pass / 0 partial / 0 fail / 0 timed out / 0 skipped / 0 incomplete evidence.
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
