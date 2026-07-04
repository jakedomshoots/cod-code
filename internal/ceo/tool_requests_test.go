package ceo

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"ceoharness/internal/jobpacket"
	"ceoharness/internal/subagent"
)

type toolRequestRunner struct{}

func (r toolRequestRunner) Run(ctx context.Context, packet subagent.TaskPacket) (subagent.Result, error) {
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
			Summary:         "used tool results: " + packet.ToolResults[0].Output,
			Evidence:        []string{"tool feedback used"},
		}, nil
	}
	requests := []subagent.ToolRequest{}
	if packet.AgentName == "scanner" {
		requests = append(
			requests,
			subagent.ToolRequest{Action: "read_workspace", Path: "app.txt"},
			subagent.ToolRequest{Action: "search_workspace", Query: "needle"},
		)
	}
	if packet.AgentName == "reviewer" {
		requests = append(requests, subagent.ToolRequest{Action: "run_checks"})
	}
	return subagent.Result{
		AgentName:       packet.AgentName,
		Role:            packet.Role,
		Status:          "pass",
		Attempts:        1,
		ContextReceived: packet.ContextMode,
		ContextBytes:    len(packet.Task),
		Summary:         "requested tools",
		ToolRequests:    requests,
		Evidence:        []string{"ok"},
	}, nil
}

func Test_Runtime_RunJob_executes_allowed_subagent_tool_requests(t *testing.T) {
	// Given
	runtime := NewRuntimeWithSubagentRunner(toolRequestRunner{})
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "app.txt"), []byte("hello needle"), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	// When
	report, err := runtime.RunJob(context.Background(), JobRequest{
		Task:         "Fix a failing test",
		WorkspaceDir: root,
		CheckCommand: []string{
			os.Args[0],
			"-test.run=Test_HelperProcess_tool_request_check",
		},
		CheckEnv: []string{"GO_WANT_TOOL_REQUEST_CHECK=1"},
	})
	// Then
	if err != nil {
		t.Fatalf("RunJob returned error: %v", err)
	}
	scanner := report.SubagentResults[0]
	if len(scanner.ToolResults) != 2 {
		t.Fatalf("scanner ToolResults length = %d, want 2", len(scanner.ToolResults))
	}
	if scanner.ToolResults[0].Status != "pass" || scanner.ToolResults[0].Output != "hello needle" {
		t.Fatalf("read tool result = %+v, want file content", scanner.ToolResults[0])
	}
	if scanner.ToolResults[1].Status != "pass" || scanner.ToolResults[1].MatchCount != 1 {
		t.Fatalf("search tool result = %+v, want one match", scanner.ToolResults[1])
	}
	if scanner.ToolFeedbackPasses != 1 {
		t.Fatalf("scanner ToolFeedbackPasses = %d, want 1", scanner.ToolFeedbackPasses)
	}
	if scanner.InitialSummary != "requested tools" {
		t.Fatalf("scanner InitialSummary = %q, want requested tools", scanner.InitialSummary)
	}
	if !strings.Contains(scanner.Summary, "hello needle") {
		t.Fatalf("scanner Summary = %q, want tool output in feedback summary", scanner.Summary)
	}
	reviewer := report.SubagentResults[2]
	if len(reviewer.ToolResults) != 1 || reviewer.ToolResults[0].Status != "pass" {
		t.Fatalf("reviewer ToolResults = %+v, want passing check tool", reviewer.ToolResults)
	}
}

type genericWorkspaceReadRunner struct{}

func (r genericWorkspaceReadRunner) Run(ctx context.Context, packet subagent.TaskPacket) (subagent.Result, error) {
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
			Summary:            "used workspace list: " + packet.ToolResults[0].Output,
			ToolFeedbackPasses: 1,
			Evidence:           []string{"workspace list used"},
		}, nil
	}
	return subagent.Result{
		AgentName:       packet.AgentName,
		Role:            packet.Role,
		Status:          "needs_input",
		Attempts:        1,
		ContextReceived: packet.ContextMode,
		ContextBytes:    len(packet.Task),
		Summary:         "request generic workspace read",
		ToolRequests: []subagent.ToolRequest{
			{Action: "read_workspace", Path: "read_workspace"},
		},
		Evidence: []string{"requested workspace list"},
	}, nil
}

