# CEO Harness

CEO Harness is a local Go CLI for coding work where one CEO loop owns the final verdict, delegates bounded work to lean subagents, keeps compact local job state, and makes every run reviewable.

First command:

```sh
go run ./cmd/ceo-packet --help
```

Verified local install:

```sh
PREFIX=/tmp/ceo-harness sh scripts/install-local.sh
/tmp/ceo-harness/bin/ceo-packet --version
```

Quick first task:

```sh
ceo-packet --quickstart /path/to/repo --format text
ceo-packet --workspace /path/to/repo --plan-only --format text -- "Fix one failing test"
```

## Product Docs

- [Install](docs/INSTALL.md)
- [Quick first task](docs/QUICK_FIRST_TASK.md)
- [Trust and release proof](docs/TRUST.md)
- [Terminal text snapshots](docs/TERMINAL_CASTS.md)
- [Product status](docs/PRODUCT_STATUS.md)
- [Dogfood gate](docs/DOGFOOD.md)
- [Release process](docs/RELEASE.md)
- [Homebrew tap plan](docs/HOMEBREW.md)
- [Roadmap](docs/ROADMAP.md)
- [Verification record](docs/VERIFICATION.md)
- [Security policy](SECURITY.md)
- [Contributing](CONTRIBUTING.md)
- [Changelog](CHANGELOG.md)

## What It Is Not

- Not a full IDE clone.
- Not a GUI-first coding assistant.
- Not a wrapper that dumps the whole repository into every prompt.
- Not several agents duplicating the same job without one owner.

## Quick Start

Install the local CLI binary:

```sh
PREFIX="$PWD/.local" sh scripts/install-local.sh
.local/bin/ceo-packet --version
```

For Go's standard install path:

```sh
go install ./cmd/ceo-packet
ceo-packet --version
```

If `ceo-packet` is not found after install, add `$(go env GOPATH)/bin` to your `PATH`.

Copy-paste first setup for a repo:

```sh
ceo-packet config explain --workspace . --format text
ceo-packet --workspace . --provider-wizard openai --http-model gpt-5 --format text
ceo-packet config doctor --workspace . --format text
ceo-packet run --workspace . --check go test ./... -- "Fix one failing test"
```

Generate shell completions:

```sh
mkdir -p /tmp/ceo-completions
ceo-packet config completions --shell zsh --output-dir /tmp/ceo-completions
```

Install a named local build into `~/.local/bin`:

```sh
VERSION=0.1.0 COMMIT=local sh scripts/install-local.sh
ceo-packet --version
```

For a temporary or custom install, set `PREFIX` or `BINDIR`:

```sh
PREFIX=/tmp/ceo-harness sh scripts/install-local.sh
```

Run the local smoke check:

```sh
sh scripts/smoke.sh
```

Run the product dogfood gate:

```sh
sh scripts/dogfood.sh
```

Run the full local gate:

```sh
make ci
make test-race
make eval-nightly
make eval-endurance
```

Build local release archives:

```sh
VERSION=0.1.0 sh scripts/release-local.sh
(cd dist && shasum -a 256 -c checksums.txt)
```

The release script writes tarballs, checksums, and a local Homebrew formula draft under `dist/homebrew/ceo-packet.rb`. It does not publish, tag, push, or create a remote release.

Create a working example workspace config:

```sh
ceo-packet --quickstart /path/to/repo
ceo-packet --workspace /path/to/repo --doctor
```

Use the shorter operator start flow when you want setup, config check, doctor, and next commands in one text report:

```sh
ceo-packet --start /path/to/repo --format text
```

Use `--quickstart /path/to/repo --format text` for a short first-run checklist.

Market gauntlet and recovery commands:

```sh
ceo-packet gauntlet --agents ceo_harness --output-dir .omo/evidence/gauntlet
ceo-packet gauntlet --suite production-core --agents ceo_harness --concurrency 4 --output-dir .omo/evidence/production-gauntlet
ceo-packet gauntlet --suite cross-language-core --agents ceo_harness --concurrency 2 --output-dir .omo/evidence/cross-language-gauntlet
sh scripts/dogfood-real.sh --repo "ceo-harness:$PWD" --repeat 3 --output-dir .omo/evidence/dogfood-real-repeat
sh scripts/dogfood-real.sh --copy-workspace --repo "ceo-harness:$PWD" --output-dir .omo/evidence/dogfood-real-copy
sh scripts/dogfood-real.sh --copy-workspace --write-probe --repo "ceo-harness:$PWD" --output-dir .omo/evidence/dogfood-real-write-probe
ceo-packet explain-failure latest --workspace .
ceo-packet retry latest --workspace .
ceo-packet rollback .ceo-harness/history/job-000001.json --workspace .
```

