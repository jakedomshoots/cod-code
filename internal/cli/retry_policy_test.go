package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func Test_Run_uses_workspace_retry_policy_for_check_retries(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	statePath := filepath.Join(t.TempDir(), "retry-state")
	configJSON := fmt.Sprintf(
		`{"check_attempts":2,"check_backoff_ms":20,"check_command":[%q,"-test.run=Test_HelperProcess_cli_retry_check"]}`,
		os.Args[0],
	)
	if err := os.WriteFile(filepath.Join(root, ".ceo-harness.json"), []byte(configJSON), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	t.Setenv("GO_WANT_CLI_HELPER_PROCESS", "retry")
	t.Setenv("GO_CLI_RETRY_STATE", statePath)

	// When
	err := Run(context.Background(), &out, []string{"--workspace", root, "Fix", "a", "failing", "test"})
	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	var body struct {
		CheckResults []struct {
			Attempt int `json:"attempt"`
		} `json:"check_results"`
	}
	if jsonErr := json.Unmarshal(out.Bytes(), &body); jsonErr != nil {
		t.Fatalf("output must be JSON: %v\n%s", jsonErr, out.String())
	}
	if len(body.CheckResults) != 2 || body.CheckResults[1].Attempt != 2 {
		t.Fatalf("check results = %#v, want two attempts from workspace retry policy", body.CheckResults)
	}
}

func Test_Run_uses_workspace_retry_policy_for_subagent_retries(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	statePath := filepath.Join(t.TempDir(), "model-retry-state")
	configJSON := `{"subagent_attempts":2,"subagent_backoff_ms":20}`
	if err := os.WriteFile(filepath.Join(root, ".ceo-harness.json"), []byte(configJSON), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	args := []string{
		"--workspace",
		root,
		"--model-command",
		os.Args[0],
		"-test.run=Test_HelperProcess_cli_model_retry",
		"--",
		"Fix",
		"a",
		"failing",
		"test",
	}
	t.Setenv("GO_WANT_CLI_MODEL_RETRY", "retry")
	t.Setenv("GO_CLI_MODEL_RETRY_STATE", statePath)

	// When
	err := Run(context.Background(), &out, args)
	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	var body struct {
		SubagentResults []struct {
			Attempts int `json:"attempts"`
		} `json:"subagent_results"`
	}
	if jsonErr := json.Unmarshal(out.Bytes(), &body); jsonErr != nil {
		t.Fatalf("output must be JSON: %v\n%s", jsonErr, out.String())
	}
	if body.SubagentResults[0].Attempts != 2 {
		t.Fatalf("scanner attempts = %d, want workspace retry policy attempts", body.SubagentResults[0].Attempts)
	}
}
