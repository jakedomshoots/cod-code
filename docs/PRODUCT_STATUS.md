# Product Status

Status date: 2026-07-03

## Where It Stands

CEO Harness is now a productized local CLI, not just a prototype folder. The core runtime is working and the repo has basic product infrastructure:

- Git repository initialized on `main`.
- Local build, install, smoke, dogfood, and release scripts.
- CI workflow for tests, vet, smoke, dogfood, race, build, and the local production gate.
- Local release workflow for versioned archives, checksums, a release manifest, verifier, and a draft Homebrew formula.
- Tag-triggered GitHub Release workflow for publishing verified archives, checksums, and the release manifest.
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
- Public-readiness operator command: `production-status` reads the latest readiness packet and reports local/public readiness plus the next launch action.
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
- The real proof still needs public release setup and additional key-backed HTTP provider proof beyond the current copied-app, Kimi, and Codex proofs.
- External/provider gauntlets can still be blocked or incomplete when a provider key, CLI login, timeout log, git status snapshot, or scorer artifact is missing.

## Market Roadmap Result 2026-07-03

- Completed the market CLI roadmap implementation through doctor, release, docs, recovery UX, gauntlet/reporting, and local proof gates.
- Full repo gates passed: `go test ./... -count=1`, `go vet ./...`, `go test -race -shuffle=on -count=1 ./...`, smoke, dogfood, release-local, and doctor.
- The first 10-task `market-parity-core` CEO Harness gauntlet produced partial results while required task evidence artifacts were being enforced. The current rerun passes 10/10 with complete evidence at `.omo/evidence/market-parity-core-ceo-r2/summary.json`.
- Early bounded cross-agent and real-model path-safety comparisons exposed missing-evidence gaps. Later runs now have complete saved evidence for focused, four-task, production-core, Kimi, and current multi-file comparisons.
- This is local-release ready with honest limitations; it is not yet a public market-win claim.

## Production Hardening Progress 2026-07-03