Quickstart auto-detects Go, Rust, package.json, pytest Python, and Makefile workspaces (`go.mod`, `Cargo.toml`, a `scripts.test` entry, pytest config, or a `test` target), writes the default test command, and enables `require_checks` so coding runs cannot pass without verification. Package workspaces use `bun test`, `pnpm test`, `yarn test`, or `npm test` based on the lockfile. Python workspaces use `uv run pytest` when `uv.lock` is present, otherwise `python -m pytest`. Makefile workspaces use `make test`.

Create a real-provider workspace config where the CEO and default subagents use an OpenAI-compatible endpoint:

```sh
export OPENAI_API_KEY=...
ceo-packet --workspace /path/to/repo \
  --provider-wizard openai \
  --http-model gpt-5 \
  --format text
```

The provider wizard creates provider `main`, routes CEO/default/fallback work to it, and prints the next secret/doctor steps.

Create a tiny golden repo for local dogfooding:

```sh
ceo-packet --init-demo-repo /tmp/ceo-harness-demo --format text
```

Build a named local binary with version metadata:

```sh
go build \
  -ldflags "-X ceoharness/internal/cli.Version=0.1.0 -X ceoharness/internal/cli.Commit=local" \
  -o ./bin/ceo-packet \
  ./cmd/ceo-packet
./bin/ceo-packet --version
```

```sh
go run ./cmd/ceo-packet Fix a failing test
```

Print CLI help:

```sh
go run ./cmd/ceo-packet --help
```

Run the built-in golden coding demo, no API key required:

```sh
go run ./cmd/ceo-packet --demo
```

Run with the bundled command-model example:

```sh
go run ./cmd/ceo-packet \
  --model-command sh examples/command-model.sh -- \
  Inspect the workspace
```

Command models receive the prompt on stdin plus lightweight metadata env vars: `CEO_MODEL_REQUEST_KIND`, `CEO_AGENT_NAME`, `CEO_AGENT_ROLE`, and `CEO_CONTEXT_MODE`.

Run the harness doctor. This reuses the golden demo and reports pass/fail health:

```sh
go run ./cmd/ceo-packet --doctor
```

Check a command-model adapter too:

```sh
go run ./cmd/ceo-packet --doctor --model-command sh examples/command-model.sh
```

Check a CEO model adapter too:

```sh
go run ./cmd/ceo-packet --doctor --ceo-model-command sh examples/ceo-model.sh
```

Check a research adapter too:

```sh
go run ./cmd/ceo-packet --doctor --research-command sh examples/research-command.sh
```

Check one configured provider without running the full doctor:

```sh
ceo-packet --workspace /path/to/repo --doctor-provider main --format text
```

Doctor adapter checks report their command `source` as `flag`, `env`, or `workspace`.
When `ceo_provider` is configured, `--doctor` also exercises that provider through the CEO delegation and final-review path. Workspace providers are checked once as `provider.<name>` so real worker routes fail fast during setup.
If `require_checks` is enabled, `--doctor` also fails when no verification command can be resolved.

Print a compact human report instead of the full JSON report:

```sh
go run ./cmd/ceo-packet --format text Fix a failing test
```

Preview the packet, provider routes, checks, and limits without running models:

```sh
go run ./cmd/ceo-packet --workspace /path/to/repo --plan-only Fix a failing test
```

Print that preview as compact text:

```sh
go run ./cmd/ceo-packet --workspace /path/to/repo --plan-only --format text Fix a failing test
```

Patch writes preview by default. Preview patch writes without changing workspace files or saving run artifacts/history:

```sh
go run ./cmd/ceo-packet \
  --workspace /path/to/repo \
  --dry-run \
  --replace app.txt old new -- \
  Patch app text
```

Use write policy presets when a repo should be safer by default:

```sh
ceo-packet --workspace /path/to/repo --write-policy dry-run --replace app.txt old new -- "Patch app text"
ceo-packet --workspace /path/to/repo --write-policy approved-write --approve-preview <digest> --replace app.txt old new -- "Patch app text"
```

Dry-run and default preview reports include `patch_approval.preview_digest`. To apply that exact preview, pass the digest back:

```sh
go run ./cmd/ceo-packet \
  --workspace /path/to/repo \
  --approve-preview <preview_digest> \
  --replace app.txt old new -- \
  Patch app text
```

Direct patch writes require an explicit trusted profile:

```sh
go run ./cmd/ceo-packet \
  --workspace /path/to/repo \
  --write-policy trusted-local \
  --replace app.txt old new -- \
  Patch app text
```

Tighten the lean task-packet budget:

```sh
go run ./cmd/ceo-packet --max-context-bytes 2048 Fix a failing test
```

The budget caps task context such as task text, assignments, tool results, prior findings, and workspace brief content; fixed harness instructions like the JSON response contract stay intact.

