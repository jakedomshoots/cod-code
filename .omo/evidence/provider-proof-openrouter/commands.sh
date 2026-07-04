#!/bin/sh
set -eu

# Export OPENROUTER_API_KEY in your shell or local secret manager before running.
# Do not paste secret values into this file or any evidence artifact.
if [ -z "${OPENROUTER_API_KEY+x}" ]; then
  printf '%s\n' 'provider setup: OPENROUTER_API_KEY is not set' >&2
  exit 2
fi
if [ -z "${OPENROUTER_API_KEY}" ]; then
  printf '%s\n' 'provider setup: OPENROUTER_API_KEY is empty' >&2
  exit 2
fi
sh scripts/provider-setup-preflight.sh --providers openrouter --output-dir .omo/evidence/provider-setup-preflight-openrouter
sh scripts/provider-proof.sh --provider openrouter --output-dir .omo/evidence/provider-proof-openrouter --timeout-seconds 600
sh scripts/production-readiness.sh --dist dist --output-dir .omo/evidence/production-readiness
