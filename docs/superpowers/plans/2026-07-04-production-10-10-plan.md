# CEO Harness Production 10/10 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Move CEO Harness from locally production-ready to a public, reliable, installable, provider-proven agentic coding harness that can be used in a real production environment.

**Architecture:** Keep the product local-first and CLI-first. Production readiness is proven by repeatable release artifacts, provider/auth evidence, real-repo coding evals, security gates, and operator docs, all surfaced through the existing `production-status`, `production-actions`, and finalizer evidence flow.

**Tech Stack:** Go CLI, shell scripts, JSON/Markdown evidence packets, GitHub Actions, local OAuth CLI wrappers, OpenAI-compatible HTTP providers, Homebrew formula draft, GitHub Releases.

---

## Current Baseline

As of 2026-07-04, the harness is locally strong but not public-production complete.

- Latest production code checkpoint: `a98ed53 feat(auth): add cli oauth provider setup`.
- Current `production-status`: `Local ready: true`, `Public ready: false`.
- Current blockers: public release readiness, OpenAI HTTP provider proof, Moonshot HTTP provider proof, final readiness rerun, and remaining finalizer setup actions.
- OAuth CLI providers ready on this Mac: `kimi`, `codex`, `claude`, `opencode`, `goose`.
- Already green: full Go tests, vet, secret scan, Kimi/MiniMax proof, local production gate, and local CLI functionality.

## 10/10 Definition

CEO Harness reaches 10/10 production only when all of these are true:

- A clean user can install it from a public release and run `ceo-packet doctor` successfully.
- `production-status --workspace . --format text` prints `Production status: pass`.
- `production-actions --workspace . --format text` has no blocked required actions.
- At least five provider paths are proven: Kimi CLI OAuth, Codex CLI OAuth, Claude CLI OAuth, OpenRouter HTTP, and one direct API provider such as OpenAI, Kimi Code, MiniMax, or Moonshot.
- Real-repo dogfood includes actual approved writes across at least five different repos, not only synthetic fixtures.
- CI blocks regressions for tests, vet, secret scan, release artifact checks, local production gate, and docs drift.
- Public docs let a new developer install, configure, run, recover, and compare the harness without reading source code.
- Secrets are never printed, committed, saved in evidence, or included in generated command files.

## File Map

- `docs/PRODUCTION_10_10.md`: create the public production readiness rubric and current scoreboard.
- `docs/PRODUCT_STATUS.md`: keep the current state honest after each gate.
- `docs/INSTALL.md`: keep install and provider setup simple for a new user.
- `README.md`: keep the first-run path, OAuth path, and release path current.
- `scripts/release-readiness.sh`: tighten release evidence blockers.
- `scripts/production-finalize.sh`: final source of truth for release/provider/comparison readiness.
- `scripts/production-readiness.sh`: aggregate readiness gate.
- `scripts/provider-proof.sh`: provider proof runner.
- `scripts/provider-setup-preflight.sh`: provider env readiness runner.
- `scripts/production-local-gate.sh`: CI-enforced local production gate.
- `.github/workflows/release.yml`: public release workflow.
- `.github/workflows/ci.yml`: source CI gate.
- `internal/cli/production_status.go`: production status output.
- `internal/cli/production_actions.go`: structured action queue output.
- `internal/cli/oauth.go`: CLI OAuth provider setup.
- `internal/cli/help.go`: top-level operator guidance.

---

### Task 1: Create The Production Scoreboard

**Files:**
- Create: `docs/PRODUCTION_10_10.md`
- Modify: `docs/PRODUCT_STATUS.md`
- Test: manual doc lint plus status command

- [ ] **Step 1: Write the scoreboard document**

Create `docs/PRODUCTION_10_10.md` with these sections:

```markdown
# CEO Harness Production 10/10 Scoreboard

Status date: 2026-07-04

## Current Score

Overall: 7.5/10

## Gates

| Gate | Required For 10/10 | Current State | Evidence |
|---|---|---|---|
| Local CLI | Full tests, vet, secret scan, local production gate pass | Pass | `go test ./... -count=1`, `go vet ./...`, `sh scripts/secret-scan.sh` |
| Public release | Installable public release with checksums and release manifest | Blocked | `.omo/evidence/production-finalize/next-actions.md` |
| Provider proof | OAuth and HTTP provider paths proven without secret leakage | Partial | Kimi/MiniMax pass; OpenAI/Moonshot blocked |
| Real repo dogfood | Five real repos with approved writes and rollback evidence | Partial | existing dogfood packets |
| CI gates | Regression gates run in CI and preserve artifacts | Pass locally, needs public workflow proof | `.github/workflows/ci.yml` |
| Docs onboarding | New user can install, configure, run, recover | Partial | `README.md`, `docs/INSTALL.md` |
| Security posture | Secret scan, no token storage, safe command files, path safety | Pass locally, needs final audit packet | `scripts/secret-scan.sh` |

## Exit Criteria

- `ceo-packet production-status --workspace . --format text` prints `Production status: pass`.
- `ceo-packet production-actions --workspace . --format text` shows zero required blocked actions.
- Public install instructions have been replayed from a clean temp directory.
- Provider proofs exist for Kimi CLI, Codex CLI, Claude CLI, OpenRouter HTTP, Kimi Code HTTP, MiniMax HTTP, OpenAI HTTP, and Moonshot HTTP where credentials are available.
```

- [ ] **Step 2: Update product status**

Add one bullet to `docs/PRODUCT_STATUS.md` under `Current Weak Spots`:

```markdown
- Public 10/10 production status is tracked in `docs/PRODUCTION_10_10.md`; the release/provider/finalizer gates must be green before making public production claims.
```

- [ ] **Step 3: Verify docs**

Run:

```sh
git diff --check
go run ./cmd/ceo-packet production-status --workspace . --format text
```

Expected:

```text
git diff --check exits 0
production-status still reports current blockers honestly
```

- [ ] **Step 4: Commit**

```sh
git add docs/PRODUCTION_10_10.md docs/PRODUCT_STATUS.md
git commit -m "docs(production): add 10 out of 10 scoreboard"
```

---

### Task 2: Close Public Release Readiness

**Files:**
- Modify: `scripts/release-readiness.sh`
- Modify: `scripts/release-preflight.sh`
- Modify: `docs/INSTALL.md`
- Verify: `.omo/evidence/release-readiness-final/index.md`

- [ ] **Step 1: Build local release artifacts**

Run:

```sh
sh scripts/release-local.sh --dist dist
sh scripts/verify-release.sh --dist dist
```

Expected:

```text
archives exist under dist/
checksums exist under dist/
dist/release-manifest.json verifies
```

- [ ] **Step 2: Generate or confirm release signing posture**

Use checksum-only mode if signing keys are not ready:

```sh
sh scripts/release-readiness.sh --dist dist --output-dir .omo/evidence/release-readiness-final
```

Expected if public release metadata is still missing:

```text
.omo/evidence/release-readiness-final/setup-actions.md exists
.omo/evidence/release-readiness-final/setup-commands.sh exists
summary.json records publish_actions_performed=false
summary.json records secret_value_saved=false
```

- [ ] **Step 3: Fill public release metadata**

Set the release metadata through existing environment variables or script inputs used by `release-preflight.sh`:

```sh
test -n "${GITHUB_REPOSITORY:-}" || { echo "set GITHUB_REPOSITORY, for example owner/repo"; exit 1; }
export RELEASE_VERSION=0.1.0
export RELEASE_PUBLIC_URL="https://github.com/${GITHUB_REPOSITORY}/releases/tag/v0.1.0"
export RELEASE_ARCHIVE_URL="https://github.com/${GITHUB_REPOSITORY}/releases/download/v0.1.0/ceo-packet_Darwin_arm64.tar.gz"
export HOMEBREW_ARCHIVE_URL="https://github.com/${GITHUB_REPOSITORY}/releases/download/v0.1.0/ceo-packet_Darwin_arm64.tar.gz"
```

