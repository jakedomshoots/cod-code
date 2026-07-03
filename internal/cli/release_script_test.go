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
