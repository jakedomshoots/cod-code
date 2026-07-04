package cli

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func Test_ReleaseLocalScript_writesManifestAndVerifyPasses(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("resolve repo root: %v", err)
	}
	dist := filepath.Join(t.TempDir(), "dist")

	release := exec.Command("sh", filepath.Join(root, "scripts", "release-local.sh"))
	release.Dir = root
	release.Env = append(release.Environ(), "DIST="+dist, "VERSION=0.2.0-test", "COMMIT=abc123")
	output, err := release.CombinedOutput()
	if err != nil {
		t.Fatalf("release-local failed: %v\n%s", err, string(output))
	}

	manifestPath := filepath.Join(dist, "release-manifest.json")
	content, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatalf("read release manifest: %v", err)
	}
	var manifest struct {
		Version   string `json:"version"`
		Commit    string `json:"commit"`
		Artifacts []struct {
			Name      string `json:"name"`
			SHA256    string `json:"sha256"`
			SizeBytes int64  `json:"size_bytes"`
		} `json:"artifacts"`
	}
	if err := json.Unmarshal(content, &manifest); err != nil {
		t.Fatalf("decode release manifest: %v", err)
	}
	if manifest.Version != "0.2.0-test" || manifest.Commit != "abc123" || len(manifest.Artifacts) != 3 {
		t.Fatalf("manifest = %#v, want version/commit and three archives", manifest)
	}
	for _, artifact := range manifest.Artifacts {
		if artifact.Name == "" || artifact.SHA256 == "" || artifact.SizeBytes <= 0 {
			t.Fatalf("manifest artifact = %#v, want complete name, sha256, size", artifact)
		}
	}

	verify := exec.Command("sh", filepath.Join(root, "scripts", "verify-release.sh"), dist)
	verify.Dir = root
	verifyOutput, err := verify.CombinedOutput()
	if err != nil {
		t.Fatalf("verify-release failed: %v\n%s", err, string(verifyOutput))
	}
}

func Test_ReleasePreflight_blocksPublicReleaseWhenRemoteProofIsMissing(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("resolve repo root: %v", err)
	}
	dist := filepath.Join(t.TempDir(), "dist")

	release := exec.Command("sh", filepath.Join(root, "scripts", "release-local.sh"))
	release.Dir = root
	release.Env = append(release.Environ(), "DIST="+dist, "VERSION=0.2.0-test", "COMMIT=abc123")
	output, err := release.CombinedOutput()
	if err != nil {
		t.Fatalf("release-local failed: %v\n%s", err, string(output))
	}

	preflight := exec.Command("sh", filepath.Join(root, "scripts", "release-preflight.sh"), dist)
	preflight.Dir = root
	preflightOutput, err := preflight.CombinedOutput()
	if err == nil {
		t.Fatalf("release-preflight unexpectedly passed:\n%s", string(preflightOutput))
	}
	body := string(preflightOutput)
	for _, want := range []string{
		"public release preflight: blocked",
		"remote_release_url",
		"artifact_signatures",
		"homebrew_formula_url",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("preflight output missing %q:\n%s", want, body)
		}
	}
}

func Test_ReleasePreflight_acceptsExplicitChecksumOnlyReleaseNotes(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("resolve repo root: %v", err)
	}
	dist := filepath.Join(t.TempDir(), "dist")

	release := exec.Command("sh", filepath.Join(root, "scripts", "release-local.sh"))
	release.Dir = root
	release.Env = append(release.Environ(), "DIST="+dist, "VERSION=0.2.0-test", "COMMIT=abc123")
	output, err := release.CombinedOutput()
	if err != nil {
		t.Fatalf("release-local failed: %v\n%s", err, string(output))
	}

	preflight := exec.Command("sh", filepath.Join(root, "scripts", "release-preflight.sh"), dist)
	preflight.Dir = root
	preflight.Env = append(
		preflight.Environ(),
		"ALLOW_CHECKSUM_ONLY_RELEASE=1",
		"CHECKSUM_ONLY_RELEASE_NOTES_URL=https://releases.ceo-harness.dev/v0.2.0",
	)
	preflightOutput, err := preflight.CombinedOutput()
	if err == nil {
		t.Fatalf("release-preflight unexpectedly passed while other public blockers remain:\n%s", string(preflightOutput))
	}
	body := string(preflightOutput)
	if !strings.Contains(body, "| artifact_signatures | pass | checksum-only release explicitly documented at https://releases.ceo-harness.dev/v0.2.0 |") {
		t.Fatalf("preflight did not accept explicit checksum-only release notes:\n%s", body)
	}
}

