package subagent

import (
	"context"
	"strings"
	"testing"
)

func Test_Runner_Run_executes_native_subagent_when_packet_is_valid(t *testing.T) {
	// Given
	packet := TaskPacket{
		Task:           "Fix a failing test",
		AgentName:      "scanner",
		Role:           "inspect scope",
		ContextMode:    "lean",
		AllowedActions: []string{"read_workspace", "search_workspace"},
	}
	runner := NewRunner()

	// When
	result, err := runner.Run(context.Background(), packet)

	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if result.AgentName != "scanner" {
		t.Fatalf("AgentName = %q, want scanner", result.AgentName)
	}
	if result.Status != "pass" {
		t.Fatalf("Status = %q, want pass", result.Status)
	}
	if result.Role != "inspect scope" {
		t.Fatalf("Role = %q, want inspect scope", result.Role)
	}
	if result.ContextReceived != "lean" {
		t.Fatalf("ContextReceived = %q, want lean", result.ContextReceived)
	}
	if len(result.AllowedActions) != 2 || result.AllowedActions[0] != "read_workspace" {
		t.Fatalf("AllowedActions = %#v, want task packet actions", result.AllowedActions)
	}
	if result.PromptBytes == 0 {
		t.Fatal("expected prompt byte count")
	}
	if len(result.Evidence) == 0 {
		t.Fatal("expected evidence")
	}
}

func Test_Runner_Run_truncates_task_when_context_budget_is_set(t *testing.T) {
	// Given
	packet := TaskPacket{
		Task:            strings.Repeat("a", 20),
		AgentName:       "scanner",
		Role:            "inspect scope",
		ContextMode:     "lean",
		MaxContextBytes: 10,
	}
	runner := NewRunner()

	// When
	result, err := runner.Run(context.Background(), packet)

	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if result.ContextBytes != 10 {
		t.Fatalf("ContextBytes = %d, want 10", result.ContextBytes)
	}
	if !result.ContextTruncated {
		t.Fatal("expected context to be truncated")
	}
}

func Test_Runner_Run_reports_context_fields_truncated_by_budget(t *testing.T) {
	// Given
	packet := TaskPacket{
		Task:            strings.Repeat("a", 12),
		AgentName:       "coder",
		Role:            "apply bounded changes",
		ContextMode:     "lean",
		PriorFindings:   "scanner prior handoff",
		MaxContextBytes: 10,
	}
	runner := NewRunner()

	// When
	result, err := runner.Run(context.Background(), packet)

	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if len(result.ContextTruncatedFields) != 2 || result.ContextTruncatedFields[0] != "task" || result.ContextTruncatedFields[1] != "prior_findings" {
		t.Fatalf("ContextTruncatedFields = %#v, want task and prior_findings", result.ContextTruncatedFields)
	}
}

func Test_Runner_Run_reports_only_prior_findings_sent_under_context_budget(t *testing.T) {
	// Given
	packet := TaskPacket{
		Task:            strings.Repeat("a", 12),
		AgentName:       "coder",
		Role:            "apply bounded changes",
		ContextMode:     "lean",
		PriorFindings:   "scanner found a very long prior handoff that should not fit",
		MaxContextBytes: 12,
	}
	runner := NewRunner()

	// When
	result, err := runner.Run(context.Background(), packet)

	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if result.PriorFindings != "" {
		t.Fatalf("PriorFindings = %q, want only budgeted prior findings", result.PriorFindings)
	}
	if !result.ContextTruncated {
		t.Fatal("expected context to be truncated")
	}
}

func Test_Runner_Run_rejects_blank_task_when_packet_task_is_empty(t *testing.T) {
	// Given
	packet := TaskPacket{
		Task:        "",
		AgentName:   "scanner",
		ContextMode: "lean",
	}
	runner := NewRunner()

	// When
	_, err := runner.Run(context.Background(), packet)

	// Then
	if err == nil {
		t.Fatal("expected blank task error")
	}
}
