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

func Test_Run_uses_agent_model_command_when_workspace_config_supplies_agent_route(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	configJSON := `{"agent_model_commands":{"scanner":[` +
		`"` + os.Args[0] + `",` +
		`"-test.run=Test_HelperProcess_cli_agent_model_command"` +
		`]}}`
	if err := os.WriteFile(filepath.Join(root, ".ceo-harness.json"), []byte(configJSON), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	t.Setenv("GO_WANT_CLI_AGENT_MODEL_COMMAND", "scanner")

	// When
	err := Run(context.Background(), &out, []string{"--workspace", root, "Fix", "a", "failing", "test"})
	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	var body struct {
		SubagentResults []struct {
			AgentName string `json:"agent_name"`
			Summary   string `json:"summary"`
		} `json:"subagent_results"`
	}
	if err := json.Unmarshal(out.Bytes(), &body); err != nil {
		t.Fatalf("output must be JSON: %v\n%s", err, out.String())
	}
	if body.SubagentResults[0].AgentName != "scanner" {
		t.Fatalf("first agent = %q, want scanner", body.SubagentResults[0].AgentName)
	}
	if !strings.Contains(body.SubagentResults[0].Summary, "scanner routed model response") {
		t.Fatalf("scanner summary = %q, want routed model response", body.SubagentResults[0].Summary)
	}
	if body.SubagentResults[1].Summary != "local deterministic model response" {
		t.Fatalf("coder summary = %q, want default local response", body.SubagentResults[1].Summary)
	}
}

func Test_HelperProcess_cli_agent_model_command(t *testing.T) {
	if os.Getenv("GO_WANT_CLI_AGENT_MODEL_COMMAND") != "scanner" {
		return
	}
	prompt, err := io.ReadAll(os.Stdin)
	if err != nil {
		t.Fatalf("read stdin: %v", err)
	}
	if !strings.Contains(string(prompt), "agent: scanner") {
		os.Stderr.WriteString("expected scanner prompt")
		os.Exit(7)
	}
	os.Stdout.WriteString("scanner routed model response")
}
