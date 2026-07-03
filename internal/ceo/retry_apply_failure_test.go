package ceo

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func Test_BadOutput_repeated_apply_failure_stops_as_no_progress(t *testing.T) {
	// Given
	root := t.TempDir()
	target := filepath.Join(root, "app.txt")
	if err := os.WriteFile(target, []byte("bad"), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	runner := &retryHistoryRunner{
		fixSummaries: []string{
			`{"patches":[{"path":"app.txt","old":"missing","new":"good"}]}`,
			`{"patches":[{"path":"app.txt","old":"missing","new":"good"}]}`,
			`{"patches":[{"path":"app.txt","old":"missing","new":"good"}]}`,
		},
	}
	runtime := NewRuntimeWithSubagentRunner(runner)

	// When
	report, err := runtime.RunJob(context.Background(), JobRequest{
		Task:              "Repair app",
		WorkspaceDir:      root,
		ApplyModelPatches: true,
		CheckFixAttempts:  5,
		NoProgressStop:    2,
		CheckCommand: []string{
			os.Args[0],
			"-test.run=Test_HelperProcess_check_fix_file",
		},
		CheckEnv: []string{"GO_WANT_CEO_CHECK_FIX=1", "GO_CEO_FIX_TARGET=" + target},
	})

	// Then
	if err != nil {
		t.Fatalf("RunJob returned error: %v", err)
	}
	if runner.fixCalls != 2 {
		t.Fatalf("fixCalls = %d, want no-progress stop after 2 repeated apply failures", runner.fixCalls)
	}
	if report.Verdict != "fail" {
		t.Fatalf("Verdict = %q, want fail", report.Verdict)
	}
	last := report.RetryHistory[len(report.RetryHistory)-1]
	if !last.NoProgressStopped || last.ModelPatchStatus != "apply_failed" {
		t.Fatalf("last retry = %#v, want no-progress apply-failure stop", last)
	}
}
