package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"ceoharness/internal/history"
)

func Test_Run_prints_history_summary_without_rows_when_summary_only_flag_is_supplied(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	store, err := history.New(root)
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	entries := []history.Entry{
		{Task: "Fix checkout retry", Verdict: "pass"},
		{Task: "CHECKOUT smoke", Verdict: "fail"},
		{Task: "Refactor parser", Verdict: "pass"},
	}
	for _, entry := range entries {
		if _, err := store.Append(context.Background(), entry); err != nil {
			t.Fatalf("Append returned error: %v", err)
		}
	}

	// When
	err = Run(context.Background(), &out, []string{"--workspace", root, "--history", "--task", "checkout", "--summary-only"})
	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	var body map[string]json.RawMessage
	if jsonErr := json.Unmarshal(out.Bytes(), &body); jsonErr != nil {
		t.Fatalf("output must be JSON: %v\n%s", jsonErr, out.String())
	}
	if _, ok := body["history"]; ok {
		t.Fatalf("history rows should be omitted in summary-only output: %s", out.String())
	}
	var summary struct {
		TotalCount    int            `json:"total_count"`
		VerdictCounts map[string]int `json:"verdict_counts"`
	}
	if jsonErr := json.Unmarshal(body["summary"], &summary); jsonErr != nil {
		t.Fatalf("summary must be JSON: %v\n%s", jsonErr, out.String())
	}
	if summary.TotalCount != 2 || summary.VerdictCounts["pass"] != 1 || summary.VerdictCounts["fail"] != 1 {
		t.Fatalf("summary = %#v, want checkout pass/fail counts", summary)
	}
	var taskFilter string
	if jsonErr := json.Unmarshal(body["task_filter"], &taskFilter); jsonErr != nil {
		t.Fatalf("task_filter must be JSON string: %v\n%s", jsonErr, out.String())
	}
	if taskFilter != "checkout" {
		t.Fatalf("task_filter = %q, want checkout", taskFilter)
	}
}

func Test_Run_prints_history_work_totals_when_summary_only_flag_is_supplied(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	store, err := history.New(root)
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	entries := []history.Entry{
		{
			Task:                          "Fix checkout retry",
			Verdict:                       "pass",
			SubagentCount:                 3,
			ReusedSubagentCount:           2,
			SubagentAttemptCount:          4,
			SubagentRetryCount:            1,
			SubagentRetriedCount:          1,
			SubagentRetryExhaustedCount:   0,
			SubagentNoProgressStopCount:   1,
			CheckCount:                    1,
			PatchCount:                    2,
			CLIPatchCount:                 1,
			ModelPatchCount:               1,
			CheckFixCount:                 1,
			ProviderErrorCount:            1,
			ProviderUnauthorizedCount:     1,
			ProviderEstimatedCostMicroUSD: 100,
		},
		{
			Task:                          "CHECKOUT smoke",
			Verdict:                       "fail",
			SubagentCount:                 3,
			ReusedSubagentCount:           3,
			SubagentAttemptCount:          6,
			SubagentRetryCount:            3,
			SubagentRetriedCount:          2,
			SubagentRetryExhaustedCount:   1,
			SubagentNoProgressStopCount:   2,
			CheckCount:                    2,
			PatchCount:                    1,
			ModelPatchCount:               1,
			ProviderErrorCount:            2,
			ProviderRateLimitedCount:      1,
			ProviderUnavailableCount:      1,
			ProviderEstimatedCostMicroUSD: 250,
			ProviderCostOverBudget:        true,
		},
		{
			Task:                          "Refactor parser",
			Verdict:                       "pass",
			SubagentCount:                 3,
			ReusedSubagentCount:           9,
			SubagentAttemptCount:          3,
			CheckCount:                    1,
			ProviderEstimatedCostMicroUSD: 999,
			ProviderCostOverBudget:        true,
		},
	}
	for _, entry := range entries {
		if _, err := store.Append(context.Background(), entry); err != nil {
			t.Fatalf("Append returned error: %v", err)
		}
	}

	// When
	err = Run(context.Background(), &out, []string{"--workspace", root, "--history", "--task", "checkout", "--summary-only"})
	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	var body struct {
		Summary struct {
			TotalCount                    int   `json:"total_count"`
			SubagentCount                 int   `json:"subagent_count"`
			ReusedSubagentCount           int   `json:"reused_subagent_count"`
			SubagentAttemptCount          int   `json:"subagent_attempt_count"`
			SubagentRetryCount            int   `json:"subagent_retry_count"`
			SubagentRetriedCount          int   `json:"subagent_retried_count"`
			SubagentRetryExhaustedCount   int   `json:"subagent_retry_exhausted_count"`
			SubagentNoProgressStopCount   int   `json:"subagent_no_progress_stop_count"`
			CheckCount                    int   `json:"check_count"`
			PatchCount                    int   `json:"patch_count"`
			CLIPatchCount                 int   `json:"cli_patch_count"`
			ModelPatchCount               int   `json:"model_patch_count"`
			CheckFixCount                 int   `json:"check_fix_count"`
			ProviderErrorCount            int   `json:"provider_error_count"`
			ProviderUnauthorizedCount     int   `json:"provider_unauthorized_count"`
			ProviderRateLimitedCount      int   `json:"provider_rate_limited_count"`
			ProviderUnavailableCount      int   `json:"provider_unavailable_count"`
			ProviderEstimatedCostMicroUSD int64 `json:"provider_estimated_cost_microusd"`
			ProviderCostOverBudgetCount   int   `json:"provider_cost_over_budget_count"`
		} `json:"summary"`
	}
	if jsonErr := json.Unmarshal(out.Bytes(), &body); jsonErr != nil {
		t.Fatalf("output must be JSON: %v\n%s", jsonErr, out.String())
	}
	if body.Summary.TotalCount != 2 ||
		body.Summary.SubagentCount != 6 ||
		body.Summary.ReusedSubagentCount != 5 ||
		body.Summary.SubagentAttemptCount != 10 ||
		body.Summary.SubagentRetryCount != 4 ||
		body.Summary.SubagentRetriedCount != 3 ||
		body.Summary.SubagentRetryExhaustedCount != 1 ||
		body.Summary.SubagentNoProgressStopCount != 3 ||
		body.Summary.CheckCount != 3 ||
		body.Summary.PatchCount != 3 ||
		body.Summary.CLIPatchCount != 1 ||
		body.Summary.ModelPatchCount != 2 ||
		body.Summary.CheckFixCount != 1 ||
		body.Summary.ProviderErrorCount != 3 ||
		body.Summary.ProviderUnauthorizedCount != 1 ||
		body.Summary.ProviderRateLimitedCount != 1 ||
		body.Summary.ProviderUnavailableCount != 1 ||
		body.Summary.ProviderEstimatedCostMicroUSD != 350 ||
		body.Summary.ProviderCostOverBudgetCount != 1 {
		t.Fatalf("summary = %#v, want checkout work totals", body.Summary)
	}
}

