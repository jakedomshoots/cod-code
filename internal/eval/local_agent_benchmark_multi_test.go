package eval

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func Test_RunLocalAgentBenchmark_runs_selected_tasks_for_each_repeat(t *testing.T) {
	// Given
	binDir := t.TempDir()
	writeExecutableContent(t, filepath.Join(binDir, "codex"), multiTaskAgentScript())
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	root := t.TempDir()
	tasksDir := filepath.Join(root, "tasks")
	writeTaskSpec(t, tasksDir, multiTaskSpec())

	// When
	summary, err := RunLocalAgentBenchmark(context.Background(), LocalAgentBenchmarkRequest{
		TasksDir:        tasksDir,
		OutputDir:       filepath.Join(root, "benchmark"),
		TimeoutSeconds:  5,
		Agents:          []string{"codex_cli"},
		BenchmarkTaskID: "docs-one,docs-two",
		RepeatCount:     2,
	})
	// Then
	if err != nil {
		t.Fatalf("RunLocalAgentBenchmark returned error: %v", err)
	}
	if summary.TaskCount != 2 || summary.RepeatCount != 2 || summary.RunCount != 4 {
		t.Fatalf("summary = %+v, want 2 tasks repeated twice", summary)
	}
	if summary.Passed != 4 || len(summary.Results) != 4 {
		t.Fatalf("summary = %+v, want four passing results", summary)
	}
	wantAttempts := map[string]int{"docs-one/1": 0, "docs-one/2": 0, "docs-two/1": 0, "docs-two/2": 0}
	for _, result := range summary.Results {
		wantAttempts[fmt.Sprintf("%s/%d", result.TaskID, result.Attempt)]++
		requireFile(t, result.ScorePath)
	}
	for key, count := range wantAttempts {
		if count != 1 {
			t.Fatalf("attempt %s count = %d, want 1", key, count)
		}
	}
}

func Test_RunLocalAgentBenchmark_runs_jobs_in_parallel_when_concurrency_is_set(t *testing.T) {
	// Given
	binDir := t.TempDir()
	barrierDir := filepath.Join(t.TempDir(), "barrier")
	writeExecutableContent(t, filepath.Join(binDir, "codex"), concurrentMultiTaskAgentScript())
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("BENCHMARK_CONCURRENCY_DIR", barrierDir)
	root := t.TempDir()
	tasksDir := filepath.Join(root, "tasks")
	writeTaskSpec(t, tasksDir, multiTaskSpec())

	// When
	summary, err := RunLocalAgentBenchmark(context.Background(), LocalAgentBenchmarkRequest{
		TasksDir:        tasksDir,
		OutputDir:       filepath.Join(root, "benchmark"),
		TimeoutSeconds:  10,
		Agents:          []string{"codex_cli"},
		BenchmarkTaskID: "docs-one,docs-two",
		Concurrency:     2,
	})
	// Then
	if err != nil {
		t.Fatalf("RunLocalAgentBenchmark returned error: %v", err)
	}
	if summary.Concurrency != 2 || summary.Passed != 2 || len(summary.Results) != 2 {
		t.Fatalf("summary = %+v, want two passing concurrent results", summary)
	}
	if summary.Results[0].TaskID != "docs-one" || summary.Results[1].TaskID != "docs-two" {
		t.Fatalf("result order = %s, %s; want planned task order", summary.Results[0].TaskID, summary.Results[1].TaskID)
	}
}

func Test_LocalAgentBenchmarkTasks_expands_market_parity_core_suite(t *testing.T) {
	// Given
	root := t.TempDir()
	tasksDir := filepath.Join(root, "tasks")
	writeTaskSpec(t, tasksDir, marketParityCoreTaskSpec())

	// When
	tasks, err := localAgentBenchmarkTasks(context.Background(), LocalAgentBenchmarkRequest{
		TasksDir:        tasksDir,
		BenchmarkTaskID: "market-parity-core",
	})
	// Then
	if err != nil {
		t.Fatalf("localAgentBenchmarkTasks returned error: %v", err)
	}
	if len(tasks) != 10 {
		t.Fatalf("tasks length = %d, want 10", len(tasks))
	}
	if tasks[0].ID != "bugfix-cli-timeout" || tasks[9].ID != "report-quality-evidence-summary" {
		t.Fatalf("suite task ids = %+v, want stable market parity core order", localAgentBenchmarkTaskIDs(tasks))
	}
}

