package cli

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"ceoharness/internal/history"
)

func Test_ExplainFailure_latest_failed_report_prints_plain_text(t *testing.T) {
	// Given
	root := t.TempDir()
	store, err := history.New(root)
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	if _, err := store.Append(context.Background(), history.Entry{Task: "Passing task", Verdict: "pass"}); err != nil {
		t.Fatalf("Append returned error: %v", err)
	}
	if _, err := store.Append(context.Background(), history.Entry{Task: "Fix broken checkout", Verdict: "fail"}); err != nil {
		t.Fatalf("Append returned error: %v", err)
	}
	if _, err := store.Append(context.Background(), history.Entry{Task: "Newest passing task", Verdict: "pass"}); err != nil {
		t.Fatalf("Append returned error: %v", err)
	}
	if _, err := store.SaveReportSnapshot(context.Background(), "job-000002", []byte(`{
		"schema_version": 1,
		"job_id": "job-000002",
		"verdict": "fail",
		"job_packet": {"task": "Fix broken checkout"},
		"run_ledger": {"next_action": "fix failing checks"},
		"check_results": [{
			"argv": ["go", "test", "./..."],
			"status": "fail",
			"exit_code": 1,
			"stderr": "checkout test failed"
		}],
		"verification_summary": {"check_fail_count": 1}
	}`)); err != nil {
		t.Fatalf("SaveReportSnapshot returned error: %v", err)
	}
	var out bytes.Buffer

	// When
	err = Run(context.Background(), &out, []string{"explain-failure", "latest", "--workspace", root})
	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	text := out.String()
	for _, want := range []string{
		"Job: job-000002",
		"Verdict: fail",
		"Likely reason: one or more checks failed",
		"Failed checks:",
		"go test ./... [fail]: checkout test failed",
		"Retryable: yes",
		"Suggested retry: cod retry job-000002 --workspace",
		"Report path: ceo-artifacts/jobs/job-000002.json",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("explain-failure output missing %q:\n%s", want, text)
		}
	}
	if strings.Contains(text, "{") {
		t.Fatalf("explain-failure output should be plain text, got JSON-looking output:\n%s", text)
	}
}

func Test_ExplainFailure_latest_blocks_when_history_has_no_failed_job(t *testing.T) {
	// Given
	root := t.TempDir()
	store, err := history.New(root)
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	if _, err := store.Append(context.Background(), history.Entry{Task: "Passing task", Verdict: "pass"}); err != nil {
		t.Fatalf("Append returned error: %v", err)
	}
	var out bytes.Buffer

	// When
	err = Run(context.Background(), &out, []string{"explain-failure", "latest", "--workspace", root})

	// Then
	if err == nil {
		t.Fatal("expected explain-failure latest to block without a failed job")
	}
	if !strings.Contains(err.Error(), "no failed jobs in history") {
		t.Fatalf("error = %q, want clear no failed jobs state", err)
	}
}
