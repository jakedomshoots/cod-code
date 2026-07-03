# Security Policy

## Supported Versions

CEO Harness is pre-1.0. Security fixes land on the active `main` line until the first tagged release policy exists.

## Reporting

For now, report security issues privately to the repository owner. Do not open public issues for exploitable bugs until a private contact is published.

## Local Trust Model

CEO Harness is local-first. It can read and modify the workspace you point it at, and provider-backed runs may send prompts or selected context to the configured provider.

Before running on sensitive code:

- Review the workspace path.
- Use `--plan-only` or `--dry-run` first.
- Check provider config with `ceo-packet config doctor --workspace <repo>`.
- Keep secrets out of prompts and committed files.

## Release Integrity

Current local releases are checksum-only. Verify archives with:

```sh
cd dist
shasum -a 256 -c checksums.txt
```

Signed releases are planned but not yet claimed.
