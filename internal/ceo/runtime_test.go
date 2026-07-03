package ceo

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func Test_Runtime_Run_delegates_to_all_native_subagents_when_task_is_valid(t *testing.T) {
	// Given
	runtime := NewRuntime()

	// When
	report, err := runtime.Run(context.Background(), "Fix a failing test")
	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if report.JobPacket.Task != "Fix a failing test" {
		t.Fatalf("Task = %q, want task", report.JobPacket.Task)
	}
	if len(report.SubagentResults) != 3 {
		t.Fatalf("SubagentResults length = %d, want 3", len(report.SubagentResults))
	}
	wantAgents := []string{"scanner", "coder", "reviewer"}
	for index, wantAgent := range wantAgents {
		if report.SubagentResults[index].AgentName != wantAgent {
			t.Fatalf("SubagentResults[%d].AgentName = %q, want %q", index, report.SubagentResults[index].AgentName, wantAgent)
		}
		if report.SubagentResults[index].ContextReceived != "lean" {
			t.Fatalf("SubagentResults[%d].ContextReceived = %q, want lean", index, report.SubagentResults[index].ContextReceived)
		}
	}
	if report.Verdict != "pass" {
		t.Fatalf("Verdict = %q, want pass", report.Verdict)
	}
}

func Test_Runtime_RunJob_writes_subagent_evidence_when_workspace_is_set(t *testing.T) {
	// Given
	runtime := NewRuntime()
	root := t.TempDir()

	// When
	report, err := runtime.RunJob(context.Background(), JobRequest{
		Task:         "Fix a failing test",
		WorkspaceDir: root,
	})
	// Then
	if err != nil {
		t.Fatalf("RunJob returned error: %v", err)
	}
	if len(report.ChangedFiles) != 4 {
		t.Fatalf("ChangedFiles length = %d, want 4", len(report.ChangedFiles))
	}
	if report.ChangedFiles[0] != "ceo-artifacts/scanner.md" {
		t.Fatalf("ChangedFiles[0] = %q, want ceo-artifacts/scanner.md", report.ChangedFiles[0])
	}
	wantFiles := []string{"scanner.md", "coder.md", "reviewer.md", "ceo-plan.md"}
	for _, wantFile := range wantFiles {
		got, err := os.ReadFile(filepath.Join(root, "ceo-artifacts", wantFile))
		if err != nil {
			t.Fatalf("read evidence file %s: %v", wantFile, err)
		}
		if string(got) == "" {
			t.Fatalf("expected evidence file content for %s", wantFile)
		}
	}
}

func Test_Runtime_RunJob_appends_history_when_workspace_is_set(t *testing.T) {
	// Given
	runtime := NewRuntime()
	root := t.TempDir()

	// When
	report, err := runtime.RunJob(context.Background(), JobRequest{
		Task:         "Fix a failing test",
		WorkspaceDir: root,
	})
	// Then
	if err != nil {
		t.Fatalf("RunJob returned error: %v", err)
	}
	if report.HistoryPath != "ceo-artifacts/jobs.jsonl" {
		t.Fatalf("HistoryPath = %q, want ceo-artifacts/jobs.jsonl", report.HistoryPath)
	}
	if report.JobID != "job-000001" {
		t.Fatalf("JobID = %q, want job-000001", report.JobID)
	}
	got, err := os.ReadFile(filepath.Join(root, "ceo-artifacts", "jobs.jsonl"))
	if err != nil {
		t.Fatalf("read history file: %v", err)
	}
	if !strings.Contains(string(got), `"task":"Fix a failing test"`) {
		t.Fatalf("history = %q, want task", string(got))
	}
	if !strings.Contains(string(got), `"id":"job-000001"`) {
		t.Fatalf("history = %q, want job id", string(got))
	}
	if !strings.Contains(string(got), `"execution_plan_step_count":4`) {
		t.Fatalf("history = %q, want execution plan step count", string(got))
	}
	if !strings.Contains(string(got), `"execution_plan_next_action":"accept"`) {
		t.Fatalf("history = %q, want execution plan next action", string(got))
	}
}

