#!/bin/sh
set -eu

root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
cd "$root"
workspace=$(mktemp -d)
provider_workspace=$(mktemp -d)
trap 'rm -rf "$workspace" "$provider_workspace"' EXIT

go run ./cmd/ceo-packet --version >/dev/null
go run ./cmd/ceo-packet \
	--doctor \
	--ceo-model-command sh examples/ceo-model.sh -- \
	--model-command sh examples/command-model.sh -- \
	--research-command sh examples/research-command.sh >/dev/null
go run ./cmd/ceo-packet --quickstart "$workspace" >/dev/null
go run ./cmd/ceo-packet \
	--workspace "$provider_workspace" \
	--init-config \
	--init-example-adapters \
	--http-provider main \
	--http-preset openai \
	--http-model gpt-5 \
	--ceo-provider main \
	--default-provider main \
	--risk-provider high=main \
	--kind-provider research=main >/dev/null