Cap tool requests per subagent:

```sh
go run ./cmd/ceo-packet --max-tool-requests 2 Fix a failing test
```

Default auto-delegation stays at 3 subagents. Raise it only when a task really needs a wider crew:

```sh
go run ./cmd/ceo-packet --max-subagents 7 Research payment database migration and deploy the fix
```

Cap the whole job wall-clock time:

```sh
go run ./cmd/ceo-packet --job-timeout-ms 120000 Fix a failing test
```

With a workspace and check command:

```sh
go run ./cmd/ceo-packet \
  --workspace /path/to/repo \
  --check go test ./... -- \
  Fix the failing test
```

When no check command is configured, reports mark verification as `unverified` instead of hiding the skipped check path.

Require a real verification command before any normal run or plan preview:

```sh
go run ./cmd/ceo-packet \
  --workspace /path/to/repo \
  --require-checks \
  --check go test ./... -- \
  Fix the failing test
```

Apply the single patch-owner model patch:

```sh
go run ./cmd/ceo-packet \
  --workspace /path/to/repo \
  --apply-model-patches \
  --max-model-patches 5 \
  --model-command your-model-command -- \
  Fix the failing test
```

Use `--dry-run --apply-model-patches` to preview the coder patch proposal in `patch_previews` without applying it or writing harness artifacts. The same dry-run report includes `patch_approval.preview_digest`, and `--approve-preview <preview_digest>` rejects writes if the current preview differs.

Patch-owner proposals can be returned as standalone JSON:

```json
{"patches":[{"path":"app.txt","old":"old","new":"new"}]}
```

To create a new file, use `content` instead of `old`/`new`:

```json
{"patches":[{"path":"docs/notes.md","content":"# Notes\n"}]}
```

Create-file patches fail if the target already exists. Preview mode returns the diff without writing the file.

Run a researcher tool request through a local command:

```sh
go run ./cmd/ceo-packet \
  --model-command your-model-command -- \
  --research-command your-research-command -- \
  Research agent harness docs
```

`network_research` commands receive the requested query on stdin and in `CEO_RESEARCH_QUERY`. Use `CEO_RESEARCH_COMMAND_JSON` or workspace `research_command` to make it reusable.

Smoke-test the research adapter with doctor:

```sh
go run ./cmd/ceo-packet --doctor --research-command sh examples/research-command.sh
```

Local model command failures are classified in reports as `provider_error_kind` values such as `command_failed`, `command_timeout`, `command_output_too_large`, `model_output_empty`, or `model_output_invalid`, so fallback and history summaries can reason about command adapters the same way they do HTTP providers. `model_output_empty` means the model returned no usable text; `model_output_invalid` means the model returned structured JSON that failed harness validation, or an HTTP provider configured with `response_format: "json_object"` returned loose prose instead of the harness JSON contract. `verification_summary.provider_error_kind_counts` totals each kind. When fallback is used, `provider_fallback_reason` keeps the typed reason when one is available.

The preferred model output shape is one structured JSON object:

```json
{"summary":"short result","confidence":0.8,"evidence":["what was checked"],"tool_requests":[{"action":"read_workspace","path":"README.md"}],"patches":[{"path":"app.txt","old":"old","new":"new"}]}
```

If a model cannot safely continue without the user, it can stop the run explicitly:

```json
{"status":"needs_input","summary":"missing target package","questions":["Which package should I change?"]}
```

Resume that saved job with a compact answer instead of carrying old chat context:

```sh
go run ./cmd/ceo-packet \
  --workspace /path/to/repo \
  --resume job-000001 \
  --answer "Use internal/cli." \
  --model-command your-model-command --
```

Continue a saved job by reusing already passed subagents and running only the remaining work:

```sh
go run ./cmd/ceo-packet \
  --workspace /path/to/repo \
  --continue-job job-000001
```

Saved-job commands accept exact IDs or `latest`/`last`, so repeated local loops can stay short:

```sh
go run ./cmd/ceo-packet --workspace /path/to/repo --job latest
go run ./cmd/ceo-packet --workspace /path/to/repo --continue-job last
```

Record your own final judgment without feeding it back into model context:

```sh
go run ./cmd/ceo-packet \
  --workspace /path/to/repo \
  --judge-job latest \
  --human-verdict accept \
  --judgment-note "Looks good."
```

Read it later with:

```sh
go run ./cmd/ceo-packet --workspace /path/to/repo --judge-job latest
```

`--job <id>` and `--history` include `human_judgment` when a judgment sidecar exists; `--history --summary-only` counts accepted and rejected human judgments.
`--continue-job` refuses to reuse a job that has a human `reject` judgment.
`--review-queue` prints only jobs that still need human attention: jobs needing input, failed/unresolved jobs, human-rejected jobs, and passing jobs that have not been accepted or rejected.
Use `--review-queue --format text` for a compact terminal inbox.
Use `--inbox` for the friendlier default: text output plus compact job details.

