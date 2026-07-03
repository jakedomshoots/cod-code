package ceo

import (
	"context"
	"strings"
	"testing"

	"ceoharness/internal/jobpacket"
	"ceoharness/internal/subagent"
)

type verifyEvidenceRunner struct{}

func (r verifyEvidenceRunner) Run(ctx context.Context, packet subagent.TaskPacket) (subagent.Result, error) {
	if err := ctx.Err(); err != nil {
		return subagent.Result{}, err
	}
	if len(packet.ToolResults) > 0 {
		return subagent.Result{
			AgentName:          packet.AgentName,
			Role:               packet.Role,
			Status:             "pass",
			Attempts:           1,
			ContextReceived:    packet.ContextMode,
			ContextBytes:       len(packet.Task),
			Summary:            "verified evidence: " + packet.ToolResults[0].Output,
			ToolFeedbackPasses: 1,
			Evidence:           []string{"evidence verified"},
		}, nil
	}
	if packet.AgentName == "reviewer" {
		return subagent.Result{
			AgentName:       packet.AgentName,
			Role:            packet.Role,
			Status:          "pass",
			Attempts:        1,
			ContextReceived: packet.ContextMode,
			ContextBytes:    len(packet.Task),
			Summary:         "requested evidence",
			ToolRequests: []subagent.ToolRequest{
				{Action: "verify_evidence"},
			},
			Evidence: []string{"review requested"},
		}, nil
	}
	return subagent.Result{
		AgentName:       packet.AgentName,
		Role:            packet.Role,
		Status:          "pass",
		Attempts:        1,
		ContextReceived: packet.ContextMode,
		ContextBytes:    len(packet.Task),
		Summary:         "scanner found evidence",
		Evidence:        []string{"app.txt contains target", "go test failed before fix"},
	}, nil
}

func Test_Runtime_RunJob_executes_allowed_verify_evidence_tool_request(t *testing.T) {
	// Given
	runtime := NewRuntimeWithSubagentRunner(verifyEvidenceRunner{})

	// When
	report, err := runtime.RunJob(context.Background(), JobRequest{
		Task: "Verify prior evidence",
		Subagents: []jobpacket.Subagent{
			{
				Name:           "scanner",
				Role:           "inspect evidence",
				Stage:          1,
				AllowedActions: []jobpacket.Action{jobpacket.ActionReadWorkspace},
			},
			{
				Name:           "reviewer",
				Role:           "verify evidence",
				Stage:          2,
				AllowedActions: []jobpacket.Action{jobpacket.ActionVerifyEvidence},
			},
		},
	})
	// Then
	if err != nil {
		t.Fatalf("RunJob returned error: %v", err)
	}
	reviewer := report.SubagentResults[1]
	if len(reviewer.ToolResults) != 1 {
		t.Fatalf("reviewer ToolResults length = %d, want 1", len(reviewer.ToolResults))
	}
	result := reviewer.ToolResults[0]
	if result.Status != "pass" {
		t.Fatalf("verify_evidence result = %+v, want pass", result)
	}
	if !strings.Contains(result.Output, "scanner(pass)") || !strings.Contains(result.Output, "app.txt contains target") {
		t.Fatalf("verify_evidence output = %q, want prior scanner evidence", result.Output)
	}
	if reviewer.ToolFeedbackPasses != 1 {
		t.Fatalf("reviewer ToolFeedbackPasses = %d, want 1", reviewer.ToolFeedbackPasses)
	}
	if !strings.Contains(reviewer.Summary, "go test failed before fix") {
		t.Fatalf("reviewer Summary = %q, want evidence feedback", reviewer.Summary)
	}
}
