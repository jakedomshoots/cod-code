# Contributing

Keep changes CLI-first, small, and verified through the real binary.

## Local Gate

Run:

```sh
go test ./... -count=1
go vet ./...
sh scripts/smoke.sh
sh scripts/dogfood.sh
```

For release or runtime/concurrency changes, also run:

```sh
go test -race -shuffle=on -count=1 ./...
VERSION=0.1.0-dev sh scripts/release-local.sh
cd dist && shasum -a 256 -c checksums.txt
```

Optional stricter tools:

- `gofumpt`
- `golangci-lint`
- `nilaway`
- `shellcheck`
- `task`

These are useful, but missing optional tools do not block source install QA.

## Docs

When adding local markdown links, run the link check documented in [Trust Surface](docs/TRUST.md).

## Changelog

Add a short entry under `0.1.0-dev - Unreleased` for user-visible changes.

## Publish Claims

Do not document remote install, Homebrew, signed artifacts, or public release availability until those exact commands have been verified.

