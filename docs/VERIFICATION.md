# Verification Record

Status date: 2026-07-02

## Passed

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

## Tooling Not Available Locally

These are documented/configured, but were not installed on PATH during this pass:

- `gofumpt`
- `golangci-lint`
- `nilaway`
- `shellcheck`
- `task`

The verified gate therefore used available local tooling: `gofmt`, `go test`, `go vet`, smoke, dogfood, release build, checksum verification, install QA, and race/shuffle tests.

These tools are optional for a source install. Missing optional tools should not block `scripts/install-local.sh`.
