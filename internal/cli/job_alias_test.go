package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"ceoharness/internal/history"
)

func Test_Run_prints_latest_history_job_when_latest_alias_is_supplied(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	store, err := history.New(root)
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	if _, err := store.Append(context.Background(), history.Entry{Task: "First job", Verdict: "pass"}); err != nil {
		t.Fatalf("Append returned error: %v", err)
	}
	if _, err := store.Append(context.Background(), history.Entry{Task: "Latest job", Verdict: "needs_input"}); err != nil {
		t.Fatalf("Append returned error: %v", err)
	}

	// When
	err = Run(context.Background(), &out, []string{"--workspace", root, "--job", "latest"})
	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	var body struct {
		Job history.Entry `json:"job"`
	}
	if err := json.Unmarshal(out.Bytes(), &body); err != nil {
		t.Fatalf("output must be JSON: %v\n%s", err, out.String())
	}
	if body.Job.ID != "job-000002" || body.Job.Task != "Latest job" {
		t.Fatalf("job = %#v, want latest stored job", body.Job)
	}
}

func Test_Run_continues_latest_job_when_last_alias_is_supplied(t *testing.T) {
	// Given
	root := t.TempDir()
	if err := Run(context.Background(), &bytes.Buffer{}, []string{"--workspace", root, "Fix", "auth", "bug"}); err != nil {
		t.Fatalf("seed Run returned error: %v", err)
	}
	var out bytes.Buffer

	// When
	err := Run(context.Background(), &out, []string{"--workspace", root, "--continue-job", "last"})
	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v\n%s", err, out.String())
	}
	var body struct {
		Continuation struct {
			JobID               string `json:"job_id"`
			ReusedSubagentCount int    `json:"reused_subagent_count"`
		} `json:"continuation"`
	}
	if err := json.Unmarshal(out.Bytes(), &body); err != nil {
		t.Fatalf("output must be JSON: %v\n%s", err, out.String())
	}
	if body.Continuation.JobID != "job-000001" || body.Continuation.ReusedSubagentCount != 3 {
		t.Fatalf("continuation = %+v, want latest job reuse", body.Continuation)
	}
}
