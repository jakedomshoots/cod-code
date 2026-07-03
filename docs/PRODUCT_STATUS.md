# Product Status

Status date: 2026-07-03

## Where It Stands

CEO Harness is now a productized local CLI, not just a prototype folder. The core runtime is working and the repo has basic product infrastructure:

- Git repository initialized on `main`.
- Local build, install, smoke, dogfood, and release scripts.
- CI workflow for tests, vet, smoke, dogfood, race, and build.
- Local release workflow for versioned archives, checksums, a release manifest, verifier, and a draft Homebrew formula.
- MIT license, changelog, task runner, Makefile fallback, editor config, and project instructions.

The core product loop already supports:

- CEO final verdict ownership.
- Bounded native subagents.
- Lean context packets and context budgets.
- Local job history and review queue.
- Resume, continue, rerun, and human final judgment.
- Preview/dry-run patch approval with digest matching.
- Provider routing, fallback, provider health, and route-decision reporting.
- JSON, compact text, and JSONL event output.
- Primary operator commands: `start`, `run`, `gauntlet`, `doctor`, `inbox`, `status`, `resume`, `retry`, `rollback`, and `explain-failure`.
- Guided start, friendly inbox/status, provider wizard, golden demo repo generation, write policy presets, shell completions, external adapter presets, and lightweight interactive `tui`.

## Positioning

CEO Harness is not trying to beat mature tools at editor polish today. Its wedge is local orchestration discipline:

- One CEO owns the final verdict.
- One primary worker owns each job.
- Subagents are bounded and role-specific.
- Context is compact by default.
- Routing can stay cheap for ordinary work and escalate only for risk.
- Saved state is local and inspectable.

## Comparison Snapshot

| Alternative | Strong At | CEO Harness Difference |
|---|---|---|
| Claude Code | Mature terminal coding agent with deep codebase workflow | CEO Harness is model/provider-agnostic and emphasizes bounded subagents plus lean saved context |
| OpenAI Codex CLI | Fast local terminal coding agent, open source, Rust-based | CEO Harness is an orchestration layer pattern: CEO, subagents, route decisions, saved review queue |
| Aider | Pair-programming edits with strong git workflow | CEO Harness is less pair-chat focused and more job-owner/review/verdict focused |
| OpenCode | Open source terminal/desktop/IDE coding agent | CEO Harness stays smaller and CLI-first while tracking CEO/subagent state explicitly |
| GitHub Copilot CLI | GitHub-native issue/PR workflow and agent modes | CEO Harness is independent of GitHub and local-first |
| Goose | General-purpose local agent with desktop, CLI, and API | CEO Harness is narrower: coding harness first, final-verdict workflow first |

## Current Weak Spots

- No public remote repository is configured yet.
- ShellCheck is not installed on this machine, so shell-script linting is skipped by the strict gate.
- The real proof still needs repeated dogfooding on more independent non-demo coding repos.
- Market gauntlet evidence can still be partial/incomplete when a provider key, CLI login, timeout log, git status snapshot, or scorer artifact is missing.
- The CLI still has many advanced flags, but the common help surface now starts with the primary operator flow.

## Market Roadmap Result 2026-07-03

- Completed the market CLI roadmap implementation through doctor, release, docs, recovery UX, gauntlet/reporting, and local proof gates.
- Full repo gates passed: `go test ./... -count=1`, `go vet ./...`, `go test -race -shuffle=on -count=1 ./...`, smoke, dogfood, release-local, and doctor.
- The 10-task `market-parity-core` CEO Harness gauntlet ran and produced 10 partial results because required task evidence artifacts were intentionally enforced instead of silently created.
- A bounded cross-agent comparison ran on `docs-roadmap-cli-first` across CEO Harness, Codex CLI, OpenCode, and Pi. All four were partial because required evidence artifacts were incomplete.
- Real Codex and Kimi model-command path-safety proofs ran. Both passed the code/test/path-safety parts but remained partial because the required task evidence file was missing.
- This is local-release ready with honest limitations; it is not yet a public market-win claim.

