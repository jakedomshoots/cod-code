BINARY := ceo-packet
PKG := ./cmd/ceo-packet
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || printf dev)
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || printf local)

.PHONY: ci fmt test test-race vet smoke dogfood build secret-scan release-local release-signatures release-homebrew-formula release-preflight release-bootstrap release-readiness provider-setup-preflight production-readiness production-finalize production-local-gate eval-nightly eval-endurance eval-provider-kimi eval-provider-codex eval-provider-openai eval-provider-openrouter eval-provider-moonshot

ci: fmt test vet smoke dogfood build

fmt:
	gofmt -w ./cmd ./internal

test:
	go test ./... -count=1

test-race:
	go test -race -shuffle=on -count=1 ./...

vet:
	go vet ./...

smoke:
	sh scripts/smoke.sh

dogfood:
	sh scripts/dogfood.sh

build:
	mkdir -p bin
	go build -trimpath -ldflags="-X ceoharness/internal/cli.Version=$(VERSION) -X ceoharness/internal/cli.Commit=$(COMMIT)" -o bin/$(BINARY) $(PKG)

secret-scan:
	sh scripts/secret-scan.sh

release-local:
	sh scripts/release-local.sh

release-signatures:
	sh scripts/release-signatures.sh --dist dist --private-key "$$RELEASE_SIGNING_KEY"

release-homebrew-formula:
	sh scripts/release-homebrew-formula.sh --dist dist --repo-url "$$REPO_URL" --homebrew-archive-base-url "$$HOMEBREW_ARCHIVE_BASE_URL"

release-preflight:
	sh scripts/release-preflight.sh dist

release-bootstrap:
	sh scripts/release-bootstrap.sh --dist dist --output-dir .omo/evidence/release-bootstrap

release-readiness:
	sh scripts/release-readiness.sh --dist dist --output-dir .omo/evidence/release-readiness

provider-setup-preflight:
	sh scripts/provider-setup-preflight.sh --output-dir .omo/evidence/provider-setup-preflight

production-readiness:
	sh scripts/production-readiness.sh --dist dist --output-dir .omo/evidence/production-readiness

production-finalize:
	sh scripts/production-finalize.sh --dist dist --output-dir .omo/evidence/production-finalize

production-local-gate:
	sh scripts/production-local-gate.sh --dist dist --output-dir .omo/evidence/production-local-gate

eval-nightly:
	mkdir -p bin
	go build -trimpath -o bin/$(BINARY) $(PKG)
	go run ./cmd/ceo-eval --benchmark-fixtures --tasks evals/tasks --output-dir .omo/evidence/nightly/benchmark-fixtures
	go run ./cmd/ceo-packet gauntlet --suite cross-language-core --agents ceo_harness --ceo-binary ./bin/$(BINARY) --tasks evals/tasks --output-dir .omo/evidence/nightly/cross-language-core --timeout-seconds 120 --concurrency 2 --ceo-benchmark-mode model-command --ceo-benchmark-model-command-json '["sh","$(CURDIR)/scripts/benchmark-model-command.sh"]'
	sh scripts/dogfood-real.sh --repo "ceo-harness:$(CURDIR)" --repeat 2 --timeout-ms 250 --output-dir .omo/evidence/nightly/dogfood-real

eval-endurance:
	sh scripts/endurance.sh --iterations 3 --output-dir .omo/evidence/endurance-local-r1

eval-provider-kimi:
	sh scripts/provider-proof.sh --provider kimi --output-dir .omo/evidence/provider-proof-kimi

eval-provider-codex:
	sh scripts/provider-proof.sh --provider codex --output-dir .omo/evidence/provider-proof-codex

eval-provider-openai:
	sh scripts/provider-proof.sh --provider openai --output-dir .omo/evidence/provider-proof-openai

eval-provider-openrouter:
	sh scripts/provider-proof.sh --provider openrouter --output-dir .omo/evidence/provider-proof-openrouter

eval-provider-moonshot:
	sh scripts/provider-proof.sh --provider moonshot --output-dir .omo/evidence/provider-proof-moonshot
