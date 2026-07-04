#!/bin/sh
set -eu

root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
dist=${DIST:-"$root/dist"}
output_dir=${OUTPUT_DIR:-"$root/.omo/evidence/release-bootstrap"}
repo_url=${REPO_URL:-}
release_url=${RELEASE_URL:-${PUBLIC_RELEASE_URL:-}}
homebrew_archive_base_url=${HOMEBREW_ARCHIVE_BASE_URL:-}
checksum_notes_url=${CHECKSUM_ONLY_RELEASE_NOTES_URL:-}
signing_mode=${SIGNING_MODE:-checksum-only}
signing_identity=${SIGNING_IDENTITY:-}
version=${VERSION:-}

usage() {
  cat <<'USAGE'
Usage: sh scripts/release-bootstrap.sh [options]

Prepares a public-release evidence packet without tagging, pushing, uploading,
publishing a tap, or creating a remote release.

Options:
  --dist DIR                         Local release dist directory.
  --output-dir DIR                   Evidence output directory.
  --repo-url URL                     Public repo HTTPS URL.
  --release-url URL                  Public release HTTPS URL.
  --homebrew-archive-base-url URL    Public HTTPS directory containing archives.
  --checksum-notes-url URL           Public checksum-only verification notes URL.
  --signing-mode MODE                checksum-only, cosign, or gpg.
  --signing-identity TEXT            Required for cosign or gpg mode.
  --version VERSION                  Release version. Defaults to manifest version.
USAGE
}

while [ "$#" -gt 0 ]; do
  case "$1" in
    --dist)
      [ "$#" -ge 2 ] || { printf '%s\n' "release-bootstrap: --dist requires a value" >&2; exit 2; }
      dist="$2"
      shift 2
      ;;
    --output-dir)
      [ "$#" -ge 2 ] || { printf '%s\n' "release-bootstrap: --output-dir requires a value" >&2; exit 2; }
      output_dir="$2"
      shift 2
      ;;
    --repo-url)
      [ "$#" -ge 2 ] || { printf '%s\n' "release-bootstrap: --repo-url requires a value" >&2; exit 2; }
      repo_url="$2"
      shift 2
      ;;
    --release-url)
      [ "$#" -ge 2 ] || { printf '%s\n' "release-bootstrap: --release-url requires a value" >&2; exit 2; }
      release_url="$2"
      shift 2
      ;;
    --homebrew-archive-base-url)
      [ "$#" -ge 2 ] || { printf '%s\n' "release-bootstrap: --homebrew-archive-base-url requires a value" >&2; exit 2; }
      homebrew_archive_base_url="$2"
      shift 2
      ;;
    --checksum-notes-url)
      [ "$#" -ge 2 ] || { printf '%s\n' "release-bootstrap: --checksum-notes-url requires a value" >&2; exit 2; }
      checksum_notes_url="$2"
      shift 2
      ;;
    --signing-mode)
      [ "$#" -ge 2 ] || { printf '%s\n' "release-bootstrap: --signing-mode requires a value" >&2; exit 2; }
      signing_mode="$2"
      shift 2
      ;;
    --signing-identity)
      [ "$#" -ge 2 ] || { printf '%s\n' "release-bootstrap: --signing-identity requires a value" >&2; exit 2; }
      signing_identity="$2"
      shift 2
      ;;
    --version)
      [ "$#" -ge 2 ] || { printf '%s\n' "release-bootstrap: --version requires a value" >&2; exit 2; }
      version="$2"
      shift 2
      ;;
    --help|-h)
      usage
      exit 0
      ;;
    *)
      printf '%s\n' "release-bootstrap: unknown argument: $1" >&2
      usage >&2
      exit 2
      ;;
  esac
done