```sh
go run ./cmd/ceo-packet --workspace /path/to/repo --inbox
go run ./cmd/ceo-packet --workspace /path/to/repo --tui
go run ./cmd/ceo-packet tui --workspace /path/to/repo --snapshot
```

The TUI dashboard stays CLI-first: snapshot mode prints deterministic text for CI/manual QA, including job list, selected job details, inbox action, provider health, patch preview, and latest check output.

Or let the CLI ask and resume automatically when a subagent needs input:

```sh
go run ./cmd/ceo-packet \
  --workspace /path/to/repo \
  --interactive \
  --model-command your-model-command -- \
  Fix the ambiguous package
```

Reports keep patch diffs in `patch_results` and patch provenance in `patch_audit`.
Workspace history stores compact patch counts, including CLI patches, model patches, and check-fix attempts. History rows and `--job` lookup also keep `run_ledger` so old jobs can be skimmed without opening the full saved report.
Use `--preview-model-patches` to inspect patch-owner diffs in `patch_previews` without changing files.
Reports with patch previews include `patch_approval`; text output prints the digest in one `Patch approval:` line.
Use `--format text` for a short human report or `--format events` for compact JSONL run events; JSON stays the default for automation. JSON and text reports include `job_owner`/`Owner` so the run names the primary work owner while CEO keeps final authority. JSON, text, and `--plan-only` output also include `run_ledger`, a compact owner/verdict/next-action/verification/changed-files/provider-routing summary for skimming a run without adding prompt context. JSON, text, and `--plan-only` output also include a compact verification contract, naming required checks and whether they are pending, passing, or failing. Event streams include provider-health reroutes as `provider_health_route` events before subagent work starts, provider fallback details on `subagent` events, and patch approval checkpoints as `patch_approval` events with digest/count fields.

Run a final model-backed CEO review:

```sh
go run ./cmd/ceo-packet \
  --ceo-model-command your-ceo-review-command -- \
  Fix the failing test
```

CEO review commands read the compact CEO prompt from stdin and must return JSON:

```json
{"recommended_verdict":"pass","summary":"short reason"}
```

The CEO review prompt includes compact subagent summaries, changed files, patch diffs, tool results, and check output, so the final verdict can inspect concrete work evidence without loading the full workspace.

Use `CEO_REVIEW_MODEL_COMMAND_JSON` or workspace `ceo_model_command` to make the reviewer reusable.

Let the CEO send one failed model review back to coder for a bounded patch revision:

```sh
go run ./cmd/ceo-packet \
  --workspace /path/to/repo \
  --apply-model-patches \
  --ceo-revision-attempts 1 \
  --ceo-model-command your-ceo-review-command \
  --model-command your-model-command -- \
  Fix the failing test
```

CEO revision attempts require a workspace and model patch application. The coder receives only the original task plus the compact CEO feedback, proposes patch JSON, checks rerun, and the CEO reviews the updated evidence again.
Use workspace `ceo_revision_attempts` to make that retry policy reusable.

Cap the CEO runtime loop. The default is `6`; the initial delegated run counts as iteration `1`, and each check-fix or CEO-revision pass consumes one more:

```sh
go run ./cmd/ceo-packet --max-ceo-iterations 3 "Fix the failing test"
```

Let the coder try one guarded fix after a failed check:

```sh
go run ./cmd/ceo-packet \
  --workspace /path/to/repo \
  --apply-model-patches \
  --check-fix-attempts 1 \
  --check go test ./... -- \
  --model-command your-model-command -- \
  Fix the failing test
```

## Init Config

Create `.ceo-harness.json` without overwriting an existing config:

```sh
go run ./cmd/ceo-packet \
  --workspace /path/to/repo \
  --init-config \
  --ceo-model-command your-ceo-review-command -- \
  --check go test ./... --
```

For a no-key local setup, use quickstart:

```sh
go run ./cmd/ceo-packet --quickstart /path/to/repo
```

This writes model, CEO, and research commands that point at the repo's `examples/` scripts, then runs `--doctor` from that workspace.

For external coding tools, initialize a thin command adapter:

```sh
go run ./cmd/ceo-packet --workspace /path/to/repo --init-config --adapter codex
```

Supported adapter presets are `codex`, `claude`, `opencode`, `aider`, and `goose`. The bundled scripts live in `examples/adapters/` and delegate to env vars such as `CEO_CODEX_ADAPTER_COMMAND` or `CEO_CLAUDE_ADAPTER_COMMAND`.

Delegate to custom subagents from `.ceo-harness.json`:

