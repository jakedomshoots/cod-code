package ceo

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func Test_Runtime_RunJob_retries_check_until_pass_when_attempts_allow(t *testing.T) {
	// Given
	runtime := NewRuntime()
	statePath := filepath.Join(t.TempDir(), "retry-state")

	// When
	report, err := runtime.RunJob(context.Background(), JobRequest{
		Task: "Fix a failing test",
		CheckCommand: []string{
			os.Args[0],
			"-test.run=Test_HelperProcess_retry_check",
		},
		CheckEnv:      []string{"GO_WANT_CEO_HELPER_PROCESS=retry", "GO_CEO_RETRY_STATE=" + statePath},
		CheckAttempts: 2,
	})

	// Then
	if err != nil {
		t.Fatalf("RunJob returned error: %v", err)
	}
	if report.Verdict != "pass" {
		t.Fatalf("Verdict = %q, want pass", report.Verdict)
	}
	if len(report.CheckResults) != 2 {
		t.Fatalf("CheckResults length = %d, want 2", len(report.CheckResults))
	}
	if report.CheckResults[0].Status != "fail" {
		t.Fatalf("first check status = %q, want fail", report.CheckResults[0].Status)
	}
	if report.CheckResults[1].Status != "pass" {
		t.Fatalf("second check status = %q, want pass", report.CheckResults[1].Status)
	}
	if report.CheckResults[0].CheckIndex != 1 || report.CheckResults[1].CheckIndex != 1 {
		t.Fatalf("check indexes = %d, %d; want 1, 1", report.CheckResults[0].CheckIndex, report.CheckResults[1].CheckIndex)
	}
	if report.CheckResults[0].Attempt != 1 || report.CheckResults[1].Attempt != 2 {
		t.Fatalf("attempts = %d, %d; want 1, 2", report.CheckResults[0].Attempt, report.CheckResults[1].Attempt)
	}
	if report.CheckResults[0].MaxAttempts != 2 || report.CheckResults[1].MaxAttempts != 2 {
		t.Fatalf("max attempts = %d, %d; want 2, 2", report.CheckResults[0].MaxAttempts, report.CheckResults[1].MaxAttempts)
	}
}
