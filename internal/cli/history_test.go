package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"
	"time"

	"ceoharness/internal/history"
)

func Test_Run_prints_job_history_when_history_flag_is_supplied(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	store, err := history.New(root)
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	if _, err := store.Append(context.Background(), history.Entry{
		Task:    "Scan repo",
		Verdict: "pass",
	}); err != nil {
		t.Fatalf("Append returned error: %v", err)
	}

	// When
	err = Run(context.Background(), &out, []string{"--workspace", root, "--history"})
	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	var body struct {
		History []struct {
			Task    string `json:"task"`
			Verdict string `json:"verdict"`
		} `json:"history"`
	}
	if jsonErr := json.Unmarshal(out.Bytes(), &body); jsonErr != nil {
		t.Fatalf("output must be JSON: %v\n%s", jsonErr, out.String())
	}
	if len(body.History) != 1 {
		t.Fatalf("history length = %d, want 1", len(body.History))
	}
	if body.History[0].Task != "Scan repo" || body.History[0].Verdict != "pass" {
		t.Fatalf("history entry = %#v, want stored job", body.History[0])
	}
}

func Test_Run_prints_created_at_when_history_entry_has_timestamp(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	fixedTime := time.Date(2026, 7, 1, 12, 30, 0, 0, time.UTC)
	store, err := history.NewWithClock(root, func() time.Time { return fixedTime })
	if err != nil {
		t.Fatalf("NewWithClock returned error: %v", err)
	}
	if _, err := store.Append(context.Background(), history.Entry{
		Task:    "Scan repo",
		Verdict: "pass",
	}); err != nil {
		t.Fatalf("Append returned error: %v", err)
	}

	// When
	err = Run(context.Background(), &out, []string{"--workspace", root, "--history"})
	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	var body struct {
		History []struct {
			CreatedAt string `json:"created_at"`
		} `json:"history"`
	}
	if jsonErr := json.Unmarshal(out.Bytes(), &body); jsonErr != nil {
		t.Fatalf("output must be JSON: %v\n%s", jsonErr, out.String())
	}
	if len(body.History) != 1 {
		t.Fatalf("history length = %d, want 1", len(body.History))
	}
	if body.History[0].CreatedAt != "2026-07-01T12:30:00Z" {
		t.Fatalf("CreatedAt = %q, want timestamp", body.History[0].CreatedAt)
	}
}

func Test_Run_history_includes_compact_recovery_fields(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	store, err := history.New(root)
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	for _, entry := range []history.Entry{
		{Task: "Fix checkout", Verdict: "fail", ExecutionPlanNextAction: "rerun after fixing tests"},
		{Task: "Ask operator", Verdict: "needs_input", ExecutionPlanNextAction: "answer question"},
		{Task: "Ready for review", Verdict: "pass", ExecutionPlanNextAction: "judge result"},
	} {
		if _, err := store.Append(context.Background(), entry); err != nil {
			t.Fatalf("Append returned error: %v", err)
		}
	}

	// When
	err = Run(context.Background(), &out, []string{"--workspace", root, "--history"})
	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	var body struct {
		History []struct {
			ID              string `json:"id"`
			RecoveryState   string `json:"recovery_state"`
			LastVerdict     string `json:"last_verdict"`
			Retryable       bool   `json:"retryable"`
			NextAction      string `json:"next_action"`
			EvidencePointer string `json:"evidence_pointer"`
		} `json:"history"`
	}
	if jsonErr := json.Unmarshal(out.Bytes(), &body); jsonErr != nil {
		t.Fatalf("output must be JSON: %v\n%s", jsonErr, out.String())
	}
	if len(body.History) != 3 {
		t.Fatalf("history length = %d, want 3", len(body.History))
	}
	if body.History[0].RecoveryState != "failed" || !body.History[0].Retryable {
		t.Fatalf("failed row = %#v, want failed retryable state", body.History[0])
	}
	if body.History[1].RecoveryState != "needs-input" || body.History[1].LastVerdict != "needs_input" {
		t.Fatalf("needs-input row = %#v, want needs-input state", body.History[1])
	}
	if body.History[2].RecoveryState != "waiting-review" {
		t.Fatalf("pass row = %#v, want waiting-review state", body.History[2])
	}
	if body.History[0].NextAction != "rerun after fixing tests" {
		t.Fatalf("NextAction = %q, want execution plan action", body.History[0].NextAction)
	}
	if body.History[0].EvidencePointer != "ceo-artifacts/jobs/job-000001.json" {
		t.Fatalf("EvidencePointer = %q, want compact report path", body.History[0].EvidencePointer)
	}
}
