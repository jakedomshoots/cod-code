# CEO Harness

## Project Shape

- This is a Go CLI product, not a TypeScript/Python orchestration shell.
- Keep the CEO/subagent runtime lean: compact context packets, local state, bounded subagent counts, and explicit final verdicts.
- Prefer CLI-first polish over GUI work until the core loop has been dogfooded on real coding tasks.

## Engineering Rules

- Run `go test ./... -count=1`, `go vet ./...`, `sh scripts/smoke.sh`, and `sh scripts/dogfood.sh` before claiming product-level completion.
- Use `go test -race -shuffle=on -count=1 ./...` for release gates or runtime/concurrency changes.
- Keep Go files below 250 non-blank, non-comment lines.
- Do not add broad dependencies or SDK frameworks unless they replace meaningful local complexity.
