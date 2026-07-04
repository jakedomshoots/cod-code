#!/bin/sh
set -eu

sh scripts/provider-setup-preflight.sh --output-dir .omo/evidence/provider-setup-preflight
# blocked command: sh scripts/provider-proof.sh --provider kimi-code
# reason: KIMI_CODE_API_KEY is missing or empty; export it before running provider proof.
# sh scripts/provider-proof.sh --provider kimi-code --output-dir .omo/evidence/provider-proof-kimi-code --timeout-seconds 600
# blocked command: sh scripts/provider-proof.sh --provider minimax
# reason: MINIMAX_API_KEY is missing or empty; export it before running provider proof.
# sh scripts/provider-proof.sh --provider minimax --output-dir .omo/evidence/provider-proof-minimax --timeout-seconds 600
