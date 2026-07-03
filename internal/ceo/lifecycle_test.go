package ceo

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func Test_Lifecycle_RunJob_progresses_to_passed_when_checks_pass(t *testing.T) {
	// Given
	runtime := NewRuntime()

	// When
	report, err := runtime.RunJob(context.Background(), JobRequest{
		Task:         "Fix a failing test",
		CheckCommand: []string{"go", "version"},
	})

	// Then
	if err != nil {
		t.Fatalf("RunJob returned error: %v", err)
	}
	if report.LifecycleState != LifecyclePassed {
		t.Fatalf("LifecycleState = %q, want %q", report.LifecycleState, LifecyclePassed)
	}
	assertLifecycleStates(t, report.LifecycleEvents, []LifecycleState{
		LifecycleCreated,
		LifecyclePlanning,
		LifecycleDelegated,
		LifecycleChecking,
		LifecycleReviewing,
		LifecyclePassed,
	})
	assertRunEventLifecycle(t, report.RunEvents, "check", LifecycleChecking)
	assertRunEventLifecycle(t, report.RunEvents, "verdict", LifecyclePassed)
}

func Test_Lifecycle_RunJob_progresses_to_failed_when_check_fails(t *testing.T) {
	// Given
	runtime := NewRuntime()

	// When
	report, err := runtime.RunJob(context.Background(), JobRequest{
		Task:         "Fix a failing test",
		CheckCommand: []string{os.Args[0], "-test.run=Test_HelperProcess_fail_check"},
		CheckEnv:     []string{"GO_WANT_CEO_HELPER_PROCESS=fail"},
	})

	// Then
	if err != nil {
		t.Fatalf("RunJob returned error: %v", err)
	}
	if report.LifecycleState != LifecycleFailed {
		t.Fatalf("LifecycleState = %q, want %q", report.LifecycleState, LifecycleFailed)
	}
	assertLifecycleStates(t, report.LifecycleEvents, []LifecycleState{
		LifecycleCreated,
		LifecyclePlanning,
		LifecycleDelegated,
		LifecycleChecking,
		LifecycleFailed,
	})
	assertRunEventLifecycle(t, report.RunEvents, "check", LifecycleChecking)
	assertRunEventLifecycle(t, report.RunEvents, "verdict", LifecycleFailed)
}

func Test_Lifecycle_RunJob_progresses_to_needs_input_when_subagent_asks_question(t *testing.T) {
	// Given
	runtime := NewRuntimeWithSubagentRunner(needsInputRunner{})

	// When
	report, err := runtime.RunJob(context.Background(), JobRequest{Task: "Fix ambiguous package"})

	// Then
	if err != nil {
		t.Fatalf("RunJob returned error: %v", err)
	}
	if report.LifecycleState != LifecycleNeedsInput {
		t.Fatalf("LifecycleState = %q, want %q", report.LifecycleState, LifecycleNeedsInput)
	}
	assertLifecycleStates(t, report.LifecycleEvents, []LifecycleState{
		LifecycleCreated,
		LifecyclePlanning,
		LifecycleDelegated,
		LifecycleNeedsInput,
	})
	assertRunEventLifecycle(t, report.RunEvents, "subagent", LifecycleDelegated)
	assertRunEventLifecycle(t, report.RunEvents, "verdict", LifecycleNeedsInput)
}

func Test_Lifecycle_RunJob_marks_patch_previewed_when_dry_run_previews_patch(t *testing.T) {
	// Given
	runtime := NewRuntime()
	root := t.TempDir()
	target := filepath.Join(root, "app.txt")
	if err := os.WriteFile(target, []byte("hello old"), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	// When
	report, err := runtime.RunJob(context.Background(), JobRequest{
		Task:         "Patch app text",
		WorkspaceDir: root,
		DryRun:       true,
		Patches: []PatchRequest{
			{Path: "app.txt", Old: "old", New: "new"},
		},
	})

	// Then
	if err != nil {
		t.Fatalf("RunJob returned error: %v", err)
	}
	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read target: %v", err)
	}
	if string(got) != "hello old" {
		t.Fatalf("content = %q, want unchanged hello old", string(got))
	}
	if report.LifecycleState != LifecyclePassed {
		t.Fatalf("LifecycleState = %q, want %q", report.LifecycleState, LifecyclePassed)
	}
	if !hasLifecycleState(report.LifecycleEvents, LifecyclePatchPreviewed) {
		t.Fatalf("LifecycleEvents = %+v, want patch_previewed", report.LifecycleEvents)
	}
	if hasLifecycleState(report.LifecycleEvents, LifecyclePatchApplied) {
		t.Fatalf("LifecycleEvents = %+v, did not want patch_applied", report.LifecycleEvents)
	}
	assertRunEventLifecycle(t, report.RunEvents, "patch_preview", LifecyclePatchPreviewed)
}

