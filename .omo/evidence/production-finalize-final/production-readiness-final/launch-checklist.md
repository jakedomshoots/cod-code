# Launch Checklist

Local production ready: true
Public production ready: true

## Required Before Public Production Claim

- No public-production blockers remain.

## Final Gate

After every item above is complete, run:

```sh
sh scripts/production-readiness.sh --dist dist --output-dir .omo/evidence/production-readiness
```
