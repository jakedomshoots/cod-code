package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func Test_Run_times_out_network_research_when_tool_timeout_flag_is_supplied(t *testing.T) {
	// Given
	var out bytes.Buffer
	t.Setenv("GO_WANT_CLI_RESEARCH_MODEL", "1")
	t.Setenv("GO_WANT_CLI_RESEARCH_COMMAND", "block")

	// When
	err := Run(context.Background(), &out, []string{
		"--tool-command-timeout-ms",
		"1",
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
			AgentName   string `json:"agent_name"`
			ToolResults []struct {
				Status string `json:"status"`
				Error  string `json:"error"`
			} `json:"tool_results"`
		} `json:"subagent_results"`
	}
	if jsonErr := json.Unmarshal(out.Bytes(), &body); jsonErr != nil {
		t.Fatalf("output must be JSON: %v\n%s", jsonErr, out.String())
	}
	if len(body.SubagentResults) == 0 || body.SubagentResults[0].AgentName != "researcher" {
		t.Fatalf("SubagentResults = %#v, want researcher first", body.SubagentResults)
	}
	tools := body.SubagentResults[0].ToolResults
	if len(tools) != 1 || tools[0].Status != "fail" {
		t.Fatalf("ToolResults = %#v, want failed research timeout", tools)
	}
	if !strings.Contains(tools[0].Error, "context deadline exceeded") {
		t.Fatalf("tool error = %q, want context deadline exceeded", tools[0].Error)
	}
}