func Test_ReleasePreflight_verifiesGitHubReleaseAssets(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("resolve repo root: %v", err)
	}
	tmp := t.TempDir()
	dist := filepath.Join(tmp, "dist")

	release := exec.Command("sh", filepath.Join(root, "scripts", "release-local.sh"))
	release.Dir = root
	release.Env = append(release.Environ(), "DIST="+dist, "VERSION=0.2.0-test", "COMMIT=abc123")
	output, err := release.CombinedOutput()
	if err != nil {
		t.Fatalf("release-local failed: %v\n%s", err, string(output))
	}

	fakeBin := filepath.Join(tmp, "bin")
	if err := os.MkdirAll(fakeBin, 0o755); err != nil {
		t.Fatalf("create fake bin: %v", err)
	}
	fakeGH := filepath.Join(fakeBin, "gh")
	fakeGHBody := `#!/bin/sh
cat <<'JSON'
{
  "url": "https://github.com/jakedom/ceo-harness/releases/tag/v0.2.0-test",
  "assets": [
    {"name": "ceo-packet_0.2.0-test_darwin_arm64.tar.gz"},
    {"name": "ceo-packet_0.2.0-test_linux_amd64.tar.gz"},
    {"name": "ceo-packet_0.2.0-test_linux_arm64.tar.gz"},
    {"name": "checksums.txt"},
    {"name": "release-manifest.json"}
  ]
}
JSON
`
	if err := os.WriteFile(fakeGH, []byte(fakeGHBody), 0o755); err != nil {
		t.Fatalf("write fake gh: %v", err)
	}

	preflight := exec.Command("sh", filepath.Join(root, "scripts", "release-preflight.sh"), dist)
	preflight.Dir = root
	preflight.Env = append(
		preflight.Environ(),
		"PATH="+fakeBin+string(os.PathListSeparator)+os.Getenv("PATH"),
		"GH_RELEASE_TAG=v0.2.0-test",
		"GH_REPO=jakedom/ceo-harness",
	)
	preflightOutput, err := preflight.CombinedOutput()
	if err == nil {
		t.Fatalf("release-preflight unexpectedly passed while other public blockers remain:\n%s", string(preflightOutput))
	}
	body := string(preflightOutput)
	for _, want := range []string{
		"| remote_release_url | pass | https://github.com/jakedom/ceo-harness/releases/tag/v0.2.0-test |",
		"| github_release_assets | pass | GitHub release v0.2.0-test has all archives, checksums.txt, and release-manifest.json |",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("preflight output missing %q:\n%s", want, body)
		}
	}
}

func Test_ReleaseReadinessScript_writesBlockedEvidencePacket(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("resolve repo root: %v", err)
	}
	tmp := t.TempDir()
	dist := filepath.Join(tmp, "dist")
	outputDir := filepath.Join(tmp, "release-readiness")

	release := exec.Command("sh", filepath.Join(root, "scripts", "release-local.sh"))
	release.Dir = root
	release.Env = append(release.Environ(), "DIST="+dist, "VERSION=0.2.0-test", "COMMIT=abc123")
	output, err := release.CombinedOutput()
	if err != nil {
		t.Fatalf("release-local failed: %v\n%s", err, string(output))
	}

	readiness := exec.Command("sh", filepath.Join(root, "scripts", "release-readiness.sh"), "--dist", dist, "--output-dir", outputDir)
	readiness.Dir = root
	readinessOutput, err := readiness.CombinedOutput()
	if err == nil {
		t.Fatalf("release-readiness unexpectedly passed without public release metadata:\n%s", string(readinessOutput))
	}
	if strings.Contains(string(readinessOutput), "command not found") {
		t.Fatalf("release-readiness output contains a shell quoting error:\n%s", string(readinessOutput))
	}

	for _, path := range []string{
		filepath.Join(outputDir, "index.md"),
		filepath.Join(outputDir, "summary.json"),
		filepath.Join(outputDir, "preflight.md"),
		filepath.Join(outputDir, "git-remote.txt"),
		filepath.Join(outputDir, "github-auth.txt"),
		filepath.Join(outputDir, "setup-actions.md"),
	} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("expected evidence file %s: %v\nscript output:\n%s", path, err, string(readinessOutput))
		}
	}

	summary, err := os.ReadFile(filepath.Join(outputDir, "summary.json"))
	if err != nil {
		t.Fatalf("read readiness summary: %v", err)
	}
	body := string(summary)
	for _, want := range []string{
		`"public_release_ready": false`,
		`"release_artifacts_verified": true`,
		`"preflight_exit_code": 1`,
		`"blocked_checks": [`,
		`"setup_actions": "setup-actions.md"`,
		`"remote_release_url"`,
		`"github_release_assets"`,
		`"homebrew_formula_url"`,
		`"artifact_signatures"`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("readiness summary missing %q:\n%s", want, body)
		}
	}

	index, err := os.ReadFile(filepath.Join(outputDir, "index.md"))
	if err != nil {
		t.Fatalf("read readiness index: %v", err)
	}
	if !strings.Contains(string(index), "release readiness: blocked") {
		t.Fatalf("readiness index did not record blocked status:\n%s", string(index))
	}
	setupActions, err := os.ReadFile(filepath.Join(outputDir, "setup-actions.md"))
	if err != nil {
		t.Fatalf("read setup actions: %v", err)
	}
	for _, want := range []string{
		"# Release Setup Actions",
		"remote_release_url: set `RELEASE_URL`",
		"github_release_assets: push a `v*` tag",
		"artifact_signatures: add `.sig`, `.minisig`, or `.asc`",
		"sh scripts/release-readiness.sh --dist dist --output-dir .omo/evidence/release-readiness-final",
	} {
		if !strings.Contains(string(setupActions), want) {
			t.Fatalf("setup-actions.md missing %q:\n%s", want, string(setupActions))
		}
	}
}

