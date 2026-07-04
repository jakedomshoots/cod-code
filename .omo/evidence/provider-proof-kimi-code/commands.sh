#!/bin/sh
set -eu

# Export KIMI_CODE_API_KEY in your shell or local secret manager before running.
# Do not paste secret values into this file or any evidence artifact.
if [ -z "${KIMI_CODE_API_KEY+x}" ]; then
  printf '%s\n' 'provider setup: KIMI_CODE_API_KEY is not set' >&2
  exit 2
fi
if [ -z "${KIMI_CODE_API_KEY}" ]; then
  printf '%s\n' 'provider setup: KIMI_CODE_API_KEY is empty' >&2
  exit 2
fi
sh scripts/provider-setup-preflight.sh --providers kimi-code --output-dir .omo/evidence/provider-setup-preflight-kimi-code
sh scripts/provider-proof.sh --provider kimi-code --output-dir .omo/evidence/provider-proof-kimi-code --timeout-seconds 600
sh scripts/production-readiness.sh --dist dist --output-dir .omo/evidence/production-readiness