```json
{
  "subagents": [
    {"name": "planner", "role": "break down work", "allowed_actions": ["read_workspace", "search_workspace"]},
    {"name": "security", "role": "review auth risks", "allowed_actions": ["read_workspace", "search_workspace", "run_checks"]}
  ]
}
```

The harness keeps the same lean task packet and caps configured delegation at 8 subagents.
If `allowed_actions` is omitted, the harness assigns a lean default from the agent name.

Without custom subagents, the CEO picks a lean native crew from the task profile:

```text
coding: scanner, coder, reviewer
planning: planner, reviewer
research: researcher, reviewer
high-risk: adds only the needed specialists: billing, database, release, security
mixed: combines planner, researcher, coder, needed specialists, reviewer
```

Native subagents run in dependency stages:

```text
stage 1: scanner, planner, researcher
stage 2: coder and risk specialists
stage 3: reviewer
```

Custom rosters are sorted into those stages before execution. A custom or CEO-created subagent can set `stage` to `1`, `2`, or `3`; omission keeps the role-name default. A subagent can also set `max_context_bytes` to use a tighter packet budget than the global context policy. Tool requests and one feedback pass run before the CEO advances to the next stage, so later stages receive compact `prior_findings` from tool-updated earlier-stage summaries. Coder/reviewer agents can build on previous work without receiving the full conversation or full repo. Reports include `stage` and `prior_findings` on each subagent result and execution-plan step.

Limit parallel model calls within each dependency stage:

```sh
go run ./cmd/ceo-packet --subagent-concurrency 1 "Fix a failing test"
```

Use workspace `subagent_concurrency` to make that cap reusable. `0` or omission keeps the default behavior: all agents in the same dependency stage may run in parallel.

Limit runtime tool calls per subagent:

```sh
go run ./cmd/ceo-packet --max-tool-requests 2 "Fix a failing test"
```

Use workspace `max_tool_requests` to make that cap reusable. Extra requests are recorded as skipped tool results, so the CEO can see what was refused without running more tools.

Stop repeated weak subagent attempts without adding prompt context:

```sh
go run ./cmd/ceo-packet --subagent-attempts 4 --no-progress-stop 2 "Fix a failing test"
```

The default guard is `2`. Use workspace `no_progress_stop` to make a different threshold reusable. When a failing subagent repeats the same weak result, the report marks it with `no_progress_stopped`.

Limit saved subagent output:

```sh
go run ./cmd/ceo-packet --max-subagent-output-bytes 800 "Fix a failing test"
```

Use workspace `max_subagent_output_bytes` to make that cap reusable. When a subagent summary, evidence item, question, or attempt error is capped, the report marks that result with `output_truncated`.

Models can also return `confidence` from `0` to `1`. Set `--min-subagent-confidence` or workspace `min_subagent_confidence` to fail low-confidence pass results, or to route them to `provider_policy.fallback_provider` when one is configured.

Subagents also carry explicit action limits:

```text
scanner/planner: read_workspace, search_workspace
researcher: read_workspace, search_workspace, network_research
coder: read_workspace, search_workspace, propose_patch
security: read_workspace, search_workspace, run_checks
reviewer: read_workspace, run_checks, verify_evidence
```

Prompts, job packets, and subagent results include `allowed_actions`. Model patches are accepted only when exactly one passing subagent has `propose_patch`; multiple patch-capable subagents are rejected so the job keeps one edit owner.

Model subagents can request bounded runtime tools through the same structured JSON:

```json
{"summary":"need file context","tool_requests":[{"action":"read_workspace","path":"README.md"},{"action":"search_workspace","query":"TODO"},{"action":"network_research","query":"agent harness docs"},{"action":"run_checks"}]}
```

The CEO runtime parses `status`, `summary`, `evidence`, `questions`, `tool_requests`, and `patches` into typed report fields. It executes only allowed actions. Supported executable actions are `read_workspace`, `search_workspace`, `network_research`, `run_checks`, and `verify_evidence`; results are recorded under each subagent's `tool_results`. `verify_evidence` gives later-stage reviewers a compact prior-evidence packet without replaying the full run history. If a tool produces feedback, that subagent gets one bounded follow-up pass with the tool results in its prompt.

When a workspace is set, the CEO runtime builds a compact `workspace_brief` before delegation work starts. The brief indexes file paths and byte sizes while skipping noisy directories like `.git`, `node_modules`, `vendor`, `dist`, `build`, and `ceo-artifacts`; subagent prompts receive that brief instead of a repo dump.

When `ceo_model_command` is configured, the CEO model first chooses which default/adaptive or custom candidate subagents should run. It can also add narrow specialist subagents with explicit action limits, as long as the final selected roster stays at 8 or fewer:

```json
{"selected_subagents":["security","db_reviewer"],"new_subagents":[{"name":"db_reviewer","role":"review migrations","provider":"premium","stage":3,"max_context_bytes":1024,"allowed_actions":["read_workspace","run_checks"]}],"assignments":{"security":"Inspect auth risks only.","db_reviewer":"Check migration risk only."},"summary":"auth work needs security and database review"}
```

Assignments are optional and are sent only to the named selected subagent. A created subagent can set `provider` to any provider configured in `.ceo-harness.json`. Created subagents must also be selected; otherwise the delegation response is rejected. The same command is called again at the end for the final CEO verdict.

Create a routed HTTP provider for `scanner`:

```sh
go run ./cmd/ceo-packet \
  --workspace /path/to/repo \
  --init-config \
  --http-provider fast \
  --http-url https://api.example.com/v1/chat/completions \
  --http-model fast-model \
  --http-api-key-env CEO_FAST_KEY \
  --http-agent scanner \
  --http-timeout-ms 2500 \
  --http-max-output-tokens 512 \
  --http-response-format json_object \
  --http-input-cost-per-million 2.5 \
  --http-output-cost-per-million 8.5 \
  --provider-health-avoid-failure-rate 0.9 \
  --provider-health-watch-failure-rate 0.5 \
  --provider-health-watch-cost-per-attempt-microusd 100
```

Use a preset when the provider is OpenAI-compatible:

```sh
go run ./cmd/ceo-packet \
  --workspace /path/to/repo \
  --init-config \
  --http-provider cheap \
  --http-preset openrouter \
  --http-model '~openai/gpt-latest' \
  --default-provider cheap \
  --http-provider premium \
  --http-preset openai \
  --http-model gpt-5.5 \
  --fallback-provider premium
```

Repeat `--http-provider` to create multiple routes in one config. Presets currently fill only the endpoint and default API-key env var: `openai` uses `OPENAI_API_KEY`, `openrouter` uses `OPENROUTER_API_KEY`, and `kimi`/`moonshot` use `MOONSHOT_API_KEY`.

Set `--ceo-provider <name>` during config init or quickstart when that provider should own CEO delegation and final review. Explicit `ceo_provider` wins over the bundled example CEO adapter in quickstarted workspaces.

Use `--risk-provider high=premium` or `--kind-provider research=premium` during config init when the default route should stay cheap but higher-risk or research work should use the stronger provider.

Route by task risk instead of hard-coding every agent:

```json
{
  "providers": {
    "cheap": {"model_command": ["cheap-model-command"]},
    "premium": {"model_command": ["premium-model-command"]}
  },
  "provider_policy": {
    "default_provider": "cheap",
    "fallback_provider": "premium",
    "risk_area_providers": {"database": "premium", "billing": "premium"},
    "risk_providers": {"high": "premium"},
    "kind_providers": {"research": "premium"}
  }
}
```

Use `risk_area_providers` when you want normal work to stay on the cheap route while only matching specialists such as `database`, `billing`, `release`, or `security` move to a stronger provider. Explicit subagent `provider` routes win, then `agent_providers`; the policy fills in routes only for agents without a fixed provider. JSON reports and `--plan-only` include `provider_route_decisions` with the chosen provider and reason for each routed subagent. If saved history marks a routed provider as `avoid` and `provider_policy.fallback_provider` is configured, the route is moved to that fallback before the next run starts. If the fallback is also `avoid`, the original route is kept. `--config-check` and normal run manifests report how many routes were moved and which providers were avoided.

## Config Check

```sh
go run ./cmd/ceo-packet --workspace /path/to/repo --config-check
```

`--config-check` reports counts and sources only. It does not print provider secret values. For provider-health routing, config-check and `run_manifest` include `provider_health_avoided_route_count` and `provider_health_avoided_providers`.
Use `--config-check --format text` for a short provider setup checklist; it prints missing env var names and the next `--doctor-provider` command without printing secret values.

Summarize provider health across stored job history:

```sh
go run ./cmd/ceo-packet --workspace /path/to/repo --provider-health
```

Filter that rollup to one provider:

```sh
go run ./cmd/ceo-packet --workspace /path/to/repo --provider-health --provider fast
```

Filter that rollup to one recommendation label:

```sh
go run ./cmd/ceo-packet --workspace /path/to/repo --provider-health --recommendation avoid
```

Filter history or provider health to matching task text:

```sh
go run ./cmd/ceo-packet --workspace /path/to/repo --history --task checkout
go run ./cmd/ceo-packet --workspace /path/to/repo --provider-health --task checkout
```

Print one compact history row, a compact resume context packet, the saved full report, or saved JSONL events for a job:

