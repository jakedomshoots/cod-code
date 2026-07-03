package ceo

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"ceoharness/internal/subagent"
)

type stageToolFeedbackRunner struct{}

func (r stageToolFeedbackRunner) Run(ctx context.Context, packet subagent.TaskPacket) (subagent.Result, error) {
	if err := ctx.Err(); err != nil {
		return subagent.Result{}, err
	}
	if packet.AgentName == "scanner" && len(packet.ToolResults) > 0 {
		return subagent.Result{
			AgentName:       packet.AgentName,
			Role:            packet.Role,
			Status:          "pass",
			Attempts:        1,
			ContextReceived: packet.ContextMode,
			ContextBytes:    len(packet.Task),
			Summary:         "scanner used tool output " + packet.ToolResults[0].Output,
			Evidence:        []string{"tool feedback used"},
		}, nil
	}
	if packet.AgentName == "scanner" {
		return subagent.Result{
			AgentName:       packet.AgentName,
			Role:            packet.Role,
			Status:          "pass",
			Attempts:        1,
			ContextReceived: packet.ContextMode,
			ContextBytes:    len(packet.Task),
			Summary:         "scanner initial",
			ToolRequests: []subagent.ToolRequest{
				{Action: "read_workspace", Path: "app.txt"},
			},
			Evidence: []string{"requested read"},
		}, nil
	}
	return subagent.Result{
		AgentName:       packet.AgentName,
		Role:            packet.Role,
		Status:          "pass",
		Attempts:        1,
		ContextReceived: packet.ContextMode,
		ContextBytes:    len(packet.Task) + len(packet.PriorFindings),
		Summary:         packet.AgentName + " saw " + packet.PriorFindings,
		Evidence:        []string{"prior findings checked"},
	}, nil
}

func Test_Runtime_RunJob_passes_stage_tool_feedback_to_later_subagents(t *testing.T) {
	// Given
	runtime := NewRuntimeWithSubagentRunner(stageToolFeedbackRunner{})
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "app.txt"), []byte("workspace clue"), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	// When
	report, err := runtime.RunJob(context.Background(), JobRequest{
		Task:         "Fix a failing test",
		WorkspaceDir: root,
	})
	// Then
	if err != nil {
		t.Fatalf("RunJob returned error: %v", err)
	}
	scanner := report.SubagentResults[0]
	if scanner.ToolFeedbackPasses != 1 {
		t.Fatalf("scanner ToolFeedbackPasses = %d, want 1", scanner.ToolFeedbackPasses)
	}
	coder := report.SubagentResults[1]
	if !strings.Contains(coder.PriorFindings, "scanner used tool output workspace clue") {
		t.Fatalf("coder PriorFindings = %q, want scanner tool feedback summary", coder.PriorFindings)
	}
	if strings.Contains(coder.PriorFindings, "scanner initial") {
		t.Fatalf("coder PriorFindings = %q, should not use stale scanner summary", coder.PriorFindings)
	}
}
