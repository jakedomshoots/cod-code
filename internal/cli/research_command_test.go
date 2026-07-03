package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func Test_Run_executes_network_research_with_research_command_flag(t *testing.T) {
	// Given
	var out bytes.Buffer
	t.Setenv("GO_WANT_CLI_RESEARCH_MODEL", "1")
	t.Setenv("GO_WANT_CLI_RESEARCH_COMMAND", "1")

	// When
	err := Run(context.Background(), &out, []string{
		"--model-command",
		os.Args[0],
		"-test.run=Test_HelperProcess_cli_research_model",
		"--",
		"--research-command",
		os.Args[0],
		"-test.run=Test_HelperProcess_cli_research_command",
		"--",
		"Research",
		"agent",
		"harness",
		"docs",
	})
	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	var body struct {
		SubagentResults []struct {
			AgentName          string `json:"agent_name"`
			Summary            string `json:"summary"`
			ToolFeedbackPasses int    `json:"tool_feedback_passes"`
			ToolResults        []struct {
				Action string `json:"action"`
				Status string `json:"status"`
				Query  string `json:"query"`
				Output string `json:"output"`
			} `json:"tool_results"`
		} `json:"subagent_results"`
	}
	if jsonErr := json.Unmarshal(out.Bytes(), &body); jsonErr != nil {
		t.Fatalf("output must be JSON: %v\n%s", jsonErr, out.String())
	}
	researcher := body.SubagentResults[0]
	if researcher.AgentName != "researcher" {
		t.Fatalf("first agent = %q, want researcher", researcher.AgentName)
	}
	if researcher.ToolFeedbackPasses != 1 {
		t.Fatalf("ToolFeedbackPasses = %d, want 1", researcher.ToolFeedbackPasses)
	}
	if len(researcher.ToolResults) != 1 || researcher.ToolResults[0].Status != "pass" {
		t.Fatalf("ToolResults = %+v, want passing research tool", researcher.ToolResults)
	}
	if !strings.Contains(researcher.Summary, "cli research result for agent harness docs") {
		t.Fatalf("Summary = %q, want research output in feedback summary", researcher.Summary)
	}
}

func Test_Run_executes_network_research_with_env_research_command(t *testing.T) {
	// Given
	var out bytes.Buffer
	t.Setenv("GO_WANT_CLI_RESEARCH_MODEL", "1")
	t.Setenv("GO_WANT_CLI_RESEARCH_COMMAND", "1")
	t.Setenv("CEO_RESEARCH_COMMAND_JSON", `["`+os.Args[0]+`","-test.run=Test_HelperProcess_cli_research_command"]`)

	// When
	err := Run(context.Background(), &out, []string{
		"--model-command",
		os.Args[0],
		"-test.run=Test_HelperProcess_cli_research_model",
		"--",
		"Research",
		"agent",
		"harness",
		"docs",
	})
	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	assertCLIResearchResult(t, out.Bytes())
}

func Test_Run_executes_network_research_with_workspace_research_command(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	t.Setenv("GO_WANT_CLI_RESEARCH_MODEL", "1")
	t.Setenv("GO_WANT_CLI_RESEARCH_COMMAND", "1")
	content := `{"research_command":["` + os.Args[0] + `","-test.run=Test_HelperProcess_cli_research_command"]}`
	if err := os.WriteFile(filepath.Join(root, ".ceo-harness.json"), []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	// When
	err := Run(context.Background(), &out, []string{
		"--workspace",
		root,
		"--model-command",
		os.Args[0],
		"-test.run=Test_HelperProcess_cli_research_model",
		"--",
		"Research",
		"agent",
		"harness",
		"docs",
	})
	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	assertCLIResearchResult(t, out.Bytes())
}

func assertCLIResearchResult(t *testing.T, output []byte) {
	t.Helper()
	var body struct {
		SubagentResults []struct {
			AgentName          string `json:"agent_name"`
			Summary            string `json:"summary"`
			ToolFeedbackPasses int    `json:"tool_feedback_passes"`
			ToolResults        []struct {
				Status string `json:"status"`
			} `json:"tool_results"`
		} `json:"subagent_results"`
	}
	if jsonErr := json.Unmarshal(output, &body); jsonErr != nil {
		t.Fatalf("output must be JSON: %v\n%s", jsonErr, string(output))
	}
	researcher := body.SubagentResults[0]
	if researcher.AgentName != "researcher" || researcher.ToolFeedbackPasses != 1 {
		t.Fatalf("researcher result = %+v, want feedback pass", researcher)
	}
	if len(researcher.ToolResults) != 1 || researcher.ToolResults[0].Status != "pass" {
		t.Fatalf("ToolResults = %+v, want passing research tool", researcher.ToolResults)
	}
	if !strings.Contains(researcher.Summary, "cli research result for agent harness docs") {
		t.Fatalf("Summary = %q, want research output in feedback summary", researcher.Summary)
	}
}

func Test_HelperProcess_cli_research_model(t *testing.T) {
	if os.Getenv("GO_WANT_CLI_RESEARCH_MODEL") != "1" {
		return
	}
	prompt, err := io.ReadAll(os.Stdin)
	if err != nil {
		t.Fatalf("read stdin: %v", err)
	}
	text := string(prompt)
	switch {
	case strings.Contains(text, "agent: researcher") && strings.Contains(text, "tool_results:"):
		os.Stdout.WriteString(`{"summary":"used cli research result for agent harness docs"}`)
	case strings.Contains(text, "agent: researcher"):
		os.Stdout.WriteString(`{"summary":"need research","tool_requests":[{"action":"network_research","query":"agent harness docs"}]}`)
	default:
		os.Stdout.WriteString(`{"summary":"ok"}`)
	}
	os.Exit(0)
}

func Test_HelperProcess_cli_research_command(t *testing.T) {
	mode := os.Getenv("GO_WANT_CLI_RESEARCH_COMMAND")
	if mode != "1" && mode != "block" {
		return
	}
	query, err := io.ReadAll(os.Stdin)
	if err != nil {
		t.Fatalf("read stdin: %v", err)
	}
	cleanQuery := strings.TrimSpace(string(query))
	if os.Getenv("CEO_RESEARCH_QUERY") != cleanQuery {
		os.Stderr.WriteString("missing research query env")
		os.Exit(2)
	}
	if mode == "block" {
		select {}
	}
	os.Stdout.WriteString("cli research result for " + cleanQuery)
	os.Exit(0)
}
