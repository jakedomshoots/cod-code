package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func Test_Run_applies_check_backoff_flag_between_retries(t *testing.T) {
	// Given
	var out bytes.Buffer
	statePath := filepath.Join(t.TempDir(), "retry-state")
	args := []string{
		"--check-attempts",
		"2",
		"--check-backoff-ms",
		"20",
		"--check",
		os.Args[0],
		"-test.run=Test_HelperProcess_cli_retry_check",
		"--",
		"Fix",
		"a",
		"failing",
		"test",
	}
	t.Setenv("GO_WANT_CLI_HELPER_PROCESS", "retry")
	t.Setenv("GO_CLI_RETRY_STATE", statePath)
	started := time.Now()

	// When
	err := Run(context.Background(), &out, args)

	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	var body struct {
		CheckResults []struct {
			Attempt    int   `json:"attempt"`
			DurationMS int64 `json:"duration_ms"`
		} `json:"check_results"`
	}
	if jsonErr := json.Unmarshal(out.Bytes(), &body); jsonErr != nil {
		t.Fatalf("output must be JSON: %v\n%s", jsonErr, out.String())
	}
	if len(body.CheckResults) != 2 {
		t.Fatalf("CheckResults length = %d, want 2", len(body.CheckResults))
	}
	if body.CheckResults[1].Attempt != 2 {
		t.Fatalf("second attempt = %d, want 2", body.CheckResults[1].Attempt)
	}
	if elapsed := time.Since(started); elapsed < 20*time.Millisecond {
		t.Fatalf("elapsed = %s, want at least configured backoff", elapsed)
	}
}
