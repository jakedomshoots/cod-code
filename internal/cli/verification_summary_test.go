package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"
)

func Test_Run_prints_verification_summary_when_check_runs(t *testing.T) {
	// Given
	var out bytes.Buffer
	args := []string{"--check", "go", "version", "--", "Fix", "a", "failing", "test"}

	// When
	err := Run(context.Background(), &out, args)

	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	var body struct {
		VerificationSummary struct {
			SubagentPassCount    int   `json:"subagent_pass_count"`
			SubagentFailCount    int   `json:"subagent_fail_count"`
			CheckAttemptCount    int   `json:"check_attempt_count"`
			CheckPassCount       int   `json:"check_pass_count"`
			CheckFailCount       int   `json:"check_fail_count"`
			CheckTotalDurationMS int64 `json:"check_total_duration_ms"`
		} `json:"verification_summary"`
	}
	if err := json.Unmarshal(out.Bytes(), &body); err != nil {
		t.Fatalf("output must be JSON: %v\n%s", err, out.String())
	}
	summary := body.VerificationSummary
	if summary.SubagentPassCount != 3 || summary.SubagentFailCount != 0 {
		t.Fatalf("subagent counts = %d pass %d fail; want 3 pass 0 fail", summary.SubagentPassCount, summary.SubagentFailCount)
	}
	if summary.CheckAttemptCount != 1 || summary.CheckPassCount != 1 || summary.CheckFailCount != 0 {
		t.Fatalf("check summary = %#v, want one passing attempt", summary)
	}
	if summary.CheckTotalDurationMS < 0 {
		t.Fatalf("CheckTotalDurationMS = %d, want nonnegative duration", summary.CheckTotalDurationMS)
	}
}