- Benchmark runner now writes a complete artifact packet for missing-agent-binary and terminal benchmark errors instead of leaving summary rows without evidence files.
- Added regression coverage for missing CEO binary evidence: command, stdout, stderr, report, score, diff, changed-files, git-status, and timing artifacts must all exist and be non-empty.
- Current `market-parity-core` CEO result: 10 tasks / 10 pass / 0 partial / 0 incomplete evidence at `.omo/evidence/market-parity-core-ceo-r2/summary.json`.
- Added `production-core`, now a 29-task gauntlet suite, available through `ceo-packet gauntlet --suite production-core`.
- Expanded full benchmark fixture scoring now covers 31 tasks, including release readiness, lean-context autonomy, and secret-safe provider proof tasks, with 31/31 pass at `.omo/evidence/benchmark-fixtures-31-r1/summary.json`.
- Synthetic CEO benchmark mode now creates task-specific required evidence artifacts, so the runner can prove a clean pass instead of stopping at partial.
- Production-core smoke result: 24 tasks / 24 pass / 0 partial / 0 incomplete evidence at `.omo/evidence/production-core-smoke-r2/summary.json`.
- Production-core model-command result: 24 tasks / 24 pass / 0 partial / 0 incomplete evidence at `.omo/evidence/production-core-model-command-r4/summary.json`.
- Expanded production-core CEO result after adding the first multi-file provider/config task: 25 tasks / 25 pass / 0 partial / 0 incomplete evidence at `.omo/evidence/production-core-25-ceo-r1/summary.json`.
- Expanded production-core CEO result after adding the four-file operator safety task: 26 tasks / 26 pass / 0 partial / 0 incomplete evidence at `.omo/evidence/production-core-26-ceo-r1/summary.json`.
- Expanded production-core CEO result after adding release-readiness, lean-context autonomy, and secret-safe provider-proof tasks: 29 tasks / 29 pass / 0 partial / 0 incomplete evidence at `.omo/evidence/production-core-29-ceo-r1/summary.json`.
- Cross-language-core CEO result after adding JavaScript and Python benchmark fixtures: 2 tasks / concurrency 2 / 2 pass / 0 partial / 0 incomplete evidence at `.omo/evidence/cross-language-core-ceo-r1/summary.json`.
- Repeated real Kimi provider path-safety result: 3 attempts / 3 pass / 18 scored checks / 0 partial / 0 fail / 0 incomplete evidence at `.omo/evidence/provider-kimi-path-safety-repeat-r7/summary.json`.
- Real Kimi provider JS app-code result: `cross-language-js-state-reducer` passed 6/6 scored checks with complete evidence at `.omo/evidence/provider-kimi-js-state-reducer-r2/summary.json` after the Kimi bridge was fixed to include JS/TS sibling tests in its compact context.
- First-class Kimi provider proof gate: `scripts/provider-proof.sh --provider kimi --output-dir .omo/evidence/provider-proof-kimi-r2` passed both cross-language real-code tasks: JS state reducer 6/6 and Python retry policy 7/7 with complete evidence.
- First-class Codex provider proof gate: `scripts/provider-proof.sh --provider codex --output-dir .omo/evidence/provider-proof-codex-r1` passed both cross-language real-code tasks: JS state reducer 6/6 and Python retry policy 7/7 with complete evidence.
- First-class HTTP provider proof gates are now wired for `openai`, `openrouter`, and `moonshot`; missing API keys are saved as blocked setup evidence with `summary.json`, `env.template`, `commands.sh`, and `setup-checklist.md` instead of failed benchmark results, with setup checklist count and SHA-256 fingerprints recorded in `summary.json`.
- Competitor smoke now covers all six required competitors and passes after local setup: Codex CLI 0.142.4, Claude Code 2.1.201, Aider 0.86.2, OpenCode 1.17.13, Goose 1.41.0, and Pi 0.80.3 at `.omo/evidence/competitor-smoke-after-installs-r1/summary.json`.
- Competitor smoke treats Pi as a required competitor, matching the final all-agent comparison set instead of checking only Codex CLI, Claude Code, Aider, OpenCode, and Goose.
- Bounded external-agent comparison result: CEO Harness, Codex CLI, OpenCode, and Pi all passed `docs-roadmap-cli-first` with complete evidence at `.omo/evidence/external-agent-one-r1/summary.json`.
- Two-task external-agent comparison result: CEO Harness, Codex CLI, OpenCode, and Pi all passed `docs-roadmap-cli-first` and `bugfix-cli-timeout` with complete evidence at `.omo/evidence/external-agent-2task-r1/summary.json`.
- Four-task external-agent comparison result: CEO Harness, Codex CLI, OpenCode, and Pi all passed `docs-roadmap-cli-first`, `bugfix-cli-timeout`, `safety-policy-path-escape`, and `recovery-resume-retry` with complete evidence at `.omo/evidence/external-agent-4task-r2/summary.json`.
- Full expanded production-core external-agent comparison result: 25 tasks x 4 agents = 100 live runs completed at `.omo/evidence/external-agent-production-core-25-r1/summary.json`.
  - CEO Harness: 25 pass / 0 partial / 0 timeout / 0 fail / 0 incomplete evidence.
  - Codex CLI: 25 pass / 0 partial / 0 timeout / 0 fail / 0 incomplete evidence.
  - OpenCode: 25 pass / 0 partial / 0 timeout / 0 fail / 0 incomplete evidence.
  - Pi: 24 pass / 0 partial / 1 timeout / 0 fail / 1 incomplete evidence.
  - Overall: 99 pass / 0 partial / 1 timeout / 0 fail / 1 incomplete evidence.
