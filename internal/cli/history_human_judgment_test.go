package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"ceoharness/internal/history"
)

func Test_Run_history_includes_human_judgment_when_present(t *testing.T) {
	// Given
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
		Verdict: "accept",
		Note:    "Ready.",
	}); err != nil {
		t.Fatalf("SaveHumanJudgment returned error: %v", err)
	}
	var out bytes.Buffer

	// When
	err = Run(context.Background(), &out, []string{"--workspace", root, "--history"})

	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v\n%s", err, out.String())
	}
	var body struct {
		History []struct {
			ID                string                 `json:"id"`
			HumanJudgment     *history.HumanJudgment `json:"human_judgment"`
			HumanJudgmentPath string                 `json:"human_judgment_path"`
		} `json:"history"`
	}
	if err := json.Unmarshal(out.Bytes(), &body); err != nil {
		t.Fatalf("output must be JSON: %v\n%s", err, out.String())
	}
	if len(body.History) != 1 {
		t.Fatalf("history length = %d, want 1", len(body.History))
	}
	row := body.History[0]
	if row.ID != "job-000001" || row.HumanJudgment == nil {
		t.Fatalf("history row = %#v, want judged job", row)
	}
	if row.HumanJudgment.Verdict != "accept" || row.HumanJudgment.Note != "Ready." {
		t.Fatalf("human judgment = %#v, want accepted note", row.HumanJudgment)
	}
	if row.HumanJudgmentPath != "ceo-artifacts/human-judgments/job-000001.json" {
		t.Fatalf("HumanJudgmentPath = %q, want sidecar path", row.HumanJudgmentPath)
	}
}

func Test_Run_history_summary_counts_human_judgments(t *testing.T) {
	// Given
	root := t.TempDir()
	store, err := history.New(root)
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	for _, task := range []string{"First", "Second", "Third"} {
		if _, err := store.Append(context.Background(), history.Entry{Task: task, Verdict: "pass"}); err != nil {
			t.Fatalf("Append returned error: %v", err)
		}
	}
	for _, judgment := range []history.HumanJudgment{
		{JobID: "job-000001", Verdict: "accept"},
		{JobID: "job-000002", Verdict: "reject"},
	} {
		if _, err := store.SaveHumanJudgment(context.Background(), judgment); err != nil {
			t.Fatalf("SaveHumanJudgment returned error: %v", err)
		}
	}
	var out bytes.Buffer

	// When
	err = Run(context.Background(), &out, []string{"--workspace", root, "--history", "--summary-only"})

	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v\n%s", err, out.String())
	}
	var body struct {
		Summary struct {
			HumanJudgmentCount int            `json:"human_judgment_count"`
			HumanVerdictCounts map[string]int `json:"human_verdict_counts"`
		} `json:"summary"`
	}
	if err := json.Unmarshal(out.Bytes(), &body); err != nil {
		t.Fatalf("output must be JSON: %v\n%s", err, out.String())
	}
	if body.Summary.HumanJudgmentCount != 2 {
		t.Fatalf("HumanJudgmentCount = %d, want 2", body.Summary.HumanJudgmentCount)
	}
	if body.Summary.HumanVerdictCounts["accept"] != 1 || body.Summary.HumanVerdictCounts["reject"] != 1 {
		t.Fatalf("HumanVerdictCounts = %#v, want accept/reject counts", body.Summary.HumanVerdictCounts)
	}
}
