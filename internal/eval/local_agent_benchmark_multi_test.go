package eval

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func Test_Benchmark_records_failed_score_checks_when_agent_misses_requirement(t *testing.T) {
	// Given
	binDir := t.TempDir()
	writeExecutableContent(t, filepath.Join(binDir, "codex"), multiTaskAgentScript())
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	root := t.TempDir()
	tasksDir := filepath.Join(root, "tasks")
	writeTaskSpec(t, tasksDir, `{
		"id":"docs-one",
		"category":"docs",
		"title":"Update one",
		"objective":"Refresh docs one.",
		"required_changed_files":["docs/ONE.md"],
		"required_diff_terms":["missing-term"]
	}`)

	// When
	summary, err := RunLocalAgentBenchmark(context.Background(), LocalAgentBenchmarkRequest{
		TasksDir:        tasksDir,
		OutputDir:       filepath.Join(root, "benchmark"),
		TimeoutSeconds:  5,
		Agents:          []string{"codex_cli"},
		BenchmarkTaskID: "docs-one",
	})
	// Then
	if err != nil {
		t.Fatalf("RunLocalAgentBenchmark returned error: %v", err)
	}
	if summary.Partial != 1 || len(summary.Results) != 1 {
		t.Fatalf("summary = %+v, want one partially failed benchmark result", summary)
	}
	failed := summary.Results[0].FailedScoreChecks
	if len(failed) != 1 || failed[0].Name != "diff_term:missing-term" || failed[0].Status != "fail" {
		t.Fatalf("FailedScoreChecks = %#v, want missing diff term failure", failed)
	}
}

func Test_RunLocalAgentBenchmark_marks_missing_required_artifact_incomplete(t *testing.T) {
	// Given
	binDir := t.TempDir()
	writeExecutableContent(t, filepath.Join(binDir, "codex"), `#!/bin/sh
cat > docs/ONE.md <<'EOF'
# Benchmark Fixture

one-term
EOF
printf 'done\n'
`)
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	root := t.TempDir()
	tasksDir := filepath.Join(root, "tasks")
	writeTaskSpec(t, tasksDir, `{
		"id":"docs-one",
		"category":"docs",
		"title":"Update one",
		"objective":"Refresh docs one.",
		"required_changed_files":["docs/ONE.md"],
		"required_artifacts":[".omo/evidence/docs-one.md"],
		"required_diff_terms":["one-term"],
		"required_report_fields":["changed_files"]
	}`)

	// When
	summary, err := RunLocalAgentBenchmark(context.Background(), LocalAgentBenchmarkRequest{
		TasksDir:        tasksDir,
		OutputDir:       filepath.Join(root, "benchmark"),
		TimeoutSeconds:  5,
		Agents:          []string{"codex_cli"},
		BenchmarkTaskID: "docs-one",
	})
	// Then
	if err != nil {
		t.Fatalf("RunLocalAgentBenchmark returned error: %v", err)
	}
	if summary.IncompleteEvidence != 1 || summary.Partial != 1 {
		t.Fatalf("summary = %+v, want one incomplete partial result", summary)
	}
	result := summary.Results[0]
	if result.EvidenceStatus != localAgentEvidenceIncomplete {
		t.Fatalf("EvidenceStatus = %q, want incomplete", result.EvidenceStatus)
	}
	if len(result.FailedScoreChecks) != 1 || result.FailedScoreChecks[0].Name != "artifact:.omo/evidence/docs-one.md" {
		t.Fatalf("FailedScoreChecks = %#v, want missing artifact failure", result.FailedScoreChecks)
	}
	if _, statErr := os.Stat(filepath.Join(result.WorkspaceDir, ".omo/evidence/docs-one.md")); !os.IsNotExist(statErr) {
		t.Fatalf("required artifact should be missing: %v", statErr)
	}
	requireFile(t, filepath.Join(root, "benchmark", "comparison-report.md"))
}

