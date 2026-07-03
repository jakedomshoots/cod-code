package ceo

import (
	"context"
	"errors"
	"strings"
	"testing"

	"ceoharness/internal/subagent"
)

type scannerErrorRunner struct{}

func (r scannerErrorRunner) Run(ctx context.Context, packet subagent.TaskPacket) (subagent.Result, error) {
	if err := ctx.Err(); err != nil {
		return subagent.Result{}, err
	}
	if packet.AgentName == "scanner" {
		return subagent.Result{}, errors.New("scanner model unavailable")
	}
	return subagent.Result{
		AgentName:       packet.AgentName,
		Role:            packet.Role,
		Status:          "pass",
		Attempts:        1,
		ContextReceived: packet.ContextMode,
		Summary:         "ok",
	}, nil
}

func Test_Runtime_RunJob_returns_failed_report_when_subagent_retries_exhaust(t *testing.T) {
	// Given
	runtime := NewRuntimeWithSubagentRunner(scannerErrorRunner{})

	// When
	report, err := runtime.RunJob(context.Background(), JobRequest{
		Task:             "Fix a failing test",
		SubagentAttempts: 2,
	})
	// Then
	if err != nil {
		t.Fatalf("RunJob returned error: %v", err)
	}
	if report.Verdict != "fail" {
		t.Fatalf("Verdict = %q, want fail", report.Verdict)
	}
	scanner := report.SubagentResults[0]
	if scanner.Status != "fail" || scanner.Attempts != 2 {
		t.Fatalf("scanner result = %#v, want failed two-attempt result", scanner)
	}
	if len(scanner.AttemptErrors) != 2 || !strings.Contains(scanner.AttemptErrors[0], "scanner model unavailable") {
		t.Fatalf("scanner attempt errors = %#v, want two model errors", scanner.AttemptErrors)
	}
	if len(scanner.AttemptRecords) != 2 || scanner.AttemptRecords[1].Status != "fail" {
		t.Fatalf("scanner attempt records = %#v, want two failed attempts", scanner.AttemptRecords)
	}
	if report.VerificationSummary.SubagentFailCount != 1 {
		t.Fatalf("SubagentFailCount = %d, want 1", report.VerificationSummary.SubagentFailCount)
	}
	if report.RunManifest.Verdict != "fail" {
		t.Fatalf("RunManifest verdict = %q, want fail", report.RunManifest.Verdict)
	}
}
