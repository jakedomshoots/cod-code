package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func Test_ReleaseWorkflow_publishesGitHubReleaseAssets(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("resolve repo root: %v", err)
	}
	content, err := os.ReadFile(filepath.Join(root, ".github", "workflows", "release.yml"))
	if err != nil {
		t.Fatalf("read release workflow: %v", err)
	}
	body := string(content)
	for _, want := range []string{
		"contents: write",
		`version="${GITHUB_REF_NAME#v}"`,
		`VERSION="$version" sh scripts/release-local.sh`,
		"sh scripts/verify-release.sh dist",
		"gh release create \"$GITHUB_REF_NAME\"",
		"dist/*.tar.gz",
		"dist/checksums.txt",
		"dist/release-manifest.json",
		"--notes-file dist/release-notes.md",
		"--verify-tag",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("release workflow missing %q:\n%s", want, body)
		}
	}
}
