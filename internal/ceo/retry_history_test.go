package ceo

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"ceoharness/internal/subagent"
)

type retryHistoryRunner struct {
	fixSummaries      []string
	revisionSummaries []string
	fixCalls          int
	revisionCalls     int
}

func (r *retryHistoryRunner) Run(ctx context.Context, packet subagent.TaskPacket) (subagent.Result, error) {
	if err := ctx.Err(); err != nil {
		return subagent.Result{}, err
	}
	summary := "initial pass"
	if packet.AgentName == "coder" && strings.Contains(packet.Task, "Verification failed") {
		summary = scriptedSummary(r.fixSummaries, r.fixCalls)
		r.fixCalls++
	}
	if packet.AgentName == "coder" && strings.Contains(packet.Task, "CEO review failed") {
		summary = scriptedSummary(r.revisionSummaries, r.revisionCalls)
		r.revisionCalls++
	}
	return subagent.Result{
		AgentName:       packet.AgentName,
		Role:            packet.Role,
		Status:          "pass",
		Attempts:        1,
		ContextReceived: packet.ContextMode,
		Summary:         summary,
		Evidence:        []string{"ran"},
	}, nil
}

func Test_RetryReport_records_check_fix_history_when_repair_runs(t *testing.T) {
	// Given
	root := t.TempDir()
	target := filepath.Join(root, "app.txt")
	if err := os.WriteFile(target, []byte("bad"), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	runner := &retryHistoryRunner{
		fixSummaries: []string{`{"patches":[{"path":"app.txt","old":"bad","new":"good"}]}`},
	}
	runtime := NewRuntimeWithSubagentRunner(runner)

	// When
	report, err := runtime.RunJob(context.Background(), JobRequest{
		Task:              "Repair app",
		WorkspaceDir:      root,
		ApplyModelPatches: true,
		CheckFixAttempts:  1,
		ScorerFailedChecks: []RepairFailureDetail{{
			Name:    "diff_term:retry_history",
			Status:  "fail",
			Message: "missing required diff term",
		}},
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
		t.Fatalf("Verdict = %q, want pass", report.Verdict)
	}
	if len(report.RetryHistory) != 1 {
		t.Fatalf("RetryHistory length = %d, want 1", len(report.RetryHistory))
	}
	entry := report.RetryHistory[0]
	if entry.Kind != "check_fix" || entry.Attempt != 1 || entry.ModelPatchStatus != "applied" || entry.FinalVerdict != "pass" {
		t.Fatalf("RetryHistory[0] = %#v, want applied passing check-fix retry", entry)
	}
	if len(entry.FailedChecks) != 2 {
		t.Fatalf("FailedChecks = %#v, want command failure plus scorer failure", entry.FailedChecks)
	}
	if !strings.Contains(entry.CorrectivePrompt, "diff_term:retry_history") {
		t.Fatalf("CorrectivePrompt missing scorer feedback: %s", entry.CorrectivePrompt)
	}
	requireRetryHistoryJSON(t, report)
}

func Test_Revision_records_retry_history_when_ceo_veto_is_repaired(t *testing.T) {
	// Given
	root := t.TempDir()
	target := filepath.Join(root, "app.txt")
	if err := os.WriteFile(target, []byte("bad"), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	runner := &retryHistoryRunner{
		revisionSummaries: []string{`{"patches":[{"path":"app.txt","old":"bad","new":"good"}]}`},
	}
	client := &sequenceCEOModelClient{
		responses: []string{
			`{"selected_subagents":["coder","reviewer"],"summary":"Use coder and reviewer."}`,
			`{"recommended_verdict":"fail","summary":"Patch app.txt before accepting."}`,
			`{"recommended_verdict":"pass","summary":"Revision accepted."}`,
		},
	}
	runtime := NewRuntimeWithSubagentRunnerAndCEOReviewer(runner, client)

	// When
	report, err := runtime.RunJob(context.Background(), JobRequest{
		Task:                "Repair app",
		WorkspaceDir:        root,
		ApplyModelPatches:   true,
		CEORevisionAttempts: 1,
		CheckCommand: []string{
			os.Args[0],
			"-test.run=Test_HelperProcess_ceo_revision_check",
		},
		CheckEnv: []string{"GO_WANT_CEO_REVISION_CHECK=1"},
	})

	// Then
	if err != nil {
		t.Fatalf("RunJob returned error: %v", err)
	}
	if report.Verdict != "pass" {
		t.Fatalf("Verdict = %q, want pass", report.Verdict)
	}
	if len(report.RetryHistory) != 1 {
		t.Fatalf("RetryHistory length = %d, want 1", len(report.RetryHistory))
	}
	entry := report.RetryHistory[0]
	if entry.Kind != "ceo_revision" || entry.ModelPatchStatus != "applied" || entry.FinalVerdict != "pass" {
		t.Fatalf("RetryHistory[0] = %#v, want applied passing CEO revision retry", entry)
	}
	if len(entry.FailedChecks) != 1 || entry.FailedChecks[0].Name != "ceo_review" {
		t.Fatalf("FailedChecks = %#v, want CEO review failure detail", entry.FailedChecks)
	}
	requireRetryHistoryJSON(t, report)
}

func Test_BadOutput_invalid_model_patch_json_does_not_crash_or_claim_pass(t *testing.T) {
	// Given
	root := t.TempDir()
	target := filepath.Join(root, "app.txt")
	if err := os.WriteFile(target, []byte("bad"), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	runner := &retryHistoryRunner{fixSummaries: []string{`{"patches":[`}}
	runtime := NewRuntimeWithSubagentRunner(runner)

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
	if report.Verdict != "fail" {
		t.Fatalf("Verdict = %q, want fail", report.Verdict)
	}
	if len(report.RetryHistory) != 1 || report.RetryHistory[0].ModelPatchStatus != "invalid" {
		t.Fatalf("RetryHistory = %#v, want invalid patch record", report.RetryHistory)
	}
}

func Test_NoProgress_repeated_noop_check_fix_stops_with_reason(t *testing.T) {
	// Given
	root := t.TempDir()
	target := filepath.Join(root, "app.txt")
	if err := os.WriteFile(target, []byte("bad"), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	runner := &retryHistoryRunner{
		fixSummaries: []string{`{"patches":[]}`, `{"patches":[]}`, `{"patches":[]}`},
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
		t.Fatalf("fixCalls = %d, want no-progress stop after 2 repeated no-op attempts", runner.fixCalls)
	}
	if report.Verdict != "fail" {
		t.Fatalf("Verdict = %q, want fail", report.Verdict)
	}
	last := report.RetryHistory[len(report.RetryHistory)-1]
	if !last.NoProgressStopped || last.ModelPatchStatus != "empty" {
		t.Fatalf("last retry = %#v, want no-progress empty-patch stop", last)
	}
}

func scriptedSummary(values []string, index int) string {
	if len(values) == 0 {
		return "ok"
	}
	if index >= len(values) {
		return values[len(values)-1]
	}
	return values[index]
}

func requireRetryHistoryJSON(t *testing.T, report Report) {
	t.Helper()
	payload, err := json.Marshal(report)
	if err != nil {
		t.Fatalf("marshal report: %v", err)
	}
	if !strings.Contains(string(payload), `"retry_history"`) {
		t.Fatalf("report JSON missing retry_history: %s", payload)
	}
}
