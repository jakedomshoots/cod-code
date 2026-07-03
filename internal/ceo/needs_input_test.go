package ceo

import (
	"context"
	"testing"

	"ceoharness/internal/subagent"
)

type needsInputRunner struct{}

func (r needsInputRunner) Run(ctx context.Context, packet subagent.TaskPacket) (subagent.Result, error) {
	if err := ctx.Err(); err != nil {
		return subagent.Result{}, err
	}
	return subagent.Result{
		AgentName:       packet.AgentName,
		Role:            packet.Role,
		Status:          "needs_input",
		Attempts:        1,
		ContextReceived: packet.ContextMode,
		ContextBytes:    len(packet.Task),
		Summary:         "missing target repo",
		Questions:       []string{"Which package should I change?"},
		Evidence:        []string{"user input required"},
	}, nil
}

func Test_Runtime_RunJob_returns_needs_input_verdict_when_subagent_needs_user_input(t *testing.T) {
	// Given
	runtime := NewRuntimeWithSubagentRunner(needsInputRunner{})

	// When
	report, err := runtime.RunJob(context.Background(), JobRequest{Task: "Fix ambiguous package"})

	// Then
	if err != nil {
		t.Fatalf("RunJob returned error: %v", err)
	}
	if report.Verdict != "needs_input" {
		t.Fatalf("Verdict = %q, want needs_input", report.Verdict)
	}
	if report.RunManifest.Verdict != "needs_input" {
		t.Fatalf("RunManifest verdict = %q, want needs_input", report.RunManifest.Verdict)
	}
	if report.SubagentResults[0].Questions[0] != "Which package should I change?" {
		t.Fatalf("Questions = %+v, want subagent question", report.SubagentResults[0].Questions)
	}
	if len(report.SubagentResults) != 1 {
		t.Fatalf("SubagentResults length = %d, want scheduler to stop after needs_input", len(report.SubagentResults))
	}
	if report.RunEvents[len(report.RunEvents)-1].Status != "needs_input" {
		t.Fatalf("final event = %+v, want needs_input status", report.RunEvents[len(report.RunEvents)-1])
	}
	if report.ExecutionPlan.NextAction != "answer subagent questions" {
		t.Fatalf("ExecutionPlan.NextAction = %q, want answer subagent questions", report.ExecutionPlan.NextAction)
	}
}
