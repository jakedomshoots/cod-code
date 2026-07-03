package ceo

import (
	"context"
	"testing"

	"ceoharness/internal/jobpacket"
	"ceoharness/internal/subagent"
)

type assignmentCaptureRunner struct {
	assignments []string
}

func (r *assignmentCaptureRunner) Run(ctx context.Context, packet subagent.TaskPacket) (subagent.Result, error) {
	if err := ctx.Err(); err != nil {
		return subagent.Result{}, err
	}
	r.assignments = append(r.assignments, packet.Assignment)
	return subagent.Result{
		AgentName:       packet.AgentName,
		Role:            packet.Role,
		Assignment:      packet.Assignment,
		Status:          "pass",
		Attempts:        1,
		ContextReceived: packet.ContextMode,
		ContextBytes:    len(packet.Task) + len(packet.Assignment),
		Summary:         "assignment received",
		Evidence:        []string{"assignment checked"},
	}, nil
}

func Test_Runtime_RunJob_passes_ceo_delegated_assignment_to_selected_subagent(t *testing.T) {
	// Given
	client := &sequenceCEOModelClient{
		responses: []string{
			`{"selected_subagents":["security"],"assignments":{"security":"Inspect auth risks only."},"summary":"Security owns the risk pass."}`,
			`{"recommended_verdict":"pass","summary":"Security assignment passed."}`,
		},
	}
	runner := &assignmentCaptureRunner{}
	runtime := NewRuntimeWithSubagentRunnerAndCEOReviewer(runner, client)

	// When
	report, err := runtime.RunJob(context.Background(), JobRequest{
		Task: "Fix auth flow",
		Subagents: []jobpacket.Subagent{
			{Name: "planner", Role: "break down work"},
			{Name: "security", Role: "review auth risks"},
		},
	})

	// Then
	if err != nil {
		t.Fatalf("RunJob returned error: %v", err)
	}
	if report.CEODelegation == nil || report.CEODelegation.Assignments["security"] != "Inspect auth risks only." {
		t.Fatalf("CEODelegation = %#v, want security assignment", report.CEODelegation)
	}
	if report.JobPacket.Subagents[0].Assignment != "Inspect auth risks only." {
		t.Fatalf("job packet assignment = %q, want delegated assignment", report.JobPacket.Subagents[0].Assignment)
	}
	if len(runner.assignments) != 1 || runner.assignments[0] != "Inspect auth risks only." {
		t.Fatalf("runner assignments = %#v, want delegated assignment", runner.assignments)
	}
	if report.SubagentResults[0].Assignment != "Inspect auth risks only." {
		t.Fatalf("result assignment = %q, want delegated assignment", report.SubagentResults[0].Assignment)
	}
}
