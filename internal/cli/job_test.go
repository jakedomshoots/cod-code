package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"ceoharness/internal/history"
)

func Test_Run_prints_history_job_when_job_flag_is_supplied(t *testing.T) {
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
	err = Run(context.Background(), &out, []string{"--workspace", root, "--job", "job-000001"})

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
	if body.Job.ID != "job-000001" || body.Job.Task != "Scan repo" {
		t.Fatalf("job = %#v, want stored job", body.Job)
	}
}

func Test_Run_prints_run_ledger_in_history_job_when_job_flag_is_supplied(t *testing.T) {
	// Given
	var runOut bytes.Buffer
	root := t.TempDir()
	err := Run(context.Background(), &runOut, []string{"--workspace", root, "Fix", "a", "failing", "test"})
	if err != nil {
		t.Fatalf("initial Run returned error: %v\n%s", err, runOut.String())
	}
	var out bytes.Buffer

	// When
	err = Run(context.Background(), &out, []string{"--workspace", root, "--job", "job-000001"})

	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	var body struct {
		Job struct {
			RunLedger struct {
				Owner              string `json:"owner"`
				Verdict            string `json:"verdict"`
				NextAction         string `json:"next_action"`
				VerificationStatus string `json:"verification_status"`
				ChangedFileCount   int    `json:"changed_file_count"`
			} `json:"run_ledger"`
		} `json:"job"`
	}
	if err := json.Unmarshal(out.Bytes(), &body); err != nil {
		t.Fatalf("output must be JSON: %v\n%s", err, out.String())
	}
	if body.Job.RunLedger.Owner != "coder" || body.Job.RunLedger.Verdict != "pass" || body.Job.RunLedger.NextAction != "accept" {
		t.Fatalf("job run ledger = %#v, want coder pass accept", body.Job.RunLedger)
	}
	if body.Job.RunLedger.VerificationStatus != "unverified" || body.Job.RunLedger.ChangedFileCount == 0 {
		t.Fatalf("job run ledger = %#v, want unverified ledger with changed files", body.Job.RunLedger)
	}
}

func Test_Run_prints_human_judgment_in_history_job_when_present(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	store, err := history.New(root)
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	if _, err := store.Append(context.Background(), history.Entry{Task: "Scan repo", Verdict: "pass"}); err != nil {
		t.Fatalf("Append returned error: %v", err)
	}
	if _, err := store.SaveHumanJudgment(context.Background(), history.HumanJudgment{
		JobID:   "job-000001",
		Verdict: "reject",
		Note:    "Needs evidence.",
	}); err != nil {
		t.Fatalf("SaveHumanJudgment returned error: %v", err)
	}

	// When
	err = Run(context.Background(), &out, []string{"--workspace", root, "--job", "job-000001"})

	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	var body struct {
		HumanJudgment *history.HumanJudgment `json:"human_judgment"`
	}
	if err := json.Unmarshal(out.Bytes(), &body); err != nil {
		t.Fatalf("output must be JSON: %v\n%s", err, out.String())
	}
	if body.HumanJudgment == nil {
		t.Fatalf("human judgment missing from job lookup:\n%s", out.String())
	}
	if body.HumanJudgment.JobID != "job-000001" || body.HumanJudgment.Verdict != "reject" || body.HumanJudgment.Note != "Needs evidence." {
		t.Fatalf("human judgment = %#v, want saved rejection", body.HumanJudgment)
	}
}

func Test_Run_prints_job_report_snapshot_when_job_report_flag_is_supplied(t *testing.T) {
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
	if _, err := store.SaveReportSnapshot(context.Background(), "job-000001", []byte(`{"job_id":"job-000001","verdict":"pass"}`)); err != nil {
		t.Fatalf("SaveReportSnapshot returned error: %v", err)
	}

	// When
	err = Run(context.Background(), &out, []string{"--workspace", root, "--job-report", "job-000001"})

	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	var body struct {
		JobID   string `json:"job_id"`
		Verdict string `json:"verdict"`
	}
	if err := json.Unmarshal(out.Bytes(), &body); err != nil {
		t.Fatalf("output must be JSON: %v\n%s", err, out.String())
	}
	if body.JobID != "job-000001" || body.Verdict != "pass" {
		t.Fatalf("report = %#v, want saved snapshot", body)
	}
}

func Test_Run_prints_legacy_schema_warning_when_job_report_latest_is_old_snapshot(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	store, err := history.New(root)
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	if _, err := store.Append(context.Background(), history.Entry{
		Task:    "Legacy report",
		Verdict: "pass",
	}); err != nil {
		t.Fatalf("Append returned error: %v", err)
	}
	if _, err := store.SaveReportSnapshot(context.Background(), "job-000001", []byte(`{"job_id":"job-000001","verdict":"pass"}`)); err != nil {
		t.Fatalf("SaveReportSnapshot returned error: %v", err)
	}

	// When
	err = Run(context.Background(), &out, []string{"--workspace", root, "--job-report", "latest"})

	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	var body struct {
		JobID               string `json:"job_id"`
		Verdict             string `json:"verdict"`
		SchemaCompatibility struct {
			Status               string `json:"status"`
			Warning              string `json:"warning"`
			AssumedSchemaVersion int    `json:"assumed_schema_version"`
			ReaderSchemaVersion  int    `json:"reader_schema_version"`
		} `json:"schema_compatibility"`
	}
	if err := json.Unmarshal(out.Bytes(), &body); err != nil {
		t.Fatalf("output must be JSON: %v\n%s", err, out.String())
	}
	if body.JobID != "job-000001" || body.Verdict != "pass" {
		t.Fatalf("report = %#v, want legacy snapshot payload", body)
	}
	if body.SchemaCompatibility.Status != "legacy" || body.SchemaCompatibility.Warning == "" {
		t.Fatalf("schema compatibility = %+v, want legacy warning", body.SchemaCompatibility)
	}
	if body.SchemaCompatibility.AssumedSchemaVersion != 0 || body.SchemaCompatibility.ReaderSchemaVersion != 1 {
		t.Fatalf("schema compatibility = %+v, want assumed v0 and reader v1", body.SchemaCompatibility)
	}
}
