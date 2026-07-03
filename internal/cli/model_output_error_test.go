package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"strings"
	"testing"
)

func Test_Run_classifies_invalid_model_output(t *testing.T) {
	// Given
	var out bytes.Buffer
	args := []string{
		"--subagent-attempts",
		"1",
		"--model-command",
		os.Args[0],
		"-test.run=Test_HelperProcess_cli_invalid_model_output",
		"--",
		"Inspect",
		"the",
		"workspace",
	}
	t.Setenv("GO_WANT_CLI_INVALID_MODEL_OUTPUT", "1")

	// When
	err := Run(context.Background(), &out, args)

	// Then
	if !errors.Is(err, ErrVerdictFailed) {
		t.Fatalf("Run error = %v, want ErrVerdictFailed", err)
	}
	var body struct {
		SubagentResults []struct {
			Status            string `json:"status"`
			ProviderErrorKind string `json:"provider_error_kind"`
			AttemptRecords    []struct {
				ProviderErrorKind string `json:"provider_error_kind"`
			} `json:"attempt_records"`
		} `json:"subagent_results"`
		VerificationSummary struct {
			ProviderErrorCount      int            `json:"provider_error_count"`
			ProviderErrorKindCounts map[string]int `json:"provider_error_kind_counts"`
		} `json:"verification_summary"`
	}
	if jsonErr := json.Unmarshal(out.Bytes(), &body); jsonErr != nil {
		t.Fatalf("output must be JSON: %v\n%s", jsonErr, out.String())
	}
	if len(body.SubagentResults) == 0 || body.SubagentResults[0].Status != "fail" {
		t.Fatalf("subagent results = %#v, want failed invalid output result", body.SubagentResults)
	}
	if body.SubagentResults[0].ProviderErrorKind != "model_output_invalid" {
		t.Fatalf("provider error kind = %q, want model_output_invalid", body.SubagentResults[0].ProviderErrorKind)
	}
	if len(body.SubagentResults[0].AttemptRecords) == 0 || body.SubagentResults[0].AttemptRecords[0].ProviderErrorKind != "model_output_invalid" {
		t.Fatalf("attempt records = %#v, want model_output_invalid", body.SubagentResults[0].AttemptRecords)
	}
	if body.VerificationSummary.ProviderErrorCount == 0 {
		t.Fatalf("ProviderErrorCount = 0, want invalid model output counted")
	}
	if body.VerificationSummary.ProviderErrorKindCounts["model_output_invalid"] == 0 {
		t.Fatalf("ProviderErrorKindCounts = %#v, want invalid model output counted", body.VerificationSummary.ProviderErrorKindCounts)
	}
}

func Test_Run_classifies_blank_model_output(t *testing.T) {
	// Given
	var out bytes.Buffer
	args := []string{
		"--subagent-attempts",
		"1",
		"--model-command",
		os.Args[0],
		"-test.run=Test_HelperProcess_cli_blank_model_output",
		"--",
		"Inspect",
		"the",
		"workspace",
	}
	t.Setenv("GO_WANT_CLI_BLANK_MODEL_OUTPUT", "1")

	// When
	err := Run(context.Background(), &out, args)

	// Then
	if !errors.Is(err, ErrVerdictFailed) {
		t.Fatalf("Run error = %v, want ErrVerdictFailed", err)
	}
	var body struct {
		SubagentResults []struct {
			Status            string `json:"status"`
			ProviderErrorKind string `json:"provider_error_kind"`
		} `json:"subagent_results"`
		VerificationSummary struct {
			ProviderErrorCount      int            `json:"provider_error_count"`
			ProviderErrorKindCounts map[string]int `json:"provider_error_kind_counts"`
		} `json:"verification_summary"`
	}
	if jsonErr := json.Unmarshal(out.Bytes(), &body); jsonErr != nil {
		t.Fatalf("output must be JSON: %v\n%s", jsonErr, out.String())
	}
	if len(body.SubagentResults) == 0 || body.SubagentResults[0].Status != "fail" {
		t.Fatalf("subagent results = %#v, want failed blank output result", body.SubagentResults)
	}
	if body.SubagentResults[0].ProviderErrorKind != "model_output_empty" {
		t.Fatalf("provider error kind = %q, want model_output_empty", body.SubagentResults[0].ProviderErrorKind)
	}
	if body.VerificationSummary.ProviderErrorCount == 0 {
		t.Fatalf("ProviderErrorCount = 0, want blank model output counted")
	}
	if body.VerificationSummary.ProviderErrorKindCounts["model_output_empty"] == 0 {
		t.Fatalf("ProviderErrorKindCounts = %#v, want blank model output counted", body.VerificationSummary.ProviderErrorKindCounts)
	}
}

func Test_HelperProcess_cli_invalid_model_output(t *testing.T) {
	if os.Getenv("GO_WANT_CLI_INVALID_MODEL_OUTPUT") != "1" {
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
	os.Stdout.WriteString(`{"summary":"bad confidence","confidence":2}`)
	os.Exit(0)
}

func Test_HelperProcess_cli_blank_model_output(t *testing.T) {
	if os.Getenv("GO_WANT_CLI_BLANK_MODEL_OUTPUT") != "1" {
		return
	}
	_, err := io.ReadAll(os.Stdin)
	if err != nil {
		t.Fatalf("read stdin: %v", err)
	}
	os.Stdout.WriteString(" \n\t ")
	os.Exit(0)
}
