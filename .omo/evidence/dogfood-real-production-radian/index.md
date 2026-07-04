# Real Repo Dogfood Evidence

- Generated: 2026-07-04T16:10:10Z
- Mode: live
- Repeat count: 1
- Workspace mode: copied
- Task: Plan a bounded real-repo fix without writing files
- Write probe: disabled
- Feature edit probe: disabled
- App-code probe: disabled
- Integrated app-code probe: disabled
- Multi-file app-code probe: enabled
- Runner: scripts/dogfood-real.sh
- Evidence root: .omo/evidence/dogfood-real-production-radian
- Secret API keys: not required for smoke path
- Real-provider path: skipped by default; this runner uses local command/dry-run surfaces unless a repo config routes providers itself

## Scenario Catalog

| Scenario | Purpose | Dry-run behavior | Live evidence |
| --- | --- | --- | --- |
| scenario-01-doctor | Build and run the no-key doctor smoke | listed only | command output, report hash |
| scenario-02-plan-only | Preview a bounded real-repo task packet | listed only | plan report, route metadata |
| scenario-03-observe-run | Run CEO Harness with a local deterministic model in observe mode | listed only | JSON report, pass/fail note |
| scenario-04-patch-preview | Capture a patch approval digest on a controlled fixture | listed only | preview report and digest |
| scenario-05-timeout-guard | Prove hung model commands do not look successful | listed only | expected-failure transcript |
| scenario-06-write-probe | Prove preview plus approved write mutates only the copied workspace and can roll back | listed only | preview digest, apply report, rollback report, after-rollback git status |
| scenario-07-feature-edit-probe | Prove a repo-specific feature note can be previewed and approved in a copied workspace | listed only | feature file, preview digest, after-state git status |
| scenario-08-app-code-probe | Prove a source-code module can be previewed and approved in a copied workspace | listed only | app-code file, preview digest, after-state git status |
| scenario-09-integrated-app-code-probe | Prove an existing source file can be previewed and approved in a copied workspace | listed only | target path, modified source file, preview digest, after-state git status |
| scenario-10-multi-file-app-code-probe | Prove two existing source files can be previewed and approved in a copied workspace | listed only | target paths, modified source files, preview digests, after-state git status |

## Repo Results

| Repo | Status | Path | Notes |
| --- | --- | --- | --- |
| radian | pass | `/Users/jakedom/Documents/cod code/.omo/evidence/dogfood-real-production-radian/repos/radian/workspace-copy` | see repos/radian/summary.md |

## Adversarial Coverage

- stale_state: live mode captures git HEAD and git status hashes before repo scenarios; dry-run records this as planned only.
- misleading_success_output: missing repos are recorded as skipped_missing_repo, and timeout probes must exit non-zero to pass.
- dirty_worktree: live mode saves git-status.txt and git-status.sha256 for review; dirty status is evidence, not an automatic pass.
- hung/long commands: live mode runs scenario-05-timeout-guard with --model-command-timeout-ms 250.
