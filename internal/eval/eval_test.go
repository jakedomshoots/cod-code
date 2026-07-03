package eval

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func Test_LoadTasks_loads_valid_task_specs_when_directory_contains_json_tasks(t *testing.T) {
	// Given
	root := t.TempDir()
	tasksDir := filepath.Join(root, "tasks")
	if err := os.MkdirAll(tasksDir, 0o755); err != nil {
		t.Fatalf("create tasks dir: %v", err)
	}
	taskJSON := `{
		"id":"bugfix-cli-timeout",
		"category":"bug_fix",
		"title":"Fix CLI timeout handling",
		"objective":"Make timed-out checks fail honestly.",
		"required_changed_files":["internal/cli/run.go"],
		"required_commands":["go test ./internal/cli -count=1"],
		"required_artifacts":["evidence/run-log.md"],
		"required_diff_terms":["timeout"]
	}`
	if err := os.WriteFile(filepath.Join(tasksDir, "bugfix-cli-timeout.json"), []byte(taskJSON), 0o644); err != nil {
		t.Fatalf("write task: %v", err)
	}

	// When
	tasks, err := LoadTasks(context.Background(), tasksDir)

	// Then
	if err != nil {
		t.Fatalf("LoadTasks returned error: %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("len(tasks) = %d, want 1", len(tasks))
	}
	if tasks[0].ID != "bugfix-cli-timeout" || tasks[0].Category != "bug_fix" {
		t.Fatalf("task = %+v, want loaded ID/category", tasks[0])
	}
}

func Test_LoadRubric_accepts_artifact_first_rubric_when_required_sections_exist(t *testing.T) {
	// Given
	root := t.TempDir()
	rubricPath := filepath.Join(root, "rubric.md")
	rubric := `# CEO Harness Evaluation Rubric

## Artifact-first scoring
Score only saved reports, diffs, command results, and artifacts.

## Self-report exclusion
Never score from model self-report or claimed success.

## Scoring dimensions
- Task fit
- Verification
- Safety

## Verdicts
pass, partial, fail

## Evidence paths
Every score points to real files.
`
	if err := os.WriteFile(rubricPath, []byte(rubric), 0o644); err != nil {
		t.Fatalf("write rubric: %v", err)
	}

	// When
	loaded, err := LoadRubric(rubricPath)

	// Then
	if err != nil {
		t.Fatalf("LoadRubric returned error: %v", err)
	}
	if loaded.Path != rubricPath {
		t.Fatalf("Path = %q, want %q", loaded.Path, rubricPath)
	}
}

func Test_ScoreReport_returns_pass_when_fixture_report_has_artifacts_diffs_and_passing_checks(t *testing.T) {
	// Given
	root := t.TempDir()
	evidenceDir := filepath.Join(root, "evidence")
	if err := os.MkdirAll(evidenceDir, 0o755); err != nil {
		t.Fatalf("create evidence dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(evidenceDir, "run-log.md"), []byte("go test output"), 0o644); err != nil {
		t.Fatalf("write evidence: %v", err)
	}
	reportPath := filepath.Join(root, "report.json")
	reportJSON := `{
		"task_id":"bugfix-cli-timeout",
		"verdict":"pass",
		"changed_files":["internal/cli/run.go","internal/cli/run_test.go"],
		"check_results":[{"argv":["go","test","./internal/cli","-count=1"],"status":"pass","exit_code":0,"stdout":"ok"}],
		"patch_results":[{"path":"internal/cli/run.go","diff":"@@\n+timeout handling returns failure"}],
		"evidence_paths":["evidence/run-log.md"],
		"verification_contract":{"status":"pass"}
	}`
	if err := os.WriteFile(reportPath, []byte(reportJSON), 0o644); err != nil {
		t.Fatalf("write report: %v", err)
	}
	task := Task{
		ID:                   "bugfix-cli-timeout",
		Category:             "bug_fix",
		Title:                "Fix CLI timeout handling",
		Objective:            "Make timed-out checks fail honestly.",
		RequiredChangedFiles: []string{"internal/cli/run.go"},
		RequiredCommands:     []string{"go test ./internal/cli -count=1"},
		RequiredArtifacts:    []string{"evidence/run-log.md"},
		RequiredDiffTerms:    []string{"timeout handling"},
		RequiredReportFields: []string{"verification_contract.status"},
	}

	// When
	result, err := ScoreReport(context.Background(), ScoreRequest{Task: task, ReportPath: reportPath})

	// Then
	if err != nil {
		t.Fatalf("ScoreReport returned error: %v", err)
	}
	if result.TaskID != "bugfix-cli-timeout" || result.Verdict != "pass" {
		t.Fatalf("result = %+v, want passing task verdict", result)
	}
	if len(result.Checks) == 0 {
		t.Fatalf("checks = %#v, want artifact/check/diff checks", result.Checks)
	}
	if len(result.EvidencePaths) != 1 || result.EvidencePaths[0] != "evidence/run-log.md" {
		t.Fatalf("evidence paths = %#v, want fixture evidence path", result.EvidencePaths)
	}
}
