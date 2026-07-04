package cli

import (
	"bytes"
	"context"
	"os"
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
		"`setup-actions.md`",
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
		"--local-agent-benchmark-timeout-retries 1",
		"--local-agent-benchmark-result-retries 1",
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
		`"setup_actions": {`,
		`"path": "setup-actions.md"`,
		`"required_action_count":`,
		`"sha256": "`,
		`"provider-openai"`,
		`"all-agent-29-comparison"`,
	} {
		if !strings.Contains(summary, want) {
			t.Fatalf("summary.json missing %q:\n%s", want, summary)
		}
	}

	nextActionsJSON := readTextFile(t, filepath.Join(outputDir, "next-actions.json"))
	for _, want := range []string{
		`"declared_evidence_files":`,
		`"field": "evidence"`,
		`"exists":`,
		`"sha256":`,
		`"go"`,
		`"run"`,
		`"./cmd/ceo-packet"`,
		`"production-finalize"`,
	} {
		if !strings.Contains(nextActionsJSON, want) {
			t.Fatalf("next-actions.json missing declared evidence metadata %q:\n%s", want, nextActionsJSON)
		}
	}

	setupActions := readTextFile(t, filepath.Join(outputDir, "setup-actions.md"))
	for _, want := range []string{
		"# Production Setup Actions",
		"## Providers",
		"openai:",
		"## Final Rerun",
		"go run ./cmd/ceo-packet production-finalize --workspace . --run-comparison",
	} {
		if !strings.Contains(setupActions, want) {
			t.Fatalf("setup-actions.md missing %q:\n%s", want, setupActions)
		}
	}
}

func Test_ProductionFinalizeScript_usesExistingCleanComparisonEvidence(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("resolve repo root: %v", err)
	}
	outputDir := filepath.Join(t.TempDir(), "production-finalize")
	evidenceRoot := filepath.Join(t.TempDir(), "evidence")
	writeProductionReadinessJSON(t, filepath.Join(evidenceRoot, "external-agent-production-core-29-final-result-retry-r1", "summary.json"), `{
  "task_count": 29,
  "agent_count": 4,
  "run_count": 116,
  "passed": 116,
  "partial": 0,
  "failed": 0,
  "timed_out": 0,
  "setup_blocked": 0,
  "skipped": 0,
  "incomplete_evidence": 0
}`)

	cmd := exec.Command(
		"sh",
		filepath.Join(root, "scripts", "production-finalize.sh"),
		"--output-dir", outputDir,
		"--evidence-root", evidenceRoot,
		"--dist", filepath.Join(root, "dist"),
		"--skip-release-readiness",
		"--skip-provider-proofs",
		"--skip-competitor-smoke",
		"--skip-production-readiness",
	)
	cmd.Dir = root
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("production-finalize failed with clean comparison evidence: %v\n%s", err, string(output))
	}

	index := readTextFile(t, filepath.Join(outputDir, "index.md"))
	for _, want := range []string{
		"Status: pass",
		"| all-agent-29-comparison | pass |",
		"Existing clean all-agent comparison evidence found",
	} {
		if !strings.Contains(index, want) {
			t.Fatalf("index.md missing %q:\n%s", want, index)
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
	binDir := t.TempDir()
	writeExecutableScript(t, filepath.Join(binDir, "opencode"), `#!/bin/sh
if [ "$1" = "--version" ]; then
  printf 'opencode 1.0.0\n'
  exit 0
fi
printf 'AI_APICallError: Token Plan usage limit reached\n' >&2
exit 1
`)

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
	cmd.Env = append(cmd.Environ(), "PATH="+binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
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
		"Open `setup-actions.md`",
	} {
		if !strings.Contains(index, want) {
			t.Fatalf("index.md missing %q:\n%s", want, index)
		}
	}

	nextActions := readTextFile(t, filepath.Join(outputDir, "next-actions.md"))
	for _, want := range []string{
		"# Production Finalize Next Actions",
		"Fix competitor setup before final comparison",
		"go run ./cmd/ceo-packet production-finalize --workspace . --dry-run",
		"production-finalize/competitor-smoke/summary.json",
	} {
		if !strings.Contains(nextActions, want) {
			t.Fatalf("next-actions.md missing %q:\n%s", want, nextActions)
		}
	}

	nextActionsJSON := readTextFile(t, filepath.Join(outputDir, "next-actions.json"))
	for _, want := range []string{
		`"required_action_count": 1`,
		`"id": "competitor-smoke"`,
		`"kind": "competitor_setup"`,
		`"inspect": "competitor-smoke/summary.json"`,
		"Fix competitor setup before final comparison",
	} {
		if !strings.Contains(nextActionsJSON, want) {
			t.Fatalf("next-actions.json missing %q:\n%s", want, nextActionsJSON)
		}
	}

	summary := readTextFile(t, filepath.Join(outputDir, "summary.json"))
	for _, want := range []string{
		`"status": "blocked"`,
		`"competitor-smoke"`,
		`"next_actions": {`,
		`"json_path": "next-actions.json"`,
		`"required_action_count": 1`,
		`"setup_actions": {`,
	} {
		if !strings.Contains(summary, want) {
			t.Fatalf("summary.json missing %q:\n%s", want, summary)
		}
	}
}
