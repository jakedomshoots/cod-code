package ceo

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func Test_Runtime_RunJob_records_context_trace_when_subagents_receive_packets(t *testing.T) {
	// Given
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "app.go"), []byte("package main"), 0o644); err != nil {
		t.Fatalf("write app fixture: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, ".env"), []byte("OPENAI_API_KEY=sk-proj-secret"), 0o644); err != nil {
		t.Fatalf("write secret fixture: %v", err)
	}
	runtime := NewRuntime()

	// When
	report, err := runtime.RunJob(context.Background(), JobRequest{
		Task:                   "Fix checkout failure",
		WorkspaceDir:           root,
		MaxContextBytes:        64,
		WorkspaceBriefExcludes: []string{".env"},
	})
	// Then
	if err != nil {
		t.Fatalf("RunJob returned error: %v", err)
	}
	if len(report.ContextTrace) != len(report.SubagentResults) {
		t.Fatalf("ContextTrace length = %d, want %d", len(report.ContextTrace), len(report.SubagentResults))
	}
	first := report.ContextTrace[0]
	if first.AgentName == "" || first.Role == "" {
		t.Fatalf("first trace = %+v, want agent identity", first)
	}
	if first.MaxContextBytes != 64 || first.BudgetUnit != "bytes" {
		t.Fatalf("first trace = %+v, want byte budget 64", first)
	}
	if first.WorkspaceBrief.FileCount != 1 || first.WorkspaceBrief.Bytes == 0 {
		t.Fatalf("workspace brief trace = %+v, want file count and bytes", first.WorkspaceBrief)
	}
	if len(first.ExcludedContent.WorkspaceExcludes) != 1 || first.ExcludedContent.WorkspaceExcludes[0] != ".env" {
		t.Fatalf("excluded content = %+v, want .env exclude", first.ExcludedContent)
	}
}

func Test_Runtime_RunJob_marks_context_trace_truncated_when_budget_is_tiny(t *testing.T) {
	// Given
	runtime := NewRuntime()
	task := strings.Repeat("tiny budget task ", 10)

	// When
	report, err := runtime.RunJob(context.Background(), JobRequest{
		Task:            task,
		MaxContextBytes: 12,
	})
	// Then
	if err != nil {
		t.Fatalf("RunJob returned error: %v", err)
	}
	if len(report.ContextTrace) == 0 {
		t.Fatal("expected context trace entries")
	}
	if !report.ContextTrace[0].ContextTruncated {
		t.Fatalf("first trace = %+v, want truncation", report.ContextTrace[0])
	}
	if !containsTraceField(report.ContextTrace[0].TruncatedFields, "task") {
		t.Fatalf("truncated fields = %#v, want task", report.ContextTrace[0].TruncatedFields)
	}
}

func containsTraceField(fields []string, want string) bool {
	for _, field := range fields {
		if field == want {
			return true
		}
	}
	return false
}
