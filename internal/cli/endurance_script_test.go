package cli

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func Test_EnduranceScript_dryRunWritesIterationEvidence(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("resolve repo root: %v", err)
	}
	outputDir := filepath.Join(t.TempDir(), "endurance")

	cmd := exec.Command(
		"sh",
		filepath.Join(root, "scripts", "endurance.sh"),
		"--dry-run",
		"--iterations", "2",
		"--output-dir", outputDir,
	)
	cmd.Dir = root
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("endurance dry-run failed: %v\n%s", err, string(output))
	}

	index := readTextFile(t, filepath.Join(outputDir, "index.md"))
	for _, want := range []string{
		"# Endurance Evidence",
		"- Mode: dry-run",
		"- Planned iterations: 2",
		"| run-01 | planned |",
		"| run-02 | planned |",
	} {
		if !strings.Contains(index, want) {
			t.Fatalf("index.md missing %q:\n%s", want, index)
		}
	}
	requireTextFile(t, filepath.Join(outputDir, "run-01", "plan.md"))
	requireTextFile(t, filepath.Join(outputDir, "run-02", "plan.md"))
}