```sh
go run ./cmd/ceo-packet --workspace /path/to/repo --review-queue
go run ./cmd/ceo-packet --workspace /path/to/repo --review-queue --review-details --format text
go run ./cmd/ceo-packet --workspace /path/to/repo --job job-000001
go run ./cmd/ceo-packet --workspace /path/to/repo --job-context job-000001
go run ./cmd/ceo-packet --workspace /path/to/repo --job-context job-000001 --format text
go run ./cmd/ceo-packet context --workspace /path/to/repo latest
go run ./cmd/ceo-packet --workspace /path/to/repo --context-trace job-000001 --format text
go run ./cmd/ceo-packet --workspace /path/to/repo --job-report job-000001
go run ./cmd/ceo-packet --workspace /path/to/repo --job-events job-000001
```

Use `--review-details` to include the compact job context directly in the review inbox: next action, questions, changed files, failed checks, and CEO review summary when a saved report snapshot is available.
Use `--job-context <id> --format text` for the same compact packet as a terminal-readable handoff. JSON job contexts include `suggested_command` when the next step can be resumed directly. Use `context <job>` or `--context-trace <id>` to inspect per-agent packet budgets, actual context bytes, truncation status, workspace brief counts, prior-finding counts, and excluded-content metadata without printing raw prompts or repo contents.

Start a new task with a compact previous-job packet instead of pasting old chat:

```sh
go run ./cmd/ceo-packet --workspace /path/to/repo --with-job-context job-000001 "Continue the checkout fix"
```

Rerun a saved job task with the current workspace config:

```sh
go run ./cmd/ceo-packet --workspace /path/to/repo --rerun job-000001
```

Continue a saved job without rerunning matching passed subagents:

```sh
go run ./cmd/ceo-packet --workspace /path/to/repo --continue-job job-000001
```

Resume a saved `needs_input` job with your answer:

```sh
go run ./cmd/ceo-packet --workspace /path/to/repo --resume job-000001 --answer "Use internal/cli."
```

Ask and resume from stdin in one command:

```sh
go run ./cmd/ceo-packet --workspace /path/to/repo --interactive "Fix ambiguous package"
```

Print compact history or provider-health counts without full rows. History summaries include verdict counts plus subagent, retry, no-progress stop, check, patch, provider failure, and provider cost totals.

```sh
go run ./cmd/ceo-packet --workspace /path/to/repo --history --summary-only
go run ./cmd/ceo-packet --workspace /path/to/repo --provider-health --summary-only
```

Show only the worst provider-health rows:

```sh
go run ./cmd/ceo-packet --workspace /path/to/repo --provider-health --top-providers 3
```

Provider-health rows include raw counts plus scan fields like `failure_rate`, `cost_per_attempt_microusd`, and `recommendation`. Recommendations are `avoid` for providers at or above 50% failures, `watch` for providers with lower failure/error signals, and `healthy` for clean providers. Rows sort worst-first by failure rate, error count, estimated cost, then provider name. The rollup also includes attempt/pass/fail/error/cost totals plus summary counts for avoid/watch/healthy/unknown providers. `--config-check` reports missing provider env var names and provider-health route avoidance without printing secret values. `--format events` and `--job-events` include a structured `provider_health_route` event when history reroutes work away from avoided providers.

## HTTP Provider Shape

```json
{
  "providers": {
    "fast": {
      "http": {
        "url": "https://api.example.com/v1/chat/completions",
        "model": "fast-model",
        "api_key_env": "CEO_FAST_KEY",
        "timeout_ms": 2500,
        "max_output_tokens": 512,
        "response_format": "json_object",
        "input_cost_per_million_tokens": 2.5,
        "output_cost_per_million_tokens": 8.5
      }
    }
  },
  "agent_providers": {
    "scanner": "fast"
  },
  "provider_policy": {
    "default_provider": "fast",
    "fallback_provider": "premium",
    "risk_area_providers": {"database": "premium", "security": "premium"},
    "risk_providers": {"high": "premium"},
    "kind_providers": {"research": "premium"}
  },
  "subagents": [
    {"name": "scanner", "role": "inspect scope", "allowed_actions": ["read_workspace", "search_workspace"]},
    {"name": "coder", "role": "apply bounded changes", "allowed_actions": ["read_workspace", "search_workspace", "propose_patch"]},
    {"name": "reviewer", "role": "verify evidence", "allowed_actions": ["read_workspace", "run_checks", "verify_evidence"]}
  ],
  "max_context_bytes": 2048,
  "max_subagent_output_bytes": 800,
  "min_subagent_confidence": 0.6,
  "job_timeout_ms": 120000,
  "max_ceo_iterations": 6,
  "max_subagents": 3,
  "subagent_concurrency": 1,
  "max_tool_requests": 2,
  "no_progress_stop": 2,
  "require_checks": true,
  "provider_cost_budget_microusd": 500,
  "ceo_revision_attempts": 1,
  "ceo_model_command": ["your-ceo-review-command"],
  "research_command": ["your-research-command"],
  "provider_health_avoid_failure_rate": 0.5,
  "provider_health_watch_failure_rate": 0.0,
  "provider_health_watch_cost_per_attempt_microusd": 100
}
```

