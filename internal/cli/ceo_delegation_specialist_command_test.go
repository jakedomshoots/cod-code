package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"
)

func Test_Run_uses_ceo_model_command_to_create_specialist_subagent(t *testing.T) {
	// Given
	var out bytes.Buffer
	args := []string{
		"--ceo-model-command",
		os.Args[0],
		"-test.run=Test_HelperProcess_cli_ceo_delegation_specialist_model",
		"--",
		"Plan",
		"a",
		"mobile",
		"checkout",
	}
	t.Setenv("GO_WANT_CLI_CEO_SPECIALIST_DELEGATION_MODEL", "1")

	// When
	err := Run(context.Background(), &out, args)
	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v\n%s", err, out.String())
	}
	var body struct {
		CEODelegation struct {
			NewSubagents []struct {
				Name string `json:"name"`
			} `json:"new_subagents"`
		} `json:"ceo_delegation"`
		JobPacket struct {
			Subagents []struct {
				Name            string   `json:"name"`
				Stage           int      `json:"stage"`
				MaxContextBytes int      `json:"max_context_bytes"`
				AllowedActions  []string `json:"allowed_actions"`
			} `json:"subagents"`
		} `json:"job_packet"`
		SubagentResults []struct {
			AgentName        string `json:"agent_name"`
			Stage            int    `json:"stage"`
			ContextBytes     int    `json:"context_bytes"`
			ContextTruncated bool   `json:"context_truncated"`
		} `json:"subagent_results"`
		Verdict string `json:"verdict"`
	}
	if jsonErr := json.Unmarshal(out.Bytes(), &body); jsonErr != nil {
		t.Fatalf("output must be JSON: %v\n%s", jsonErr, out.String())
	}
	if len(body.CEODelegation.NewSubagents) != 1 || body.CEODelegation.NewSubagents[0].Name != "ux_reviewer" {
		t.Fatalf("new subagents = %#v, want ux_reviewer", body.CEODelegation.NewSubagents)
	}
	if len(body.JobPacket.Subagents) != 2 || body.JobPacket.Subagents[1].Name != "ux_reviewer" {
		t.Fatalf("job packet subagents = %#v, want planner and ux_reviewer", body.JobPacket.Subagents)
	}
	if body.JobPacket.Subagents[1].Stage != 3 {
		t.Fatalf("job packet stage = %d, want 3", body.JobPacket.Subagents[1].Stage)
	}
	if body.JobPacket.Subagents[1].MaxContextBytes != 20 {
		t.Fatalf("job packet max context = %d, want 20", body.JobPacket.Subagents[1].MaxContextBytes)
	}
	if strings.Join(body.JobPacket.Subagents[1].AllowedActions, ",") != "read_workspace,run_checks" {
		t.Fatalf("allowed actions = %#v, want read_workspace and run_checks", body.JobPacket.Subagents[1].AllowedActions)
	}
	if len(body.SubagentResults) != 2 || body.SubagentResults[1].AgentName != "ux_reviewer" {
		t.Fatalf("subagent results = %#v, want ux_reviewer to run", body.SubagentResults)
	}
	if body.SubagentResults[1].Stage != 3 {
		t.Fatalf("result stage = %d, want 3", body.SubagentResults[1].Stage)
	}
	if body.SubagentResults[1].ContextBytes != 20 || !body.SubagentResults[1].ContextTruncated {
		t.Fatalf("context = %d truncated=%v, want 20 truncated", body.SubagentResults[1].ContextBytes, body.SubagentResults[1].ContextTruncated)
	}
	if body.Verdict != "pass" {
		t.Fatalf("Verdict = %q, want pass", body.Verdict)
	}
}

func Test_HelperProcess_cli_ceo_delegation_specialist_model(t *testing.T) {
	if os.Getenv("GO_WANT_CLI_CEO_SPECIALIST_DELEGATION_MODEL") != "1" {
		return
	}
	prompt, err := io.ReadAll(os.Stdin)
	if err != nil {
		t.Fatalf("read stdin: %v", err)
	}
	text := string(prompt)
	if strings.Contains(text, "candidate_subagents") {
		os.Stdout.WriteString(`{"selected_subagents":["planner","ux_reviewer"],"new_subagents":[{"name":"ux_reviewer","role":"review mobile checkout UX","stage":3,"max_context_bytes":20,"allowed_actions":["read_workspace","run_checks"]}],"summary":"Checkout planning needs UX review."}`)
		os.Exit(0)
	}
	if strings.Contains(text, "guard_verdict: pass") {
		os.Stdout.WriteString(`{"recommended_verdict":"pass","summary":"Specialist lane passed."}`)
		os.Exit(0)
	}
	os.Stderr.WriteString("unexpected CEO prompt")
	os.Exit(15)
}
