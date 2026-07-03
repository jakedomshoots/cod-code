package ceo

import (
	"context"
	"testing"
	"time"
)

func Test_Runtime_RunJob_waits_between_subagent_retries_when_backoff_is_set(t *testing.T) {
	// Given
	runner := &failOnceRunner{calls: map[string]int{}}
	runtime := NewRuntimeWithSubagentRunner(runner)
	started := time.Now()

	// When
	report, err := runtime.RunJob(context.Background(), JobRequest{
		Task:              "Fix a failing test",
		SubagentAttempts:  2,
		SubagentBackoffMS: 20,
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