func Test_LocalAgentBenchmarkTasks_expands_production_core_suite(t *testing.T) {
	// Given
	root := t.TempDir()
	tasksDir := filepath.Join(root, "tasks")
	payload, err := os.ReadFile(filepath.Join("..", "..", "evals", "tasks", "benchmark_tasks.json"))
	if err != nil {
		t.Fatalf("read benchmark tasks: %v", err)
	}
	writeTaskSpec(t, tasksDir, string(payload))

	// When
	tasks, err := localAgentBenchmarkTasks(context.Background(), LocalAgentBenchmarkRequest{
		TasksDir:        tasksDir,
		BenchmarkTaskID: "production-core",
	})
	// Then
	if err != nil {
		t.Fatalf("localAgentBenchmarkTasks returned error: %v", err)
	}
	if len(tasks) != 25 {
		t.Fatalf("tasks length = %d, want 25 production tasks", len(tasks))
	}
	if tasks[0].ID != "bugfix-cli-timeout" || tasks[24].ID != "report-quality-evidence-summary" {
		t.Fatalf("suite task ids = %+v, want stable production core order", localAgentBenchmarkTaskIDs(tasks))
	}
}

func Test_LocalAgentBenchmarkTasks_expands_cross_language_core_suite(t *testing.T) {
	// Given
	root := t.TempDir()
	tasksDir := filepath.Join(root, "tasks")
	payload, err := os.ReadFile(filepath.Join("..", "..", "evals", "tasks", "benchmark_tasks.json"))
	if err != nil {
		t.Fatalf("read benchmark tasks: %v", err)
	}
	writeTaskSpec(t, tasksDir, string(payload))

	// When
	tasks, err := localAgentBenchmarkTasks(context.Background(), LocalAgentBenchmarkRequest{
		TasksDir:        tasksDir,
		BenchmarkTaskID: "cross-language-core",
	})
	// Then
	if err != nil {
		t.Fatalf("localAgentBenchmarkTasks returned error: %v", err)
	}
	if len(tasks) != 2 {
		t.Fatalf("tasks length = %d, want 2 cross-language tasks", len(tasks))
	}
	if tasks[0].ID != "cross-language-js-state-reducer" || tasks[1].ID != "cross-language-python-retry-policy" {
		t.Fatalf("suite task ids = %+v, want stable cross-language core order", localAgentBenchmarkTaskIDs(tasks))
	}
}

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

func concurrentMultiTaskAgentScript() string {
	return `#!/bin/sh
set -eu
barrier="${BENCHMARK_CONCURRENCY_DIR:?}"
mkdir -p "$barrier"
: > "$barrier/started-$$"
deadline=$(( $(date +%s) + 5 ))
while [ "$(find "$barrier" -name 'started-*' -type f | wc -l | tr -d ' ')" -lt 2 ]; do
  if [ "$(date +%s)" -ge "$deadline" ]; then
    echo "concurrency barrier timed out" >&2
    exit 2
  fi
  sleep 0.1
done
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

func marketParityCoreTaskSpec() string {
	return `[
		{"id":"bugfix-cli-timeout","category":"bug_fix","title":"Timeout","objective":"Fix timeout.","required_changed_files":["internal/cli/run.go"]},
		{"id":"docs-roadmap-cli-first","category":"docs","title":"Roadmap","objective":"Keep CLI first.","required_changed_files":["docs/ROADMAP.md"]},
		{"id":"refactor-model-selection-split","category":"refactor","title":"Model split","objective":"Split model selection.","required_changed_files":["internal/cli/model_selection.go"]},
		{"id":"test-repair-require-checks","category":"test_repair","title":"Require checks","objective":"Repair require checks.","required_changed_files":["internal/cli/require_checks_test.go"]},
		{"id":"provider-config-openai-compatible","category":"provider_config","title":"OpenAI compatible","objective":"Configure OpenAI compatible provider.","required_changed_files":["internal/cli/http_provider.go"]},
		{"id":"safety-policy-observe-no-write","category":"safety_policy","title":"Observe no write","objective":"Prove observe write policy leaves files unchanged.","required_changed_files":["internal/cli/write_policy.go"]},
		{"id":"safety-policy-path-escape","category":"safety_policy","title":"Path escape","objective":"Reject path escape writes.","required_changed_files":["internal/workspace/workspace.go"]},
		{"id":"recovery-resume-retry","category":"recovery","title":"Resume retry recovery","objective":"Repair resume/retry recovery reporting.","required_changed_files":["internal/cli/resume.go"]},
		{"id":"safety-policy-rollback-report","category":"rollback","title":"Rollback report","objective":"Prove rollback reports restore files honestly.","required_changed_files":["internal/cli/rollback.go"]},
		{"id":"report-quality-evidence-summary","category":"report_quality","title":"Report quality","objective":"Improve report quality evidence summaries.","required_changed_files":["docs/REPORT_SCHEMA.md"]}
	]`
}
