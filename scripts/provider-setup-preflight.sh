#!/bin/sh
set -eu

root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
output_dir=${OUTPUT_DIR:-"$root/.omo/evidence/provider-setup-preflight"}
providers=${PROVIDERS:-"openai openrouter moonshot"}

usage() {
  cat <<'USAGE'
Usage: sh scripts/provider-setup-preflight.sh [options]

Checks whether paid HTTP provider environment variables are present and
non-empty without printing or saving secret values.

Options:
  --output-dir DIR       Evidence output directory.
  --providers LIST       Space or comma separated providers. Default: openai openrouter moonshot.
  --help                 Show this help.
USAGE
}

while [ "$#" -gt 0 ]; do
  case "$1" in
    --output-dir)
      [ "$#" -ge 2 ] || { printf '%s\n' "provider-setup-preflight: --output-dir requires a value" >&2; exit 2; }
      output_dir="$2"
      shift 2
      ;;
    --providers)
      [ "$#" -ge 2 ] || { printf '%s\n' "provider-setup-preflight: --providers requires a value" >&2; exit 2; }
      providers="$2"
      shift 2
      ;;
    --help|-h)
      usage
      exit 0
      ;;
    *)
      printf '%s\n' "provider-setup-preflight: unknown argument: $1" >&2
      usage >&2
      exit 2
      ;;
  esac
done

case "$output_dir" in
  /*) ;;
  *) output_dir="$(pwd)/$output_dir" ;;
esac

mkdir -p "$output_dir"
providers_normalized=$(printf '%s' "$providers" | tr ',' ' ')
provider_list_file="$output_dir/providers.txt"
: >"$provider_list_file"

for provider in $providers_normalized; do
  case "$provider" in
    openai|openrouter|moonshot)
      printf '%s\n' "$provider" >>"$provider_list_file"
      ;;
    "")
      ;;
    *)
      printf '%s\n' "provider-setup-preflight: unsupported provider: $provider" >&2
      exit 2
      ;;
  esac
done

if [ ! -s "$provider_list_file" ]; then
  printf '%s\n' "provider-setup-preflight: no providers selected" >&2
  exit 2
fi

provider_env() {
  case "$1" in
    openai) printf '%s\n' "OPENAI_API_KEY" ;;
    openrouter) printf '%s\n' "OPENROUTER_API_KEY" ;;
    moonshot) printf '%s\n' "MOONSHOT_API_KEY" ;;
  esac
}

provider_model() {
  case "$1" in
    openai) printf '%s\n' "${CEO_PROVIDER_PROOF_OPENAI_MODEL:-gpt-5}" ;;
    openrouter) printf '%s\n' "${CEO_PROVIDER_PROOF_OPENROUTER_MODEL:-openai/gpt-5-mini}" ;;
    moonshot) printf '%s\n' "${CEO_PROVIDER_PROOF_MOONSHOT_MODEL:-moonshot-v1-128k}" ;;
  esac
}

json_escape() {
  printf '%s' "$1" | sed 's/\\/\\\\/g; s/"/\\"/g'
}

results_file="$output_dir/results.jsonl"
blocked_file="$output_dir/blocked-providers.txt"
: >"$results_file"
: >"$blocked_file"

ready_count=0
blocked_count=0
provider_count=0

while IFS= read -r provider; do
  [ -n "$provider" ] || continue
  provider_count=$((provider_count + 1))
  env_name=$(provider_env "$provider")
  model=$(provider_model "$provider")
  eval "env_present=\${$env_name+x}"
  eval "env_value=\${$env_name:-}"
  if [ -z "$env_value" ]; then
    blocked_count=$((blocked_count + 1))
    printf '%s\n' "$provider" >>"$blocked_file"
    if [ "$env_present" = "x" ]; then
      status="empty_env"
      reason="required env is set but empty"
    else
      status="missing_env"
      reason="required env is not set"
    fi
  else
    ready_count=$((ready_count + 1))
    status="ready"
    reason="required env is present and non-empty"
  fi
  printf '{"provider":"%s","api_key_env":"%s","model":"%s","status":"%s","reason":"%s","secret_value_saved":false}\n' \
    "$(json_escape "$provider")" \
    "$(json_escape "$env_name")" \
    "$(json_escape "$model")" \
    "$(json_escape "$status")" \
    "$(json_escape "$reason")" >>"$results_file"
done <"$provider_list_file"

if [ "$blocked_count" -eq 0 ]; then
  overall_status=pass
else
  overall_status=blocked
fi

results_json=$(python3 - "$results_file" <<'PY'
import json
import sys

rows = []
with open(sys.argv[1], "r", encoding="utf-8") as handle:
    for line in handle:
        line = line.strip()
        if line:
            rows.append(json.loads(line))
print(json.dumps(rows, indent=2))
PY
)

blocked_json=$(python3 - "$blocked_file" <<'PY'
import json
import sys

with open(sys.argv[1], "r", encoding="utf-8") as handle:
    print(json.dumps([line.strip() for line in handle if line.strip()]))
PY
)

cat >"$output_dir/commands.sh" <<'EOF'
#!/bin/sh
set -eu

sh scripts/provider-setup-preflight.sh --output-dir .omo/evidence/provider-setup-preflight
sh scripts/provider-proof.sh --provider openai --output-dir .omo/evidence/provider-proof-openai --timeout-seconds 600
sh scripts/provider-proof.sh --provider openrouter --output-dir .omo/evidence/provider-proof-openrouter --timeout-seconds 600
sh scripts/provider-proof.sh --provider moonshot --output-dir .omo/evidence/provider-proof-moonshot --timeout-seconds 600
EOF
chmod +x "$output_dir/commands.sh"

commands_sha256=$(python3 - "$output_dir/commands.sh" <<'PY'
import hashlib
import sys

with open(sys.argv[1], "rb") as handle:
    print(hashlib.sha256(handle.read()).hexdigest())
PY
)

cat >"$output_dir/summary.json" <<JSON
{
  "schema_version": 1,
  "status": "$overall_status",
  "provider_count": $provider_count,
  "ready_count": $ready_count,
  "blocked_count": $blocked_count,
  "blocked_providers": $blocked_json,
  "command_script_secret_policy": "no_secret_assignment",
  "secret_value_saved": false,
  "commands_sha256": "$commands_sha256",
  "results": $results_json,
  "artifacts": {
    "index": "index.md",
    "summary": "summary.json",
    "commands": "commands.sh",
    "blocked_providers": "blocked-providers.txt"
  }
}
JSON

{
  printf '%s\n' "# Provider Setup Preflight"
  printf '\n'
  printf '%s\n' "Status: $overall_status"
  printf '\n'
  printf '%s\n' '| Provider | Status | Env | Model |'
  printf '%s\n' '| --- | --- | --- | --- |'
  python3 - "$results_file" <<'PY'
import json
import sys

with open(sys.argv[1], "r", encoding="utf-8") as handle:
    for line in handle:
        row = json.loads(line)
        print(f"| {row['provider']} | {row['status']} | `{row['api_key_env']}` | `{row['model']}` |")
PY
  printf '\n'
  printf '%s\n' "Secret values were not printed or saved."
} >"$output_dir/index.md"

printf '%s\n' "provider-setup-preflight: wrote $output_dir/index.md"
printf '%s\n' "provider-setup-preflight: $overall_status"

if [ "$overall_status" = "pass" ]; then
  exit 0
fi
exit 1
