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

func Test_Run_uses_ceo_revision_attempt_after_model_veto(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "app.txt"), []byte("bad"), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	args := []string{
		"--workspace", root,
		"--apply-model-patches",
		"--ceo-revision-attempts", "1",
		"--model-command",
		os.Args[0],
		"-test.run=Test_HelperProcess_cli_ceo_revision_subagent_model",
		"--",
		"--ceo-model-command",
		os.Args[0],
		"-test.run=Test_HelperProcess_cli_ceo_revision_ceo_model",
		"--",
		"Repair",
		"app",
	}
	t.Setenv("GO_WANT_CLI_CEO_REVISION_SUBAGENT_MODEL", "1")
	t.Setenv("GO_WANT_CLI_CEO_REVISION_CEO_MODEL", "1")

	// When
	err := Run(context.Background(), &out, args)
	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	got, readErr := os.ReadFile(filepath.Join(root, "app.txt"))
	if readErr != nil {
		t.Fatalf("read fixed file: %v", readErr)
	}
	if string(got) != "good" {
		t.Fatalf("content = %q, want good", string(got))
	}
	var body struct {
		SubagentResults []struct {
			AgentName string `json:"agent_name"`
			Role      string `json:"role"`
		} `json:"subagent_results"`
		CEOReview struct {
			Summary string `json:"summary"`
		} `json:"ceo_review"`
		Verdict string `json:"verdict"`
	}
	if jsonErr := json.Unmarshal(out.Bytes(), &body); jsonErr != nil {
		t.Fatalf("output must be JSON: %v\n%s", jsonErr, out.String())
	}
	if body.Verdict != "pass" || body.CEOReview.Summary != "Revision accepted." {
		t.Fatalf("verdict = %q CEO review = %#v, want final pass review", body.Verdict, body.CEOReview)
	}
	if !hasSubagentRole(body.SubagentResults, "coder", "revise work after CEO review feedback") {
		t.Fatalf("subagent results = %#v, want CEO revision coder", body.SubagentResults)
	}
}

func Test_Run_uses_workspace_ceo_revision_attempt_default_after_model_veto(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "app.txt"), []byte("bad"), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	configJSON := `{"ceo_revision_attempts":1,"model_command":[` +
		`"` + os.Args[0] + `",` +
		`"-test.run=Test_HelperProcess_cli_ceo_revision_subagent_model"` +
		`],"ceo_model_command":[` +
		`"` + os.Args[0] + `",` +
		`"-test.run=Test_HelperProcess_cli_ceo_revision_ceo_model"` +
		`]}`
	if err := os.WriteFile(filepath.Join(root, ".ceo-harness.json"), []byte(configJSON), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	t.Setenv("GO_WANT_CLI_CEO_REVISION_SUBAGENT_MODEL", "1")
	t.Setenv("GO_WANT_CLI_CEO_REVISION_CEO_MODEL", "1")

	// When
	err := Run(context.Background(), &out, []string{"--workspace", root, "--apply-model-patches", "Repair", "app"})
	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	got, readErr := os.ReadFile(filepath.Join(root, "app.txt"))
	if readErr != nil {
		t.Fatalf("read fixed file: %v", readErr)
	}
	if string(got) != "good" {
		t.Fatalf("content = %q, want good", string(got))
	}
}

func Test_HelperProcess_cli_ceo_revision_subagent_model(t *testing.T) {
	if os.Getenv("GO_WANT_CLI_CEO_REVISION_SUBAGENT_MODEL") != "1" {
		return
	}
	prompt, err := io.ReadAll(os.Stdin)
	if err != nil {
		t.Fatalf("read stdin: %v", err)
	}
	if strings.Contains(string(prompt), "CEO review failed") {
		os.Stdout.WriteString(`{"summary":"revision patch","patches":[{"path":"app.txt","old":"bad","new":"good"}]}`)
		os.Exit(0)
	}
	os.Stdout.WriteString(`{"summary":"initial pass","evidence":["ran"]}`)
	os.Exit(0)
}

func Test_HelperProcess_cli_ceo_revision_ceo_model(t *testing.T) {
	if os.Getenv("GO_WANT_CLI_CEO_REVISION_CEO_MODEL") != "1" {
		return
	}
	prompt, err := io.ReadAll(os.Stdin)
	if err != nil {
		t.Fatalf("read stdin: %v", err)
	}
	text := string(prompt)
	if strings.Contains(text, "candidate_subagents") {
		os.Stdout.WriteString(`{"selected_subagents":["coder","reviewer"],"summary":"Use coder and reviewer."}`)
		os.Exit(0)
	}
	if strings.Contains(text, "+good") {
		os.Stdout.WriteString(`{"recommended_verdict":"pass","summary":"Revision accepted."}`)
		os.Exit(0)
	}
	os.Stdout.WriteString(`{"recommended_verdict":"fail","summary":"Patch app.txt before accepting."}`)
	os.Exit(0)
}

func hasSubagentRole(results []struct {
	AgentName string `json:"agent_name"`
	Role      string `json:"role"`
}, agentName string, role string,
) bool {
	for _, result := range results {
		if result.AgentName == agentName && result.Role == role {
			return true
		}
	}
	return false
}
