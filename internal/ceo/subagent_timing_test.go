package ceo

import (
	"context"
	"testing"
	"time"

	"ceoharness/internal/subagent"
)

type slowRunner struct{}

func (r slowRunner) Run(ctx context.Context, packet subagent.TaskPacket) (subagent.Result, error) {
	if err := ctx.Err(); err != nil {
		return subagent.Result{}, err
	}
	time.Sleep(5 * time.Millisecond)
	return subagent.Result{
		AgentName:       packet.AgentName,
		Role:            packet.Role,
		Status:          "pass",
		Attempts:        1,
		ContextReceived: packet.ContextMode,
		Summary:         "ok",
	}, nil
}

func Test_Runtime_RunJob_records_subagent_duration(t *testing.T) {
	// Given
	runtime := NewRuntimeWithSubagentRunner(slowRunner{})

	// When
	report, err := runtime.RunJob(context.Background(), JobRequest{
		Task: "Fix a failing test",
	})

	// Then
	if err != nil {
		t.Fatalf("RunJob returned error: %v", err)
	}
	if report.SubagentResults[0].DurationMS <= 0 {
		t.Fatalf("DurationMS = %d, want positive duration", report.SubagentResults[0].DurationMS)
	}
}
