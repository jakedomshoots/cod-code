package cli

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func Test_ProductionLocalGateScript_passesWhenOnlyPublicBlockersRemain(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("resolve repo root: %v", err)
	}
	tmp := t.TempDir()
	dist := filepath.Join(tmp, "dist")
	outputDir := filepath.Join(tmp, "production-local-gate")

	release := exec.Command("sh", filepath.Join(root, "scripts", "release-local.sh"))
	release.Dir = root
	release.Env = append(release.Environ(), "DIST="+dist, "VERSION=0.2.0-test", "COMMIT=abc123")
	output, err := release.CombinedOutput()
	if err != nil {
		t.Fatalf("release-local failed: %v\n%s", err, string(output))
	}

	cmd := exec.Command(
		"sh",
		filepath.Join(root, "scripts", "production-local-gate.sh"),
		"--dist", dist,
		"--output-dir", outputDir,
	)
	cmd.Dir = root
	output, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("production-local-gate failed: %v\n%s", err, string(output))
	}
	body := string(output)
	for _, want := range []string{
		"production-local-gate: pass local_production_ready=true public_production_ready=false",
		"production-local-gate: blocked_count=",
		"production-local-gate: checklist_actions=",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("production-local-gate output missing %q:\n%s", want, body)
		}
	}
	for _, path := range []string{
		filepath.Join(outputDir, "summary.json"),
		filepath.Join(outputDir, "index.md"),
		filepath.Join(outputDir, "launch-checklist.md"),
		filepath.Join(outputDir, "production-readiness.stdout.txt"),
		filepath.Join(outputDir, "production-readiness.stderr.txt"),
	} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("expected evidence file %s: %v", path, err)
		}
	}
}

func Test_ProductionLocalGateScript_failsWhenLocalReadinessIsFalse(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("resolve repo root: %v", err)
	}
	tmp := t.TempDir()
	outputDir := filepath.Join(tmp, "production-local-gate")

	cmd := exec.Command(
		"sh",
		filepath.Join(root, "scripts", "production-local-gate.sh"),
		"--dist", filepath.Join(root, "dist"),
		"--evidence-root", filepath.Join(tmp, "empty-evidence"),
		"--output-dir", outputDir,
	)
	cmd.Dir = root
	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("production-local-gate unexpectedly passed with missing dist:\n%s", string(output))
	}
	if !strings.Contains(string(output), "production-local-gate: fail local_production_ready=false") {
		t.Fatalf("production-local-gate output missing local failure:\n%s", string(output))
	}
}
