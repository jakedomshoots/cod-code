package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"ceoharness/internal/history"
)

func Test_Run_prints_provider_health_summary_without_rows_when_summary_only_flag_is_supplied(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	store, err := history.New(root)
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	entries := []history.Entry{
		{
			Task: "Fix checkout retry",
			ProviderHealth: []history.ProviderHealth{
				{ProviderName: "fast", ModelSource: "http", AttemptCount: 2, PassCount: 2, EstimatedCostMicroUSD: 100},
			},
		},
		{
			Task: "Refactor parser",
			ProviderHealth: []history.ProviderHealth{
				{ProviderName: "cheap", ModelSource: "http", AttemptCount: 1, FailCount: 1, ErrorCount: 1, EstimatedCostMicroUSD: 50},
			},
		},
	}
	for _, entry := range entries {
		if _, err := store.Append(context.Background(), entry); err != nil {
			t.Fatalf("Append returned error: %v", err)
		}
	}

	// When
	err = Run(context.Background(), &out, []string{"--workspace", root, "--provider-health", "--summary-only"})
	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v\n%s", err, out.String())
	}
	var body map[string]json.RawMessage
	if jsonErr := json.Unmarshal(out.Bytes(), &body); jsonErr != nil {
		t.Fatalf("output must be JSON: %v\n%s", jsonErr, out.String())
	}
	if _, ok := body["provider_health"]; ok {
		t.Fatalf("provider health rows should be omitted in summary-only output: %s", out.String())
	}
	var summary struct {
		AvoidCount            int   `json:"avoid_count"`
		HealthyCount          int   `json:"healthy_count"`
		ProviderCount         int   `json:"provider_count"`
		AttemptCount          int   `json:"attempt_count"`
		PassCount             int   `json:"pass_count"`
		FailCount             int   `json:"fail_count"`
		ErrorCount            int   `json:"error_count"`
		EstimatedCostMicroUSD int64 `json:"estimated_cost_microusd"`
	}
	if jsonErr := json.Unmarshal(body["summary"], &summary); jsonErr != nil {
		t.Fatalf("summary must be JSON: %v\n%s", jsonErr, out.String())
	}
	if summary.AvoidCount != 1 || summary.HealthyCount != 1 {
		t.Fatalf("summary = %#v, want one avoid and one healthy", summary)
	}
	if summary.ProviderCount != 2 ||
		summary.AttemptCount != 3 ||
		summary.PassCount != 2 ||
		summary.FailCount != 1 ||
		summary.ErrorCount != 1 ||
		summary.EstimatedCostMicroUSD != 150 {
		t.Fatalf("summary totals = %#v, want provider health totals", summary)
	}
}
