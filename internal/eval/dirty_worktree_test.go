package eval

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"testing"
)

func Test_ScoreReport_returns_partial_when_dirty_worktree_sensitive_report_missing_git_status_evidence(t *testing.T) {
	// Given
	root := t.TempDir()
	task := loadDirtySensitiveTask(t, root)
	workspace := newDirtyEvalWorkspace(t)
	reportPath := writeEvalReport(t, root, `{
		"task_id":"bugfix-cli-timeout",
		"verdict":"pass",
		"changed_files":["internal/cli/run.go","internal/cli/run_test.go"],
		"check_results":[{"argv":["go","test","./internal/cli","-count=1"],"status":"pass","exit_code":0,"stdout":"ok"}],
		"patch_results":[{"path":"internal/cli/run.go","diff":"@@\n+timeout handling returns failure"}],
		"evidence_paths":[".omo/evidence/bugfix-cli-timeout.md"],
		"verification_contract":{"status":"pass"}
	}`)

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
	check := requireScoreCheck(t, result, "dirty_worktree_evidence")
	if check.Status != "fail" {
		t.Fatalf("dirty worktree evidence check = %+v, want fail", check)
	}
}

func Test_ScoreReport_returns_partial_when_dirty_worktree_sensitive_report_claims_clean_status_for_dirty_workspace(t *testing.T) {
	// Given
	root := t.TempDir()
	task := loadDirtySensitiveTask(t, root)
	workspace := newDirtyEvalWorkspace(t)
	reportPath := writeEvalReport(t, root, `{
		"task_id":"bugfix-cli-timeout",
		"verdict":"pass",
		"changed_files":["internal/cli/run.go","internal/cli/run_test.go"],
		"check_results":[{"argv":["go","test","./internal/cli","-count=1"],"status":"pass","exit_code":0,"stdout":"ok"}],
		"patch_results":[{"path":"internal/cli/run.go","diff":"@@\n+timeout handling returns failure"}],
		"evidence_paths":[".omo/evidence/bugfix-cli-timeout.md"],
		"verification_contract":{"status":"pass"},
		"worktree_status":{
			"source":"git",
			"command":["git","status","--porcelain=v1","--untracked-files=all"],
			"payload":"",
			"sha256":"`+sha256Hex("")+`"
		}
	}`)

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
	check := requireScoreCheck(t, result, "dirty_worktree_status")
	if check.Status != "fail" {
		t.Fatalf("dirty worktree status check = %+v, want fail", check)
	}
}

func Test_ScoreReport_returns_partial_when_dirty_worktree_status_contains_unreported_file(t *testing.T) {
	// Given
	root := t.TempDir()
	task := loadDirtySensitiveTask(t, root)
	workspace := newDirtyEvalWorkspace(t)
	statusPayload := gitStatusPayload(t, workspace)
	reportPath := writeEvalReport(t, root, `{
		"task_id":"bugfix-cli-timeout",
		"verdict":"pass",
		"changed_files":["internal/cli/run.go"],
		"check_results":[{"argv":["go","test","./internal/cli","-count=1"],"status":"pass","exit_code":0,"stdout":"ok"}],
		"patch_results":[{"path":"internal/cli/run.go","diff":"@@\n+timeout handling returns failure"}],
		"evidence_paths":[".omo/evidence/bugfix-cli-timeout.md"],
		"verification_contract":{"status":"pass"},
		"worktree_status":{
			"source":"git",
			"command":["git","status","--porcelain=v1","--untracked-files=all"],
			"payload":`+strconv.Quote(statusPayload)+`,
			"sha256":"`+sha256Hex(statusPayload)+`"
		}
	}`)

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

func loadDirtySensitiveTask(t *testing.T, root string) Task {
	t.Helper()
	tasksDir := filepath.Join(root, "tasks")
	if err := os.MkdirAll(tasksDir, 0o755); err != nil {
		t.Fatalf("create tasks dir: %v", err)
	}
	taskJSON := `{
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
	}`
	if err := os.WriteFile(filepath.Join(tasksDir, "bugfix-cli-timeout.json"), []byte(taskJSON), 0o644); err != nil {
		t.Fatalf("write task: %v", err)
	}
	tasks, err := LoadTasks(context.Background(), tasksDir)
	if err != nil {
		t.Fatalf("LoadTasks returned error: %v", err)
	}
	task, err := FindTask(tasks, "bugfix-cli-timeout")
	if err != nil {
		t.Fatalf("FindTask returned error: %v", err)
	}
	return task
}

func newDirtyEvalWorkspace(t *testing.T) string {
	t.Helper()
	workspace := t.TempDir()
	runGit(t, workspace, "init")
	runGit(t, workspace, "config", "user.email", "eval@example.com")
	runGit(t, workspace, "config", "user.name", "Eval Test")
	writeWorkspaceFile(t, workspace, "internal/cli/run.go", "package cli\n")
	writeWorkspaceFile(t, workspace, ".omo/evidence/bugfix-cli-timeout.md", "evidence\n")
	runGit(t, workspace, "add", ".")
	runGit(t, workspace, "commit", "-m", "base")
	writeWorkspaceFile(t, workspace, "internal/cli/run.go", "package cli\n\n// timeout handling returns failure\n")
	writeWorkspaceFile(t, workspace, "internal/cli/run_test.go", "package cli\n")
	return workspace
}

func writeEvalReport(t *testing.T, root string, content string) string {
	t.Helper()
	reportPath := filepath.Join(root, "report.json")
	if err := os.WriteFile(reportPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write report: %v", err)
	}
	return reportPath
}

func writeWorkspaceFile(t *testing.T, root string, name string, content string) {
	t.Helper()
	path := filepath.Join(root, name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("create parent for %s: %v", name, err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v failed: %v\n%s", args, err, output)
	}
}

func gitStatusPayload(t *testing.T, dir string) string {
	t.Helper()
	cmd := exec.Command("git", "status", "--porcelain=v1", "--untracked-files=all")
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git status failed: %v\n%s", err, output)
	}
	return string(output)
}

func sha256Hex(payload string) string {
	sum := sha256.Sum256([]byte(payload))
	return hex.EncodeToString(sum[:])
}

func requireScoreCheck(t *testing.T, result ScoreResult, name string) CheckResult {
	t.Helper()
	for _, check := range result.Checks {
		if check.Name == name {
			return check
		}
	}
	t.Fatalf("checks = %+v, missing %s", result.Checks, name)
	return CheckResult{}
}
