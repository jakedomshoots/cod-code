package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func Test_Run_uses_command_model_when_model_command_flag_is_supplied(t *testing.T) {
	// Given
	var out bytes.Buffer
	args := []string{
		"--model-command",
		os.Args[0],
		"-test.run=Test_HelperProcess_cli_model_command",
		"--",
		"Fix",
		"a",
		"failing",
		"test",
	}
	t.Setenv("GO_WANT_CLI_MODEL_COMMAND", "echo")

	// When
	err := Run(context.Background(), &out, args)
	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	var body struct {
		SubagentResults []struct {
			Summary string `json:"summary"`
		} `json:"subagent_results"`
	}
	if err := json.Unmarshal(out.Bytes(), &body); err != nil {
		t.Fatalf("output must be JSON: %v\n%s", err, out.String())
	}
	if len(body.SubagentResults) != 3 {
		t.Fatalf("SubagentResults length = %d, want 3", len(body.SubagentResults))
	}
	if body.SubagentResults[0].Summary != "cli command model response" {
		t.Fatalf("Summary = %q, want command model response", body.SubagentResults[0].Summary)
	}
}

func Test_Run_uses_command_model_when_env_command_is_supplied(t *testing.T) {
	// Given
	var out bytes.Buffer
	commandJSON := `[` +
		`"` + os.Args[0] + `",` +
		`"-test.run=Test_HelperProcess_cli_model_command"` +
		`]`
	t.Setenv("CEO_MODEL_COMMAND_JSON", commandJSON)
	t.Setenv("GO_WANT_CLI_MODEL_COMMAND", "echo")

	// When
	err := Run(context.Background(), &out, []string{"Fix", "a", "failing", "test"})
	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	var body struct {
		SubagentResults []struct {
			Summary string `json:"summary"`
		} `json:"subagent_results"`
	}
	if err := json.Unmarshal(out.Bytes(), &body); err != nil {
		t.Fatalf("output must be JSON: %v\n%s", err, out.String())
	}
	if body.SubagentResults[0].Summary != "cli command model response" {
		t.Fatalf("Summary = %q, want command model response", body.SubagentResults[0].Summary)
	}
}

func Test_Run_uses_command_model_when_workspace_config_is_supplied(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	commandJSON := `{"model_command":[` +
		`"` + os.Args[0] + `",` +
		`"-test.run=Test_HelperProcess_cli_model_command"` +
		`]}`
	if err := os.WriteFile(filepath.Join(root, ".ceo-harness.json"), []byte(commandJSON), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	t.Setenv("GO_WANT_CLI_MODEL_COMMAND", "echo")

	// When
	err := Run(context.Background(), &out, []string{"--workspace", root, "Fix", "a", "failing", "test"})
	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	var body struct {
		SubagentResults []struct {
			Summary string `json:"summary"`
		} `json:"subagent_results"`
	}
	if err := json.Unmarshal(out.Bytes(), &body); err != nil {
		t.Fatalf("output must be JSON: %v\n%s", err, out.String())
	}
	if body.SubagentResults[0].Summary != "cli command model response" {
		t.Fatalf("Summary = %q, want command model response", body.SubagentResults[0].Summary)
	}
}

func Test_Run_times_out_model_command_when_timeout_flag_is_supplied(t *testing.T) {
	// Given
	var out bytes.Buffer
	args := []string{
		"--model-command-timeout-ms",
		"1",
		"--model-command",
		os.Args[0],
		"-test.run=Test_HelperProcess_cli_model_command",
		"--",
		"Fix",
		"a",
		"failing",
		"test",
	}
	t.Setenv("GO_WANT_CLI_MODEL_COMMAND", "block")

	// When
	err := Run(context.Background(), &out, args)

	// Then
	if !errors.Is(err, ErrVerdictFailed) {
		t.Fatalf("Run error = %v, want ErrVerdictFailed", err)
	}
	var body struct {
		SubagentResults []struct {
			Status            string   `json:"status"`
			ProviderErrorKind string   `json:"provider_error_kind"`
			AttemptErrors     []string `json:"attempt_errors"`
		} `json:"subagent_results"`
		VerificationSummary struct {
			ProviderErrorCount int `json:"provider_error_count"`
		} `json:"verification_summary"`
	}
	if jsonErr := json.Unmarshal(out.Bytes(), &body); jsonErr != nil {
		t.Fatalf("output must be JSON: %v\n%s", jsonErr, out.String())
	}
	if len(body.SubagentResults) == 0 || body.SubagentResults[0].Status != "fail" {
		t.Fatalf("subagent results = %#v, want failed timeout result", body.SubagentResults)
	}
	if len(body.SubagentResults[0].AttemptErrors) == 0 || !strings.Contains(body.SubagentResults[0].AttemptErrors[0], "context deadline exceeded") {
		t.Fatalf("attempt errors = %#v, want context deadline exceeded", body.SubagentResults[0].AttemptErrors)
	}
	if body.SubagentResults[0].ProviderErrorKind != "command_timeout" {
		t.Fatalf("provider error kind = %q, want command_timeout", body.SubagentResults[0].ProviderErrorKind)
	}
	if body.VerificationSummary.ProviderErrorCount == 0 {
		t.Fatalf("ProviderErrorCount = 0, want command timeout counted")
	}
}

func Test_HelperProcess_cli_model_command(t *testing.T) {
	switch os.Getenv("GO_WANT_CLI_MODEL_COMMAND") {
	case "echo":
	case "block":
		select {}
	default:
		return
	}
	prompt, err := io.ReadAll(os.Stdin)
	if err != nil {
		t.Fatalf("read stdin: %v", err)
	}
	if !strings.Contains(string(prompt), "role:") {
		os.Stderr.WriteString("missing role in prompt")
		os.Exit(2)
	}
	os.Stdout.WriteString("cli command model response")
	os.Exit(0)
}
