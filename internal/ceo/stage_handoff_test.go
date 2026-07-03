package ceo

import (
	"context"
	"strings"
	"testing"

	"ceoharness/internal/subagent"
)

type stageHandoffRunner struct{}

func (r stageHandoffRunner) Run(ctx context.Context, packet subagent.TaskPacket) (subagent.Result, error) {
	if err := ctx.Err(); err != nil {
		return subagent.Result{}, err
	}
	summary := packet.AgentName + " summary"
	if packet.PriorFindings != "" {
		summary = packet.AgentName + " saw " + packet.PriorFindings
	}
	return subagent.Result{
		AgentName:       packet.AgentName,
		Role:            packet.Role,
		Status:          "pass",
		Attempts:        1,
		ContextReceived: packet.ContextMode,
		ContextBytes:    len(packet.Task) + len(packet.PriorFindings),
		Summary:         summary,
		Evidence:        []string{"stage handoff checked"},
	}, nil
}

func Test_Runtime_RunJob_passes_prior_stage_findings_to_later_subagents(t *testing.T) {
	// Given
	runtime := NewRuntimeWithSubagentRunner(stageHandoffRunner{})

	// When
	report, err := runtime.RunJob(context.Background(), JobRequest{Task: "Fix a failing test"})

	// Then
	if err != nil {
		t.Fatalf("RunJob returned error: %v", err)
	}
	coder := report.SubagentResults[1]
	if !strings.Contains(coder.PriorFindings, "scanner(pass): scanner summary") {
		t.Fatalf("coder PriorFindings = %q, want scanner summary", coder.PriorFindings)
	}
	reviewer := report.SubagentResults[2]
	if !strings.Contains(reviewer.PriorFindings, "coder(pass): coder saw") {
		t.Fatalf("reviewer PriorFindings = %q, want coder summary", reviewer.PriorFindings)
	}
}
