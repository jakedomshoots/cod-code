# Install

CEO Harness currently supports verified local installs from this checkout. There is no published remote install URL yet.

## One-Line Local Install

```sh
PREFIX="$PWD/.local" sh scripts/install-local.sh
```

Then run:

```sh
.local/bin/ceo-packet --version
```

Use a temp prefix for QA:

```sh
PREFIX="$(mktemp -d)" sh scripts/install-local.sh
```

The script builds `./cmd/ceo-packet`, installs only the `ceo-packet` binary, prints `ceo-packet <version> commit=<commit>`, and does not require Homebrew, Task, ShellCheck, golangci-lint, nilaway, or gofumpt.

## First Run

Start with a real repo and keep the first run read-only:

```sh
ceo-packet start /path/to/repo
ceo-packet config explain --workspace /path/to/repo --format text
ceo-packet run --workspace /path/to/repo --plan-only --format text -- "Fix one failing test"
```

For a guarded repair run, require checks and use the standard repair preset:

```sh
ceo-packet run --workspace /path/to/repo \
  --repair-preset standard \
  --check go test ./... -- \
  "Fix one failing test"
```

## Real Provider Setup

Codex CLI and Kimi CLI use adapter presets:

```sh
ceo-packet config init --workspace /path/to/repo --adapter codex
ceo-packet config init --workspace /path/to/repo --adapter kimi
```

OpenRouter uses the HTTP provider wizard and needs a real key:

```sh
export OPENROUTER_API_KEY=...
ceo-packet --workspace /path/to/repo --provider-wizard openrouter --format text
ceo-packet config doctor --workspace /path/to/repo --format text
```

Codex/Kimi/OpenRouter missing key or missing login states are blocked setup, not proof that the harness failed a benchmark. Save the exact command output before scoring the run.

## Market Gauntlet

Use the gauntlet command to create market evidence:

```sh
ceo-packet gauntlet --agents ceo_harness --output-dir .omo/evidence/gauntlet
```

Use the stricter 25-task production suite when you want production-readiness evidence instead of the faster 10-task market smoke:

```sh
ceo-packet gauntlet --suite production-core --agents ceo_harness --output-dir .omo/evidence/production-gauntlet
```

The gauntlet can report partial/incomplete evidence when an agent, key, login, timeout log, git status, or scorer artifact is missing. Treat that as a setup or evidence gap until the raw artifacts prove a pass or fail.

## Recovery Commands

```sh
ceo-packet status --workspace /path/to/repo
ceo-packet explain-failure latest --workspace /path/to/repo
ceo-packet retry latest --workspace /path/to/repo
ceo-packet rollback .ceo-harness/history/job-000001.json --workspace /path/to/repo
```

`rollback` only supports saved JSON reports with supported patch shapes. For anything more complex, inspect the saved report and git diff before trusting the rollback.

## Requirements

Required:

- Go 1.23 or newer.
- POSIX `sh`.

Optional:

- `task` for Taskfile shortcuts.
- `gofumpt`, `golangci-lint`, and `nilaway` for stricter release gates.
- `shellcheck` for shell linting.

If optional tools are missing, use the documented fallback commands in [Verification](VERIFICATION.md).

## Known Limits

- No remote install URL or Homebrew tap is published yet.
- Local archives are checksum-only; they are not signed.
- External provider quality depends on the configured provider, model, login, and key.
- Current market comparison evidence is useful but still narrow. Do not describe prototype areas as proven unless the saved command logs and artifacts show it.
