package ceo

import (
	"context"
	"testing"

	"ceoharness/internal/jobpacket"
	"ceoharness/internal/subagent"
)

type budgetCaptureRunner struct{}

func (r *budgetCaptureRunner) Run(ctx context.Context, packet subagent.TaskPacket) (subagent.Result, error) {
	if err := ctx.Err(); err != nil {
		return subagent.Result{}, err
	}
	return subagent.Result{
		AgentName:       packet.AgentName,
		Status:          "pass",
		ContextReceived: packet.ContextMode,
		ContextBytes:    packet.MaxContextBytes,
		Summary:         "captured context budget",
	}, nil
}

func Test_Runtime_RunJob_uses_subagent_context_budget_when_configured(t *testing.T) {
	// Given
	runner := &budgetCaptureRunner{}
	runtime := NewRuntimeWithSubagentRunner(runner)

	// When
	report, err := runtime.RunJob(context.Background(), JobRequest{
		Task: "Fix a failing test",
		Subagents: []jobpacket.Subagent{
			{Name: "scanner", Role: "inspect scope", MaxContextBytes: 512},
			{Name: "reviewer", Role: "verify evidence"},
		},
	})
	// Then
	if err != nil {
		t.Fatalf("RunJob returned error: %v", err)
	}
	if report.SubagentResults[0].ContextBytes != 512 {
		t.Fatalf("scanner ContextBytes = %d, want 512", report.SubagentResults[0].ContextBytes)
	}
	if report.SubagentResults[1].ContextBytes != report.JobPacket.ContextPolicy.MaxBytes {
		t.Fatalf("reviewer ContextBytes = %d, want global budget", report.SubagentResults[1].ContextBytes)
	}
}

func Test_Runtime_RunJob_passes_context_budget_to_subagents_when_packet_is_built(t *testing.T) {
	// Given
	runner := &budgetCaptureRunner{}
	runtime := NewRuntimeWithSubagentRunner(runner)

	// When
	report, err := runtime.RunJob(context.Background(), JobRequest{
		Task: "Fix a failing test",
	})
	// Then
	if err != nil {
		t.Fatalf("RunJob returned error: %v", err)
	}
	if report.JobPacket.ContextPolicy.MaxBytes != 4096 {
		t.Fatalf("MaxBytes = %d, want 4096", report.JobPacket.ContextPolicy.MaxBytes)
	}
	if len(report.SubagentResults) != 3 {
		t.Fatalf("SubagentResults length = %d, want 3", len(report.SubagentResults))
	}
	for index, result := range report.SubagentResults {
		if result.ContextBytes != report.JobPacket.ContextPolicy.MaxBytes {
			t.Fatalf("SubagentResults[%d].ContextBytes = %d, want packet max", index, result.ContextBytes)
		}
	}
}

func Test_Runtime_RunJob_reports_only_prior_findings_sent_under_context_budget(t *testing.T) {
	// Given
	runtime := NewRuntime()

	// When
	report, err := runtime.RunJob(context.Background(), JobRequest{
		Task:            "Fix a very long failing checkout workflow",
		MaxContextBytes: 10,
	})
	// Then
	if err != nil {
		t.Fatalf("RunJob returned error: %v", err)
	}
	if len(report.SubagentResults) < 2 {
		t.Fatalf("SubagentResults length = %d, want at least 2", len(report.SubagentResults))
	}
	coder := report.SubagentResults[1]
	if coder.PriorFindings != "" {
		t.Fatalf("coder PriorFindings = %q, want only budgeted prior findings", coder.PriorFindings)
	}
	if !coder.ContextTruncated {
		t.Fatal("expected coder context to be truncated")
	}
}
