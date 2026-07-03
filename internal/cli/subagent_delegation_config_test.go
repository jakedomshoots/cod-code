package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func Test_Run_uses_workspace_subagents_when_config_supplies_delegation(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	content := `{"subagents":[{"name":"planner","role":"break down work","allowed_actions":["read_workspace"]},{"name":"security","role":"review auth risks"}]}`
	if err := os.WriteFile(filepath.Join(root, ".ceo-harness.json"), []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	// When
	err := Run(context.Background(), &out, []string{"--workspace", root, "Fix", "auth", "flow"})

	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	var body struct {
		JobPacket struct {
			MaxSubagents int `json:"max_subagents"`
			Subagents    []struct {
				Name           string   `json:"name"`
				Role           string   `json:"role"`
				AllowedActions []string `json:"allowed_actions"`
			} `json:"subagents"`
		} `json:"job_packet"`
		SubagentResults []struct {
			AgentName      string   `json:"agent_name"`
			Role           string   `json:"role"`
			AllowedActions []string `json:"allowed_actions"`
		} `json:"subagent_results"`
	}
	if jsonErr := json.Unmarshal(out.Bytes(), &body); jsonErr != nil {
		t.Fatalf("output must be JSON: %v\n%s", jsonErr, out.String())
	}
	if body.JobPacket.MaxSubagents != 2 || len(body.JobPacket.Subagents) != 2 {
		t.Fatalf("job packet delegation = %#v, want 2 custom subagents", body.JobPacket)
	}
	if body.JobPacket.Subagents[0].Name != "planner" || body.JobPacket.Subagents[1].Role != "review auth risks" {
		t.Fatalf("job packet subagents = %#v, want configured roles", body.JobPacket.Subagents)
	}
	if len(body.JobPacket.Subagents[0].AllowedActions) != 1 || body.JobPacket.Subagents[0].AllowedActions[0] != "read_workspace" {
		t.Fatalf("job packet allowed actions = %#v, want configured action", body.JobPacket.Subagents[0].AllowedActions)
	}
	if len(body.SubagentResults) != 2 || body.SubagentResults[1].AgentName != "security" {
		t.Fatalf("subagent results = %#v, want configured delegation", body.SubagentResults)
	}
	if len(body.SubagentResults[0].AllowedActions) != 1 || body.SubagentResults[0].AllowedActions[0] != "read_workspace" {
		t.Fatalf("result allowed actions = %#v, want configured action", body.SubagentResults[0].AllowedActions)
	}
}

func Test_Run_reports_workspace_subagents_when_config_check_flag_is_supplied(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	content := `{"subagents":[{"name":"planner","role":"break down work"}]}`
	if err := os.WriteFile(filepath.Join(root, ".ceo-harness.json"), []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	// When
	err := Run(context.Background(), &out, []string{"--workspace", root, "--config-check"})

	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	var body struct {
		ModelCommandSource     string `json:"model_command_source"`
		DelegatedSubagentCount int    `json:"delegated_subagent_count"`
	}
	if jsonErr := json.Unmarshal(out.Bytes(), &body); jsonErr != nil {
		t.Fatalf("output must be JSON: %v\n%s", jsonErr, out.String())
	}
	if body.DelegatedSubagentCount != 1 {
		t.Fatalf("DelegatedSubagentCount = %d, want 1", body.DelegatedSubagentCount)
	}
	if body.ModelCommandSource != "default" {
		t.Fatalf("ModelCommandSource = %q, want default", body.ModelCommandSource)
	}
}
