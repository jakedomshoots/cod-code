package eval

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func Test_PrepareLocalAgentBenchmarkWorkspace_writes_compile_safe_go_fixture(t *testing.T) {
	// Given
	workspaceDir := filepath.Join(t.TempDir(), "workspace")
	task := Task{
		ID:                   "bugfix-cli-timeout",
		Category:             "bug_fix",
		Title:                "Timeout failure is reported honestly",
		Objective:            "Fix command timeout handling.",
		RequiredChangedFiles: []string{"internal/cli/run.go"},
		RequiredCommands:     []string{"go test ./internal/cli -count=1"},
		RequiredDiffTerms:    []string{"timeout"},
	}

	// When
	err := prepareLocalAgentBenchmarkWorkspace(context.Background(), workspaceDir, task, nil)

	// Then
	if err != nil {
		t.Fatalf("prepareLocalAgentBenchmarkWorkspace returned error: %v", err)
	}
	content, readErr := os.ReadFile(filepath.Join(workspaceDir, "internal/cli/run.go"))
	if readErr != nil {
		t.Fatalf("read Go fixture: %v", readErr)
	}
	if !strings.Contains(string(content), "package cli") {
		t.Fatalf("fixture = %q, want package cli", string(content))
	}
	run := runLocalAgentCommand(context.Background(), []string{"go", "test", "./internal/cli", "-count=1"}, workspaceDir, nil, localAgentTimeout(5))
	if run.exitCode != 0 || run.errText != "" {
		t.Fatalf("go fixture did not compile: exit=%d err=%q stderr=%q", run.exitCode, run.errText, run.stderr)
	}
}

func Test_PrepareLocalAgentBenchmarkWorkspace_writes_real_path_escape_fixture(t *testing.T) {
	// Given
	workspaceDir := filepath.Join(t.TempDir(), "workspace")
	task := Task{
		ID:                   "safety-policy-path-escape",
		Category:             "safety_policy",
		Title:                "Reject path escape writes",
		Objective:            "Ensure patch/create requests cannot write outside the workspace root.",
		RequiredChangedFiles: []string{"internal/workspace/workspace.go"},
		RequiredCommands:     []string{"go test ./internal/workspace -run Test_.*[Pp]ath -count=1"},
		RequiredArtifacts:    []string{".omo/evidence/safety-policy-path-escape.md"},
		RequiredDiffTerms:    []string{"path escapes workspace"},
	}

	// When
	err := prepareLocalAgentBenchmarkWorkspace(context.Background(), workspaceDir, task, nil)

	// Then
	if err != nil {
		t.Fatalf("prepareLocalAgentBenchmarkWorkspace returned error: %v", err)
	}
	baselineRun := runLocalAgentCommand(context.Background(), []string{"go", "test", "./internal/workspace", "-run", "Test_.*[Pp]ath", "-count=1"}, workspaceDir, nil, localAgentTimeout(5))
	if baselineRun.exitCode == 0 {
		t.Fatalf("baseline safety fixture unexpectedly passed: stdout=%q stderr=%q", baselineRun.stdout, baselineRun.stderr)
	}
	if err := os.WriteFile(filepath.Join(workspaceDir, "internal/workspace/workspace.go"), []byte(benchmarkExpectedText(task, "internal/workspace/workspace.go")), 0o644); err != nil {
		t.Fatalf("write expected safety fixture: %v", err)
	}
	fixedRun := runLocalAgentCommand(context.Background(), []string{"go", "test", "./internal/workspace", "-run", "Test_.*[Pp]ath", "-count=1"}, workspaceDir, nil, localAgentTimeout(5))
	if fixedRun.exitCode != 0 || fixedRun.errText != "" {
		t.Fatalf("fixed safety fixture failed: exit=%d err=%q stderr=%q", fixedRun.exitCode, fixedRun.errText, fixedRun.stderr)
	}
}

func Test_PrepareLocalAgentBenchmarkWorkspace_can_rerun_same_workspace(t *testing.T) {
	// Given
	workspaceDir := filepath.Join(t.TempDir(), "workspace")
	task := Task{
		ID:                   "safety-policy-path-escape",
		Category:             "safety_policy",
		Title:                "Reject path escape writes",
		Objective:            "Ensure patch/create requests cannot write outside the workspace root.",
		RequiredChangedFiles: []string{"internal/workspace/workspace.go"},
		RequiredCommands:     []string{"go test ./internal/workspace -run Test_.*[Pp]ath -count=1"},
		RequiredArtifacts:    []string{".omo/evidence/safety-policy-path-escape.md"},
		RequiredDiffTerms:    []string{"path escapes workspace"},
	}

	// When
	firstErr := prepareLocalAgentBenchmarkWorkspace(context.Background(), workspaceDir, task, nil)
	secondErr := prepareLocalAgentBenchmarkWorkspace(context.Background(), workspaceDir, task, nil)

	// Then
	if firstErr != nil {
		t.Fatalf("first prepareLocalAgentBenchmarkWorkspace returned error: %v", firstErr)
	}
	if secondErr != nil {
		t.Fatalf("second prepareLocalAgentBenchmarkWorkspace returned error: %v", secondErr)
	}
}

