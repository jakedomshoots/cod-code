package eval

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func Test_RunBenchmarkFixtures_writes_scores_for_all_tasks_when_specs_are_valid(t *testing.T) {
	// Given
	root := t.TempDir()
	tasksDir := filepath.Join(root, "tasks")
	writeTaskSpec(t, tasksDir, `[
		{
			"id":"bugfix-cli-timeout",
			"category":"bug_fix",
			"title":"Fix CLI timeout handling",
			"objective":"Make timed-out checks fail honestly.",
			"dirty_worktree_sensitive":true,
			"required_changed_files":["internal/cli/run.go"],
			"required_commands":["go test ./internal/cli -count=1"],
			"required_artifacts":[".omo/evidence/bugfix-cli-timeout.md"],
			"required_diff_terms":["timeout"],
			"required_report_fields":["verification_contract.status"]
		},
		{
			"id":"docs-verification-record",
			"category":"docs",
			"title":"Update verification record",
			"objective":"Document validation results.",
			"required_changed_files":["docs/VERIFICATION.md"],
			"required_commands":["go test ./internal/cli -count=1"],
			"required_artifacts":[".omo/evidence/docs-verification-record.md"],
			"required_diff_terms":["Tooling Not Available"],
			"required_report_fields":["check_results"]
		}
	]`)
	outputDir := filepath.Join(root, "out")

	// When
	summary, err := RunBenchmarkFixtures(context.Background(), BenchmarkFixtureRequest{
		TasksDir:   tasksDir,
		OutputDir:  outputDir,
		ReportMode: "deterministic_fixture_scoring",
	})
	// Then
	if err != nil {
		t.Fatalf("RunBenchmarkFixtures returned error: %v", err)
	}
	if summary.TaskCount != 2 || summary.Passed != 2 || summary.Failed != 0 {
		t.Fatalf("summary = %+v, want 2 passing task scores", summary)
	}
	for _, id := range []string{"bugfix-cli-timeout", "docs-verification-record"} {
		requireFile(t, filepath.Join(outputDir, id, "report.json"))
		requireFile(t, filepath.Join(outputDir, id, "score.json"))
		requireFile(t, filepath.Join(outputDir, id, "score.log"))
	}
}

func Test_RunBenchmarkFixtures_can_rerun_dirty_workspace_output_dir(t *testing.T) {
	// Given
	root := t.TempDir()
	tasksDir := filepath.Join(root, "tasks")
	writeTaskSpec(t, tasksDir, `{
		"id":"bugfix-cli-timeout",
		"category":"bug_fix",
		"title":"Fix CLI timeout handling",
		"objective":"Make timed-out checks fail honestly.",
		"dirty_worktree_sensitive":true,
		"required_changed_files":["internal/cli/run.go"],
		"required_commands":["go test ./internal/cli -count=1"],
		"required_artifacts":[".omo/evidence/bugfix-cli-timeout.md"],
		"required_diff_terms":["timeout"],
		"required_report_fields":["verification_contract.status"]
	}`)
	outputDir := filepath.Join(root, "out")
	req := BenchmarkFixtureRequest{
		TasksDir:   tasksDir,
		OutputDir:  outputDir,
		ReportMode: "deterministic_fixture_scoring",
	}

	// When
	first, firstErr := RunBenchmarkFixtures(context.Background(), req)
	second, secondErr := RunBenchmarkFixtures(context.Background(), req)

	// Then
	if firstErr != nil {
		t.Fatalf("first RunBenchmarkFixtures returned error: %v", firstErr)
	}
	if secondErr != nil {
		t.Fatalf("second RunBenchmarkFixtures returned error: %v", secondErr)
	}
	if first.Passed != 1 || second.Passed != 1 || second.Skipped != 0 {
		t.Fatalf("first = %+v second = %+v, want both reruns passing", first, second)
	}
}

func Test_RunCLI_runs_benchmark_fixtures_when_flag_is_set(t *testing.T) {
	// Given
	root := t.TempDir()
	tasksDir := filepath.Join(root, "tasks")
	writeTaskSpec(t, tasksDir, `{
		"id":"docs-roadmap-cli-first",
		"category":"docs",
		"title":"Keep roadmap CLI-first",
		"objective":"Refresh roadmap wording.",
		"required_changed_files":["docs/ROADMAP.md"],
		"required_commands":["go test ./internal/cli -count=1"],
		"required_artifacts":[".omo/evidence/docs-roadmap-cli-first.md"],
		"required_diff_terms":["CLI-first"],
		"required_report_fields":["changed_files"]
	}`)
	outputDir := filepath.Join(root, "benchmark")

	// When
	err := RunCLI(context.Background(), os.Stdout, os.Stderr, []string{
		"--benchmark-fixtures",
		"--tasks", tasksDir,
		"--output-dir", outputDir,
	})
	// Then
	if err != nil {
		t.Fatalf("RunCLI returned error: %v", err)
	}
	requireFile(t, filepath.Join(outputDir, "summary.json"))
	requireFile(t, filepath.Join(outputDir, "docs-roadmap-cli-first", "score.json"))
}

func writeTaskSpec(t *testing.T, tasksDir string, content string) {
	t.Helper()
	if err := os.MkdirAll(tasksDir, 0o755); err != nil {
		t.Fatalf("create tasks dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tasksDir, "tasks.json"), []byte(content), 0o644); err != nil {
		t.Fatalf("write task spec: %v", err)
	}
}

func requireFile(t *testing.T, path string) {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat %s: %v", path, err)
	}
	if info.IsDir() {
		t.Fatalf("%s is a directory, want file", path)
	}
	if info.Size() == 0 {
		t.Fatalf("%s is empty", path)
	}
}