- Focused current-suite external-agent comparison for the 26th multi-file operator safety task: CEO Harness, Codex CLI, OpenCode, and Pi all passed with complete evidence at `.omo/evidence/external-agent-operator-safety-flow-r1/summary.json`.
- Current 29-task external-agent comparison result: 29 tasks x 4 agents = 116 live runs completed at `.omo/evidence/external-agent-production-core-29-r1/summary.json`; CEO Harness stayed clean, overall result was 115 pass / 1 partial / 0 timeout / 0 incomplete evidence.
- The one partial was a brittle path-safety diff-term rubric on Codex CLI, not a behavior failure; the rubric now accepts the fixture's actual `ErrPathEscapesWorkspace` addition, and the focused all-agent rerun passed 4/4 at `.omo/evidence/external-agent-path-escape-rubric-r1/summary.json`.
- A fresh 29-task all-agent rerun with the corrected rubric completed at `.omo/evidence/external-agent-production-core-29-r2/summary.json`; CEO Harness again stayed clean, but external tools produced 6 timeouts, so the all-agent comparison gate stayed blocked at that point.
- Latest final 29-task all-agent comparison with timeout/result retries completed at `.omo/evidence/external-agent-production-core-29-final-result-retry-r1/summary.json`: 116 runs / 116 pass / 0 partial / 0 fail / 0 timeout / 0 incomplete evidence. One OpenCode partial was retried once and then passed with the prior attempt preserved.
- Added bounded timeout retries for local-agent benchmark runs through `--local-agent-benchmark-timeout-retries`; prior timed-out attempts are kept in `prior_attempts` and separate `attempt-XX` evidence folders.
- Focused retry proof over five timeout-heavy current-suite tasks completed at `.omo/evidence/external-agent-timeout-retry-r1/summary.json`: CEO Harness 5/5 pass, Codex CLI 5/5 pass, Pi 5/5 pass, OpenCode 0/5 pass after exhausting both attempts.
- Added per-agent timeout overrides through `--local-agent-benchmark-agent-timeouts`; focused OpenCode proof with `opencode=600` still timed out at `.omo/evidence/opencode-agent-timeout-r1/summary.json`, proving the current blocker is OpenCode execution behavior rather than the harness timeout ceiling.
- Local-agent comparison reports now include a readiness decision that separates a clean CEO Harness result from external-agent blockers such as competitor timeouts or incomplete evidence.
- Tightened local-agent benchmark prompts to tell competitors to avoid unrelated file inspection and broad test suites, then reran the five timeout-heavy tasks at `.omo/evidence/external-agent-timeout-prompt-discipline-r1/summary.json`: Codex CLI passed 5/5 and Pi passed 5/5, while OpenCode still timed out 5/5.
- Added per-agent model overrides through `--local-agent-benchmark-agent-models` and enabled OpenCode provider logs in benchmark evidence; short OpenCode proof at `.omo/evidence/opencode-setup-blocked-r1/summary.json` now reports `setup_blocked` with complete evidence and exposes the blocker as MiniMax token-plan quota in `stderr.log`.
- Benchmark gauntlets now support bounded parallelism with `--local-agent-benchmark-concurrency` and the product alias `ceo-packet gauntlet --concurrency`.
- Live real-repo dogfood result: `scripts/dogfood-real.sh --repo ceo-harness:<repo> --timeout-ms 250` passed all five scenarios, including expected timeout failure evidence, at `.omo/evidence/dogfood-real/index.md`.
- Repeated real-repo dogfood result: `scripts/dogfood-real.sh --repo ceo-harness-repeat:<repo> --repeat 3 --timeout-ms 250 --output-dir .omo/evidence/dogfood-real-repeat-self-r1` produced 3 live passes / 0 fails.
- Copied-workspace dogfood result: `scripts/dogfood-real.sh --copy-workspace --repo ceo-harness-copy:<repo> --timeout-ms 250 --output-dir .omo/evidence/dogfood-real-copy-self-r1` passed all five scenarios without using the source checkout as the writable workspace.
- Independent copied-workspace dogfood result: ChemCheck and Axis Health both passed all five no-key scenarios in copied workspaces at `.omo/evidence/dogfood-real-independent-r1/index.md`.
- Expanded independent copied-workspace dogfood result: Clicky, DPS Internal Comms, Janus, and Radian all passed all five no-key scenarios in copied workspaces at `.omo/evidence/dogfood-real-independent-r2/index.md`.
- Task-specific copied-workspace dogfood result: ChemCheck and Axis Health both passed all five no-key scenarios using a custom onboarding-docs cleanup task at `.omo/evidence/dogfood-real-task-specific-r1/index.md`.
- Copied-workspace write-probe dogfood result: ChemCheck and Axis Health both passed six scenarios, including preview plus approved write against copied workspaces only, at `.omo/evidence/dogfood-real-write-probe-r1/index.md`; source checkouts stayed without `ceo-dogfood-write-probe.txt`.
- Copied-workspace feature-edit dogfood result: ChemCheck and Axis Health both passed feature-note edits through preview plus approved write at `.omo/evidence/dogfood-real-feature-edit-r2/index.md`; source checkouts stayed without `ceo-dogfood-feature.md`.
- Copied-workspace app-code dogfood result: ChemCheck and Axis Health both passed source-module edits through preview plus approved write at `.omo/evidence/dogfood-real-app-code-r1/index.md`; source checkouts stayed without `src/ceoDogfoodProbe.mjs`.
- Copied-workspace integrated app-code dogfood result: ChemCheck and Axis Health both passed existing-source-file edits through preview plus approved write at `.omo/evidence/dogfood-real-integrated-app-code-r1/index.md`; source checkouts stayed without `ceoDogfoodIntegratedProbe`.
- Copied-workspace multi-file app-code dogfood result: ChemCheck and Axis Health both passed two existing-source-file edits through preview plus approved write at `.omo/evidence/dogfood-real-multi-file-app-code-r1/index.md`; source checkouts stayed without `ceoDogfoodMultiFileProbe`.
- Broadened copied-workspace multi-file app-code dogfood result: Janus passed two existing-source-file edits through preview plus approved write at `.omo/evidence/dogfood-real-multi-file-janus-r1/index.md`; the Janus source checkout stayed without `ceoDogfoodMultiFileProbe`.
- Broader app-shaped copied-workspace dogfood result: Family OS, Pools, DPS frontend, and Janus Mobile all passed two existing-source-file edits through preview plus approved write at `.omo/evidence/dogfood-real-broader-apps-r1/index.md`; all source checkouts stayed without `ceoDogfoodMultiFileProbe`.
- Added no-key nightly eval automation through `make eval-nightly` and `.github/workflows/nightly-evals.yml`.
- Added endurance eval runner through `scripts/endurance.sh`, `make eval-endurance`, and `task eval:endurance`; short local proof produced 3/3 passing iterations at `.omo/evidence/endurance-local-r1/index.md`.
- Longer local endurance proof produced 10/10 passing iterations, each running build, 28-task fixture scoring, cross-language gauntlet, and real-repo dogfood, at `.omo/evidence/endurance-local-r2/index.md`.
- Extended local endurance proof produced 30/30 passing iterations in 102 seconds at `.omo/evidence/endurance-local-r3/index.md`.
- Added release manifest and verifier through `dist/release-manifest.json` and `scripts/verify-release.sh`.
- Added public release bootstrap evidence through `scripts/release-bootstrap.sh`; it prepares commands, env, checklist, and a remote Homebrew formula draft without publishing anything.
- Added public release preflight through `scripts/release-preflight.sh`; it blocks public claims until remote URL, Homebrew URL, and signature/checksum posture are explicit.
- Added public release readiness evidence through `scripts/release-readiness.sh`; it writes `index.md`, `summary.json`, preflight output, git remote state, and GitHub auth state without publishing anything.
- Release readiness now writes `setup-actions.md` when public release blockers remain and records its action count plus SHA-256 in `summary.json`, so missing remote, release URL, GitHub release assets, Homebrew URL, and signing/checksum policy work is listed in one actionable, fingerprinted file.
- Added tag-triggered GitHub Release publishing; pushing a `v*` tag builds from the tag version, verifies archives, creates the GitHub Release, and attaches tarballs, checksums, and the manifest.
- Added production-readiness aggregate evidence through `scripts/production-readiness.sh`; it summarizes release, provider, eval, security, endurance, and all-agent comparison proof in one packet without publishing or calling paid providers, writes `launch-checklist.md` with the exact public-production actions left, and fingerprints that checklist in `summary.json`.
- Production-readiness now selects the newest matching evidence packet by file timestamp instead of filename order, so a fresh `external-agent-production-core-29-final-result-retry-r1` packet is not hidden behind older `-r*` packets.
- Added guarded finalization runner through `scripts/production-finalize.sh` and `ceo-packet production-finalize`; it sequences release readiness, HTTP provider proofs, competitor smoke, optional 29-task comparison, and final production readiness without publishing or saving secrets.
- Finalizer command replay is now shell-quoted, so generated `commands.sh` remains safe to inspect or rerun when output/evidence paths contain spaces.
- Finalizer competitor smoke now validates `summary.json`, so setup-blocked or failed competitors block the finalizer even when the smoke command exits 0.
- Finalizer now writes `next-actions.md` and records the required action count in `summary.json`, giving one place to see the exact remaining release/provider/comparison commands.
- `production-status` now surfaces the latest finalizer `next-actions.md` when public readiness is blocked, so the operator sees the most actionable next step first.
- `production-status` ignores partial finalizer packets with skipped steps, preventing a narrow smoke-only run from hiding release and provider blockers.
- `production-status` now prints both the human `next-actions.md` path and the automation-ready `next-actions.json` path from the latest complete finalizer packet.
- `production-status` now summarizes finalizer action states, runnable/blocked command counts, and declared-evidence match counts from `next-actions.json`, so the blocked status output shows missing env, empty env, setup-blocked, waiting, command safety, and evidence drift without opening the action queue.
- `production-status` now prints the finalizer `setup-actions.md` path when present, giving one consolidated public-readiness checklist for release, provider, competitor, comparison, and final gate work.
- Added `production-actions`, a read-only CLI command that prints the latest structured finalizer action checklist in JSON, text, event, or `--commands-only` format, with `--action-id`, `--action-kind`, `--action-provider`, `--action-state`, `--env-ready-only`, `--ready-only`, and `--next` filters for operator queues; advanced help includes production-action examples, shell completions include production action state/kind/provider values, `--action-state` rejects unknown states, provider env readiness is reported by non-empty presence only and never includes secret values, blank env vars are surfaced as `empty_env`, release/competitor setup blockers are surfaced as `setup_blocked`, release setup markdown is parsed into structured `setup_action_items`, provider setup checklists are parsed into structured `checklist_items`, finalizer-declared evidence files are fingerprinted with SHA-256 and size metadata, current reports show whether evidence still matches the declared fingerprint, `--commands-only` prints setup items as safe shell comments and comments out blocked commands, reports include `action_state_counts`, `runnable_command_count`, and `blocked_command_count`, each action gets an `action_state`, `Ready now` excludes setup-blocked and dependency-blocked actions, release and competitor setup actions summarize saved blocker evidence directly, downstream actions show what they are waiting on, and text output includes shell-quoted `Command:` lines.
- Provider-proof actions now surface each HTTP provider's blocked reason, selected model, setup checklist, and setup command file directly from saved evidence without printing secret values.
- Latest complete finalizer next-actions result: `.omo/evidence/production-finalize-after-clean-comparison-r1/next-actions.md` lists five remaining actions across release proof, HTTP providers, and final readiness using repo-relative commands; `.omo/evidence/production-finalize-after-clean-comparison-r1/next-actions.json` records structured action ids, kinds, commands, provider names, required env vars, and evidence paths for automation. Its root `.omo/evidence/production-finalize-after-clean-comparison-r1/setup-actions.md` links release, provider, and final rerun steps in one checklist.
- Latest guarded finalization result: `scripts/production-finalize.sh --output-dir .omo/evidence/production-finalize-after-clean-comparison-r1 --dist dist` wrote a blocked evidence packet, passed all six competitor smoke checks, recognized clean all-agent comparison evidence, and saved release/provider setup blockers without publishing or saving secret values.
- Added `ceo-packet production-status`, a read-only operator command that reports the latest local/public production-readiness state from saved evidence.
- Added `scripts/production-local-gate.sh` and CI artifact upload so source CI fails on local production regressions while preserving public-production blockers as evidence; the gate also requires the production action queue while public blockers remain, saves `production-actions.json` plus a paste-safe command script, and checks runnable/blocked command counts.
- Strict checks now run `sh -n` across shell scripts even when ShellCheck is not installed; ShellCheck remains an optional deeper lint layer.
- Rollback now covers created-file model patches as well as normal replacement patches; created-file rollback refuses to delete if the file content changed after creation.
- Default `--help` is now compact and points advanced users to `--help-advanced`; the full reference remains available without loading the first screen with every flag.
- Latest verification: `go test ./... -count=1`, `go vet ./...`, `sh scripts/smoke.sh`, `sh scripts/dogfood.sh`, `sh scripts/release-local.sh`, `task ci`, `golangci-lint run ./...`, `nilaway ./...`, and `sh scripts/strict-checks.sh`.
- First product baseline commit: `8509a4b Initial CEO Harness production baseline`.
- Remaining evidence gap: add deeper task-specific real-repo jobs with real writes, more provider families, and overnight or truly long-duration endurance runs before making broad market-win claims.

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
