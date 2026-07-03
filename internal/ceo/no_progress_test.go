package ceo

import (
	"context"
	"sync"
	"testing"

	"ceoharness/internal/subagent"
)

type weakProgressRunner struct {
	mu        sync.Mutex
	summaries map[string][]string
	calls     map[string]int
}

func (r *weakProgressRunner) Run(ctx context.Context, packet subagent.TaskPacket) (subagent.Result, error) {
	if err := ctx.Err(); err != nil {
		return subagent.Result{}, err
	}
	r.mu.Lock()
	r.calls[packet.AgentName]++
	call := r.calls[packet.AgentName]
	summaries := append([]string(nil), r.summaries[packet.AgentName]...)
	r.mu.Unlock()
	if len(summaries) == 0 {
		return subagent.Result{
			AgentName:       packet.AgentName,
			Status:          "pass",
			ContextReceived: packet.ContextMode,
			Summary:         "ok",
		}, nil
	}
	summary := summaries[min(call, len(summaries))-1]
	status := "fail"
	if summary == "recovered" {
		status = "pass"
	}
	return subagent.Result{
		AgentName:       packet.AgentName,
		Status:          status,
		ContextReceived: packet.ContextMode,
		Summary:         summary,
	}, nil
}

func (r *weakProgressRunner) callCount(agentName string) int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.calls[agentName]
}

func Test_Runtime_RunJob_stops_repeated_weak_subagent_after_no_progress_limit(t *testing.T) {
	// Given
	runner := &weakProgressRunner{
		summaries: map[string][]string{"scanner": {"same weak result", "same weak result", "same weak result"}},
		calls:     map[string]int{},
	}
	runtime := NewRuntimeWithSubagentRunner(runner)

	// When
	report, err := runtime.RunJob(context.Background(), JobRequest{
		Task:             "Fix a failing test",
		SubagentAttempts: 4,
		NoProgressStop:   2,
	})

	// Then
	if err != nil {
		t.Fatalf("RunJob returned error: %v", err)
	}
	scanner := report.SubagentResults[0]
	if scanner.Attempts != 2 || !scanner.NoProgressStopped {
		t.Fatalf("scanner result = %#v, want stopped after two repeated weak attempts", scanner)
	}
	if runner.callCount("scanner") != 2 {
		t.Fatalf("scanner calls = %d, want 2", runner.callCount("scanner"))
	}
	if report.VerificationSummary.SubagentNoProgressStopCount != 1 {
		t.Fatalf("SubagentNoProgressStopCount = %d, want 1", report.VerificationSummary.SubagentNoProgressStopCount)
	}
	if report.RunManifest.NoProgressStop != 2 {
		t.Fatalf("RunManifest.NoProgressStop = %d, want 2", report.RunManifest.NoProgressStop)
	}
}

func Test_Runtime_RunJob_continues_when_weak_subagent_makes_progress(t *testing.T) {
	// Given
	runner := &weakProgressRunner{
		summaries: map[string][]string{"scanner": {"first weak result", "new weak result", "recovered"}},
		calls:     map[string]int{},
	}
	runtime := NewRuntimeWithSubagentRunner(runner)

	// When
	report, err := runtime.RunJob(context.Background(), JobRequest{
		Task:             "Fix a failing test",
		SubagentAttempts: 3,
		NoProgressStop:   2,
	})

	// Then
	if err != nil {
		t.Fatalf("RunJob returned error: %v", err)
	}
	if report.Verdict != "pass" {
		t.Fatalf("Verdict = %q, want pass", report.Verdict)
	}
	if runner.callCount("scanner") != 3 {
		t.Fatalf("scanner calls = %d, want 3", runner.callCount("scanner"))
	}
	if report.SubagentResults[0].NoProgressStopped {
		t.Fatalf("scanner result = %#v, want no no-progress stop", report.SubagentResults[0])
	}
}
