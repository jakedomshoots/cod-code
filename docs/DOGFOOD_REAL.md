# Real Repo Dogfood

`scripts/dogfood-real.sh` writes durable evidence for selected local repos without requiring API keys.

## Smoke Path

```sh
sh scripts/dogfood-real.sh --dry-run
```

Dry-run writes `.omo/evidence/dogfood-real/index.md`, lists five scenarios, and does not run commands against external repos.

## Selected Repos

```sh
sh scripts/dogfood-real.sh --repo "ceo-harness:/path/to/repo"
```

Live mode builds a local `ceo-packet` binary, records command transcripts, hashes report output, captures git state, uses the local example model for no-key runs, and captures a patch preview digest on a controlled fixture.

Missing repos are recorded as `skipped_missing_repo`, not pass.

## Evidence

Evidence is saved under `.omo/evidence/dogfood-real/`:

- `index.md`: run summary, scenario catalog, repo status, adversarial notes.
- `repos/<name>/summary.md`: per-repo pass/fail notes.
- `repos/<name>/scenario-*/command.argv`: exact argv.
- `repos/<name>/scenario-*/stdout.txt`: report output.
- `repos/<name>/scenario-*/stdout.sha256`: report digest.
- `repos/<name>/scenario-04-patch-preview/preview-digest.txt`: patch approval digest.
- `_archive/<run>/`: previous run evidence preserved before the latest run is written.

Real-provider runs are intentionally skipped by default. The smoke path stays local and keyless.
