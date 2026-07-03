package ceo

import (
	"context"
	"testing"

	"ceoharness/internal/subagent"
)

type roleCaptureRunner struct{}

func (r *roleCaptureRunner) Run(ctx context.Context, packet subagent.TaskPacket) (subagent.Result, error) {
	if err := ctx.Err(); err != nil {
		return subagent.Result{}, err
	}
	return subagent.Result{
		AgentName:       packet.AgentName,
		Role:            packet.Role,
		Status:          "pass",
		ContextReceived: packet.ContextMode,
		Summary:         "captured role",
	}, nil
}

func Test_Runtime_RunJob_passes_role_to_each_subagent_when_packet_is_built(t *testing.T) {
	// Given
	runner := &roleCaptureRunner{}
	runtime := NewRuntimeWithSubagentRunner(runner)

	// When
	report, err := runtime.RunJob(context.Background(), JobRequest{
		Task: "Fix a failing test",
	})
	// Then
	if err != nil {
		t.Fatalf("RunJob returned error: %v", err)
	}
	wantRoles := []string{"inspect scope", "apply bounded changes", "verify evidence"}
	if len(report.SubagentResults) != len(wantRoles) {
		t.Fatalf("SubagentResults length = %d, want %d", len(report.SubagentResults), len(wantRoles))
	}
	for index, wantRole := range wantRoles {
		if report.SubagentResults[index].Role != wantRole {
			t.Fatalf("SubagentResults[%d].Role = %q, want %q", index, report.SubagentResults[index].Role, wantRole)
		}
	}
}
