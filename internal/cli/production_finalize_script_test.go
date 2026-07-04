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
	evidenceRoot := filepath.Join(t.TempDir(), "evidence root")

	cmd := exec.Command(
		"sh",
		filepath.Join(root, "scripts", "production-finalize.sh"),
		"--dry-run",
		"--output-dir", outputDir,
		"--evidence-root", evidenceRoot,
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
		"evidence root/provider-proof-openai",
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

func Test_ProductionFinalizeScript_marksSetupBlockedCompetitorSmokeBlocked(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("resolve repo root: %v", err)
	}
	outputDir := filepath.Join(t.TempDir(), "production-finalize")

	cmd := exec.Command(
		"sh",
		filepath.Join(root, "scripts", "production-finalize.sh"),
		"--output-dir", outputDir,
		"--dist", filepath.Join(root, "dist"),
		"--skip-release-readiness",
		"--skip-provider-proofs",
		"--skip-production-readiness",
	)
	cmd.Dir = root
	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("production-finalize unexpectedly passed with setup-blocked smoke:\n%s", string(output))
	}

	index := readTextFile(t, filepath.Join(outputDir, "index.md"))
	for _, want := range []string{
		"| competitor-smoke-command | pass |",
		"| competitor-smoke | blocked |",
		"Smoke summary has failed or setup-blocked competitors",
		"Open `next-actions.md`",
	} {
		if !strings.Contains(index, want) {
			t.Fatalf("index.md missing %q:\n%s", want, index)
		}
	}

	nextActions := readTextFile(t, filepath.Join(outputDir, "next-actions.md"))
	for _, want := range []string{
		"# Production Finalize Next Actions",
		"Fix competitor setup before final comparison",
		"ceo-packet production-finalize --workspace . --run-comparison",
	} {
		if !strings.Contains(nextActions, want) {
			t.Fatalf("next-actions.md missing %q:\n%s", want, nextActions)
		}
	}

	summary := readTextFile(t, filepath.Join(outputDir, "summary.json"))
	for _, want := range []string{
		`"status": "blocked"`,
		`"competitor-smoke"`,
		`"next_actions": {`,
		`"required_action_count": 2`,
	} {
		if !strings.Contains(summary, want) {
			t.Fatalf("summary.json missing %q:\n%s", want, summary)
		}
	}
}