func Test_Lifecycle_RunJob_marks_patch_applied_when_approved_preview_is_written(t *testing.T) {
	// Given
	runtime := NewRuntime()
	root := t.TempDir()
	target := filepath.Join(root, "app.txt")
	if err := os.WriteFile(target, []byte("hello old"), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	preview, err := runtime.RunJob(context.Background(), JobRequest{
		Task:         "Patch app text",
		WorkspaceDir: root,
		DryRun:       true,
		Patches: []PatchRequest{
			{Path: "app.txt", Old: "old", New: "new"},
		},
	})
	if err != nil {
		t.Fatalf("preview RunJob returned error: %v", err)
	}

	// When
	report, err := runtime.RunJob(context.Background(), JobRequest{
		Task:                  "Patch app text",
		WorkspaceDir:          root,
		ApprovedPreviewDigest: preview.PatchApproval.PreviewDigest,
		Patches: []PatchRequest{
			{Path: "app.txt", Old: "old", New: "new"},
		},
	})

	// Then
	if err != nil {
		t.Fatalf("RunJob returned error: %v", err)
	}
	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read target: %v", err)
	}
	if string(got) != "hello new" {
		t.Fatalf("content = %q, want hello new", string(got))
	}
	assertLifecycleStates(t, report.LifecycleEvents, []LifecycleState{
		LifecyclePatchPreviewed,
		LifecyclePatchApplied,
		LifecyclePassed,
	})
	assertRunEventLifecycle(t, report.RunEvents, "patch_preview", LifecyclePatchPreviewed)
	assertRunEventLifecycle(t, report.RunEvents, "patch", LifecyclePatchApplied)
}

func Test_LifecycleMachine_cancels_when_context_is_canceled(t *testing.T) {
	// Given
	machine := NewLifecycleMachine()
	if err := machine.Transition(LifecycleCreated, "job created"); err != nil {
		t.Fatalf("created transition returned error: %v", err)
	}
	if err := machine.Transition(LifecyclePlanning, "job planned"); err != nil {
		t.Fatalf("planning transition returned error: %v", err)
	}

	// When
	err := machine.Cancel(context.Canceled)

	// Then
	if err != nil {
		t.Fatalf("Cancel returned error: %v", err)
	}
	if err := machine.Cancel(context.Canceled); err != nil {
		t.Fatalf("second Cancel returned error: %v", err)
	}
	if machine.State() != LifecycleCanceled {
		t.Fatalf("State = %q, want %q", machine.State(), LifecycleCanceled)
	}
	if !hasLifecycleState(machine.Events(), LifecycleCanceled) {
		t.Fatalf("Events = %+v, want canceled", machine.Events())
	}
}

func assertLifecycleStates(t *testing.T, events []LifecycleEvent, want []LifecycleState) {
	t.Helper()
	next := 0
	for _, event := range events {
		if next < len(want) && event.State == want[next] {
			next++
		}
	}
	if next != len(want) {
		t.Fatalf("LifecycleEvents = %+v, want ordered states %v", events, want)
	}
}

func assertRunEventLifecycle(t *testing.T, events []RunEvent, kind string, want LifecycleState) {
	t.Helper()
	for _, event := range events {
		if event.Kind == kind {
			if event.LifecycleState != want {
				t.Fatalf("%s lifecycle = %q, want %q in event %+v", kind, event.LifecycleState, want, event)
			}
			return
		}
	}
	t.Fatalf("RunEvents = %+v, want kind %s", events, kind)
}

func hasLifecycleState(events []LifecycleEvent, want LifecycleState) bool {
	for _, event := range events {
		if event.State == want {
			return true
		}
	}
	return false
}