This checkout currently has no git remote configured, so `GITHUB_REPOSITORY` must be set explicitly before public release readiness can pass.

- [ ] **Step 4: Rerun release readiness**

Run:

```sh
sh scripts/release-readiness.sh --dist dist --output-dir .omo/evidence/release-readiness-final
```

Expected:

```text
.omo/evidence/release-readiness-final/index.md reports pass
.omo/evidence/release-readiness-final/summary.json reports public_release_ready=true
```

- [ ] **Step 5: Commit**

```sh
git add scripts/release-readiness.sh scripts/release-preflight.sh docs/INSTALL.md .omo/evidence/release-readiness-final
git commit -m "chore(release): prove public release readiness"
```

---

### Task 3: Complete Provider Proof Matrix

**Files:**
- Modify: `scripts/provider-proof.sh`
- Modify: `scripts/provider-setup-preflight.sh`
- Modify: `docs/INSTALL.md`
- Verify: `.omo/evidence/provider-proof-*`

- [ ] **Step 1: Prove OAuth CLI providers**

Run:

```sh
go run ./cmd/ceo-packet oauth doctor --format text
sh scripts/provider-proof.sh --provider kimi --output-dir .omo/evidence/provider-proof-kimi --timeout-seconds 600
sh scripts/provider-proof.sh --provider codex --output-dir .omo/evidence/provider-proof-codex --timeout-seconds 600
```

Expected:

```text
Kimi OAuth and Codex OAuth pass provider proof
summary.json for each provider reports pass
no secret values appear in evidence
```

- [ ] **Step 2: Add provider proof targets for Claude, OpenCode, and Goose**

Update `scripts/provider-proof.sh` so `--provider claude`, `--provider opencode`, and `--provider goose` use the command wrappers:

```sh
scripts/claude-model-command.sh
scripts/opencode-model-command.sh
scripts/goose-model-command.sh
```

The command shape should match the existing Kimi/Codex command-provider proof flow.

- [ ] **Step 3: Prove the new OAuth wrappers**

Run:

```sh
sh scripts/provider-proof.sh --provider claude --output-dir .omo/evidence/provider-proof-claude --timeout-seconds 600
sh scripts/provider-proof.sh --provider opencode --output-dir .omo/evidence/provider-proof-opencode --timeout-seconds 600
sh scripts/provider-proof.sh --provider goose --output-dir .omo/evidence/provider-proof-goose --timeout-seconds 600
```

Expected:

```text
Each provider writes index.md and summary.json
Each summary reports pass or setup_blocked with complete evidence
No provider token or API key is printed
```

- [ ] **Step 4: Prove HTTP providers**

Run only after each env var is present:

```sh
sh scripts/provider-setup-preflight.sh --providers openai --output-dir .omo/evidence/provider-setup-preflight-openai
sh scripts/provider-proof.sh --provider openai --output-dir .omo/evidence/provider-proof-openai --timeout-seconds 600

sh scripts/provider-setup-preflight.sh --providers openrouter --output-dir .omo/evidence/provider-setup-preflight-openrouter
sh scripts/provider-proof.sh --provider openrouter --output-dir .omo/evidence/provider-proof-openrouter --timeout-seconds 600

sh scripts/provider-setup-preflight.sh --providers moonshot --output-dir .omo/evidence/provider-setup-preflight-moonshot
sh scripts/provider-proof.sh --provider moonshot --output-dir .omo/evidence/provider-proof-moonshot --timeout-seconds 600

sh scripts/provider-setup-preflight.sh --providers kimi-code minimax --output-dir .omo/evidence/provider-setup-preflight-kimi-minimax
sh scripts/provider-proof.sh --provider kimi-code --output-dir .omo/evidence/provider-proof-kimi-code --timeout-seconds 600
sh scripts/provider-proof.sh --provider minimax --output-dir .omo/evidence/provider-proof-minimax --timeout-seconds 600
```

