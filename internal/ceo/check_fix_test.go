package ceo

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"ceoharness/internal/subagent"
)

type checkFixRunner struct{}

func (r checkFixRunner) Run(ctx context.Context, packet subagent.TaskPacket) (subagent.Result, error) {
	summary := "ok"
	if packet.AgentName == "coder" && strings.Contains(packet.Task, "Verification failed") {
		summary = `{"patches":[{"path":"app.txt","old":"bad","new":"good"}]}`
	}
	return subagent.Result{
		AgentName:       packet.AgentName,
		Role:            packet.Role,
		Status:          "pass",
		Attempts:        1,
		ContextReceived: packet.ContextMode,
		ContextBytes:    len(packet.Task),
		Summary:         summary,
		Evidence:        []string{"ok"},
	}, nil
}

func Test_Runtime_RunJob_runs_bounded_check_fix_after_failed_check(t *testing.T) {
	// Given
	runtime := NewRuntimeWithSubagentRunner(checkFixRunner{})
	root := t.TempDir()
	target := filepath.Join(root, "app.txt")
	if err := os.WriteFile(target, []byte("bad"), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	// When
	report, err := runtime.RunJob(context.Background(), JobRequest{
		Task:              "Repair app",
		WorkspaceDir:      root,
		ApplyModelPatches: true,
		CheckFixAttempts:  1,
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
	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read fixed file: %v", err)
	}
	if string(got) != "good" {
		t.Fatalf("content = %q, want good", string(got))
	}
	if report.Verdict != "pass" {
		t.Fatalf("Verdict = %q, want pass", report.Verdict)
	}
	if len(report.CheckResults) != 2 {
		t.Fatalf("CheckResults length = %d, want 2", len(report.CheckResults))
	}
	if report.CheckResults[0].Status != "fail" || report.CheckResults[1].Status != "pass" {
		t.Fatalf("check statuses = %q, %q; want fail, pass", report.CheckResults[0].Status, report.CheckResults[1].Status)
	}
	if len(report.PatchAudit) != 1 || report.PatchAudit[0].Source != "model" {
		t.Fatalf("PatchAudit = %+v, want one model patch", report.PatchAudit)
	}
	if !containsString(report.ChangedFiles, "ceo-artifacts/coder-fix-1.md") {
		t.Fatalf("ChangedFiles = %+v, want coder fix evidence", report.ChangedFiles)
	}
}

func Test_Runtime_RunJob_skips_check_fix_when_max_ceo_iterations_is_exhausted(t *testing.T) {
	// Given
	runtime := NewRuntimeWithSubagentRunner(checkFixRunner{})
	root := t.TempDir()
	target := filepath.Join(root, "app.txt")
	if err := os.WriteFile(target, []byte("bad"), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	// When
	report, err := runtime.RunJob(context.Background(), JobRequest{
		Task:              "Repair app",
		WorkspaceDir:      root,
		ApplyModelPatches: true,
		CheckFixAttempts:  1,
		MaxCEOIterations:  1,
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
	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read target: %v", err)
	}
	if string(got) != "bad" {
		t.Fatalf("content = %q, want bad because check-fix was skipped", string(got))
	}
	if report.Verdict != "fail" {
		t.Fatalf("Verdict = %q, want fail", report.Verdict)
	}
	if len(report.CheckResults) != 1 {
		t.Fatalf("CheckResults length = %d, want only the initial check", len(report.CheckResults))
	}
	if report.RunManifest.MaxCEOIterations != 1 ||
		report.RunManifest.CEOIterationCount != 1 ||
		!report.RunManifest.CEOIterationExhausted {
		t.Fatalf("RunManifest = %#v, want exhausted one-iteration budget", report.RunManifest)
	}
}

func Test_HelperProcess_check_fix_file(t *testing.T) {
	if os.Getenv("GO_WANT_CEO_CHECK_FIX") != "1" {
		return
	}
	content, err := os.ReadFile(os.Getenv("GO_CEO_FIX_TARGET"))
	if err != nil {
		t.Fatalf("read target: %v", err)
	}
	if strings.TrimSpace(string(content)) == "good" {
		os.Stdout.WriteString("file fixed\n")
		os.Exit(0)
	}
	os.Stderr.WriteString("file still bad\n")
	os.Exit(4)
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
