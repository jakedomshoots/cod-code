package ceo

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func Test_Runtime_RunJob_waits_between_check_retries_when_backoff_is_set(t *testing.T) {
	// Given
	runtime := NewRuntime()
	statePath := filepath.Join(t.TempDir(), "retry-state")
	started := time.Now()

	// When
	report, err := runtime.RunJob(context.Background(), JobRequest{
		Task: "Fix a failing test",
		CheckCommand: []string{
			os.Args[0],
			"-test.run=Test_HelperProcess_retry_check",
		},
		CheckEnv:       []string{"GO_WANT_CEO_HELPER_PROCESS=retry", "GO_CEO_RETRY_STATE=" + statePath},
		CheckAttempts:  2,
		CheckBackoffMS: 20,
	})

	// Then
	if err != nil {
		t.Fatalf("RunJob returned error: %v", err)
	}
	if report.Verdict != "pass" {
		t.Fatalf("Verdict = %q, want pass", report.Verdict)
	}
	if elapsed := time.Since(started); elapsed < 20*time.Millisecond {
		t.Fatalf("elapsed = %s, want at least configured backoff", elapsed)
	}
}