Expected:

```text
ready providers pass
missing providers save blocked setup evidence
blank env vars are reported as empty_env
secret_value_saved=false in every summary
```

- [ ] **Step 5: Commit**

```sh
git add scripts/provider-proof.sh scripts/provider-setup-preflight.sh docs/INSTALL.md .omo/evidence/provider-proof-* .omo/evidence/provider-setup-preflight-*
git commit -m "test(providers): complete production provider proof matrix"
```

---

### Task 4: Add Real Production Workload Evals

**Files:**
- Modify: `scripts/dogfood-real.sh`
- Modify: `internal/eval/local_agent_benchmark_spec.go`
- Modify: `docs/COMPARISON.md`
- Verify: `.omo/evidence/dogfood-real-production-*`

- [ ] **Step 1: Pick five real repo lanes**

Use copied workspaces only. Confirm the source checkout state before running:

```sh
git -C /Users/jakedom/Documents/Codex/2026-06-30/new-chat/work/ceo-harness status --short
```

The production eval set should include these concrete local paths:

```text
ceo-harness self-edit: /Users/jakedom/Documents/Codex/2026-06-30/new-chat/work/ceo-harness
Janus app edit: /Users/jakedom/Documents/janus-code
Radian notes edit: /Users/jakedom/Documents/Radian notes app 
Axis Health edit: /Users/jakedom/Documents/Axis health
ChemCheck edit: /Users/jakedom/Documents/chemcheck-main
```

- [ ] **Step 2: Run copied-workspace write evals**

Run:

```sh
sh scripts/dogfood-real.sh --copy-workspace --write-probe --repeat 3 --repo "ceo-harness:$PWD" --output-dir .omo/evidence/dogfood-real-production-self
```

Expected:

```text
source checkout is not modified
copied workspace receives approved write
rollback evidence exists
index.md reports pass
```

- [ ] **Step 3: Run five-repo app-code eval**

Run one command per repo using copied workspaces:

```sh
sh scripts/dogfood-real.sh --copy-workspace --multi-file-app-code-probe --repo "ceo-harness:/Users/jakedom/Documents/Codex/2026-06-30/new-chat/work/ceo-harness" --output-dir ".omo/evidence/dogfood-real-production-ceo-harness"
sh scripts/dogfood-real.sh --copy-workspace --multi-file-app-code-probe --repo "janus:/Users/jakedom/Documents/janus-code" --output-dir ".omo/evidence/dogfood-real-production-janus"
sh scripts/dogfood-real.sh --copy-workspace --multi-file-app-code-probe --repo "radian:/Users/jakedom/Documents/Radian notes app " --output-dir ".omo/evidence/dogfood-real-production-radian"
sh scripts/dogfood-real.sh --copy-workspace --multi-file-app-code-probe --repo "axis-health:/Users/jakedom/Documents/Axis health" --output-dir ".omo/evidence/dogfood-real-production-axis-health"
sh scripts/dogfood-real.sh --copy-workspace --multi-file-app-code-probe --repo "chemcheck:/Users/jakedom/Documents/chemcheck-main" --output-dir ".omo/evidence/dogfood-real-production-chemcheck"
```

Expected:

```text
each repo produces index.md
each repo reports pass
each source checkout remains clean
```

- [ ] **Step 4: Add a harder benchmark task**

Add a production-grade benchmark task in `internal/eval/local_agent_benchmark_spec.go` with:

```text
multi-file source change
required test command
required evidence artifact
required diff terms
failure mode that needs check-fix
```

- [ ] **Step 5: Verify benchmark scoring**

Run:

```sh
go test ./internal/eval -run Test_LocalAgentBenchmark -count=1
go run ./cmd/ceo-packet gauntlet --suite production-core --agents ceo_harness --output-dir .omo/evidence/production-core-final
```

