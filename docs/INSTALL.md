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

Provider proof gates are available for both CLI-backed and HTTP-backed providers:

```sh
sh scripts/provider-proof.sh --provider kimi --output-dir .omo/evidence/provider-proof-kimi
sh scripts/provider-proof.sh --provider codex --output-dir .omo/evidence/provider-proof-codex
sh scripts/provider-proof.sh --provider openai --output-dir .omo/evidence/provider-proof-openai
sh scripts/provider-proof.sh --provider openrouter --output-dir .omo/evidence/provider-proof-openrouter
sh scripts/provider-proof.sh --provider moonshot --output-dir .omo/evidence/provider-proof-moonshot
```

HTTP proof gates require non-empty `OPENAI_API_KEY`, `OPENROUTER_API_KEY`, or `MOONSHOT_API_KEY`. Missing keys are recorded as `blocked_missing_key`; blank keys are recorded as `blocked_empty_key`.
When a key is missing, the proof gate also writes `summary.json`, `env.template`, `commands.sh`, and `setup-checklist.md` so setup blockers can be resolved without saving secret values. The summary records the setup checklist count and SHA-256 fingerprints for the setup artifacts.

## Market Gauntlet

Use the gauntlet command to create market evidence:

```sh
ceo-packet gauntlet --agents ceo_harness --output-dir .omo/evidence/gauntlet
```

Use the stricter 29-task production suite when you want production-readiness evidence instead of the faster 10-task market smoke:

```sh
ceo-packet gauntlet --suite production-core --agents ceo_harness --concurrency 4 --output-dir .omo/evidence/production-gauntlet
```

The gauntlet can report partial/incomplete evidence when an agent, key, login, timeout log, git status, or scorer artifact is missing. Treat that as a setup or evidence gap until the raw artifacts prove a pass or fail.

## Recovery Commands

```sh
ceo-packet status --workspace /path/to/repo
ceo-packet production-status --workspace /path/to/repo --format text
ceo-packet explain-failure latest --workspace /path/to/repo
ceo-packet retry latest --workspace /path/to/repo
ceo-packet rollback .ceo-harness/history/job-000001.json --workspace /path/to/repo
```

`production-status` reads the latest `.omo/evidence/production-readiness*/summary.json` packet and reports local readiness, public readiness, blockers, and the launch checklist next action.

For the final public-production evidence sequence, run:

```sh
ceo-packet production-finalize --workspace . --dry-run
```

Remove `--dry-run` after release metadata and provider key environment variables are ready. The script writes evidence and command files, but it does not publish, tag, upload, or save secret values.

To list the remaining production actions without opening files:

```sh
ceo-packet production-actions --workspace . --format text
ceo-packet production-actions --workspace . --format text --action-id provider-openai
ceo-packet production-actions --workspace . --format text --action-kind release_proof
ceo-packet production-actions --workspace . --format text --action-kind provider_proof
ceo-packet production-actions --workspace . --format text --action-provider openai
ceo-packet production-actions --workspace . --format text --action-state missing_env
ceo-packet production-actions --workspace . --format text --action-state empty_env
ceo-packet production-actions --workspace . --format text --action-state setup_blocked
ceo-packet production-actions --workspace . --format text --env-ready-only
ceo-packet production-actions --workspace . --format text --ready-only
ceo-packet production-actions --workspace . --format text --next
ceo-packet production-actions --workspace . --format text --action-kind competitor_setup
ceo-packet production-actions --workspace . --format text --action-kind final_readiness
ceo-packet production-actions --workspace . --action-id provider-openai --commands-only
```

`--commands-only` is paste-safe: actions that are missing environment variables, setup-blocked, or waiting on another action are emitted as commented `# blocked command:` lines. Use `--ready-only --commands-only` when you want only immediately runnable commands.

`rollback` supports saved JSON reports for normal replacements and created files produced by CEO Harness. It refuses to remove a created file if the file content no longer matches the saved report.

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
- Public release readiness is tracked with `sh scripts/release-readiness.sh --dist dist --output-dir .omo/evidence/release-readiness`.
- Public production readiness can be checked with `ceo-packet production-status --workspace . --format text` after running `scripts/production-readiness.sh`.
- External provider quality depends on the configured provider, model, login, and key.
- Current market comparison evidence is useful but still narrow. Do not describe prototype areas as proven unless the saved command logs and artifacts show it.
