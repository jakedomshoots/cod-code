#!/bin/sh
set -eu

sh scripts/provider-setup-preflight.sh --output-dir .omo/evidence/provider-setup-preflight
# blocked command: sh scripts/provider-proof.sh --provider moonshot
# reason: MOONSHOT_API_KEY is missing or empty; export it before running provider proof.
# sh scripts/provider-proof.sh --provider moonshot --output-dir .omo/evidence/provider-proof-moonshot --timeout-seconds 600
