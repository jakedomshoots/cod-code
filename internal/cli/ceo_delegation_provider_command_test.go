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

func Test_Run_routes_ceo_created_subagent_to_requested_provider(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	configJSON := `{"ceo_model_command":[` +
		`"` + os.Args[0] + `",` +
		`"-test.run=Test_HelperProcess_cli_ceo_delegation_provider_model"` +
		`],"providers":{"premium":{"model_command":["sh","-c","cat >/dev/null; printf premium-specialist"]}}}`
	if err := os.WriteFile(filepath.Join(root, ".ceo-harness.json"), []byte(configJSON), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	t.Setenv("GO_WANT_CLI_CEO_PROVIDER_DELEGATION_MODEL", "1")

	// When
	err := Run(context.Background(), &out, []string{"--workspace", root, "Plan", "checkout", "UX"})
	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v\n%s", err, out.String())
	}
	var body struct {
		JobPacket struct {
			Subagents []struct {
				Name     string `json:"name"`
				Provider string `json:"provider"`
			} `json:"subagents"`
		} `json:"job_packet"`
		SubagentResults []struct {
			AgentName    string `json:"agent_name"`
			ProviderName string `json:"provider_name"`
			Summary      string `json:"summary"`
		} `json:"subagent_results"`
	}
	if jsonErr := json.Unmarshal(out.Bytes(), &body); jsonErr != nil {
		t.Fatalf("output must be JSON: %v\n%s", jsonErr, out.String())
	}
	if len(body.JobPacket.Subagents) != 2 || body.JobPacket.Subagents[1].Provider != "premium" {
		t.Fatalf("job packet subagents = %#v, want ux_reviewer provider premium", body.JobPacket.Subagents)
	}
	if len(body.SubagentResults) != 2 || body.SubagentResults[1].AgentName != "ux_reviewer" {
		t.Fatalf("subagent results = %#v, want ux_reviewer", body.SubagentResults)
	}
	if body.SubagentResults[1].ProviderName != "premium" || body.SubagentResults[1].Summary != "premium-specialist" {
		t.Fatalf("provider result = %#v, want premium specialist", body.SubagentResults[1])
	}
}

func Test_HelperProcess_cli_ceo_delegation_provider_model(t *testing.T) {
	if os.Getenv("GO_WANT_CLI_CEO_PROVIDER_DELEGATION_MODEL") != "1" {
		return
	}
	prompt, err := io.ReadAll(os.Stdin)
	if err != nil {
		t.Fatalf("read stdin: %v", err)
	}
	text := string(prompt)
	if strings.Contains(text, "candidate_subagents") {
		os.Stdout.WriteString(`{"selected_subagents":["planner","ux_reviewer"],"new_subagents":[{"name":"ux_reviewer","role":"review checkout UX","stage":3,"provider":"premium"}],"summary":"Use premium for UX specialist."}`)
		os.Exit(0)
	}
	if strings.Contains(text, "guard_verdict: pass") {
		os.Stdout.WriteString(`{"recommended_verdict":"pass","summary":"Premium specialist passed."}`)
		os.Exit(0)
	}
	os.Stderr.WriteString("unexpected CEO prompt")
	os.Exit(16)
}