Expected:

```text
eval tests pass
production-core-final summary reports pass with complete evidence
```

- [ ] **Step 6: Commit**

```sh
git add scripts/dogfood-real.sh internal/eval/local_agent_benchmark_spec.go docs/COMPARISON.md .omo/evidence/dogfood-real-production-* .omo/evidence/production-core-final
git commit -m "test(eval): add production-grade real repo workloads"
```

---

### Task 5: Harden Security And Secret Handling

**Files:**
- Modify: `scripts/secret-scan.sh`
- Modify: `scripts/production-local-gate.sh`
- Modify: `docs/TRUST.md`
- Verify: `.omo/evidence/security-production-final`

- [ ] **Step 1: Extend secret scan patterns**

Update `scripts/secret-scan.sh` to scan:

```text
OPENAI_API_KEY assignments
OPENROUTER_API_KEY assignments
MOONSHOT_API_KEY assignments
KIMI_CODE_API_KEY assignments
MINIMAX_API_KEY assignments
OAuth token file names
Authorization bearer values
provider command files under .omo/evidence
```

- [ ] **Step 2: Add evidence secret scan**

Run:

```sh
sh scripts/secret-scan.sh
rg -n "sk-|Authorization: Bearer|OPENAI_API_KEY=|OPENROUTER_API_KEY=|MOONSHOT_API_KEY=|KIMI_CODE_API_KEY=|MINIMAX_API_KEY=" .omo/evidence || true
```

Expected:

```text
secret-scan ok
manual evidence scan prints no secret assignments
```

- [ ] **Step 3: Document trust boundary**

Add to `docs/TRUST.md`:

```markdown
## OAuth Boundary

CEO Harness does not store OAuth tokens. CLI OAuth providers use local vendor CLIs and whatever login state those tools already manage. Harness config stores only command paths such as `scripts/kimi-model-command.sh`.

## Evidence Boundary

Evidence may store provider names, env var names, command exit codes, and SHA-256 fingerprints. Evidence must not store secret values, bearer tokens, OAuth refresh tokens, or copied key files.
```

- [ ] **Step 4: Verify production gate**

Run:

```sh
sh scripts/production-local-gate.sh --workspace . --output-dir .omo/evidence/security-production-final
```

Expected:

```text
local gate passes
security evidence packet exists
no secret value is saved
```

- [ ] **Step 5: Commit**

```sh
git add scripts/secret-scan.sh scripts/production-local-gate.sh docs/TRUST.md .omo/evidence/security-production-final
git commit -m "chore(security): harden production secret gates"
```

---

### Task 6: Make Install And First Run Boring

**Files:**
- Modify: `README.md`
- Modify: `docs/INSTALL.md`
- Modify: `internal/cli/help.go`
- Modify: `internal/cli/start.go`
- Test: `internal/cli/help_test.go`, `internal/cli/start_test.go`

- [ ] **Step 1: Write failing tests for first-run instructions**

Add test expectations:

```go
// internal/cli/help_test.go
"ceo-packet oauth doctor --format text"
"ceo-packet oauth init kimi --workspace . --format text"
"ceo-packet production-status --workspace . --format text"
```

Expected failure:

```text
help output missing new first-run strings
```

- [ ] **Step 2: Update help and start output**

Make compact help show only the recommended path:

```text
1. ceo-packet oauth doctor --format text
2. ceo-packet oauth init kimi --workspace . --format text
3. ceo-packet run --workspace . --check go test ./... -- "Fix one real task"
4. ceo-packet production-status --workspace . --format text
```

- [ ] **Step 3: Replay install from a clean temp dir**

Run:

```sh
tmp=$(mktemp -d)
GOBIN="$tmp/bin" go install ./cmd/ceo-packet
"$tmp/bin/ceo-packet" --version
"$tmp/bin/ceo-packet" oauth list --format text
"$tmp/bin/ceo-packet" oauth doctor --format text
```

Expected:

