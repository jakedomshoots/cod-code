package ceo

import (
	"context"
	"strings"
	"testing"

	"ceoharness/internal/jobpacket"
	"ceoharness/internal/subagent"
)

type bloatedSubagentRunner struct{}

func (bloatedSubagentRunner) Run(_ context.Context, packet subagent.TaskPacket) (subagent.Result, error) {
	if packet.AgentName == "coder" {
		return subagent.Result{
			AgentName:       packet.AgentName,
			Role:            packet.Role,
			Status:          "pass",
			ContextReceived: packet.ContextMode,
			PriorFindings:   packet.PriorFindings,
			Summary:         "coder saw " + packet.PriorFindings,
			Evidence:        []string{"coder evidence"},
		}, nil
	}
	return subagent.Result{
		AgentName:       packet.AgentName,
		Role:            packet.Role,
		Status:          "pass",
		ContextReceived: packet.ContextMode,
		Summary:         strings.Repeat("scanner-detail-", 20) + "RAW_TAIL",
		Evidence:        []string{strings.Repeat("evidence-detail-", 20) + "RAW_TAIL"},
		Questions:       []string{strings.Repeat("question-detail-", 20) + "RAW_TAIL"},
	}, nil
}

func Test_Runtime_RunJob_caps_subagent_output_before_prior_findings(t *testing.T) {
	// Given
	runtime := NewRuntimeWithSubagentRunner(bloatedSubagentRunner{})

	// When
	report, err := runtime.RunJob(context.Background(), JobRequest{
		Task:                   "Fix a checkout bug",
		MaxSubagentOutputBytes: 40,
		Subagents: []jobpacket.Subagent{
			{Name: "scanner", Role: "inspect scope"},
			{Name: "coder", Role: "apply patch"},
		},
	})
	// Then
	if err != nil {
		t.Fatalf("RunJob returned error: %v", err)
	}
	if len(report.SubagentResults) != 2 {
		t.Fatalf("SubagentResults length = %d, want 2", len(report.SubagentResults))
	}
	scanner := report.SubagentResults[0]
	if !scanner.OutputTruncated {
		t.Fatal("scanner OutputTruncated = false, want true")
	}
	if !strings.Contains(scanner.Summary, "[truncated]") {
		t.Fatalf("scanner summary = %q, want truncated marker", scanner.Summary)
	}
	if strings.Contains(scanner.Summary, "RAW_TAIL") {
		t.Fatalf("scanner summary kept raw tail: %q", scanner.Summary)
	}
	coder := report.SubagentResults[1]
	if !strings.Contains(coder.PriorFindings, "[truncated]") {
		t.Fatalf("coder prior findings = %q, want truncated marker", coder.PriorFindings)
	}
	if strings.Contains(coder.PriorFindings, "RAW_TAIL") {
		t.Fatalf("coder prior findings kept raw tail: %q", coder.PriorFindings)
	}
}