case "$dist" in
  /*) ;;
  *) dist="$(pwd)/$dist" ;;
esac

case "$output_dir" in
  /*) ;;
  *) output_dir="$(pwd)/$output_dir" ;;
esac

mkdir -p "$output_dir"

json_escape() {
  printf '%s' "$1" | sed 's/\\/\\\\/g; s/"/\\"/g'
}

is_public_https_url() {
  candidate="$1"
  case "$candidate" in
    https://*example.invalid*|https://example.com*|https://example.org*|"")
      return 1
      ;;
    https://*)
      return 0
      ;;
    *)
      return 1
      ;;
  esac
}

slash_trim() {
  printf '%s' "$1" | sed 's:/*$::'
}

if [ -z "$version" ] && [ -f "$dist/release-manifest.json" ]; then
  version=$(python3 - "$dist/release-manifest.json" <<'PY'
import json
import sys
with open(sys.argv[1], "r", encoding="utf-8") as handle:
    print(json.load(handle).get("version", ""))
PY
)
fi

if [ -z "$version" ]; then
  version="unknown"
fi

local_artifacts_status=blocked
if sh "$root/scripts/verify-release.sh" "$dist" >"$output_dir/verify-release.txt" 2>&1; then
  local_artifacts_status=pass
fi

blocked_file="$output_dir/blocked-checks.txt"
: >"$blocked_file"

check_status() {
  name="$1"
  status="$2"
  if [ "$status" = "blocked" ]; then
    printf '%s\n' "$name" >>"$blocked_file"
  fi
}

check_status local_release_artifacts "$local_artifacts_status"
if is_public_https_url "$repo_url"; then repo_status=pass; else repo_status=blocked; fi
check_status public_repo_url "$repo_status"
if is_public_https_url "$release_url"; then release_status=pass; else release_status=blocked; fi
check_status public_release_url "$release_status"
if is_public_https_url "$homebrew_archive_base_url"; then homebrew_status=pass; else homebrew_status=blocked; fi
check_status homebrew_archive_base_url "$homebrew_status"

case "$signing_mode" in
  checksum-only)
    if is_public_https_url "$checksum_notes_url"; then signing_status=pass; else signing_status=blocked; fi
    ;;
  cosign|gpg)
    if [ -n "$signing_identity" ]; then signing_status=pass; else signing_status=blocked; fi
    ;;
  *)
    signing_status=blocked
    ;;
esac
check_status signing_or_checksum_policy "$signing_status"

archive_name="ceo-packet_${version}_darwin_arm64.tar.gz"
archive_sha=""
if [ -f "$dist/checksums.txt" ]; then
  archive_sha=$(awk -v archive="$archive_name" '$2 == archive {print $1}' "$dist/checksums.txt")
fi

if [ "$homebrew_status" = "pass" ] && [ -n "$archive_sha" ]; then
  archive_base=$(slash_trim "$homebrew_archive_base_url")
  cat >"$output_dir/remote-homebrew-formula.rb" <<EOF
class CeoPacket < Formula
  desc "Local CEO/subagent coding harness"
  homepage "$repo_url"
  url "$archive_base/$archive_name"
  sha256 "$archive_sha"
  version "$version"

  def install
    bin.install "ceo-packet"
  end

  test do
    assert_match "ceo-packet $version", shell_output("#{bin}/ceo-packet --version")
  end
end
EOF
else
  cat >"$output_dir/remote-homebrew-formula.rb" <<'EOF'
# Blocked: provide --homebrew-archive-base-url and verified local checksums first.
EOF
fi

cat >"$output_dir/env.template" <<EOF
REPO_URL=$repo_url
RELEASE_URL=$release_url
HOMEBREW_ARCHIVE_BASE_URL=$homebrew_archive_base_url
SIGNING_MODE=$signing_mode
SIGNING_IDENTITY=$signing_identity
CHECKSUM_ONLY_RELEASE_NOTES_URL=$checksum_notes_url
VERSION=$version
EOF

cat >"$output_dir/commands.sh" <<EOF
#!/bin/sh
set -eu

# This file is a release operator checklist. Review before running.
VERSION="$version" sh scripts/release-local.sh
sh scripts/verify-release.sh dist

# Upload dist/*.tar.gz, dist/checksums.txt, and dist/release-manifest.json to:
# $release_url

# After remote artifacts exist, update the Homebrew tap formula from:
# $output_dir/remote-homebrew-formula.rb

ALLOW_CHECKSUM_ONLY_RELEASE=$(if [ "$signing_mode" = "checksum-only" ]; then printf 1; else printf 0; fi) \\
CHECKSUM_ONLY_RELEASE_NOTES_URL="$checksum_notes_url" \\
RELEASE_URL="$release_url" \\
sh scripts/release-preflight.sh dist

sh scripts/release-readiness.sh --dist dist --output-dir .omo/evidence/release-readiness
sh scripts/production-readiness.sh --dist dist --output-dir .omo/evidence/production-readiness
EOF
chmod +x "$output_dir/commands.sh"

cat >"$output_dir/release-checklist.md" <<EOF
# Public Release Bootstrap Checklist

Status inputs are captured in \`summary.json\`. This packet does not publish anything.

1. Confirm the public repo URL: \`$repo_url\`
2. Build local artifacts: \`VERSION=$version sh scripts/release-local.sh\`
3. Verify local artifacts: \`sh scripts/verify-release.sh dist\`
4. Upload archives, checksums, and release manifest to: \`$release_url\`
5. Use \`remote-homebrew-formula.rb\` as the tap formula after remote artifacts exist.
6. Document verification policy: \`$signing_mode\`
7. Re-run release and production readiness gates.

Publish boundary: do not tag, push, upload, publish a tap, or announce installer support until the readiness gates pass against public URLs.
EOF

cat >"$output_dir/release-handoff.md" <<EOF
# Public Release Handoff

This handoff is safe to share with a release operator. It does not publish anything.

## Inputs

- Version: \`$version\`
- Repo URL: \`$repo_url\`
- Release URL: \`$release_url\`
- Archive base URL: \`$homebrew_archive_base_url\`
- Signing mode: \`$signing_mode\`
- Checksum notes URL: \`$checksum_notes_url\`

## Required Public Assets

- \`dist/checksums.txt\`
- \`dist/release-manifest.json\`
EOF

if [ -f "$dist/release-manifest.json" ]; then
  python3 - "$dist/release-manifest.json" >>"$output_dir/release-handoff.md" <<'PY'
import json
import sys

with open(sys.argv[1], "r", encoding="utf-8") as handle:
    manifest = json.load(handle)

for artifact in manifest.get("artifacts", []):
    name = artifact.get("name", "")
    sha = artifact.get("sha256", "")
    if name:
        print(f"- `dist/{name}` sha256 `{sha}`")
PY
else
  printf '%s\n' "- Blocked: release manifest is missing." >>"$output_dir/release-handoff.md"
fi

cat >>"$output_dir/release-handoff.md" <<EOF

## Operator Boundary

- Do not paste secrets into repo files or evidence folders.
- Do not claim public production readiness until \`release-readiness\`, provider proof, and \`production-readiness\` pass.
- Publishing is intentionally manual from this handoff; this script did not tag, push, upload, create a GitHub release, or publish a tap.

## Verification Commands After Publishing

\`\`\`sh
ALLOW_CHECKSUM_ONLY_RELEASE=$(if [ "$signing_mode" = "checksum-only" ]; then printf 1; else printf 0; fi) \\
CHECKSUM_ONLY_RELEASE_NOTES_URL="$checksum_notes_url" \\
RELEASE_URL="$release_url" \\
sh scripts/release-preflight.sh dist
sh scripts/release-readiness.sh --dist dist --output-dir .omo/evidence/release-readiness-final
go run ./cmd/ceo-packet production-finalize --workspace . --dry-run
\`\`\`
EOF

blocked_count=$(wc -l <"$blocked_file" | tr -d ' ')
if [ "$blocked_count" -eq 0 ]; then
  overall_status=pass
  release_bootstrap_ready=true
else
  overall_status=blocked
  release_bootstrap_ready=false
fi

blocked_json=$(awk '
  BEGIN { first=1; printf "[" }
  {
    gsub(/\\/,"\\\\")
    gsub(/"/,"\\\"")
    if (!first) {
      printf ", "
    }
    printf "\"%s\"", $0
    first=0
  }
  END { printf "]" }
' "$blocked_file")

release_checklist_item_count=$(awk '/^[0-9][0-9]*[.] / { count++ } END { print count + 0 }' "$output_dir/release-checklist.md")
bootstrap_artifacts_sha256=$(python3 - "$output_dir" <<'PY'
import hashlib
import json
import pathlib
import sys

output_dir = pathlib.Path(sys.argv[1])
artifacts = [
    "blocked-checks.txt",
    "commands.sh",
    "env.template",
    "release-handoff.md",
    "release-checklist.md",
    "remote-homebrew-formula.rb",
    "verify-release.txt",
]

digests = {}
for artifact in artifacts:
    path = output_dir / artifact
    if path.exists():
        digests[artifact] = hashlib.sha256(path.read_bytes()).hexdigest()

print(json.dumps(digests, sort_keys=True))
PY
)

cat >"$output_dir/summary.json" <<JSON
{
  "schema_version": 1,
  "status": "$overall_status",
  "release_bootstrap_ready": $release_bootstrap_ready,
  "version": "$(json_escape "$version")",
  "dist": "$(json_escape "$dist")",
  "repo_url": "$(json_escape "$repo_url")",
  "release_url": "$(json_escape "$release_url")",
  "homebrew_archive_base_url": "$(json_escape "$homebrew_archive_base_url")",
  "signing_mode": "$(json_escape "$signing_mode")",
  "local_release_artifacts": "$local_artifacts_status",
  "public_repo_url": "$repo_status",
  "public_release_url": "$release_status",
  "homebrew_archive_base_url_status": "$homebrew_status",
  "signing_or_checksum_policy": "$signing_status",
  "blocked_count": $blocked_count,
  "blocked_checks": $blocked_json,
  "release_checklist_item_count": $release_checklist_item_count,
  "bootstrap_artifacts_sha256": $bootstrap_artifacts_sha256,
  "artifacts": {
    "index": "index.md",
    "summary": "summary.json",
    "commands": "commands.sh",
    "env_template": "env.template",
    "handoff": "release-handoff.md",
    "checklist": "release-checklist.md",
    "homebrew_formula": "remote-homebrew-formula.rb",
    "verify_release": "verify-release.txt"
  }
}
JSON

{
  printf '%s\n' "# Release Bootstrap Evidence"
  printf '\n'
  printf '%s\n' "Status: release bootstrap: $overall_status"
  printf '\n'
  printf '| Check | Status | Evidence |\n'
  printf '| --- | --- | --- |\n'
  printf '| local_release_artifacts | %s | `verify-release.txt` |\n' "$local_artifacts_status"
  printf '| public_repo_url | %s | `env.template` |\n' "$repo_status"
  printf '| public_release_url | %s | `env.template` |\n' "$release_status"
  printf '| homebrew_archive_base_url | %s | `remote-homebrew-formula.rb` |\n' "$homebrew_status"
  printf '| signing_or_checksum_policy | %s | `release-checklist.md` |\n' "$signing_status"
  printf '\n'
  if [ "$blocked_count" -gt 0 ]; then
    printf '%s\n' "## Blocked Checks"
    printf '\n'
    while IFS= read -r check; do
      [ -n "$check" ] || continue
      printf -- '- `%s`\n' "$check"
    done <"$blocked_file"
    printf '\n'
  fi
  printf '%s\n' "## Artifacts"
  printf '\n'
  printf '%s\n' '- `commands.sh`'
  printf '%s\n' '- `env.template`'
  printf '%s\n' '- `release-handoff.md`'
  printf '%s\n' '- `release-checklist.md`'
  printf '%s\n' '- `remote-homebrew-formula.rb`'
  printf '\n'
  printf '%s\n' "No publishing action was performed."
} >"$output_dir/index.md"

printf '%s\n' "release-bootstrap: wrote $output_dir/index.md"
printf '%s\n' "release-bootstrap: $overall_status"

if [ "$overall_status" = "pass" ]; then
  exit 0
fi
exit 1