func Test_ReleaseBootstrapScript_writesBlockedEvidencePacket(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("resolve repo root: %v", err)
	}
	tmp := t.TempDir()
	dist := filepath.Join(tmp, "dist")
	outputDir := filepath.Join(tmp, "release-bootstrap")

	release := exec.Command("sh", filepath.Join(root, "scripts", "release-local.sh"))
	release.Dir = root
	release.Env = append(release.Environ(), "DIST="+dist, "VERSION=0.2.0-test", "COMMIT=abc123")
	output, err := release.CombinedOutput()
	if err != nil {
		t.Fatalf("release-local failed: %v\n%s", err, string(output))
	}

	bootstrap := exec.Command("sh", filepath.Join(root, "scripts", "release-bootstrap.sh"), "--dist", dist, "--output-dir", outputDir)
	bootstrap.Dir = root
	bootstrapOutput, err := bootstrap.CombinedOutput()
	if err == nil {
		t.Fatalf("release-bootstrap unexpectedly passed without public metadata:\n%s", string(bootstrapOutput))
	}

	for _, path := range []string{
		filepath.Join(outputDir, "index.md"),
		filepath.Join(outputDir, "summary.json"),
		filepath.Join(outputDir, "commands.sh"),
		filepath.Join(outputDir, "env.template"),
		filepath.Join(outputDir, "release-checklist.md"),
		filepath.Join(outputDir, "remote-homebrew-formula.rb"),
		filepath.Join(outputDir, "verify-release.txt"),
	} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("expected evidence file %s: %v\nscript output:\n%s", path, err, string(bootstrapOutput))
		}
	}

	summary := readTextFile(t, filepath.Join(outputDir, "summary.json"))
	for _, want := range []string{
		`"status": "blocked"`,
		`"release_bootstrap_ready": false`,
		`"local_release_artifacts": "pass"`,
		`"public_repo_url"`,
		`"public_release_url"`,
		`"homebrew_archive_base_url"`,
		`"signing_or_checksum_policy"`,
	} {
		if !strings.Contains(summary, want) {
			t.Fatalf("bootstrap summary missing %q:\n%s", want, summary)
		}
	}
}

func Test_ReleaseBootstrapScript_passesWithPublicMetadata(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("resolve repo root: %v", err)
	}
	tmp := t.TempDir()
	dist := filepath.Join(tmp, "dist")
	outputDir := filepath.Join(tmp, "release-bootstrap")

	release := exec.Command("sh", filepath.Join(root, "scripts", "release-local.sh"))
	release.Dir = root
	release.Env = append(release.Environ(), "DIST="+dist, "VERSION=0.2.0-test", "COMMIT=abc123")
	output, err := release.CombinedOutput()
	if err != nil {
		t.Fatalf("release-local failed: %v\n%s", err, string(output))
	}

	bootstrap := exec.Command(
		"sh",
		filepath.Join(root, "scripts", "release-bootstrap.sh"),
		"--dist", dist,
		"--output-dir", outputDir,
		"--repo-url", "https://github.com/acme/ceo-harness",
		"--release-url", "https://github.com/acme/ceo-harness/releases/tag/v0.2.0-test",
		"--homebrew-archive-base-url", "https://github.com/acme/ceo-harness/releases/download/v0.2.0-test",
		"--checksum-notes-url", "https://github.com/acme/ceo-harness/releases/tag/v0.2.0-test",
	)
	bootstrap.Dir = root
	bootstrapOutput, err := bootstrap.CombinedOutput()
	if err != nil {
		t.Fatalf("release-bootstrap failed with complete metadata: %v\n%s", err, string(bootstrapOutput))
	}

	summary := readTextFile(t, filepath.Join(outputDir, "summary.json"))
	for _, want := range []string{
		`"status": "pass"`,
		`"release_bootstrap_ready": true`,
		`"version": "0.2.0-test"`,
		`"blocked_count": 0`,
	} {
		if !strings.Contains(summary, want) {
			t.Fatalf("bootstrap summary missing %q:\n%s", want, summary)
		}
	}

	formula := readTextFile(t, filepath.Join(outputDir, "remote-homebrew-formula.rb"))
	for _, want := range []string{
		`homepage "https://github.com/acme/ceo-harness"`,
		`url "https://github.com/acme/ceo-harness/releases/download/v0.2.0-test/ceo-packet_0.2.0-test_darwin_arm64.tar.gz"`,
		`version "0.2.0-test"`,
	} {
		if !strings.Contains(formula, want) {
			t.Fatalf("remote formula missing %q:\n%s", want, formula)
		}
	}
}
