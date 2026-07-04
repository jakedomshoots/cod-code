package cli

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func Test_ProductionReadinessScript_reportsCurrentPublicState(t *testing.T) {
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
	if err != nil {
		t.Fatalf("production readiness failed with current public evidence: %v\n%s", err, string(output))
	}

	index := readTextFile(t, filepath.Join(outputDir, "index.md"))
	for _, want := range []string{
		"# Production Readiness Evidence",
		"Local production ready: true",
		"Public production ready: true",
		"| comparison | all_agent_29_task_comparison | pass |",
		"| provider | openrouter_http_provider |",
		"| provider | kimi-code_http_provider |",
		"| provider | minimax_http_provider |",
	} {
		if !strings.Contains(index, want) {
			t.Fatalf("index.md missing %q:\n%s", want, index)
		}
	}

	checklist := readTextFile(t, filepath.Join(outputDir, "launch-checklist.md"))
	for _, want := range []string{
		"# Launch Checklist",
		"Public production ready: true",
	} {
		if !strings.Contains(checklist, want) {
			t.Fatalf("launch-checklist.md missing %q:\n%s", want, checklist)
		}
	}

	summary := readTextFile(t, filepath.Join(outputDir, "summary.json"))
	for _, want := range []string{
		`"status": "pass"`,
		`"local_production_ready": true`,
		`"public_production_ready": true`,
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
	writeProviderIndex(t, filepath.Join(evidenceRoot, "provider-proof-openrouter", "index.md"), "pass")
	writeProviderIndex(t, filepath.Join(evidenceRoot, "provider-proof-kimi-code", "index.md"), "pass")
	writeProviderIndex(t, filepath.Join(evidenceRoot, "provider-proof-minimax", "index.md"), "pass")
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
		"| provider | minimax_http_provider | pass |",
	} {
		if !strings.Contains(index, want) {
			t.Fatalf("index.md missing %q:\n%s", want, index)
		}
	}
}

func Test_ProductionReadinessScript_usesCanonicalBlockedProviderEvidence(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("resolve repo root: %v", err)
	}
	evidenceRoot := filepath.Join(t.TempDir(), "evidence")
	outputDir := filepath.Join(t.TempDir(), "production-readiness")

	writeCompleteProductionReadinessEvidence(t, evidenceRoot)
	writeProviderIndex(t, filepath.Join(evidenceRoot, "provider-proof-openrouter", "index.md"), "blocked")
	writeProductionReadinessJSON(t, filepath.Join(evidenceRoot, "provider-proof-openrouter", "summary.json"), `{
  "status": "blocked",
  "blocked_reason": "empty_api_key_env",
  "command_script_secret_policy": "no_secret_assignment",
  "secret_value_saved": false
}`)

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
	if err == nil {
		t.Fatalf("production readiness unexpectedly passed with blocked provider:\n%s", string(output))
	}

	summary := readTextFile(t, filepath.Join(outputDir, "summary.json"))
	for _, want := range []string{
		`"public_production_ready": false`,
		`"provider.openrouter_http_provider"`,
		filepath.ToSlash(filepath.Join(evidenceRoot, "provider-proof-openrouter", "index.md")),
		`"detail": "openrouter HTTP provider proof is blocked by setup"`,
	} {
		if !strings.Contains(summary, want) {
			t.Fatalf("summary.json missing %q:\n%s", want, summary)
		}
	}
	if strings.Contains(summary, "provider-proof-openrouter-blocked-r1") {
		t.Fatalf("summary.json used stale blocked provider folder:\n%s", summary)
	}
}

func Test_ProductionReadinessScript_requiresSafeBlockedReleaseEvidence(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("resolve repo root: %v", err)
	}
	evidenceRoot := filepath.Join(t.TempDir(), "evidence")
	outputDir := filepath.Join(t.TempDir(), "production-readiness")

	writeCompleteProductionReadinessEvidence(t, evidenceRoot)
	writeProductionReadinessJSON(t, filepath.Join(evidenceRoot, "release-readiness-final", "summary.json"), `{
  "status": "blocked",
  "public_release_ready": false,
  "publish_actions_performed": false,
  "secret_value_saved": false
}`)

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
	if err == nil {
		t.Fatalf("production readiness unexpectedly passed with unsafe blocked release evidence:\n%s", string(output))
	}

	index := readTextFile(t, filepath.Join(outputDir, "index.md"))
	if !strings.Contains(index, "| release | public_release_ready | blocked |") {
		t.Fatalf("index.md missing blocked release row:\n%s", index)
	}
	summary := readTextFile(t, filepath.Join(outputDir, "summary.json"))
	if !strings.Contains(summary, "Release evidence missing setup safety policy") {
		t.Fatalf("summary.json missing unsafe release detail:\n%s", summary)
	}
}

