package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"

	"ceoharness/internal/history"
)

func Test_Run_review_queue_lists_jobs_that_need_human_attention(t *testing.T) {
	// Given
	root := t.TempDir()
	store, err := history.New(root)
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	entries := []history.Entry{
		{Task: "Passing unjudged job", Verdict: "pass"},
		{Task: "Passing accepted job", Verdict: "pass"},
		{Task: "Passing rejected job", Verdict: "pass"},
		{Task: "Failing job", Verdict: "fail"},
		{Task: "Needs answer", Verdict: "needs_input"},
	}
	for _, entry := range entries {
		if _, err := store.Append(context.Background(), entry); err != nil {
			t.Fatalf("Append returned error: %v", err)
		}
	}
	for _, judgment := range []history.HumanJudgment{
		{JobID: "job-000002", Verdict: "accept"},
		{JobID: "job-000003", Verdict: "reject"},
	} {
		if _, err := store.SaveHumanJudgment(context.Background(), judgment); err != nil {
			t.Fatalf("SaveHumanJudgment returned error: %v", err)
		}
	}
	var out bytes.Buffer

	// When
	err = Run(context.Background(), &out, []string{"--workspace", root, "--review-queue"})
	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v\n%s", err, out.String())
	}
	var body struct {
		TotalCount int `json:"total_count"`
		Queue      []struct {
			ID               string                 `json:"id"`
			Task             string                 `json:"task"`
			ReviewReason     string                 `json:"review_reason"`
			HumanJudgment    *history.HumanJudgment `json:"human_judgment"`
			SuggestedCommand string                 `json:"suggested_command"`
		} `json:"review_queue"`
	}
	if err := json.Unmarshal(out.Bytes(), &body); err != nil {
		t.Fatalf("output must be JSON: %v\n%s", err, out.String())
	}
	if body.TotalCount != 4 || len(body.Queue) != 4 {
		t.Fatalf("queue count = %d/%d, want 4", body.TotalCount, len(body.Queue))
	}
	wantReasons := map[string]string{
		"job-000001": "awaiting_human_judgment",
		"job-000003": "human_rejected",
		"job-000004": "failed_or_unresolved",
		"job-000005": "needs_input",
	}
	for _, row := range body.Queue {
		if row.ReviewReason != wantReasons[row.ID] {
			t.Fatalf("row %#v reason = %q, want %q", row, row.ReviewReason, wantReasons[row.ID])
		}
		if row.SuggestedCommand == "" {
			t.Fatalf("row %#v missing suggested command", row)
		}
	}
	if body.Queue[1].HumanJudgment == nil || body.Queue[1].HumanJudgment.Verdict != "reject" {
		t.Fatalf("rejected job row = %#v, want reject sidecar", body.Queue[1])
	}
}

func Test_Run_review_queue_includes_compact_recovery_fields(t *testing.T) {
	// Given
	root := t.TempDir()
	store, err := history.New(root)
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	for _, entry := range []history.Entry{
		{Task: "Failing job", Verdict: "fail", ExecutionPlanNextAction: "rerun"},
		{Task: "Needs answer", Verdict: "needs_input", ExecutionPlanNextAction: "answer"},
		{Task: "Passing unjudged job", Verdict: "pass", ExecutionPlanNextAction: "judge"},
		{Task: "Rejected job", Verdict: "pass", ExecutionPlanNextAction: "rerun after rejection"},
	} {
		if _, err := store.Append(context.Background(), entry); err != nil {
			t.Fatalf("Append returned error: %v", err)
		}
	}
	if _, err := store.SaveHumanJudgment(context.Background(), history.HumanJudgment{
		JobID:   "job-000004",
		Verdict: "reject",
	}); err != nil {
		t.Fatalf("SaveHumanJudgment returned error: %v", err)
	}
	var out bytes.Buffer

	// When
	err = Run(context.Background(), &out, []string{"--workspace", root, "--review-queue"})
	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v\n%s", err, out.String())
	}
	var body struct {
		Queue []struct {
			ID              string `json:"id"`
			RecoveryState   string `json:"recovery_state"`
			LastVerdict     string `json:"last_verdict"`
			Retryable       bool   `json:"retryable"`
			NextAction      string `json:"next_action"`
			EvidencePointer string `json:"evidence_pointer"`
		} `json:"review_queue"`
	}
	if err := json.Unmarshal(out.Bytes(), &body); err != nil {
		t.Fatalf("output must be JSON: %v\n%s", err, out.String())
	}
	if len(body.Queue) != 4 {
		t.Fatalf("queue count = %d, want 4", len(body.Queue))
	}
	if body.Queue[0].RecoveryState != "failed" || !body.Queue[0].Retryable {
		t.Fatalf("failed row = %#v, want failed retryable state", body.Queue[0])
	}
	if body.Queue[1].RecoveryState != "needs-input" || body.Queue[1].LastVerdict != "needs_input" {
		t.Fatalf("needs-input row = %#v, want needs-input state", body.Queue[1])
	}
	if body.Queue[2].RecoveryState != "waiting-review" {
		t.Fatalf("waiting row = %#v, want waiting-review state", body.Queue[2])
	}
	if body.Queue[3].RecoveryState != "rejected" {
		t.Fatalf("rejected row = %#v, want rejected state", body.Queue[3])
	}
	if body.Queue[0].NextAction == "" || body.Queue[0].EvidencePointer != "ceo-artifacts/jobs/job-000001.json" {
		t.Fatalf("operator pointers = %#v, want compact action and evidence", body.Queue[0])
	}
}

