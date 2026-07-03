# Product Status

Status date: 2026-07-03

## Where It Stands

CEO Harness is now a productized local CLI, not just a prototype folder. The core runtime is working and the repo has basic product infrastructure:

- Git repository initialized on `main`.
- Local build, install, smoke, dogfood, and release scripts.
- CI workflow for tests, vet, smoke, dogfood, race, and build.
- Local release workflow for versioned archives, checksums, and a draft Homebrew formula.
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
- Strict external tools (`gofumpt`, `golangci-lint`, `nilaway`) are configured/documented but not installed on this machine.
- The real proof still needs repeated dogfooding on non-demo coding tasks.
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
- Added `production-core`, a 24-task gauntlet suite, available through `ceo-packet gauntlet --suite production-core`.
- Synthetic CEO benchmark mode now creates task-specific required evidence artifacts, so the runner can prove a clean pass instead of stopping at partial.
- Production-core smoke result: 24 tasks / 24 pass / 0 partial / 0 incomplete evidence at `.omo/evidence/production-core-smoke-r2/summary.json`.
- Production-core model-command result: 24 tasks / 24 pass / 0 partial / 0 incomplete evidence at `.omo/evidence/production-core-model-command-r4/summary.json`.
- Bounded external-agent comparison result: CEO Harness, Codex CLI, OpenCode, and Pi all passed `docs-roadmap-cli-first` with complete evidence at `.omo/evidence/external-agent-one-r1/summary.json`.
- Two-task external-agent comparison result: CEO Harness, Codex CLI, OpenCode, and Pi all passed `docs-roadmap-cli-first` and `bugfix-cli-timeout` with complete evidence at `.omo/evidence/external-agent-2task-r1/summary.json`.
- Latest verification: `go test ./... -count=1`, `go vet ./...`, `sh scripts/smoke.sh`, and `sh scripts/dogfood.sh`.
- First product baseline commit: `d9f3055 Initial CEO Harness production baseline`.
- Remaining evidence gap: broaden the external-agent comparison beyond two tasks before making market-wide claims.

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

1. Dogfood the primary operator flow on a real repo and tighten awkward output.
2. Run a multi-task external-agent comparison, then the full 24-task suite when runtime cost/time is acceptable.
3. Use the external-agent results to fix remaining cross-agent prompt/setup gaps before making market claims.

## Current References

- OpenAI Codex CLI: https://developers.openai.com/codex/cli
- Claude Code: https://github.com/anthropics/claude-code
- Aider: https://github.com/Aider-AI/aider
- OpenCode: https://opencode.ai/docs/
- GitHub Copilot CLI: https://docs.github.com/en/copilot/how-tos/copilot-cli/use-copilot-cli/overview
- Goose: https://goose-docs.ai/
