package ceo

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func Test_Runtime_RunJob_runs_check_command_from_workspace(t *testing.T) {
	// Given
	runtime := NewRuntime()
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "check-marker.txt"), []byte("ok\n"), 0o644); err != nil {
		t.Fatalf("write marker: %v", err)
	}

	// When
	report, err := runtime.RunJob(context.Background(), JobRequest{
		Task:         "Run workspace check",
		WorkspaceDir: root,
		CheckCommand: []string{
			os.Args[0],
			"-test.run=Test_HelperProcess_workspace_check",
		},
		CheckEnv: []string{"GO_WANT_CEO_HELPER_PROCESS=workspace_check"},
	})
	// Then
	if err != nil {
		t.Fatalf("RunJob returned error: %v", err)
	}
	if report.Verdict != "pass" {
		t.Fatalf("Verdict = %q, want pass", report.Verdict)
	}
	if len(report.CheckResults) != 1 || report.CheckResults[0].Status != "pass" {
		t.Fatalf("CheckResults = %#v, want one passing check", report.CheckResults)
	}
}

func Test_HelperProcess_workspace_check(t *testing.T) {
	if os.Getenv("GO_WANT_CEO_HELPER_PROCESS") != "workspace_check" {
		return
	}
	if _, err := os.Stat("check-marker.txt"); err != nil {
		os.Stderr.WriteString(err.Error() + "\n")
		os.Exit(8)
	}
	os.Exit(0)
}
