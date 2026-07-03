package ceo

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"ceoharness/internal/jobpacket"
	"ceoharness/internal/subagent"
)

type cappedToolRequestRunner struct{}

func (r cappedToolRequestRunner) Run(ctx context.Context, packet subagent.TaskPacket) (subagent.Result, error) {
	if err := ctx.Err(); err != nil {
		return subagent.Result{}, err
	}
	if len(packet.ToolResults) > 0 {
		return subagent.Result{
			AgentName:       packet.AgentName,
			Role:            packet.Role,
			Status:          "pass",
			Attempts:        1,
			ContextReceived: packet.ContextMode,
			ContextBytes:    len(packet.Task),
			Summary:         "used capped tool results",
			Evidence:        []string{"tool cap honored"},
		}, nil
	}
	return subagent.Result{
		AgentName:       packet.AgentName,
		Role:            packet.Role,
		Status:          "pass",
		Attempts:        1,
		ContextReceived: packet.ContextMode,
		ContextBytes:    len(packet.Task),
		Summary:         "requested too many tools",
		ToolRequests: []subagent.ToolRequest{
			{Action: "read_workspace", Path: "app.txt"},
			{Action: "search_workspace", Query: "needle"},
			{Action: "read_workspace", Path: "other.txt"},
		},
		Evidence: []string{"ok"},
	}, nil
}

func Test_Runtime_RunJob_skips_tool_requests_after_configured_limit(t *testing.T) {
	// Given
	runtime := NewRuntimeWithSubagentRunner(cappedToolRequestRunner{})
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "app.txt"), []byte("hello needle"), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	// When
	report, err := runtime.RunJob(context.Background(), JobRequest{
		Task:            "Inspect files",
		WorkspaceDir:    root,
		MaxToolRequests: 1,
		Subagents: []jobpacket.Subagent{
			{
				Name:           "scanner",
				Role:           "inspect files",
				AllowedActions: []jobpacket.Action{jobpacket.ActionReadWorkspace, jobpacket.ActionSearchWorkspace},
			},
		},
	})

	// Then
	if err != nil {
		t.Fatalf("RunJob returned error: %v", err)
	}
	results := report.SubagentResults[0].ToolResults
	if len(results) != 3 {
		t.Fatalf("ToolResults length = %d, want 3", len(results))
	}
	if results[0].Status != "pass" || results[0].Output != "hello needle" {
		t.Fatalf("ToolResults[0] = %+v, want executed read", results[0])
	}
	for index := 1; index < len(results); index++ {
		if results[index].Status != "skipped" || !strings.Contains(results[index].Error, "tool request limit") {
			t.Fatalf("ToolResults[%d] = %+v, want skipped tool limit result", index, results[index])
		}
	}
	if report.SubagentResults[0].ToolFeedbackPasses != 1 {
		t.Fatalf("ToolFeedbackPasses = %d, want feedback from executed tool", report.SubagentResults[0].ToolFeedbackPasses)
	}
}
