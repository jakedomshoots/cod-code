package cli

import (
	"bytes"
	"context"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func Test_ProductionFinalizeScript_dryRunWritesGuardedPlan(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("resolve repo root: %v", err)
	}
	outputDir := filepath.Join(t.TempDir(), "production-finalize")

	cmd := exec.Command(
		"sh",
		filepath.Join(root, "scripts", "production-finalize.sh"),
		"--dry-run",
		"--output-dir", outputDir,
		"--dist", filepath.Join(root, "dist"),
	)
	cmd.Dir = root
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("production-finalize dry-run failed: %v\n%s", err, string(output))
	}

	index := readTextFile(t, filepath.Join(outputDir, "index.md"))
	for _, want := range []string{
		"# Production Finalize Evidence",
		"Status: planned",
		"Publishes or tags: false",
		"Secret values saved: false",
		"| provider-openai | planned |",
		"| provider-openrouter | planned |",
		"| provider-moonshot | planned |",
		"| all-agent-29-comparison | planned |",
		"`commands.sh`",
	} {
		if !strings.Contains(index, want) {
			t.Fatalf("index.md missing %q:\n%s", want, index)
		}
	}

	commands := readTextFile(t, filepath.Join(outputDir, "commands.sh"))
	for _, want := range []string{
		"scripts/release-readiness.sh",
		"scripts/provider-proof.sh --provider openai",
		"scripts/provider-proof.sh --provider openrouter",
		"scripts/provider-proof.sh --provider moonshot",
		"--comparison-smoke",
		"--local-agent-benchmark-task production-core",
		"scripts/production-readiness.sh",
	} {
		if !strings.Contains(commands, want) {
			t.Fatalf("commands.sh missing %q:\n%s", want, commands)
		}
	}

	summary := readTextFile(t, filepath.Join(outputDir, "summary.json"))
	for _, want := range []string{
		`"status": "planned"`,
		`"secret_value_saved": false`,
		`"publish_actions_performed": false`,
		`"provider-openai"`,
		`"all-agent-29-comparison"`,
	} {
		if !strings.Contains(summary, want) {
			t.Fatalf("summary.json missing %q:\n%s", want, summary)
		}
	}
}

func Test_Run_productionFinalizeVerbRunsDryRun(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("resolve repo root: %v", err)
	}
	outputDir := filepath.Join(t.TempDir(), "production-finalize")
	var out bytes.Buffer

	err = Run(context.Background(), &out, []string{
		"production-finalize",
		"--workspace", root,
		"--dry-run",
		"--output-dir", outputDir,
		"--dist", filepath.Join(root, "dist"),
	})
	if err != nil {
		t.Fatalf("Run production-finalize returned error: %v\n%s", err, out.String())
	}
	if !strings.Contains(out.String(), "production-finalize: planned") {
		t.Fatalf("output = %q, want planned finalizer", out.String())
	}
	summary := readTextFile(t, filepath.Join(outputDir, "summary.json"))
	if !strings.Contains(summary, `"status": "planned"`) {
		t.Fatalf("summary missing planned status:\n%s", summary)
	}
}