HTTP providers use an OpenAI-style chat completion request and read `choices[0].message.content`. A rate-limited HTTP response gets one automatic retry even when no subagent retry flag is set, and `Retry-After` is honored when it is longer than the configured retry backoff. `provider_policy.fallback_provider` lets a stronger provider take over when the routed provider fails or returns confidence below `min_subagent_confidence`; successful fallback results include `provider_fallback_from` and `provider_fallback_reason`. Saved provider health also affects future routing: providers with recommendation `avoid` are pre-routed to the fallback when the fallback is not also avoided. `provider_cost_budget_microusd` fails the CEO verdict when the estimated provider cost exceeds the workspace budget. Provider-health recommendation thresholds can be tuned in JSON with `provider_health_avoid_failure_rate`, `provider_health_watch_failure_rate`, and `provider_health_watch_cost_per_attempt_microusd`; the same fields can also be scaffolded with `--provider-health-avoid-failure-rate`, `--provider-health-watch-failure-rate`, and `--provider-health-watch-cost-per-attempt-microusd`.

## Lean Context

Subagents receive compact task packets, not full chat history or repo dumps. Workspace runs add a bounded `workspace_brief` so agents can see the shape of the repo before requesting specific tools. Staged runs add `prior_findings`, a compact summary handoff from completed earlier-stage agents after their allowed tools and feedback pass run. Model JSON output is parsed once into typed `status`, `summary`, `confidence`, `evidence`, `questions`, `tool_requests`, and `patches`, so the CEO no longer has to scrape raw JSON strings for normal model responses. Tool feedback keeps saved results intact but caps large stdout/stderr when rendering the follow-up prompt, `max_tool_requests` can stop tool-call sprawl per subagent, `no_progress_stop` can stop repeated weak attempts, and `max_subagent_output_bytes` can cap saved subagent summary/evidence/question text before later agents or the CEO see it. `max_ceo_iterations` caps the initial run plus bounded correction passes. If any stage returns `needs_input`, later stages are skipped and the CEO verdict becomes `needs_input`; `--resume` loads that saved report and injects only the prior question plus your answer into the next task packet. `--job-context` prints a smaller packet for a saved job with task, verdict, next action, run ledger, questions, changed files, failed checks, and subagent summaries. `--with-job-context` injects that compact packet into a new task so a follow-up run can continue without carrying the full old report; when a saved run ledger exists, it uses one `previous_run_ledger` line instead of separate verdict/next-action lines. `--interactive` uses that same lean resume path after reading your answer from stdin. Reports include metadata like context bytes, provider route reason, request id, token usage, cost estimate, and structured provider error fields when providers return them. CEO delegation and final review also report `model_source` and `provider_name` when the CEO is routed through a configured provider. CEO-created specialists can set `provider` for one configured route and can own patches when they are the only selected subagent with `propose_patch`. Provider-health history can pre-route future work away from avoided providers without adding context to the prompt. CEO review and check-fix prompts stay compact while including capped check output with command, attempt, exit code, and duration metadata. If `--ceo-revision-attempts` is set and the CEO model vetoes an otherwise passing run, coder receives the compact CEO feedback for a bounded patch pass before checks and CEO review run again. `verification_summary` includes subagent attempt, retry, and no-progress stop totals. `verification_summary.provider_health` breaks provider attempts, pass/fail counts, error counts, estimated cost, and recommendation down by provider name, sorted worst-first. `--history` preserves the same retry, no-progress stop, and provider health rows for past jobs.

Reports also include `execution_plan`, a compact CEO-owned plan with delegated subagent steps, optional check step, final CEO verdict step, and `next_action`. They also include `run_events`, a short ordered event log for packet creation, delegation, subagents, tool requests/results, feedback passes, checks, patches, CEO review, and verdict. Workspace runs write the same plan to `ceo-artifacts/ceo-plan.md`.

Workspace runs also save full report snapshots under `ceo-artifacts/jobs/<job-id>.json`, so later sessions can replay the exact CEO packet, workspace brief, event log, subagent results, checks, delegation, review, and verdict without carrying old chat context.

Every job packet includes a small `task_profile` with `kind` (`coding`, `planning`, `research`, or `mixed`), `risk_level` (`low`, `medium`, or `high`), and optional `risk_areas` such as `billing`, `database`, `release`, or `security`. The profile is copied into the run manifest and history summary so later routing can stay cost-aware without adding chat history.
