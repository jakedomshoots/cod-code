package ceo

import (
	"context"
	"sync"
	"testing"

	"ceoharness/internal/subagent"
)

type failOnceRunner struct {
	mu    sync.Mutex
	calls map[string]int
}

func (r *failOnceRunner) Run(ctx context.Context, packet subagent.TaskPacket) (subagent.Result, error) {
	if err := ctx.Err(); err != nil {
		return subagent.Result{}, err
	}
	r.mu.Lock()
	r.calls[packet.AgentName]++
	calls := r.calls[packet.AgentName]
	r.mu.Unlock()
	status := "pass"
	if packet.AgentName == "scanner" && calls == 1 {
		status = "fail"
	}
	return subagent.Result{
		AgentName:       packet.AgentName,
		Status:          status,
		ContextReceived: packet.ContextMode,
		Summary:         "fake subagent result",
	}, nil
}

func (r *failOnceRunner) callCount(agentName string) int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.calls[agentName]
}

func Test_Runtime_RunJob_retries_failed_subagent_when_attempts_allow(t *testing.T) {
	// Given
	runner := &failOnceRunner{calls: map[string]int{}}
	runtime := NewRuntimeWithSubagentRunner(runner)

	// When
	report, err := runtime.RunJob(context.Background(), JobRequest{
		Task:             "Fix a failing test",
		SubagentAttempts: 2,
	})
	// Then
	if err != nil {
		t.Fatalf("RunJob returned error: %v", err)
	}
	if report.Verdict != "pass" {
		t.Fatalf("Verdict = %q, want pass", report.Verdict)
	}
	if runner.callCount("scanner") != 2 {
		t.Fatalf("scanner calls = %d, want 2", runner.callCount("scanner"))
	}
	if report.SubagentResults[0].Attempts != 2 {
		t.Fatalf("scanner attempts = %d, want 2", report.SubagentResults[0].Attempts)
	}
	records := report.SubagentResults[0].AttemptRecords
	if len(records) != 2 {
		t.Fatalf("scanner attempt records length = %d, want 2", len(records))
	}
	if records[0].Attempt != 1 || records[0].Status != "fail" {
		t.Fatalf("first attempt record = %#v, want failed attempt 1", records[0])
	}
	if records[1].Attempt != 2 || records[1].Status != "pass" {
		t.Fatalf("second attempt record = %#v, want passing attempt 2", records[1])
	}
}
