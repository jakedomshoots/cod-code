package ceo

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func Test_Runtime_RunJob_reports_compact_run_events(t *testing.T) {
	// Given
	runtime := NewRuntimeWithSubagentRunner(toolRequestRunner{})
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "app.txt"), []byte("hello needle"), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	// When
	report, err := runtime.RunJob(context.Background(), JobRequest{
		Task:         "Fix a failing test",
		WorkspaceDir: root,
		CheckCommand: []string{
			os.Args[0],
			"-test.run=Test_HelperProcess_tool_request_check",
		},
		CheckEnv: []string{"GO_WANT_TOOL_REQUEST_CHECK=1"},
	})
	// Then
	if err != nil {
		t.Fatalf("RunJob returned error: %v", err)
	}
	wantKinds := []string{
		"job_packet",
		"subagent",
		"tool_request",
		"tool_result",
		"tool_feedback",
		"check",
		"verdict",
	}
	assertEventKinds(t, report.RunEvents, wantKinds)
	for index, event := range report.RunEvents {
		if event.Index != index+1 {
			t.Fatalf("RunEvents[%d].Index = %d, want %d", index, event.Index, index+1)
		}
	}
	if report.RunEvents[len(report.RunEvents)-1].Status != "pass" {
		t.Fatalf("final event status = %q, want pass", report.RunEvents[len(report.RunEvents)-1].Status)
	}
}

func assertEventKinds(t *testing.T, events []RunEvent, want []string) {
	t.Helper()
	got := make([]string, 0, len(events))
	for _, event := range events {
		got = append(got, event.Kind)
	}
	next := 0
	for _, kind := range got {
		if next < len(want) && kind == want[next] {
			next++
		}
	}
	if next != len(want) {
		t.Fatalf("event kinds = %v, want ordered subsequence %v", got, want)
	}
}
