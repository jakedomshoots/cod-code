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

func Test_Run_passes_ceo_assignment_to_selected_subagent_model(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	content := `{"ceo_model_command":[` +
		`"` + os.Args[0] + `",` +
		`"-test.run=Test_HelperProcess_cli_ceo_assignment_model"` +
		`],"model_command":[` +
		`"` + os.Args[0] + `",` +
		`"-test.run=Test_HelperProcess_cli_assignment_subagent_model"` +
		`],"subagents":[{"name":"security","role":"review auth risks"}]}`
	if err := os.WriteFile(filepath.Join(root, ".ceo-harness.json"), []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	t.Setenv("GO_WANT_CLI_CEO_ASSIGNMENT_MODEL", "1")
	t.Setenv("GO_WANT_CLI_ASSIGNMENT_SUBAGENT_MODEL", "1")

	// When
	err := Run(context.Background(), &out, []string{"--workspace", root, "Fix", "auth", "flow"})
	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	var body struct {
		CEODelegation struct {
			Assignments map[string]string `json:"assignments"`
		} `json:"ceo_delegation"`
		SubagentResults []struct {
			AgentName  string `json:"agent_name"`
			Assignment string `json:"assignment"`
			Summary    string `json:"summary"`
		} `json:"subagent_results"`
	}
	if jsonErr := json.Unmarshal(out.Bytes(), &body); jsonErr != nil {
		t.Fatalf("output must be JSON: %v\n%s", jsonErr, out.String())
	}
	if body.CEODelegation.Assignments["security"] != "Inspect auth risks only." {
		t.Fatalf("CEO assignments = %#v, want security assignment", body.CEODelegation.Assignments)
	}
	if len(body.SubagentResults) != 1 || body.SubagentResults[0].AgentName != "security" {
		t.Fatalf("SubagentResults = %#v, want one security result", body.SubagentResults)
	}
	if body.SubagentResults[0].Assignment != "Inspect auth risks only." {
		t.Fatalf("Assignment = %q, want delegated assignment", body.SubagentResults[0].Assignment)
	}
	if body.SubagentResults[0].Summary != "assignment prompt received" {
		t.Fatalf("Summary = %q, want assignment-aware subagent response", body.SubagentResults[0].Summary)
	}
}

func Test_HelperProcess_cli_ceo_assignment_model(t *testing.T) {
	if os.Getenv("GO_WANT_CLI_CEO_ASSIGNMENT_MODEL") != "1" {
		return
	}
	prompt, err := io.ReadAll(os.Stdin)
	if err != nil {
		t.Fatalf("read stdin: %v", err)
	}
	text := string(prompt)
	if strings.Contains(text, "candidate_subagents") {
		os.Stdout.WriteString(`{"selected_subagents":["security"],"assignments":{"security":"Inspect auth risks only."},"summary":"Security owns auth risk."}`)
		os.Exit(0)
	}
	if strings.Contains(text, "guard_verdict: pass") {
		os.Stdout.WriteString(`{"recommended_verdict":"pass","summary":"Assignment-aware lane passed."}`)
		os.Exit(0)
	}
	os.Stderr.WriteString("unexpected CEO prompt")
	os.Exit(15)
}

func Test_HelperProcess_cli_assignment_subagent_model(t *testing.T) {
	if os.Getenv("GO_WANT_CLI_ASSIGNMENT_SUBAGENT_MODEL") != "1" {
		return
	}
	prompt, err := io.ReadAll(os.Stdin)
	if err != nil {
		t.Fatalf("read stdin: %v", err)
	}
	if !strings.Contains(string(prompt), "assignment: Inspect auth risks only.") {
		os.Stderr.WriteString("missing delegated assignment")
		os.Exit(16)
	}
	os.Stdout.WriteString(`{"summary":"assignment prompt received"}`)
	os.Exit(0)
}
