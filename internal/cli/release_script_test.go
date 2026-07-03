package cli

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
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
