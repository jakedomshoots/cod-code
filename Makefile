BINARY := ceo-packet
PKG := ./cmd/ceo-packet
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || printf dev)
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || printf local)

.PHONY: ci fmt test test-race vet smoke dogfood build release-local

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
