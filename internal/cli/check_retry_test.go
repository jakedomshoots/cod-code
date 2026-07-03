package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func Test_Run_retries_check_until_pass_when_attempts_flag_is_supplied(t *testing.T) {
	// Given
	var out bytes.Buffer
	statePath := filepath.Join(t.TempDir(), "retry-state")
	args := []string{"--check-attempts", "2", "--check", os.Args[0], "-test.run=Test_HelperProcess_cli_retry_check", "--", "Fix", "a", "failing", "test"}
	t.Setenv("GO_WANT_CLI_HELPER_PROCESS", "retry")
	t.Setenv("GO_CLI_RETRY_STATE", statePath)

	// When
	err := Run(context.Background(), &out, args)
	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	var body struct {
		Verdict      string `json:"verdict"`
		CheckResults []struct {
			Status      string `json:"status"`
			CheckIndex  int    `json:"check_index"`
			Attempt     int    `json:"attempt"`
			MaxAttempts int    `json:"max_attempts"`
		} `json:"check_results"`
	}
	if jsonErr := json.Unmarshal(out.Bytes(), &body); jsonErr != nil {
		t.Fatalf("output must be JSON: %v\n%s", jsonErr, out.String())
	}
	if body.Verdict != "pass" {
		t.Fatalf("Verdict = %q, want pass", body.Verdict)
	}
	if len(body.CheckResults) != 2 {
		t.Fatalf("CheckResults length = %d, want 2", len(body.CheckResults))
	}
	if body.CheckResults[0].Status != "fail" || body.CheckResults[1].Status != "pass" {
		t.Fatalf("statuses = %q, %q; want fail, pass", body.CheckResults[0].Status, body.CheckResults[1].Status)
	}
	if body.CheckResults[0].CheckIndex != 1 || body.CheckResults[1].CheckIndex != 1 {
		t.Fatalf("check indexes = %d, %d; want 1, 1", body.CheckResults[0].CheckIndex, body.CheckResults[1].CheckIndex)
	}
	if body.CheckResults[0].Attempt != 1 || body.CheckResults[1].Attempt != 2 {
		t.Fatalf("attempts = %d, %d; want 1, 2", body.CheckResults[0].Attempt, body.CheckResults[1].Attempt)
	}
	if body.CheckResults[0].MaxAttempts != 2 || body.CheckResults[1].MaxAttempts != 2 {
		t.Fatalf("max attempts = %d, %d; want 2, 2", body.CheckResults[0].MaxAttempts, body.CheckResults[1].MaxAttempts)
	}
}

func Test_HelperProcess_cli_retry_check(t *testing.T) {
	if os.Getenv("GO_WANT_CLI_HELPER_PROCESS") != "retry" {
		return
	}
	statePath := os.Getenv("GO_CLI_RETRY_STATE")
	if _, err := os.Stat(statePath); os.IsNotExist(err) {
		if writeErr := os.WriteFile(statePath, []byte("failed once"), 0o644); writeErr != nil {
			t.Fatalf("write retry state: %v", writeErr)
		}
		os.Stderr.WriteString("first cli attempt failed\n")
		os.Exit(5)
	}
	os.Stdout.WriteString("second cli attempt passed\n")
	os.Exit(0)
}