func Test_RunLocalAgentBenchmark_scores_dirty_worktree_go_task_when_agent_changes_fixture(t *testing.T) {
	// Given
	binDir := t.TempDir()
	writeExecutableContent(t, filepath.Join(binDir, "codex"), `#!/bin/sh
cat > internal/cli/run.go <<'EOF'
package cli

const benchmarkFixture = "timeout"
EOF
mkdir -p .omo/evidence
printf 'agent evidence\n' > .omo/evidence/bugfix-cli-timeout.md
printf 'done\n'
`)
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	root := t.TempDir()
	tasksDir := filepath.Join(root, "tasks")
	writeTaskSpec(t, tasksDir, dirtyGoTaskSpec())

	// When
	summary, err := RunLocalAgentBenchmark(context.Background(), LocalAgentBenchmarkRequest{
		TasksDir:        tasksDir,
		OutputDir:       filepath.Join(root, "benchmark"),
		TimeoutSeconds:  5,
		Agents:          []string{"codex_cli"},
		BenchmarkTaskID: "bugfix-cli-timeout",
	})

	// Then
	if err != nil {
		t.Fatalf("RunLocalAgentBenchmark returned error: %v", err)
	}
	result := summary.Results[0]
	if result.Status != localAgentStatusPass || result.PassedChecks != result.TotalChecks {
		t.Fatalf("result = %+v, want dirty Go task to score pass", result)
	}
}

func Test_RunLocalAgentBenchmark_writes_evidence_when_workspace_prepare_fails(t *testing.T) {
	// Given
	root := t.TempDir()
	tasksDir := filepath.Join(root, "tasks")
	writeTaskSpec(t, tasksDir, `{
		"id":"bad-workspace",
		"category":"docs",
		"title":"Bad workspace fixture",
		"objective":"Trigger workspace preparation failure.",
		"required_changed_files":["../escape.md"],
		"required_diff_terms":["escape"]
	}`)

	// When
	summary, err := RunLocalAgentBenchmark(context.Background(), LocalAgentBenchmarkRequest{
		TasksDir:        tasksDir,
		OutputDir:       filepath.Join(root, "benchmark"),
		TimeoutSeconds:  5,
		Agents:          []string{"ceo_harness"},
		BenchmarkTaskID: "bad-workspace",
	})

	// Then
	if err != nil {
		t.Fatalf("RunLocalAgentBenchmark returned error: %v", err)
	}
	if summary.Failed != 1 || summary.IncompleteEvidence != 1 {
		t.Fatalf("summary = %+v, want one failed incomplete-evidence result", summary)
	}
	result := summary.Results[0]
	for _, path := range []string{
		result.CommandPath,
		result.StdoutPath,
		result.StderrPath,
		result.ReportPath,
		result.ScorePath,
		result.DiffPath,
		result.ChangedFilesPath,
		result.GitBeforePath,
		result.GitAfterPath,
		result.TimingPath,
	} {
		requireFile(t, path)
	}
}

func Test_RunLocalAgentBenchmark_writes_evidence_when_agent_binary_is_missing(t *testing.T) {
	// Given
	root := t.TempDir()
	tasksDir := filepath.Join(root, "tasks")
	writeTaskSpec(t, tasksDir, `{
		"id":"docs-one",
		"category":"docs",
		"title":"Docs task",
		"objective":"Refresh docs.",
		"required_changed_files":["docs/ONE.md"],
		"required_artifacts":[".omo/evidence/docs-one.md"],
		"required_diff_terms":["one-term"]
	}`)

	// When
	summary, err := RunLocalAgentBenchmark(context.Background(), LocalAgentBenchmarkRequest{
		TasksDir:         tasksDir,
		OutputDir:        filepath.Join(root, "benchmark"),
		TimeoutSeconds:   5,
		Agents:           []string{"ceo_harness"},
		CEOHarnessBinary: filepath.Join(root, "missing-ceo-packet"),
		BenchmarkTaskID:  "docs-one",
	})

	// Then
	if err != nil {
		t.Fatalf("RunLocalAgentBenchmark returned error: %v", err)
	}
	if summary.Skipped != 1 || summary.IncompleteEvidence != 1 {
		t.Fatalf("summary = %+v, want one skipped incomplete-evidence result", summary)
	}
	result := summary.Results[0]
	if result.EvidenceStatus != localAgentEvidenceIncomplete {
		t.Fatalf("EvidenceStatus = %q, want incomplete", result.EvidenceStatus)
	}
	for _, path := range []string{
		result.CommandPath,
		result.StdoutPath,
		result.StderrPath,
		result.ReportPath,
		result.ScorePath,
		result.DiffPath,
		result.ChangedFilesPath,
		result.GitBeforePath,
		result.GitAfterPath,
		result.TimingPath,
	} {
		requireFile(t, path)
	}
}

func dirtyGoTaskSpec() string {
	return `{
		"id":"bugfix-cli-timeout",
		"category":"bug_fix",
		"title":"Timeout failure is reported honestly",
		"objective":"Fix command timeout handling.",
		"dirty_worktree_sensitive":true,
		"required_changed_files":["internal/cli/run.go"],
		"required_commands":["go test ./internal/cli -count=1"],
		"required_artifacts":[".omo/evidence/bugfix-cli-timeout.md"],
		"required_diff_terms":["timeout"],
		"required_report_fields":["verification_contract.status"]
	}`
}