```text
version prints
oauth list prints providers
oauth doctor prints ready or missing_cli without panic
```

- [ ] **Step 4: Commit**

```sh
git add README.md docs/INSTALL.md internal/cli/help.go internal/cli/start.go internal/cli/help_test.go internal/cli/start_test.go
git commit -m "docs(onboarding): simplify production first run"
```

---

### Task 7: Finish CI And Release Enforcement

**Files:**
- Modify: `.github/workflows/ci.yml`
- Modify: `.github/workflows/release.yml`
- Modify: `scripts/production-local-gate.sh`
- Modify: `docs/VERIFICATION.md`

- [ ] **Step 1: Ensure CI runs required gates**

CI must run:

```sh
go test ./... -count=1
go vet ./...
sh scripts/secret-scan.sh
sh scripts/smoke.sh
sh scripts/production-local-gate.sh --workspace . --output-dir .omo/evidence/ci-production-local-gate
```

- [ ] **Step 2: Ensure release workflow verifies artifacts**

Release workflow must run:

```sh
sh scripts/release-local.sh --dist dist
sh scripts/verify-release.sh --dist dist
sh scripts/release-readiness.sh --dist dist --output-dir .omo/evidence/release-readiness-ci
```

- [ ] **Step 3: Verify workflow syntax locally**

Run:

```sh
git diff --check
go test ./internal/cli -run 'Test_Run_production|Test_Run_config_completions' -count=1
```

Expected:

```text
diff check exits 0
targeted CLI tests pass
```

- [ ] **Step 4: Commit**

```sh
git add .github/workflows/ci.yml .github/workflows/release.yml scripts/production-local-gate.sh docs/VERIFICATION.md
git commit -m "ci(production): enforce release and local production gates"
```

---

### Task 8: Run Finalizer To Green

**Files:**
- Verify: `.omo/evidence/production-finalize-final`
- Verify: `.omo/evidence/production-readiness-final`

- [ ] **Step 1: Dry-run finalizer**

Run:

```sh
go run ./cmd/ceo-packet production-finalize --workspace . --dry-run --evidence-root .omo/evidence/production-finalize-final
```

Expected:

```text
next-actions.md has only actions that truly require external setup
no declared evidence mismatch
no secret value saved
```

- [ ] **Step 2: Full finalizer with comparison**

Run:

```sh
go run ./cmd/ceo-packet production-finalize --workspace . --run-comparison --evidence-root .omo/evidence/production-finalize-final
```

Expected:

```text
release readiness pass
provider proofs pass or have accepted setup-blocked evidence
competitor smoke pass
comparison pass
final readiness pass
```

- [ ] **Step 3: Final aggregate**

Run:

```sh
sh scripts/production-readiness.sh --dist dist --output-dir .omo/evidence/production-readiness-final
go run ./cmd/ceo-packet production-status --workspace . --format text
go run ./cmd/ceo-packet production-actions --workspace . --format text
```

Expected:

```text
Production status: pass
Local ready: true
Public ready: true
Blocked checks: 0
production-actions has no required blocked actions
```

- [ ] **Step 4: Commit**

```sh
git add .omo/evidence/production-finalize-final .omo/evidence/production-readiness-final docs/PRODUCT_STATUS.md
git commit -m "chore(production): record final readiness evidence"
```

---

### Task 9: Public Release And Install Verification

**Files:**
- Modify: `CHANGELOG.md`
- Modify: `docs/INSTALL.md`
- Verify: public GitHub release assets

- [ ] **Step 1: Prepare release commit**

Run:

```sh
git status --short
go test ./... -count=1
go vet ./...
sh scripts/secret-scan.sh
```

Expected:

```text
working tree clean before tag
all checks pass
```

- [ ] **Step 2: Tag release**

Run:

```sh
git tag -a v0.1.0 -m "CEO Harness v0.1.0"
git push origin main
git push origin v0.1.0
```

Expected:

