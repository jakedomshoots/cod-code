package cli

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func Test_DogfoodRealScript_dryRunRepeatWritesIsolatedEvidence(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("resolve repo root: %v", err)
	}
	repoDir := filepath.Join(t.TempDir(), "repo")
	if err := os.MkdirAll(repoDir, 0o755); err != nil {
		t.Fatalf("create temp repo: %v", err)
	}
	outputDir := filepath.Join(t.TempDir(), "dogfood-real")

	cmd := exec.Command(
		"sh",
		filepath.Join(root, "scripts", "dogfood-real.sh"),
		"--dry-run",
		"--output-dir", outputDir,
		"--repeat", "2",
		"--repo", "sample:"+repoDir,
	)
	cmd.Dir = root
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("dogfood-real dry-run failed: %v\n%s", err, string(output))
	}

	index := readTextFile(t, filepath.Join(outputDir, "index.md"))
	for _, want := range []string{
		"- Evidence root: " + outputDir,
		"- Repeat count: 2",
		"| sample run-01 | planned |",
		"| sample run-02 | planned |",
	} {
		if !strings.Contains(index, want) {
			t.Fatalf("index.md missing %q:\n%s", want, index)
		}
	}
	requireTextFile(t, filepath.Join(outputDir, "repos", "sample", "summary.md"))
	requireTextFile(t, filepath.Join(outputDir, "repos", "sample", "run-01", "plan.md"))
	requireTextFile(t, filepath.Join(outputDir, "repos", "sample", "run-02", "plan.md"))
}

func readTextFile(t *testing.T, path string) string {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(content)
}

func requireTextFile(t *testing.T, path string) {
	t.Helper()
	content := readTextFile(t, path)
	if strings.TrimSpace(content) == "" {
		t.Fatalf("%s is empty", path)
	}
}
