package ceo

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"ceoharness/internal/model"
	"ceoharness/internal/subagent"
)

type rateLimitedOnceRunner struct {
	mu    sync.Mutex
	calls map[string]int
}

func (r *rateLimitedOnceRunner) Run(ctx context.Context, packet subagent.TaskPacket) (subagent.Result, error) {
	if err := ctx.Err(); err != nil {
		return subagent.Result{}, err
	}
	r.mu.Lock()
	r.calls[packet.AgentName]++
	calls := r.calls[packet.AgentName]
	r.mu.Unlock()
	if packet.AgentName == "scanner" && calls == 1 {
		return subagent.Result{}, fmt.Errorf("complete prompt: %w", model.ErrHTTPRateLimited)
	}
	return subagent.Result{
		AgentName:       packet.AgentName,
		Role:            packet.Role,
		Status:          "pass",
		ContextReceived: packet.ContextMode,
		Summary:         "ok",
	}, nil
}

func (r *rateLimitedOnceRunner) callCount(agentName string) int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.calls[agentName]
}

func Test_Runtime_RunJob_retries_rate_limited_subagent_once_without_retry_config(t *testing.T) {
	// Given
	runner := &rateLimitedOnceRunner{calls: map[string]int{}}
	runtime := NewRuntimeWithSubagentRunner(runner)

	// When
	report, err := runtime.RunJob(context.Background(), JobRequest{
		Task: "Fix a rate limited provider call",
	})
	// Then
	if err != nil {
		t.Fatalf("RunJob returned error: %v", err)
	}
	if report.Verdict != "pass" {
		t.Fatalf("Verdict = %q, want pass", report.Verdict)
	}
	scanner := report.SubagentResults[0]
	if scanner.Attempts != 2 {
		t.Fatalf("scanner attempts = %d, want one automatic retry", scanner.Attempts)
	}
	if runner.callCount("scanner") != 2 {
		t.Fatalf("scanner calls = %d, want 2", runner.callCount("scanner"))
	}
	if len(scanner.AttemptRecords) != 2 || scanner.AttemptRecords[0].Status != "fail" || scanner.AttemptRecords[1].Status != "pass" {
		t.Fatalf("scanner attempt records = %#v, want fail then pass", scanner.AttemptRecords)
	}
}

func Test_retryBackoffForError_uses_retry_after_when_larger_than_configured_backoff(t *testing.T) {
	// Given
	err := fmt.Errorf("complete prompt: %w", &model.HTTPStatusError{
		StatusCode:   429,
		Kind:         model.HTTPErrorKindRateLimited,
		RetryAfterMS: 25,
	})

	// When
	got := retryBackoffForError(err, 5*time.Millisecond)

	// Then
	if got != 25*time.Millisecond {
		t.Fatalf("retry backoff = %s, want 25ms", got)
	}
}
