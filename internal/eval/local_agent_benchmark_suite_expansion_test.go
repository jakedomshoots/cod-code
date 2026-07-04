package eval

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

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
	if len(tasks) != 30 {
		t.Fatalf("tasks length = %d, want 30 production tasks", len(tasks))
	}
	if tasks[0].ID != "bugfix-cli-timeout" ||
		tasks[24].ID != "multi-file-operator-safety-flow" ||
		tasks[25].ID != "multi-file-release-readiness-publish-boundary" ||
		tasks[26].ID != "multi-file-lean-context-autonomy" ||
		tasks[27].ID != "multi-file-secret-safe-provider-proof" ||
		tasks[28].ID != "multi-file-finalizer-check-fix" ||
		tasks[29].ID != "report-quality-evidence-summary" {
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