## Production Hardening Progress 2026-07-03

- Benchmark runner now writes a complete artifact packet for missing-agent-binary and terminal benchmark errors instead of leaving summary rows without evidence files.
- Added regression coverage for missing CEO binary evidence: command, stdout, stderr, report, score, diff, changed-files, git-status, and timing artifacts must all exist and be non-empty.
- Added `production-core`, now a 26-task gauntlet suite, available through `ceo-packet gauntlet --suite production-core`.
- Synthetic CEO benchmark mode now creates task-specific required evidence artifacts, so the runner can prove a clean pass instead of stopping at partial.
- Production-core smoke result: 24 tasks / 24 pass / 0 partial / 0 incomplete evidence at `.omo/evidence/production-core-smoke-r2/summary.json`.
- Production-core model-command result: 24 tasks / 24 pass / 0 partial / 0 incomplete evidence at `.omo/evidence/production-core-model-command-r4/summary.json`.
- Expanded production-core CEO result after adding the first multi-file provider/config task: 25 tasks / 25 pass / 0 partial / 0 incomplete evidence at `.omo/evidence/production-core-25-ceo-r1/summary.json`.
- Expanded production-core CEO result after adding the four-file operator safety task: 26 tasks / 26 pass / 0 partial / 0 incomplete evidence at `.omo/evidence/production-core-26-ceo-r1/summary.json`.
- Cross-language-core CEO result after adding JavaScript and Python benchmark fixtures: 2 tasks / concurrency 2 / 2 pass / 0 partial / 0 incomplete evidence at `.omo/evidence/cross-language-core-ceo-r1/summary.json`.
- Repeated real Kimi provider path-safety result: 3 attempts / 3 pass / 18 scored checks / 0 partial / 0 fail / 0 incomplete evidence at `.omo/evidence/provider-kimi-path-safety-repeat-r7/summary.json`.
- Bounded external-agent comparison result: CEO Harness, Codex CLI, OpenCode, and Pi all passed `docs-roadmap-cli-first` with complete evidence at `.omo/evidence/external-agent-one-r1/summary.json`.
- Two-task external-agent comparison result: CEO Harness, Codex CLI, OpenCode, and Pi all passed `docs-roadmap-cli-first` and `bugfix-cli-timeout` with complete evidence at `.omo/evidence/external-agent-2task-r1/summary.json`.
- Four-task external-agent comparison result: CEO Harness, Codex CLI, OpenCode, and Pi all passed `docs-roadmap-cli-first`, `bugfix-cli-timeout`, `safety-policy-path-escape`, and `recovery-resume-retry` with complete evidence at `.omo/evidence/external-agent-4task-r2/summary.json`.
- Full expanded production-core external-agent comparison result: 25 tasks x 4 agents = 100 live runs completed at `.omo/evidence/external-agent-production-core-25-r1/summary.json`.
  - CEO Harness: 25 pass / 0 partial / 0 timeout / 0 fail / 0 incomplete evidence.
  - Codex CLI: 25 pass / 0 partial / 0 timeout / 0 fail / 0 incomplete evidence.
  - OpenCode: 25 pass / 0 partial / 0 timeout / 0 fail / 0 incomplete evidence.
  - Pi: 24 pass / 0 partial / 1 timeout / 0 fail / 1 incomplete evidence.
  - Overall: 99 pass / 0 partial / 1 timeout / 0 fail / 1 incomplete evidence.