func Test_RunLocalAgentBenchmark_marks_provider_quota_as_setup_blocked(t *testing.T) {
	// Given
	binDir := t.TempDir()
	writeExecutableContent(t, filepath.Join(binDir, "opencode"), `#!/bin/sh
printf 'AI_APICallError: Token Plan usage limit reached\n' >&2
exit 1
`)
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	root := t.TempDir()
	tasksDir := filepath.Join(root, "tasks")
	writeTaskSpec(t, tasksDir, `{
		"id":"docs-one",
		"category":"docs",
		"title":"Update one",
		"objective":"Refresh docs one.",
		"required_changed_files":["docs/ONE.md"],
		"required_artifacts":[".omo/evidence/docs-one.md"],
		"required_diff_terms":["one-term"],
		"required_report_fields":["changed_files"]
	}`)

	// When
	summary, err := RunLocalAgentBenchmark(context.Background(), LocalAgentBenchmarkRequest{
		TasksDir:        tasksDir,
		OutputDir:       filepath.Join(root, "benchmark"),
		TimeoutSeconds:  1,
		Agents:          []string{"opencode"},
		BenchmarkTaskID: "docs-one",
	})
	// Then
	if err != nil {
		t.Fatalf("RunLocalAgentBenchmark returned error: %v", err)
	}
	if summary.SetupBlocked != 1 || summary.TimedOut != 0 || summary.IncompleteEvidence != 0 {
		t.Fatalf("summary = %+v, want setup-blocked with complete evidence", summary)
	}
	result := summary.Results[0]
	if result.Status != localAgentStatusSetupBlocked || result.EvidenceStatus != localAgentEvidenceComplete {
		t.Fatalf("result status=%q evidence=%q, want setup blocked complete", result.Status, result.EvidenceStatus)
	}
	stderr, err := os.ReadFile(filepath.Join(root, "benchmark", "opencode", "stderr.log"))
	if err != nil {
		t.Fatalf("read stderr evidence: %v", err)
	}
	if !strings.Contains(string(stderr), "Token Plan usage limit reached") {
		t.Fatalf("stderr evidence missing provider quota error:\n%s", stderr)
	}
}

func Test_RunCLI_runs_local_agent_benchmark_with_repeat_flag(t *testing.T) {
	// Given
	binDir := t.TempDir()
	writeExecutableContent(t, filepath.Join(binDir, "codex"), multiTaskAgentScript())
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	root := t.TempDir()
	tasksDir := filepath.Join(root, "tasks")
	writeTaskSpec(t, tasksDir, multiTaskSpec())
	outputDir := filepath.Join(root, "benchmark")

	// When
	err := RunCLI(context.Background(), os.Stdout, os.Stderr, []string{
		"--local-agent-benchmark",
		"--local-agents", "codex_cli",
		"--local-agent-benchmark-task", "docs-one",
		"--local-agent-benchmark-repeat", "2",
		"--local-agent-benchmark-concurrency", "2",
		"--tasks", tasksDir,
		"--output-dir", outputDir,
		"--timeout-seconds", "5",
	})
	// Then
	if err != nil {
		t.Fatalf("RunCLI returned error: %v", err)
	}
	payload, readErr := os.ReadFile(filepath.Join(outputDir, "summary.json"))
	if readErr != nil {
		t.Fatalf("read summary: %v", readErr)
	}
	var summary LocalAgentBenchmarkSummary
	if err := json.Unmarshal(payload, &summary); err != nil {
		t.Fatalf("decode summary: %v", err)
	}
	if summary.RepeatCount != 2 || summary.Concurrency != 2 || summary.RunCount != 2 || summary.Passed != 2 {
		t.Fatalf("summary = %+v, want repeat count applied", summary)
	}
}

func multiTaskAgentScript() string {
	return `#!/bin/sh
if [ -f docs/ONE.md ]; then
  printf '# Benchmark Fixture\n\none-term\n' > docs/ONE.md
  mkdir -p .omo/evidence
  printf 'agent evidence\n' > .omo/evidence/docs-one.md
fi
if [ -f docs/TWO.md ]; then
  printf '# Benchmark Fixture\n\ntwo-term\n' > docs/TWO.md
  mkdir -p .omo/evidence
  printf 'agent evidence\n' > .omo/evidence/docs-two.md
fi
printf 'done\n'
`
}

func multiTaskSpec() string {
	return `[
		{
			"id":"docs-one",
			"category":"docs",
			"title":"Update one",
			"objective":"Refresh docs one.",
			"required_changed_files":["docs/ONE.md"],
			"required_commands":["go test ./internal/cli -count=1"],
			"required_artifacts":[".omo/evidence/docs-one.md"],
			"required_diff_terms":["one-term"],
			"required_report_fields":["changed_files"]
		},
		{
			"id":"docs-two",
			"category":"docs",
			"title":"Update two",
			"objective":"Refresh docs two.",
			"required_changed_files":["docs/TWO.md"],
			"required_commands":["go test ./internal/cli -count=1"],
			"required_artifacts":[".omo/evidence/docs-two.md"],
			"required_diff_terms":["two-term"],
			"required_report_fields":["changed_files"]
		}
	]`
}