func Test_Run_review_queue_prints_text_when_format_text_is_supplied(t *testing.T) {
	// Given
	root := t.TempDir()
	store, err := history.New(root)
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	if _, err := store.Append(context.Background(), history.Entry{Task: "Needs answer", Verdict: "needs_input"}); err != nil {
		t.Fatalf("Append returned error: %v", err)
	}
	var out bytes.Buffer

	// When
	err = Run(context.Background(), &out, []string{"--workspace", root, "--review-queue", "--format", "text"})
	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v\n%s", err, out.String())
	}
	text := out.String()
	for _, want := range []string{
		"Review queue: 1 job",
		"- job-000001 [needs-input] Needs answer",
		"Verdict: needs_input",
		"Retryable: no",
		"Evidence: ceo-artifacts/jobs/job-000001.json",
		"Next: cod",
		"--resume job-000001",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("text output = %q, want %q", text, want)
		}
	}
}

func Test_Run_review_queue_rejects_events_format(t *testing.T) {
	// Given
	root := t.TempDir()
	var out bytes.Buffer

	// When
	err := Run(context.Background(), &out, []string{"--workspace", root, "--review-queue", "--format", "events"})

	// Then
	if err == nil {
		t.Fatal("expected events format error")
	}
	if !strings.Contains(err.Error(), "only available for run reports") {
		t.Fatalf("error = %q, want run reports guidance", err.Error())
	}
}

func Test_Run_review_queue_respects_task_filter_and_limit_after_filtering(t *testing.T) {
	// Given
	root := t.TempDir()
	store, err := history.New(root)
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	for _, entry := range []history.Entry{
		{Task: "Auth pass", Verdict: "pass"},
		{Task: "Billing fail", Verdict: "fail"},
		{Task: "Auth needs input", Verdict: "needs_input"},
		{Task: "Auth fail", Verdict: "fail"},
		{Task: "Auth accepted", Verdict: "pass"},
	} {
		if _, err := store.Append(context.Background(), entry); err != nil {
			t.Fatalf("Append returned error: %v", err)
		}
	}
	if _, err := store.SaveHumanJudgment(context.Background(), history.HumanJudgment{
		JobID:   "job-000005",
		Verdict: "accept",
	}); err != nil {
		t.Fatalf("SaveHumanJudgment returned error: %v", err)
	}
	var out bytes.Buffer

	// When
	err = Run(context.Background(), &out, []string{
		"--workspace", root,
		"--review-queue",
		"--task", "auth",
		"--limit", "2",
	})
	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v\n%s", err, out.String())
	}
	var body struct {
		TotalCount int    `json:"total_count"`
		TaskFilter string `json:"task_filter"`
		Queue      []struct {
			Task string `json:"task"`
		} `json:"review_queue"`
	}
	if err := json.Unmarshal(out.Bytes(), &body); err != nil {
		t.Fatalf("output must be JSON: %v\n%s", err, out.String())
	}
	if body.TaskFilter != "auth" {
		t.Fatalf("task_filter = %q, want auth", body.TaskFilter)
	}
	if body.TotalCount != 2 || len(body.Queue) != 2 {
		t.Fatalf("queue count = %d/%d, want 2", body.TotalCount, len(body.Queue))
	}
	if body.Queue[0].Task != "Auth needs input" || body.Queue[1].Task != "Auth fail" {
		t.Fatalf("queue = %#v, want latest two auth rows", body.Queue)
	}
}
