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

func Test_Run_retries_subagent_error_when_subagent_attempts_flag_is_supplied(t *testing.T) {
	// Given
	var out bytes.Buffer
	statePath := filepath.Join(t.TempDir(), "model-retry-state")
	args := []string{
		"--subagent-attempts",
		"2",
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
			Attempts       int      `json:"attempts"`
			Summary        string   `json:"summary"`
			AttemptErrors  []string `json:"attempt_errors"`
			AttemptRecords []struct {
				Attempt int    `json:"attempt"`
				Status  string `json:"status"`
				Error   string `json:"error"`
			} `json:"attempt_records"`
		} `json:"subagent_results"`
		VerificationSummary struct {
			SubagentAttemptCount        int `json:"subagent_attempt_count"`
			SubagentRetryCount          int `json:"subagent_retry_count"`
			SubagentRetriedCount        int `json:"subagent_retried_count"`
			SubagentRetryExhaustedCount int `json:"subagent_retry_exhausted_count"`
		} `json:"verification_summary"`
		Verdict string `json:"verdict"`
	}
	if err := json.Unmarshal(out.Bytes(), &body); err != nil {
		t.Fatalf("output must be JSON: %v\n%s", err, out.String())
	}
	if body.Verdict != "pass" {
		t.Fatalf("Verdict = %q, want pass", body.Verdict)
	}
	if body.SubagentResults[0].Attempts != 2 {
		t.Fatalf("first subagent attempts = %d, want 2", body.SubagentResults[0].Attempts)
	}
	if !strings.Contains(body.SubagentResults[0].Summary, "retry model response") {
		t.Fatalf("Summary = %q, want retry model response", body.SubagentResults[0].Summary)
	}
	if len(body.SubagentResults[0].AttemptErrors) != 1 {
		t.Fatalf("attempt errors length = %d, want 1", len(body.SubagentResults[0].AttemptErrors))
	}
	records := body.SubagentResults[0].AttemptRecords
	if len(records) != 2 {
		t.Fatalf("attempt records length = %d, want 2", len(records))
	}
	if records[0].Attempt != 1 || records[0].Status != "fail" || !strings.Contains(records[0].Error, "first model attempt failed") {
		t.Fatalf("first attempt record = %#v, want failed first model attempt", records[0])
	}
	if records[1].Attempt != 2 || records[1].Status != "pass" || records[1].Error != "" {
		t.Fatalf("second attempt record = %#v, want clean passing retry", records[1])
	}
	if body.VerificationSummary.SubagentAttemptCount != 4 ||
		body.VerificationSummary.SubagentRetryCount != 1 ||
		body.VerificationSummary.SubagentRetriedCount != 1 ||
		body.VerificationSummary.SubagentRetryExhaustedCount != 0 {
		t.Fatalf("verification summary = %#v, want one successful retry across four attempts", body.VerificationSummary)
	}
}

func Test_HelperProcess_cli_model_retry(t *testing.T) {
	if os.Getenv("GO_WANT_CLI_MODEL_RETRY") != "retry" {
		return
	}
	prompt, err := io.ReadAll(os.Stdin)
	if err != nil {
		t.Fatalf("read stdin: %v", err)
	}
	isScanner := strings.Contains(string(prompt), "agent: scanner")
	statePath := os.Getenv("GO_CLI_MODEL_RETRY_STATE")
	if isScanner {
		_, statErr := os.Stat(statePath)
		if statErr == nil {
			os.Stdout.WriteString("retry model response")
			return
		}
		if !os.IsNotExist(statErr) {
			t.Fatalf("stat retry state: %v", statErr)
		}
		if writeErr := os.WriteFile(statePath, []byte("failed once"), 0o644); writeErr != nil {
			t.Fatalf("write retry state: %v", writeErr)
		}
		os.Stderr.WriteString("first model attempt failed")
		os.Exit(6)
	}
	os.Stdout.WriteString("retry model response")
}