```text
GitHub release workflow starts
release assets upload
release manifest and checksums are attached
```

- [ ] **Step 3: Verify clean install from release**

Run from a clean temp directory:

```sh
test -n "${GITHUB_REPOSITORY:-}" || { echo "set GITHUB_REPOSITORY, for example owner/repo"; exit 1; }
tmp=$(mktemp -d)
cd "$tmp"
curl -L -o ceo-packet.tar.gz "https://github.com/${GITHUB_REPOSITORY}/releases/download/v0.1.0/ceo-packet_Darwin_arm64.tar.gz"
tar -xzf ceo-packet.tar.gz
./ceo-packet --version
./ceo-packet oauth list --format text
./ceo-packet doctor --format text
```

Expected:

```text
binary runs
version is v0.1.0
oauth list works
doctor works
```

- [ ] **Step 4: Commit post-release docs**

```sh
git add CHANGELOG.md docs/INSTALL.md docs/PRODUCT_STATUS.md
git commit -m "docs(release): record public install verification"
```

---

### Task 10: Final 10/10 Review

**Files:**
- Modify: `docs/PRODUCTION_10_10.md`
- Modify: `docs/PRODUCT_STATUS.md`
- Modify: `README.md`

- [ ] **Step 1: Update scoreboard to final state**

Set:

```markdown
Overall: 10/10
```

Only do this after `production-status` passes.

- [ ] **Step 2: Final command proof**

Run:

```sh
git status --short
go test ./... -count=1
go vet ./...
sh scripts/secret-scan.sh
go run ./cmd/ceo-packet production-status --workspace . --format text
go run ./cmd/ceo-packet production-actions --workspace . --format text
go run ./cmd/ceo-packet oauth doctor --format text
```

Expected:

```text
git status is clean before final docs commit
tests pass
vet passes
secret scan ok
production status pass
OAuth doctor reports providers without token storage
```

- [ ] **Step 3: Write final public wording**

Use this exact claim style in `README.md` and `docs/PRODUCT_STATUS.md`:

```markdown
CEO Harness is production-ready for local-first CLI agentic coding workflows. It has public release artifacts, reproducible install verification, provider proof evidence, real-repo dogfood evidence, CI production gates, and secret-safe evidence handling.
```

- [ ] **Step 4: Commit**

```sh
git add docs/PRODUCTION_10_10.md docs/PRODUCT_STATUS.md README.md
git commit -m "docs(production): mark 10 out of 10 readiness"
```

---

## Final Verification Bundle

Before calling this complete, run:

```sh
go test ./... -count=1
go vet ./...
sh scripts/secret-scan.sh
git diff --check
sh scripts/production-local-gate.sh --workspace . --output-dir .omo/evidence/production-local-gate-final
go run ./cmd/ceo-packet oauth doctor --format text
go run ./cmd/ceo-packet production-status --workspace . --format text
go run ./cmd/ceo-packet production-actions --workspace . --format text
```

Required final output:

```text
go test passes
go vet passes
secret-scan ok
git diff --check exits 0
production local gate passes
OAuth doctor reports no harness token storage
Production status: pass
no required production actions are blocked
```

## Risk Notes

- Do not claim public production until public release evidence is green.
- Do not print or commit API keys while proving OpenAI, OpenRouter, Moonshot, Kimi Code, or MiniMax.
- Do not treat competitor provider quota failures as harness failures; record them as setup-blocked evidence.
- Do not weaken context limits to pass evals; the product wedge is lean context.
- Do not add a GUI before CLI install, provider setup, and finalizer are boring.

## Self-Review

- Spec coverage: release, provider proof, real-repo evals, security, CI, install, docs, and finalizer gates are all covered.
- Placeholder scan: no deferred implementation language is required to execute this plan; each task has concrete files, commands, and expected evidence.
- Type consistency: commands use current CLI names: `oauth`, `production-status`, `production-actions`, `production-finalize`, `provider-proof`, and `provider-setup-preflight`.