func Test_Run_status_summary_counts_operator_recovery_states(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	store, err := history.New(root)
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	for _, entry := range []history.Entry{
		{Task: "Failing job", Verdict: "fail"},
		{Task: "Needs answer", Verdict: "needs_input"},
		{Task: "Waiting review", Verdict: "pass"},
		{Task: "Accepted job", Verdict: "pass"},
		{Task: "Rejected job", Verdict: "pass"},
	} {
		if _, err := store.Append(context.Background(), entry); err != nil {
			t.Fatalf("Append returned error: %v", err)
		}
	}
	for _, judgment := range []history.HumanJudgment{
		{JobID: "job-000004", Verdict: "accept"},
		{JobID: "job-000005", Verdict: "reject"},
	} {
		if _, err := store.SaveHumanJudgment(context.Background(), judgment); err != nil {
			t.Fatalf("SaveHumanJudgment returned error: %v", err)
		}
	}

	// When
	err = Run(context.Background(), &out, []string{"status", "--workspace", root})
	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v\n%s", err, out.String())
	}
	var body struct {
		Summary struct {
			RecoveryStateCounts map[string]int `json:"recovery_state_counts"`
			RetryableCount      int            `json:"retryable_count"`
		} `json:"summary"`
	}
	if jsonErr := json.Unmarshal(out.Bytes(), &body); jsonErr != nil {
		t.Fatalf("output must be JSON: %v\n%s", jsonErr, out.String())
	}
	wantCounts := map[string]int{
		"failed":         1,
		"needs-input":    1,
		"waiting-review": 1,
		"accepted":       1,
		"rejected":       1,
	}
	for state, want := range wantCounts {
		if body.Summary.RecoveryStateCounts[state] != want {
			t.Fatalf("RecoveryStateCounts[%q] = %d, want %d in %#v", state, body.Summary.RecoveryStateCounts[state], want, body.Summary.RecoveryStateCounts)
		}
	}
	if body.Summary.RetryableCount != 1 {
		t.Fatalf("RetryableCount = %d, want 1", body.Summary.RetryableCount)
	}
}