func Test_Runtime_RunJob_returns_workspace_brief_for_generic_read_workspace_request(t *testing.T) {
	// Given
	runtime := NewRuntimeWithSubagentRunner(genericWorkspaceReadRunner{})
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "frontend.js"), []byte("hello"), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	// When
	report, err := runtime.RunJob(context.Background(), JobRequest{
		Task:         "Inspect workspace",
		WorkspaceDir: root,
		Subagents: []jobpacket.Subagent{
			{
				Name:           "scanner",
				Role:           "read only",
				AllowedActions: []jobpacket.Action{jobpacket.ActionReadWorkspace},
			},
		},
	})
	// Then
	if err != nil {
		t.Fatalf("RunJob returned error: %v", err)
	}
	result := report.SubagentResults[0]
	if result.ToolFeedbackPasses != 1 {
		t.Fatalf("ToolFeedbackPasses = %d, want 1", result.ToolFeedbackPasses)
	}
	if len(result.ToolResults) != 1 || result.ToolResults[0].Status != "pass" || result.ToolResults[0].Path != "workspace" {
		t.Fatalf("ToolResults = %+v, want passing workspace brief", result.ToolResults)
	}
	if !strings.Contains(result.Summary, "frontend.js") || !strings.Contains(result.Summary, "hello") {
		t.Fatalf("Summary = %q, want workspace file list and small file content", result.Summary)
	}
}

type inferredWorkspaceReadRunner struct{}

func (r inferredWorkspaceReadRunner) Run(ctx context.Context, packet subagent.TaskPacket) (subagent.Result, error) {
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
			Summary:            "read inferred files: " + packet.ToolResults[0].Output,
			ToolFeedbackPasses: 1,
			Evidence:           []string{"inferred read used"},
		}, nil
	}
	return subagent.Result{
		AgentName:       packet.AgentName,
		Role:            packet.Role,
		Status:          "needs_input",
		Attempts:        1,
		ContextReceived: packet.ContextMode,
		ContextBytes:    len(packet.Task),
		Summary:         "Need frontend/state.js before patching.",
		Questions:       []string{"Please provide frontend/state.js."},
		Evidence:        []string{"asked for file"},
	}, nil
}

func Test_Runtime_RunJob_infers_read_workspace_request_from_needs_input_file_question(t *testing.T) {
	// Given
	runtime := NewRuntimeWithSubagentRunner(inferredWorkspaceReadRunner{})
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "frontend"), 0o755); err != nil {
		t.Fatalf("mkdir fixture: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "frontend", "state.js"), []byte("module.exports = 'state'"), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	// When
	report, err := runtime.RunJob(context.Background(), JobRequest{
		Task:         "Patch state",
		WorkspaceDir: root,
		Subagents: []jobpacket.Subagent{
			{
				Name:           "coder",
				Role:           "patch",
				AllowedActions: []jobpacket.Action{jobpacket.ActionReadWorkspace, jobpacket.ActionProposePatch},
			},
		},
	})
	// Then
	if err != nil {
		t.Fatalf("RunJob returned error: %v", err)
	}
	result := report.SubagentResults[0]
	if result.Status != "pass" || result.ToolFeedbackPasses != 1 {
		t.Fatalf("result = %+v, want pass after inferred file read", result)
	}
	if len(result.ToolRequests) != 1 || result.ToolRequests[0].Path != "frontend/state.js" {
		t.Fatalf("ToolRequests = %+v, want inferred frontend/state.js read", result.ToolRequests)
	}
	if !strings.Contains(result.Summary, "module.exports") {
		t.Fatalf("Summary = %q, want file content from inferred read", result.Summary)
	}
}

type deniedToolRequestRunner struct{}

func (r deniedToolRequestRunner) Run(ctx context.Context, packet subagent.TaskPacket) (subagent.Result, error) {
	if err := ctx.Err(); err != nil {
		return subagent.Result{}, err
	}
	return subagent.Result{
		AgentName:       packet.AgentName,
		Role:            packet.Role,
		Status:          "pass",
		Attempts:        1,
		ContextReceived: packet.ContextMode,
		ContextBytes:    len(packet.Task),
		Summary:         "requested search",
		ToolRequests: []subagent.ToolRequest{
			{Action: "search_workspace", Query: "needle"},
		},
		Evidence: []string{"ok"},
	}, nil
}

func Test_Runtime_RunJob_denies_tool_request_when_action_is_not_allowed(t *testing.T) {
	// Given
	runtime := NewRuntimeWithSubagentRunner(deniedToolRequestRunner{})
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "app.txt"), []byte("hello needle"), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	// When
	report, err := runtime.RunJob(context.Background(), JobRequest{
		Task:         "Inspect files",
		WorkspaceDir: root,
		Subagents: []jobpacket.Subagent{
			{
				Name:           "scanner",
				Role:           "read only",
				AllowedActions: []jobpacket.Action{jobpacket.ActionReadWorkspace},
			},
		},
	})
	// Then
	if err != nil {
		t.Fatalf("RunJob returned error: %v", err)
	}
	results := report.SubagentResults[0].ToolResults
	if len(results) != 1 {
		t.Fatalf("ToolResults length = %d, want 1", len(results))
	}
	if results[0].Status != "denied" {
		t.Fatalf("ToolResults[0] = %+v, want denied status", results[0])
	}
	if report.SubagentResults[0].ToolFeedbackPasses != 0 {
		t.Fatalf("ToolFeedbackPasses = %d, want no feedback pass for denied tool", report.SubagentResults[0].ToolFeedbackPasses)
	}
}

