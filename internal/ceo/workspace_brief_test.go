package ceo

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"ceoharness/internal/subagent"
)

type workspaceBriefRunner struct{}

func (r workspaceBriefRunner) Run(ctx context.Context, packet subagent.TaskPacket) (subagent.Result, error) {
	if err := ctx.Err(); err != nil {
		return subagent.Result{}, err
	}
	return subagent.Result{
		AgentName:       packet.AgentName,
		Role:            packet.Role,
		Status:          "pass",
		Attempts:        1,
		ContextReceived: packet.ContextMode,
		ContextBytes:    len(packet.Task) + len(packet.WorkspaceBrief),
		Summary:         packet.WorkspaceBrief,
		Evidence:        []string{"workspace brief received"},
	}, nil
}

func Test_Runtime_RunJob_sends_workspace_brief_to_subagents_when_workspace_is_set(t *testing.T) {
	// Given
	runtime := NewRuntimeWithSubagentRunner(workspaceBriefRunner{})
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "app.go"), []byte("package main"), 0o644); err != nil {
		t.Fatalf("write app fixture: %v", err)
	}

	// When
	report, err := runtime.RunJob(context.Background(), JobRequest{
		Task:         "Fix workspace bug",
		WorkspaceDir: root,
	})

	// Then
	if err != nil {
		t.Fatalf("RunJob returned error: %v", err)
	}
	if report.WorkspaceBrief == nil {
		t.Fatal("WorkspaceBrief is nil")
	}
	if !strings.Contains(report.SubagentResults[0].Summary, "app.go") {
		t.Fatalf("subagent summary = %q, want workspace brief path", report.SubagentResults[0].Summary)
	}
	if !hasRunEvent(report.RunEvents, "workspace_brief") {
		t.Fatalf("RunEvents = %+v, want workspace_brief event", report.RunEvents)
	}
}

func Test_Runtime_RunJob_respects_workspace_brief_max_files(t *testing.T) {
	// Given
	runtime := NewRuntimeWithSubagentRunner(workspaceBriefRunner{})
	root := t.TempDir()
	for _, name := range []string{"app.go", "README.md", "notes.txt"} {
		if err := os.WriteFile(filepath.Join(root, name), []byte("content"), 0o644); err != nil {
			t.Fatalf("write %s fixture: %v", name, err)
		}
	}

	// When
	report, err := runtime.RunJob(context.Background(), JobRequest{
		Task:                   "Fix workspace bug",
		WorkspaceDir:           root,
		WorkspaceBriefMaxFiles: 1,
	})

	// Then
	if err != nil {
		t.Fatalf("RunJob returned error: %v", err)
	}
	if report.WorkspaceBrief == nil {
		t.Fatal("WorkspaceBrief is nil")
	}
	if len(report.WorkspaceBrief.Files) != 1 || !report.WorkspaceBrief.Truncated {
		t.Fatalf("WorkspaceBrief = %+v, want one shown file and truncated", report.WorkspaceBrief)
	}
	if !strings.Contains(report.SubagentResults[0].Summary, "shown=1") {
		t.Fatalf("subagent summary = %q, want shown=1", report.SubagentResults[0].Summary)
	}
}

func hasRunEvent(events []RunEvent, kind string) bool {
	for _, event := range events {
		if event.Kind == kind {
			return true
		}
	}
	return false
}
