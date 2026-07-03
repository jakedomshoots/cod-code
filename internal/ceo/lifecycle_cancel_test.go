package ceo

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"ceoharness/internal/history"
	"ceoharness/internal/subagent"
)

type cancelingRunner struct {
	cancel context.CancelFunc
}

func (r cancelingRunner) Run(ctx context.Context, packet subagent.TaskPacket) (subagent.Result, error) {
	if r.cancel != nil {
		r.cancel()
	}
	<-ctx.Done()
	return subagent.Result{}, ctx.Err()
}

func Test_Lifecycle_RunJob_persists_canceled_recovered_report_history_and_events_when_context_cancels(t *testing.T) {
	// Given
	ctx, cancel := context.WithCancel(context.Background())
	runtime := NewRuntimeWithSubagentRunner(cancelingRunner{cancel: cancel})
	root := t.TempDir()

	// When
	report, err := runtime.RunJob(ctx, JobRequest{
		Task:         "Resume canceled lifecycle proof",
		WorkspaceDir: root,
		Resume:       &ResumeContext{JobID: "job-000099"},
	})

	// Then
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("RunJob error = %v, want context canceled", err)
	}
	if report.LifecycleState != LifecycleCanceled {
		t.Fatalf("LifecycleState = %q, want %q", report.LifecycleState, LifecycleCanceled)
	}
	assertLifecycleStates(t, report.LifecycleEvents, []LifecycleState{
		LifecycleCreated,
		LifecycleRecovered,
		LifecyclePlanning,
		LifecycleDelegated,
		LifecycleCanceled,
	})
	assertRunEventLifecycle(t, report.RunEvents, "verdict", LifecycleCanceled)
	if report.Verdict != "canceled" {
		t.Fatalf("Verdict = %q, want canceled", report.Verdict)
	}
	store, storeErr := history.New(root)
	if storeErr != nil {
		t.Fatalf("open history: %v", storeErr)
	}
	entries, readErr := store.ReadAll(context.Background())
	if readErr != nil {
		t.Fatalf("read history: %v", readErr)
	}
	if len(entries) != 1 {
		t.Fatalf("history entries length = %d, want 1", len(entries))
	}
	if entries[0].LifecycleState != "canceled" {
		t.Fatalf("history lifecycle_state = %q, want canceled", entries[0].LifecycleState)
	}
	if !historyLifecycleStatesInclude(entries[0].LifecycleEvents, "recovered", "canceled") {
		t.Fatalf("history lifecycle_events = %+v, want recovered then canceled", entries[0].LifecycleEvents)
	}
	snapshot, readSnapshotErr := os.ReadFile(filepath.Join(root, "ceo-artifacts", "jobs", "job-000001.json"))
	if readSnapshotErr != nil {
		t.Fatalf("read report snapshot: %v", readSnapshotErr)
	}
	if !strings.Contains(string(snapshot), `"lifecycle_state": "canceled"`) {
		t.Fatalf("snapshot = %q, want canceled lifecycle state", string(snapshot))
	}
	if !strings.Contains(string(snapshot), `"state": "recovered"`) {
		t.Fatalf("snapshot = %q, want recovered lifecycle event", string(snapshot))
	}
}

func Test_Lifecycle_RunJob_persists_canceled_recoverable_report_when_check_context_expires(t *testing.T) {
	// Given
	ctx, cancel := context.WithTimeout(context.Background(), 20_000_000)
	defer cancel()
	runtime := NewRuntime()
	root := t.TempDir()

	// When
	report, err := runtime.RunJob(ctx, JobRequest{
		Task:         "Cancel long-running helper command",
		WorkspaceDir: root,
		CheckCommand: []string{
			"sh",
			"-c",
			"sleep 10",
		},
	})

	// Then
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("RunJob error = %v, want context deadline exceeded", err)
	}
	if report.Verdict != "canceled" {
		t.Fatalf("Verdict = %q, want canceled", report.Verdict)
	}
	if report.LifecycleState != LifecycleCanceled {
		t.Fatalf("LifecycleState = %q, want %q", report.LifecycleState, LifecycleCanceled)
	}
	assertLifecycleStates(t, report.LifecycleEvents, []LifecycleState{
		LifecycleCreated,
		LifecyclePlanning,
		LifecycleDelegated,
		LifecycleCanceled,
	})
	store, storeErr := history.New(root)
	if storeErr != nil {
		t.Fatalf("open history: %v", storeErr)
	}
	entries, readErr := store.ReadAll(context.Background())
	if readErr != nil {
		t.Fatalf("read history: %v", readErr)
	}
	if len(entries) != 1 {
		t.Fatalf("history entries length = %d, want 1", len(entries))
	}
	if entries[0].Verdict != "canceled" || entries[0].LifecycleState != "canceled" {
		t.Fatalf("history entry = %+v, want canceled recoverable job", entries[0])
	}
}

func historyLifecycleStatesInclude(events []history.LifecycleEvent, want ...string) bool {
	next := 0
	for _, event := range events {
		if next < len(want) && event.State == want[next] {
			next++
		}
	}
	return next == len(want)
}