- Benchmark gauntlets now support bounded parallelism with `--local-agent-benchmark-concurrency` and the product alias `ceo-packet gauntlet --concurrency`.
- Live real-repo dogfood result: `scripts/dogfood-real.sh --repo ceo-harness:<repo> --timeout-ms 250` passed all five scenarios, including expected timeout failure evidence, at `.omo/evidence/dogfood-real/index.md`.
- Repeated real-repo dogfood result: `scripts/dogfood-real.sh --repo ceo-harness-repeat:<repo> --repeat 3 --timeout-ms 250 --output-dir .omo/evidence/dogfood-real-repeat-self-r1` produced 3 live passes / 0 fails.
- Copied-workspace dogfood result: `scripts/dogfood-real.sh --copy-workspace --repo ceo-harness-copy:<repo> --timeout-ms 250 --output-dir .omo/evidence/dogfood-real-copy-self-r1` passed all five scenarios without using the source checkout as the writable workspace.
- Independent copied-workspace dogfood result: ChemCheck and Axis Health both passed all five no-key scenarios in copied workspaces at `.omo/evidence/dogfood-real-independent-r1/index.md`.
- Expanded independent copied-workspace dogfood result: Clicky, DPS Internal Comms, Janus, and Radian all passed all five no-key scenarios in copied workspaces at `.omo/evidence/dogfood-real-independent-r2/index.md`.
- Added no-key nightly eval automation through `make eval-nightly` and `.github/workflows/nightly-evals.yml`.
- Added endurance eval runner through `scripts/endurance.sh`, `make eval-endurance`, and `task eval:endurance`; short local proof produced 3/3 passing iterations at `.omo/evidence/endurance-local-r1/index.md`.
- Longer local endurance proof produced 10/10 passing iterations, each running build, 28-task fixture scoring, cross-language gauntlet, and real-repo dogfood, at `.omo/evidence/endurance-local-r2/index.md`.
- Added release manifest and verifier through `dist/release-manifest.json` and `scripts/verify-release.sh`.
- Added public release preflight through `scripts/release-preflight.sh`; it blocks public claims until remote URL, Homebrew URL, and signature/checksum posture are explicit.
- Rollback now covers created-file model patches as well as normal replacement patches; created-file rollback refuses to delete if the file content changed after creation.
- Latest verification: `go test ./... -count=1`, `go vet ./...`, `sh scripts/smoke.sh`, `sh scripts/dogfood.sh`, `sh scripts/release-local.sh`, `task ci`, `golangci-lint run ./...`, `nilaway ./...`, and `sh scripts/strict-checks.sh`.
- First product baseline commit: `8509a4b Initial CEO Harness production baseline`.
- Remaining evidence gap: add deeper task-specific real-repo jobs beyond copied no-key dogfood, harder multi-file jobs, and overnight or truly long-duration endurance runs before making broad market-win claims.

## Additions Completed 2026-07-02

1. `ceo-packet --start <path>` guided setup/check/doctor flow.
2. `ceo-packet --inbox` review queue alias with text details.
3. `ceo-packet --provider-wizard <preset>` for OpenAI-compatible provider setup.
4. `ceo-packet --init-demo-repo <path>` golden demo repo generator.
5. Safer write policies: `observe`, `preview`, `dry-run`, `approved-write`, `trusted-local`.
   - Patch write intent previews by default.
   - `preview` is the non-writing patch review mode.
   - `dry-run` remains a compatibility alias for preview.
   - `approved-write` applies writes only with a matching `--approve-preview` digest.
   - `trusted-local` is the explicit local opt-in for direct writes.
6. External worker adapter presets: Codex CLI, Claude Code, OpenCode, Aider, Goose.
7. Lightweight `ceo-packet --tui` operator dashboard with deterministic snapshot mode plus stdin-driven navigation/action dispatch.
8. Wave 3 operator polish: simpler first-screen help, primary command aliases, zsh/bash/fish completions, clearer compact text summaries, and provider setup wording for Codex CLI, Kimi CLI, and OpenRouter blocked-key states.

## Best Next Features

1. Dogfood the primary operator flow on more real repos and tighten awkward output.
2. Add harder real-repo tasks beyond the controlled benchmark fixtures.
3. Keep tightening external-agent timeout/setup gaps, especially around Pi.

## Current References

- OpenAI Codex CLI: https://developers.openai.com/codex/cli
- Claude Code: https://github.com/anthropics/claude-code
- Aider: https://github.com/Aider-AI/aider
- OpenCode: https://opencode.ai/docs/
- GitHub Copilot CLI: https://docs.github.com/en/copilot/how-tos/copilot-cli/use-copilot-cli/overview
- Goose: https://goose-docs.ai/
