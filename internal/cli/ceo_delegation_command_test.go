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

func Test_Run_uses_ceo_model_command_to_select_workspace_subagents(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	content := `{"ceo_model_command":[` +
		`"` + os.Args[0] + `",` +
		`"-test.run=Test_HelperProcess_cli_ceo_delegation_model"` +
		`],"subagents":[{"name":"planner","role":"break down work"},{"name":"security","role":"review auth risks"}]}`
	if err := os.WriteFile(filepath.Join(root, ".ceo-harness.json"), []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	t.Setenv("GO_WANT_CLI_CEO_DELEGATION_MODEL", "1")

	// When
	err := Run(context.Background(), &out, []string{"--workspace", root, "Fix", "auth", "flow"})
	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	var body struct {
		CEODelegation struct {
			SelectedSubagents []string `json:"selected_subagents"`
		} `json:"ceo_delegation"`
		SubagentResults []struct {
			AgentName string `json:"agent_name"`
		} `json:"subagent_results"`
		Verdict string `json:"verdict"`
	}
	if jsonErr := json.Unmarshal(out.Bytes(), &body); jsonErr != nil {
		t.Fatalf("output must be JSON: %v\n%s", jsonErr, out.String())
	}
	if len(body.CEODelegation.SelectedSubagents) != 1 || body.CEODelegation.SelectedSubagents[0] != "security" {
		t.Fatalf("CEO delegation = %#v, want security", body.CEODelegation)
	}
	if len(body.SubagentResults) != 1 || body.SubagentResults[0].AgentName != "security" {
		t.Fatalf("subagent results = %#v, want only security", body.SubagentResults)
	}
	if body.Verdict != "pass" {
		t.Fatalf("Verdict = %q, want pass", body.Verdict)
	}
}

func Test_Run_uses_ceo_model_command_to_select_default_subagents(t *testing.T) {
	// Given
	var out bytes.Buffer
	args := []string{
		"--ceo-model-command",
		os.Args[0],
		"-test.run=Test_HelperProcess_cli_ceo_delegation_default_model",
		"--",
		"Fix",
		"a",
		"failing",
		"test",
	}
	t.Setenv("GO_WANT_CLI_CEO_DEFAULT_DELEGATION_MODEL", "1")

	// When
	err := Run(context.Background(), &out, args)
	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	var body struct {
		CEODelegation struct {
			SelectedSubagents []string `json:"selected_subagents"`
		} `json:"ceo_delegation"`
		SubagentResults []struct {
			AgentName string `json:"agent_name"`
		} `json:"subagent_results"`
		Verdict string `json:"verdict"`
	}
	if jsonErr := json.Unmarshal(out.Bytes(), &body); jsonErr != nil {
		t.Fatalf("output must be JSON: %v\n%s", jsonErr, out.String())
	}
	if len(body.CEODelegation.SelectedSubagents) != 1 || body.CEODelegation.SelectedSubagents[0] != "coder" {
		t.Fatalf("CEO delegation = %#v, want coder", body.CEODelegation)
	}
	if len(body.SubagentResults) != 1 || body.SubagentResults[0].AgentName != "coder" {
		t.Fatalf("subagent results = %#v, want only coder", body.SubagentResults)
	}
	if body.Verdict != "pass" {
		t.Fatalf("Verdict = %q, want pass", body.Verdict)
	}
}

func Test_HelperProcess_cli_ceo_delegation_default_model(t *testing.T) {
	if os.Getenv("GO_WANT_CLI_CEO_DEFAULT_DELEGATION_MODEL") != "1" {
		return
	}
	prompt, err := io.ReadAll(os.Stdin)
	if err != nil {
		t.Fatalf("read stdin: %v", err)
	}
	text := string(prompt)
	if strings.Contains(text, "candidate_subagents") {
		if !strings.Contains(text, "scanner") || !strings.Contains(text, "coder") || !strings.Contains(text, "reviewer") {
			os.Stderr.WriteString("missing default candidates")
			os.Exit(13)
		}
		os.Stdout.WriteString(`{"selected_subagents":["coder"],"summary":"Coding task only needs coder."}`)
		os.Exit(0)
	}
	if strings.Contains(text, "guard_verdict: pass") {
		os.Stdout.WriteString(`{"recommended_verdict":"pass","summary":"Selected default lane passed."}`)
		os.Exit(0)
	}
	os.Stderr.WriteString("unexpected CEO prompt")
	os.Exit(14)
}

func Test_HelperProcess_cli_ceo_delegation_model(t *testing.T) {
	if os.Getenv("GO_WANT_CLI_CEO_DELEGATION_MODEL") != "1" {
		return
	}
	prompt, err := io.ReadAll(os.Stdin)
	if err != nil {
		t.Fatalf("read stdin: %v", err)
	}
	if strings.Contains(string(prompt), "candidate_subagents") {
		os.Stdout.WriteString(`{"selected_subagents":["security"],"summary":"Auth work needs security review."}`)
		os.Exit(0)
	}
	if strings.Contains(string(prompt), "guard_verdict: pass") {
		os.Stdout.WriteString(`{"recommended_verdict":"pass","summary":"Selected lane passed."}`)
		os.Exit(0)
	}
	os.Stderr.WriteString("unexpected CEO prompt")
	os.Exit(12)
}