func Test_ProductionReadinessScript_usesNewestSkippedReleaseReadinessEvidence(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("resolve repo root: %v", err)
	}
	evidenceRoot := filepath.Join(t.TempDir(), "evidence")
	outputDir := filepath.Join(t.TempDir(), "production-readiness")

	writeCompleteProductionReadinessEvidence(t, evidenceRoot)
	oldRelease := filepath.Join(evidenceRoot, "release-readiness-r1", "summary.json")
	newRelease := filepath.Join(evidenceRoot, "release-readiness-final", "summary.json")
	writeProductionReadinessJSON(t, oldRelease, `{
  "status": "blocked",
  "public_release_ready": false
}`)
	writeProductionReadinessJSON(t, newRelease, `{
  "status": "pass",
  "public_release_ready": true
}`)
	writeProductionReadinessJSON(t, filepath.Join(evidenceRoot, "external-agent-production-core-29-final", "summary.json"), `{
  "task_count": 29,
  "agent_count": 4,
  "failed": 0,
  "partial": 0,
  "timed_out": 0,
  "incomplete_evidence": 0
}`)
	oldTime := time.Now().Add(-2 * time.Hour)
	newTime := time.Now()
	if err := os.Chtimes(oldRelease, oldTime, oldTime); err != nil {
		t.Fatalf("touch old release: %v", err)
	}
	if err := os.Chtimes(newRelease, newTime, newTime); err != nil {
		t.Fatalf("touch new release: %v", err)
	}

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
		t.Fatalf("production readiness failed with newer passing release evidence: %v\n%s", err, string(output))
	}

	index := readTextFile(t, filepath.Join(outputDir, "index.md"))
	for _, want := range []string{
		"| release | public_release_readiness_run | skipped |",
		"| release | public_release_ready | pass |",
		newRelease,
	} {
		if !strings.Contains(index, want) {
			t.Fatalf("index.md missing %q:\n%s", want, index)
		}
	}
	if strings.Contains(index, oldRelease) {
		t.Fatalf("index.md used stale release-readiness evidence:\n%s", index)
	}
}

func Test_ProductionReadinessScript_usesNewestComparisonEvidence(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("resolve repo root: %v", err)
	}
	evidenceRoot := filepath.Join(t.TempDir(), "evidence")
	outputDir := filepath.Join(t.TempDir(), "production-readiness")

	writeCompleteProductionReadinessEvidence(t, evidenceRoot)
	oldComparison := filepath.Join(evidenceRoot, "external-agent-production-core-29-r9", "summary.json")
	newComparison := filepath.Join(evidenceRoot, "external-agent-production-core-29-final", "summary.json")
	writeProductionReadinessJSON(t, oldComparison, `{
  "task_count": 29,
  "agent_count": 4,
  "failed": 0,
  "partial": 0,
  "timed_out": 0,
  "incomplete_evidence": 0
}`)
	writeProductionReadinessJSON(t, newComparison, `{
  "task_count": 29,
  "agent_count": 4,
  "failed": 0,
  "partial": 0,
  "timed_out": 2,
  "incomplete_evidence": 1
}`)
	oldTime := time.Now().Add(-2 * time.Hour)
	newTime := time.Now()
	if err := os.Chtimes(oldComparison, oldTime, oldTime); err != nil {
		t.Fatalf("touch old comparison: %v", err)
	}
	if err := os.Chtimes(newComparison, newTime, newTime); err != nil {
		t.Fatalf("touch new comparison: %v", err)
	}

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
	if err == nil {
		t.Fatalf("production readiness unexpectedly passed with newer blocked comparison:\n%s", string(output))
	}

	index := readTextFile(t, filepath.Join(outputDir, "index.md"))
	for _, want := range []string{
		"| comparison | all_agent_29_task_comparison | blocked |",
		newComparison,
	} {
		if !strings.Contains(index, want) {
			t.Fatalf("index.md missing %q:\n%s", want, index)
		}
	}
}

func writeCompleteProductionReadinessEvidence(t *testing.T, evidenceRoot string) {
	t.Helper()
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
	writeProviderIndex(t, filepath.Join(evidenceRoot, "provider-proof-kimi-r2", "index.md"), "pass")
	writeProviderIndex(t, filepath.Join(evidenceRoot, "provider-proof-codex-r1", "index.md"), "pass")
	writeProviderIndex(t, filepath.Join(evidenceRoot, "provider-proof-openrouter", "index.md"), "pass")
	writeProviderIndex(t, filepath.Join(evidenceRoot, "provider-proof-kimi-code", "index.md"), "pass")
	writeProviderIndex(t, filepath.Join(evidenceRoot, "provider-proof-minimax", "index.md"), "pass")
	writeProductionReadinessText(t, filepath.Join(evidenceRoot, "endurance-local-r3", "index.md"), "# Endurance Evidence\n\n## Summary\n\n- Overall: pass\n- Completed iterations: 30\n")
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
