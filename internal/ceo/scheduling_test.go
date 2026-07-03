package ceo

import (
	"context"
	"testing"
	"time"

	"ceoharness/internal/jobpacket"
	"ceoharness/internal/subagent"
)

type blockingRunner struct {
	started chan string
	release map[string]chan struct{}
}

func (r *blockingRunner) Run(ctx context.Context, packet subagent.TaskPacket) (subagent.Result, error) {
	if err := ctx.Err(); err != nil {
		return subagent.Result{}, err
	}
	r.started <- packet.AgentName
	select {
	case <-ctx.Done():
		return subagent.Result{}, ctx.Err()
	case <-r.release[packet.AgentName]:
		return subagent.Result{
			AgentName:       packet.AgentName,
			Status:          "pass",
			ContextReceived: packet.ContextMode,
			Summary:         "parallel result",
		}, nil
	}
}

func Test_Runtime_RunJob_runs_subagents_by_dependency_stage(t *testing.T) {
	// Given
	runner := &blockingRunner{
		started: make(chan string, 3),
		release: map[string]chan struct{}{
			"scanner":  make(chan struct{}),
			"coder":    make(chan struct{}),
			"reviewer": make(chan struct{}),
		},
	}
	runtime := NewRuntimeWithSubagentRunner(runner)
	type runResult struct {
		report Report
		err    error
	}
	done := make(chan runResult, 1)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// When
	go func() {
		report, err := runtime.RunJob(ctx, JobRequest{Task: "Fix a failing test"})
		done <- runResult{report: report, err: err}
	}()

	// Then
	requireNextStarted(t, runner.started, "scanner")
	requireNoStartBeforeRelease(t, runner.started)
	close(runner.release["scanner"])
	requireNextStarted(t, runner.started, "coder")
	requireNoStartBeforeRelease(t, runner.started)
	close(runner.release["coder"])
	requireNextStarted(t, runner.started, "reviewer")
	close(runner.release["reviewer"])
	select {
	case result := <-done:
		if result.err != nil {
			t.Fatalf("RunJob returned error: %v", result.err)
		}
		wantAgents := []string{"scanner", "coder", "reviewer"}
		for index, wantAgent := range wantAgents {
			if result.report.SubagentResults[index].AgentName != wantAgent {
				t.Fatalf("SubagentResults[%d].AgentName = %q, want %q", index, result.report.SubagentResults[index].AgentName, wantAgent)
			}
			if result.report.SubagentResults[index].Stage != index+1 {
				t.Fatalf("SubagentResults[%d].Stage = %d, want %d", index, result.report.SubagentResults[index].Stage, index+1)
			}
		}
	case <-time.After(time.Second):
		t.Fatal("RunJob did not return after releasing subagents")
	}
}

func Test_Runtime_RunJob_limits_parallel_subagents_when_concurrency_is_set(t *testing.T) {
	// Given
	runner := &blockingRunner{
		started: make(chan string, 2),
		release: map[string]chan struct{}{
			"scanner":  make(chan struct{}),
			"planner":  make(chan struct{}),
			"reviewer": make(chan struct{}),
		},
	}
	runtime := NewRuntimeWithSubagentRunner(runner)
	type runResult struct {
		report Report
		err    error
	}
	done := make(chan runResult, 1)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// When
	go func() {
		report, err := runtime.RunJob(ctx, JobRequest{
			Task:                "Plan scanner work",
			SubagentConcurrency: 1,
			Subagents: []jobpacket.Subagent{
				{Name: "scanner", Role: "inspect scope"},
				{Name: "planner", Role: "plan work"},
				{Name: "reviewer", Role: "verify evidence"},
			},
		})
		done <- runResult{report: report, err: err}
	}()

	// Then
	requireNextStarted(t, runner.started, "scanner")
	requireNoStartBeforeRelease(t, runner.started)
	close(runner.release["scanner"])
	requireNextStarted(t, runner.started, "planner")
	requireNoStartBeforeRelease(t, runner.started)
	close(runner.release["planner"])
	requireNextStarted(t, runner.started, "reviewer")
	close(runner.release["reviewer"])
	select {
	case result := <-done:
		if result.err != nil {
			t.Fatalf("RunJob returned error: %v", result.err)
		}
		if len(result.report.SubagentResults) != 3 {
			t.Fatalf("SubagentResults length = %d, want 3", len(result.report.SubagentResults))
		}
	case <-time.After(time.Second):
		t.Fatal("RunJob did not return after releasing subagents")
	}
}

func Test_StagedSubagents_orders_custom_agents_by_dependency_stage(t *testing.T) {
	// Given
	agents := []jobpacket.Subagent{
		{Name: "reviewer", Role: "verify evidence"},
		{Name: "coder", Role: "implement fix"},
		{Name: "scanner", Role: "inspect workspace"},
	}

	// When
	stages := stagedSubagents(agents)

	// Then
	wantStages := []int{1, 2, 3}
	for index, wantStage := range wantStages {
		if stages[index].index != wantStage {
			t.Fatalf("stages[%d].index = %d, want %d", index, stages[index].index, wantStage)
		}
	}
	if stages[0].agents[0].agent.Name != "scanner" {
		t.Fatalf("stage 1 agent = %q, want scanner", stages[0].agents[0].agent.Name)
	}
}

func Test_StagedSubagents_uses_explicit_custom_stage(t *testing.T) {
	// Given
	agents := []jobpacket.Subagent{
		{Name: "ux_reviewer", Role: "review UX", Stage: 3},
		{Name: "planner", Role: "plan work"},
	}

	// When
	stages := stagedSubagents(agents)

	// Then
	if len(stages) != 2 {
		t.Fatalf("stage count = %d, want 2", len(stages))
	}
	if stages[1].index != 3 || stages[1].agents[0].agent.Name != "ux_reviewer" {
		t.Fatalf("stage 3 = %#v, want ux_reviewer", stages[1])
	}
}

func requireNextStarted(t *testing.T, started <-chan string, want string) {
	t.Helper()
	select {
	case got := <-started:
		if got != want {
			t.Fatalf("started = %q, want %q", got, want)
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatalf("timed out waiting for %s to start", want)
	}
}

func requireNoStartBeforeRelease(t *testing.T, started <-chan string) {
	t.Helper()
	select {
	case got := <-started:
		t.Fatalf("started %q before previous stage released", got)
	case <-time.After(50 * time.Millisecond):
	}
}
