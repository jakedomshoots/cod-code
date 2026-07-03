package ceo

import (
	"context"
	"testing"

	"ceoharness/internal/subagent"
)

type configurableBudgetRunner struct{}

func (r configurableBudgetRunner) Run(ctx context.Context, packet subagent.TaskPacket) (subagent.Result, error) {
	if err := ctx.Err(); err != nil {
		return subagent.Result{}, err
	}
	return subagent.Result{
		AgentName:       packet.AgentName,
		Status:          "pass",
		ContextReceived: packet.ContextMode,
		ContextBytes:    packet.MaxContextBytes,
		Summary:         "captured configured context budget",
	}, nil
}

func Test_Runtime_RunJob_uses_configured_context_budget(t *testing.T) {
	// Given
	runtime := NewRuntimeWithSubagentRunner(configurableBudgetRunner{})

	// When
	report, err := runtime.RunJob(context.Background(), JobRequest{
		Task:            "Fix a failing test",
		MaxContextBytes: 1024,
	})
	// Then
	if err != nil {
		t.Fatalf("RunJob returned error: %v", err)
	}
	if report.JobPacket.ContextPolicy.MaxBytes != 1024 {
		t.Fatalf("MaxBytes = %d, want configured budget", report.JobPacket.ContextPolicy.MaxBytes)
	}
	for index, result := range report.SubagentResults {
		if result.ContextBytes != 1024 {
			t.Fatalf("SubagentResults[%d].ContextBytes = %d, want configured budget", index, result.ContextBytes)
		}
	}
}
