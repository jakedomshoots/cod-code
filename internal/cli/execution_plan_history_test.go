package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"
)

func Test_Run_history_includes_execution_plan_metadata(t *testing.T) {
	// Given
	var runOut bytes.Buffer
	root := t.TempDir()
	err := Run(context.Background(), &runOut, []string{"--workspace", root, "Fix", "a", "failing", "test"})
	if err != nil {
		t.Fatalf("initial Run returned error: %v\n%s", err, runOut.String())
	}
	var historyOut bytes.Buffer

	// When
	err = Run(context.Background(), &historyOut, []string{"--workspace", root, "--history"})
	// Then
	if err != nil {
		t.Fatalf("history Run returned error: %v\n%s", err, historyOut.String())
	}
	var body struct {
		History []struct {
			ExecutionPlanStepCount  int    `json:"execution_plan_step_count"`
			ExecutionPlanNextAction string `json:"execution_plan_next_action"`
			RunLedger               struct {
				Owner              string `json:"owner"`
				Verdict            string `json:"verdict"`
				NextAction         string `json:"next_action"`
				VerificationStatus string `json:"verification_status"`
				ChangedFileCount   int    `json:"changed_file_count"`
			} `json:"run_ledger"`
		} `json:"history"`
	}
	if jsonErr := json.Unmarshal(historyOut.Bytes(), &body); jsonErr != nil {
		t.Fatalf("output must be JSON: %v\n%s", jsonErr, historyOut.String())
	}
	if len(body.History) != 1 {
		t.Fatalf("history length = %d, want 1", len(body.History))
	}
	if body.History[0].ExecutionPlanStepCount != 4 || body.History[0].ExecutionPlanNextAction != "accept" {
		t.Fatalf("history execution plan metadata = %#v, want 4 accept", body.History[0])
	}
	if body.History[0].RunLedger.Owner != "coder" || body.History[0].RunLedger.Verdict != "pass" || body.History[0].RunLedger.NextAction != "accept" {
		t.Fatalf("history run ledger = %#v, want coder pass accept", body.History[0].RunLedger)
	}
	if body.History[0].RunLedger.VerificationStatus != "unverified" || body.History[0].RunLedger.ChangedFileCount == 0 {
		t.Fatalf("history run ledger = %#v, want unverified ledger with changed files", body.History[0].RunLedger)
	}
}
