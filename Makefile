BINARY := ceo-packet
PKG := ./cmd/ceo-packet
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || printf dev)
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || printf local)

.PHONY: ci fmt test test-race vet smoke dogfood build release-local release-preflight eval-nightly eval-endurance eval-provider-kimi eval-provider-codex

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

release-local:
	sh scripts/release-local.sh

release-preflight:
	sh scripts/release-preflight.sh dist

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
