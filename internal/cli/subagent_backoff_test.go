package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func Test_Run_applies_subagent_backoff_flag_between_retries(t *testing.T) {
	// Given
	var out bytes.Buffer
	statePath := filepath.Join(t.TempDir(), "model-retry-state")
	args := []string{
		"--subagent-attempts",
		"2",
		"--subagent-backoff-ms",
		"20",
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
			Attempts   int   `json:"attempts"`
			DurationMS int64 `json:"duration_ms"`
		} `json:"subagent_results"`
	}
	if jsonErr := json.Unmarshal(out.Bytes(), &body); jsonErr != nil {
		t.Fatalf("output must be JSON: %v\n%s", jsonErr, out.String())
	}
	if body.SubagentResults[0].Attempts != 2 {
		t.Fatalf("scanner attempts = %d, want 2", body.SubagentResults[0].Attempts)
	}
	if body.SubagentResults[0].DurationMS < 20 {
		t.Fatalf("scanner duration_ms = %d, want at least configured backoff", body.SubagentResults[0].DurationMS)
	}
}
