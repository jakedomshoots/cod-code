package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"ceoharness/internal/history"
)

func Test_Run_reruns_history_job_when_rerun_flag_is_supplied(t *testing.T) {
	// Given
	root := t.TempDir()
	if err := Run(context.Background(), &bytes.Buffer{}, []string{"--workspace", root, "Fix", "auth", "bug"}); err != nil {
		t.Fatalf("seed Run returned error: %v", err)
	}
	var out bytes.Buffer

	// When
	err := Run(context.Background(), &out, []string{"--workspace", root, "--rerun", "job-000001"})

	// Then
	if err != nil {
		t.Fatalf("rerun returned error: %v", err)
	}
	var body struct {
		JobID     string `json:"job_id"`
		JobPacket struct {
			Task string `json:"task"`
		} `json:"job_packet"`
	}
	if err := json.Unmarshal(out.Bytes(), &body); err != nil {
		t.Fatalf("output must be JSON: %v\n%s", err, out.String())
	}
	if body.JobID != "job-000002" {
		t.Fatalf("JobID = %q, want job-000002", body.JobID)
	}
	if body.JobPacket.Task != "Fix auth bug" {
		t.Fatalf("rerun task = %q, want original task", body.JobPacket.Task)
	}
}

func Test_Run_rejects_rerun_when_task_args_are_also_supplied(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()

	// When
	err := Run(context.Background(), &out, []string{"--workspace", root, "--rerun", "job-000001", "New", "task"})

	// Then
	if err == nil {
		t.Fatal("expected rerun/task conflict error")
	}
}

func Test_Retry_latest_reruns_latest_failed_job_with_prior_context(t *testing.T) {
	// Given
	root := t.TempDir()
	store, err := history.New(root)
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	if _, err := store.Append(context.Background(), history.Entry{Task: "Older pass", Verdict: "pass"}); err != nil {
		t.Fatalf("Append returned error: %v", err)
	}
	if _, err := store.Append(context.Background(), history.Entry{Task: "Fix failed checkout", Verdict: "fail"}); err != nil {
		t.Fatalf("Append returned error: %v", err)
	}
	if _, err := store.Append(context.Background(), history.Entry{Task: "Newest pass", Verdict: "pass"}); err != nil {
		t.Fatalf("Append returned error: %v", err)
	}
	if _, err := store.SaveReportSnapshot(context.Background(), "job-000002", []byte(`{
		"schema_version": 1,
		"job_id": "job-000002",
		"verdict": "fail",
		"job_packet": {"task": "Fix failed checkout"},
		"check_results": [{
			"argv": ["go", "test", "./..."],
			"status": "fail",
			"exit_code": 1,
			"stderr": "checkout test failed"
		}]
	}`)); err != nil {
		t.Fatalf("SaveReportSnapshot returned error: %v", err)
	}
	var out bytes.Buffer

	// When
	err = Run(context.Background(), &out, []string{"retry", "latest", "--workspace", root})

	// Then
	if err != nil {
		t.Fatalf("retry latest returned error: %v", err)
	}
	var body struct {
		JobPacket struct {
			Task string `json:"task"`
		} `json:"job_packet"`
	}
	if err := json.Unmarshal(out.Bytes(), &body); err != nil {
		t.Fatalf("output must be JSON: %v\n%s", err, out.String())
	}
	if !strings.Contains(body.JobPacket.Task, "Fix failed checkout") {
		t.Fatalf("rerun task = %q, want latest failed job task", body.JobPacket.Task)
	}
	if !strings.Contains(body.JobPacket.Task, "prior_job_context:") ||
		!strings.Contains(body.JobPacket.Task, "previous_failed_check: go test ./... [fail]") {
		t.Fatalf("rerun task = %q, want preserved failed-check context", body.JobPacket.Task)
	}
	if strings.Contains(body.JobPacket.Task, "Newest pass") {
		t.Fatalf("rerun task = %q, should not use latest passing job", body.JobPacket.Task)
	}
}

func Test_Retry_latest_keeps_prior_context_compact_when_retrying_failed_retry(t *testing.T) {
	// Given
	root := t.TempDir()
	configJSON := `{"check_command":[` + strconv.Quote(os.Args[0]) + `,"-test.run=Test_HelperProcess_cli_fail_check"]}`
	if err := os.WriteFile(filepath.Join(root, ".ceo-harness.json"), []byte(configJSON), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	t.Setenv("GO_WANT_CLI_HELPER_PROCESS", "fail")
	if err := Run(context.Background(), &bytes.Buffer{}, []string{"--workspace", root, "Fix", "fixture", "failure"}); err == nil {
		t.Fatal("expected seed run to fail")
	}
	var firstRetry bytes.Buffer
	if err := Run(context.Background(), &firstRetry, []string{"retry", "latest", "--workspace", root}); err == nil {
		t.Fatal("expected first retry to fail")
	}
	var secondRetry bytes.Buffer

	// When
	err := Run(context.Background(), &secondRetry, []string{"retry", "latest", "--workspace", root})

	// Then
	if err == nil {
		t.Fatal("expected second retry to fail")
	}
	var body struct {
		JobPacket struct {
			Task string `json:"task"`
		} `json:"job_packet"`
	}
	if jsonErr := json.Unmarshal(secondRetry.Bytes(), &body); jsonErr != nil {
		t.Fatalf("output must be JSON: %v\n%s", jsonErr, secondRetry.String())
	}
	if got := strings.Count(body.JobPacket.Task, "prior_job_context:"); got != 1 {
		t.Fatalf("prior context count = %d, want 1 in task:\n%s", got, body.JobPacket.Task)
	}
	if got := strings.Count(body.JobPacket.Task, "previous_failed_check:"); got != 1 {
		t.Fatalf("previous failed check count = %d, want 1 in task:\n%s", got, body.JobPacket.Task)
	}
	if !strings.Contains(body.JobPacket.Task, "previous_job: job-000002") {
		t.Fatalf("task = %q, want immediate failed retry job context", body.JobPacket.Task)
	}
}