func Test_Runtime_RunJob_saves_report_snapshot_when_workspace_is_set(t *testing.T) {
	// Given
	runtime := NewRuntime()
	root := t.TempDir()

	// When
	report, err := runtime.RunJob(context.Background(), JobRequest{
		Task:         "Fix a failing test",
		WorkspaceDir: root,
	})
	// Then
	if err != nil {
		t.Fatalf("RunJob returned error: %v", err)
	}
	got, err := os.ReadFile(filepath.Join(root, "ceo-artifacts", "jobs", "job-000001.json"))
	if err != nil {
		t.Fatalf("read report snapshot: %v", err)
	}
	if !strings.Contains(string(got), `"job_id": "job-000001"`) {
		t.Fatalf("snapshot = %q, want job id", string(got))
	}
	if !strings.Contains(string(got), `"verdict": "pass"`) {
		t.Fatalf("snapshot = %q, want verdict", string(got))
	}
	if report.JobID != "job-000001" {
		t.Fatalf("JobID = %q, want job-000001", report.JobID)
	}
}

func Test_Runtime_RunJob_includes_check_result_when_check_command_is_set(t *testing.T) {
	// Given
	runtime := NewRuntime()

	// When
	report, err := runtime.RunJob(context.Background(), JobRequest{
		Task: "Fix a failing test",
		CheckCommand: []string{
			"go",
			"version",
		},
	})
	// Then
	if err != nil {
		t.Fatalf("RunJob returned error: %v", err)
	}
	if len(report.CheckResults) != 1 {
		t.Fatalf("CheckResults length = %d, want 1", len(report.CheckResults))
	}
	if report.CheckResults[0].Status != "pass" {
		t.Fatalf("Check status = %q, want pass", report.CheckResults[0].Status)
	}
}

func Test_Runtime_RunJob_marks_verdict_fail_when_check_command_fails(t *testing.T) {
	// Given
	runtime := NewRuntime()

	// When
	report, err := runtime.RunJob(context.Background(), JobRequest{
		Task: "Fix a failing test",
		CheckCommand: []string{
			os.Args[0],
			"-test.run=Test_HelperProcess_fail_check",
		},
		CheckEnv: []string{"GO_WANT_CEO_HELPER_PROCESS=fail"},
	})
	// Then
	if err != nil {
		t.Fatalf("RunJob returned error: %v", err)
	}
	if report.Verdict != "fail" {
		t.Fatalf("Verdict = %q, want fail", report.Verdict)
	}
	if len(report.CheckResults) != 1 {
		t.Fatalf("CheckResults length = %d, want 1", len(report.CheckResults))
	}
	if report.CheckResults[0].Status != "fail" {
		t.Fatalf("Check status = %q, want fail", report.CheckResults[0].Status)
	}
}

func Test_Runtime_RunJob_applies_text_patch_when_patch_request_is_set(t *testing.T) {
	// Given
	runtime := NewRuntime()
	root := t.TempDir()
	target := filepath.Join(root, "app.txt")
	if err := os.WriteFile(target, []byte("hello old"), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	// When
	report, err := runtime.RunJob(context.Background(), JobRequest{
		Task:         "Patch app text",
		WorkspaceDir: root,
		Patches: []PatchRequest{
			{Path: "app.txt", Old: "old", New: "new"},
		},
	})
	// Then
	if err != nil {
		t.Fatalf("RunJob returned error: %v", err)
	}
	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read patched file: %v", err)
	}
	if string(got) != "hello new" {
		t.Fatalf("content = %q, want hello new", string(got))
	}
	if len(report.PatchResults) != 1 {
		t.Fatalf("PatchResults length = %d, want 1", len(report.PatchResults))
	}
	if report.PatchResults[0].Path != "app.txt" {
		t.Fatalf("PatchResults[0].Path = %q, want app.txt", report.PatchResults[0].Path)
	}
	if report.PatchResults[0].Diff == "" {
		t.Fatal("expected patch diff")
	}
}

func Test_HelperProcess_fail_check(t *testing.T) {
	if os.Getenv("GO_WANT_CEO_HELPER_PROCESS") != "fail" {
		return
	}
	os.Stderr.WriteString("check failed\n")
	os.Exit(9)
}

func Test_HelperProcess_retry_check(t *testing.T) {
	if os.Getenv("GO_WANT_CEO_HELPER_PROCESS") != "retry" {
		return
	}
	statePath := os.Getenv("GO_CEO_RETRY_STATE")
	if _, err := os.Stat(statePath); os.IsNotExist(err) {
		if writeErr := os.WriteFile(statePath, []byte("failed once"), 0o644); writeErr != nil {
			t.Fatalf("write retry state: %v", writeErr)
		}
		os.Stderr.WriteString("first attempt failed\n")
		os.Exit(6)
	}
	os.Stdout.WriteString("second attempt passed\n")
	os.Exit(0)
}