type networkResearchRunner struct{}

func (r networkResearchRunner) Run(ctx context.Context, packet subagent.TaskPacket) (subagent.Result, error) {
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
			Summary:            "used research: " + packet.ToolResults[0].Output,
			ToolFeedbackPasses: 1,
			Evidence:           []string{"research used"},
		}, nil
	}
	requests := []subagent.ToolRequest{}
	if packet.AgentName == "researcher" {
		requests = append(requests, subagent.ToolRequest{Action: "network_research", Query: "agent harness docs"})
	}
	return subagent.Result{
		AgentName:       packet.AgentName,
		Role:            packet.Role,
		Status:          "pass",
		Attempts:        1,
		ContextReceived: packet.ContextMode,
		ContextBytes:    len(packet.Task),
		Summary:         "requested research",
		ToolRequests:    requests,
		Evidence:        []string{"ok"},
	}, nil
}

func Test_Runtime_RunJob_executes_allowed_network_research_tool_request(t *testing.T) {
	// Given
	runtime := NewRuntimeWithSubagentRunner(networkResearchRunner{})
	t.Setenv("GO_WANT_NETWORK_RESEARCH_TOOL", "1")

	// When
	report, err := runtime.RunJob(context.Background(), JobRequest{
		Task: "Research agent harness docs",
		ResearchCommand: []string{
			os.Args[0],
			"-test.run=Test_HelperProcess_network_research_tool",
		},
	})
	// Then
	if err != nil {
		t.Fatalf("RunJob returned error: %v", err)
	}
	researcher := report.SubagentResults[0]
	if len(researcher.ToolResults) != 1 {
		t.Fatalf("researcher ToolResults length = %d, want 1", len(researcher.ToolResults))
	}
	result := researcher.ToolResults[0]
	if result.Status != "pass" || result.Query != "agent harness docs" {
		t.Fatalf("network research result = %+v, want passing query result", result)
	}
	if !strings.Contains(result.Output, "research result for agent harness docs") {
		t.Fatalf("network research output = %q, want research result", result.Output)
	}
	if researcher.ToolFeedbackPasses != 1 {
		t.Fatalf("researcher ToolFeedbackPasses = %d, want 1", researcher.ToolFeedbackPasses)
	}
}

func Test_HelperProcess_tool_request_check(t *testing.T) {
	if os.Getenv("GO_WANT_TOOL_REQUEST_CHECK") != "1" {
		return
	}
	os.Stdout.WriteString("tool check passed\n")
	os.Exit(0)
}

func Test_HelperProcess_network_research_tool(t *testing.T) {
	if os.Getenv("GO_WANT_NETWORK_RESEARCH_TOOL") != "1" {
		return
	}
	query, err := io.ReadAll(os.Stdin)
	if err != nil {
		t.Fatalf("read stdin: %v", err)
	}
	if os.Getenv("CEO_RESEARCH_QUERY") != strings.TrimSpace(string(query)) {
		os.Stderr.WriteString("missing query env")
		os.Exit(2)
	}
	os.Stdout.WriteString("research result for " + strings.TrimSpace(string(query)))
	os.Exit(0)
}
