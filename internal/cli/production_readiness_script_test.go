package cli

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func Test_ProductionReadinessScript_reportsCurrentPublicBlockers(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("resolve repo root: %v", err)
	}
	outputDir := filepath.Join(t.TempDir(), "production-readiness")

	cmd := exec.Command(
		"sh",
		filepath.Join(root, "scripts", "production-readiness.sh"),
		"--output-dir", outputDir,
		"--skip-release-readiness",
		"--skip-secret-scan",
	)
	cmd.Dir = root
	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("production readiness unexpectedly passed with known public blockers:\n%s", string(output))
	}

	index := readTextFile(t, filepath.Join(outputDir, "index.md"))
	for _, want := range []string{
		"# Production Readiness Evidence",
		"Local production ready: true",
		"Public production ready: false",
		"Next public-production actions are in `launch-checklist.md`.",
		"| comparison | all_agent_29_task_comparison | blocked |",
		"| provider | openai_http_provider | blocked |",
		"| provider | openrouter_http_provider | blocked |",
		"| provider | moonshot_http_provider | blocked |",
	} {
		if !strings.Contains(index, want) {
			t.Fatalf("index.md missing %q:\n%s", want, index)
		}
	}

	checklist := readTextFile(t, filepath.Join(outputDir, "launch-checklist.md"))
	for _, want := range []string{
		"# Launch Checklist",
		"Public production ready: false",
		"Publish release proof",
		"push an explicit `v*` tag",
		"GitHub release workflow publishes verified tarballs",
		"`checksums.txt`",
		"`release-manifest.json`",
		"Refresh market comparison",
		"Prove OpenAI provider",
		"Prove OpenRouter provider",
		"Prove Moonshot provider",
		"sh scripts/production-readiness.sh --dist dist --output-dir .omo/evidence/production-readiness",
	} {
		if !strings.Contains(checklist, want) {
			t.Fatalf("launch-checklist.md missing %q:\n%s", want, checklist)
		}
	}

	summary := readTextFile(t, filepath.Join(outputDir, "summary.json"))
	for _, want := range []string{
		`"status": "blocked"`,
		`"local_production_ready": true`,
		`"public_production_ready": false`,
		`"comparison.all_agent_29_task_comparison"`,
		`"launch_checklist": {`,
		`"path": "launch-checklist.md"`,
		`"required_action_count":`,
		`"status": "pass"`,
	} {
		if !strings.Contains(summary, want) {
			t.Fatalf("summary.json missing %q:\n%s", want, summary)
		}
	}
}

func Test_ProductionReadinessScript_passesWithCompleteEvidence(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("resolve repo root: %v", err)
	}
	evidenceRoot := filepath.Join(t.TempDir(), "evidence")
	outputDir := filepath.Join(t.TempDir(), "production-readiness")

	writeProductionReadinessJSON(t, filepath.Join(evidenceRoot, "release-readiness-r1", "summary.json"), `{
  "status": "pass",
  "public_release_ready": true
}`)
	writeProductionReadinessJSON(t, filepath.Join(evidenceRoot, "production-core-29-ceo-r1", "summary.json"), `{
  "passed": 29,
  "failed": 0,
  "partial": 0,
  "timed_out": 0,
  "incomplete_evidence": 0
}`)
	writeProductionReadinessJSON(t, filepath.Join(evidenceRoot, "benchmark-fixtures-31-r1", "summary.json"), `{
  "task_count": 31,
  "passed": 31,
  "failed": 0,
  "partial": 0
}`)
	writeProductionReadinessJSON(t, filepath.Join(evidenceRoot, "external-agent-production-core-29-r1", "summary.json"), `{
  "task_count": 29,
  "agent_count": 4,
  "failed": 0,
  "partial": 0,
  "timed_out": 0,
  "incomplete_evidence": 0
}`)
	writeProviderIndex(t, filepath.Join(evidenceRoot, "provider-proof-kimi-r2", "index.md"), "pass")
	writeProviderIndex(t, filepath.Join(evidenceRoot, "provider-proof-codex-r1", "index.md"), "pass")
	writeProviderIndex(t, filepath.Join(evidenceRoot, "provider-proof-openai", "index.md"), "pass")
	writeProviderIndex(t, filepath.Join(evidenceRoot, "provider-proof-openrouter", "index.md"), "pass")
	writeProviderIndex(t, filepath.Join(evidenceRoot, "provider-proof-moonshot", "index.md"), "pass")
	writeProductionReadinessText(t, filepath.Join(evidenceRoot, "endurance-local-r3", "index.md"), "# Endurance Evidence\n\n## Summary\n\n- Overall: pass\n- Completed iterations: 30\n")

	cmd := exec.Command(
		"sh",
		filepath.Join(root, "scripts", "production-readiness.sh"),
		"--evidence-root", evidenceRoot,
		"--output-dir", outputDir,
		"--skip-release-readiness",
		"--skip-secret-scan",
	)
	cmd.Dir = root
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("production readiness failed with complete evidence: %v\n%s", err, string(output))
	}

	index := readTextFile(t, filepath.Join(outputDir, "index.md"))
	for _, want := range []string{
		"Status: pass",
		"Local production ready: true",
		"Public production ready: true",
		"| comparison | all_agent_29_task_comparison | pass |",
		"| provider | moonshot_http_provider | pass |",
	} {
		if !strings.Contains(index, want) {
			t.Fatalf("index.md missing %q:\n%s", want, index)
		}
	}
}

func writeProviderIndex(t *testing.T, path string, status string) {
	t.Helper()
	writeProductionReadinessText(t, path, "# Provider Proof Evidence\n\n## Summary\n\n- Overall: "+status+"\n")
}

func writeProductionReadinessJSON(t *testing.T, path string, body string) {
	t.Helper()
	writeProductionReadinessText(t, path, body+"\n")
}

func writeProductionReadinessText(t *testing.T, path string, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("create dir for %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
