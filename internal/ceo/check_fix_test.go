package ceo

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"ceoharness/internal/checkrunner"
	"ceoharness/internal/jobpacket"
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

type requiredFileCheckFixRunner struct {
	packets []subagent.TaskPacket
}

func (r *requiredFileCheckFixRunner) Run(ctx context.Context, packet subagent.TaskPacket) (subagent.Result, error) {
	r.packets = append(r.packets, packet)
	summary := "ok"
	if strings.Contains(packet.Task, "Verification failed") {
		summary = `{"patches":[{"path":"frontend/state.js","old":"TODO: update this benchmark fixture","new":"optimistic update keeps rollback evidence"}]}`
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

type badInitialPatchThenFixRunner struct{}

func (r badInitialPatchThenFixRunner) Run(ctx context.Context, packet subagent.TaskPacket) (subagent.Result, error) {
	summary := `{"patches":[{"path":"app.txt","old":"missing","new":"good"}]}`
	if strings.Contains(packet.Task, "Verification failed") {
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

func Test_Runtime_RunJob_repairs_after_initial_model_patch_text_not_found(t *testing.T) {
	// Given
	runtime := NewRuntimeWithSubagentRunner(badInitialPatchThenFixRunner{})
	root := t.TempDir()
	target := filepath.Join(root, "app.txt")
	if err := os.WriteFile(target, []byte("bad"), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	// When
	report, err := runtime.RunJob(context.Background(), JobRequest{
		Task:              "Repair app.\nRequired changed files: app.txt.",
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
	if report.Verdict != "pass" {
		t.Fatalf("Verdict = %q, want repaired pass", report.Verdict)
	}
	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read target: %v", err)
	}
	if string(got) != "good" {
		t.Fatalf("content = %q, want good", string(got))
	}
}

func Test_Runtime_RunJob_check_fix_includes_required_file_contents(t *testing.T) {
	// Given
	runner := &requiredFileCheckFixRunner{}
	runtime := NewRuntimeWithSubagentRunner(runner)
	root := t.TempDir()
	target := filepath.Join(root, "frontend", "state.js")
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		t.Fatalf("mkdir fixture: %v", err)
	}
	if err := os.WriteFile(target, []byte(`module.exports = { benchmarkFixture: "TODO: update this benchmark fixture" };`), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	// When
	report, err := runtime.RunJob(context.Background(), JobRequest{
		Task: "Repair JS fixture.\n" +
			"Required changed files: frontend/state.js.\n",
		WorkspaceDir:      root,
		ApplyModelPatches: true,
		CheckFixAttempts:  1,
		CheckCommand: []string{
			os.Args[0],
			"-test.run=Test_HelperProcess_check_fix_required_file",
		},
		CheckEnv: []string{"GO_WANT_CEO_CHECK_FIX_REQUIRED=1", "GO_CEO_FIX_TARGET=" + target},
	})
	// Then
	if err != nil {
		t.Fatalf("RunJob returned error: %v", err)
	}
	if report.Verdict != "pass" {
		t.Fatalf("Verdict = %q, want pass", report.Verdict)
	}
	if len(runner.packets) < 2 {
		t.Fatalf("packets length = %d, want initial and check-fix packets", len(runner.packets))
	}
	fixPacket := runner.packets[len(runner.packets)-1]
	if strings.Join(fixPacket.AllowedActions, ",") != "propose_patch" {
		t.Fatalf("AllowedActions = %+v, want patch-only check-fix", fixPacket.AllowedActions)
	}
	if len(fixPacket.ToolResults) != 1 || !strings.Contains(fixPacket.ToolResults[0].Output, "TODO: update") {
		t.Fatalf("ToolResults = %+v, want required file content", fixPacket.ToolResults)
	}
}

func Test_checkFixContextPaths_includes_required_and_command_files(t *testing.T) {
	// Given
	req := checkFixRequest{
		Packet: jobpacket.Packet{
			Task: "Required changed files: frontend/state.js.",
		},
		Checks: []checkrunner.Result{
			{Argv: []string{"sh", "-c", "node frontend/state.test.js"}},
		},
	}

	// When
	paths := checkFixContextPaths(req)

	// Then
	if !containsString(paths, "frontend/state.js") || !containsString(paths, "frontend/state.test.js") {
		t.Fatalf("paths = %+v, want state file and test file", paths)
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

func Test_HelperProcess_check_fix_required_file(t *testing.T) {
	if os.Getenv("GO_WANT_CEO_CHECK_FIX_REQUIRED") != "1" {
		return
	}
	content, err := os.ReadFile(os.Getenv("GO_CEO_FIX_TARGET"))
	if err != nil {
		t.Fatalf("read target: %v", err)
	}
	text := string(content)
	if strings.Contains(text, "optimistic update") && strings.Contains(text, "rollback") {
		os.Exit(0)
	}
	os.Stderr.WriteString("required terms missing\n")
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
