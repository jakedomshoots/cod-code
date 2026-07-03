package eval

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
)

func Test_ScoreReport_returns_pass_for_happy_fixture(t *testing.T) {
	// Given
	task := fixtureTask()
	workspace := newFixtureWorkspace(t, true)
	reportPath := filepath.Join("testdata", "happy", "report.json")

	// When
	result, err := ScoreReport(context.Background(), ScoreRequest{
		Task:         task,
		ReportPath:   reportPath,
		WorkspaceDir: workspace,
	})
	// Then
	if err != nil {
		t.Fatalf("ScoreReport returned error: %v", err)
	}
	if result.Verdict != "pass" {
		t.Fatalf("Verdict = %q, want pass; checks = %+v", result.Verdict, result.Checks)
	}
}

func Test_ScoreReport_returns_partial_for_stale_artifact_fixture(t *testing.T) {
	// Given
	task := fixtureTask()
	workspace := newFixtureWorkspace(t, false)
	reportPath := filepath.Join("testdata", "adversarial", "stale-artifact", "report.json")

	// When
	result, err := ScoreReport(context.Background(), ScoreRequest{
		Task:         task,
		ReportPath:   reportPath,
		WorkspaceDir: workspace,
	})
	// Then
	if err != nil {
		t.Fatalf("ScoreReport returned error: %v", err)
	}
	if result.Verdict != "partial" {
		t.Fatalf("Verdict = %q, want partial", result.Verdict)
	}
	check := requireScoreCheck(t, result, "artifact:.omo/evidence/bugfix-cli-timeout.md")
	if check.Status != "fail" {
		t.Fatalf("artifact check = %+v, want fail", check)
	}
}

func Test_ScoreReport_returns_partial_for_misleading_success_fixture(t *testing.T) {
	// Given
	task := fixtureTask()
	workspace := newFixtureWorkspace(t, true)
	reportPath := filepath.Join("testdata", "adversarial", "misleading-success", "report.json")

	// When
	result, err := ScoreReport(context.Background(), ScoreRequest{
		Task:         task,
		ReportPath:   reportPath,
		WorkspaceDir: workspace,
	})
	// Then
	if err != nil {
		t.Fatalf("ScoreReport returned error: %v", err)
	}
	if result.Verdict != "partial" {
		t.Fatalf("Verdict = %q, want partial", result.Verdict)
	}
	check := requireScoreCheck(t, result, "command:go test ./internal/cli -count=1")
	if check.Status != "fail" {
		t.Fatalf("command check = %+v, want fail", check)
	}
}

func Test_ScoreReport_returns_invalid_report_error_for_corrupt_fixture(t *testing.T) {
	// Given
	task := fixtureTask()
	reportPath := filepath.Join("testdata", "corrupt", "report.json")

	// When
	_, err := ScoreReport(context.Background(), ScoreRequest{Task: task, ReportPath: reportPath})

	// Then
	if !errors.Is(err, ErrInvalidReport) {
		t.Fatalf("error = %v, want ErrInvalidReport", err)
	}
}

func Test_ScoreReport_returns_pass_for_dirty_worktree_happy_fixture(t *testing.T) {
	// Given
	task := dirtyFixtureTask()
	workspace := newDirtyEvalWorkspace(t)
	reportPath := filepath.Join("testdata", "dirty-worktree", "happy", "report.json")

	// When
	result, err := ScoreReport(context.Background(), ScoreRequest{
		Task:         task,
		ReportPath:   reportPath,
		WorkspaceDir: workspace,
	})
	// Then
	if err != nil {
		t.Fatalf("ScoreReport returned error: %v", err)
	}
	if result.Verdict != "pass" {
		t.Fatalf("Verdict = %q, want pass; checks = %+v", result.Verdict, result.Checks)
	}
}

func Test_ScoreReport_returns_partial_for_dirty_worktree_unreported_fixture(t *testing.T) {
	// Given
	task := dirtyFixtureTask()
	workspace := newDirtyEvalWorkspace(t)
	reportPath := filepath.Join("testdata", "dirty-worktree", "unreported", "report.json")

	// When
	result, err := ScoreReport(context.Background(), ScoreRequest{
		Task:         task,
		ReportPath:   reportPath,
		WorkspaceDir: workspace,
	})
	// Then
	if err != nil {
		t.Fatalf("ScoreReport returned error: %v", err)
	}
	if result.Verdict != "partial" {
		t.Fatalf("Verdict = %q, want partial", result.Verdict)
	}
	check := requireScoreCheck(t, result, "dirty_worktree_reported_files")
	if check.Status != "fail" {
		t.Fatalf("dirty worktree reported files check = %+v, want fail", check)
	}
}

func fixtureTask() Task {
	return Task{
		ID:                   "bugfix-cli-timeout",
		Category:             "bug_fix",
		Title:                "Fix CLI timeout handling",
		Objective:            "Make timed-out checks fail honestly.",
		RequiredChangedFiles: []string{"internal/cli/run.go"},
		RequiredCommands:     []string{"go test ./internal/cli -count=1"},
		RequiredArtifacts:    []string{".omo/evidence/bugfix-cli-timeout.md"},
		RequiredDiffTerms:    []string{"timeout"},
		RequiredReportFields: []string{"verification_contract.status"},
	}
}

func dirtyFixtureTask() Task {
	task := fixtureTask()
	task.DirtyWorktreeSensitive = true
	return task
}

func newFixtureWorkspace(t *testing.T, includeArtifact bool) string {
	t.Helper()
	workspace := t.TempDir()
	if includeArtifact {
		writeWorkspaceFile(t, workspace, ".omo/evidence/bugfix-cli-timeout.md", "evidence\n")
	}
	return workspace
}
