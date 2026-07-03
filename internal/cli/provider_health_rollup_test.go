package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"ceoharness/internal/history"
)

func Test_Run_prints_provider_health_rollup_when_provider_health_flag_is_supplied(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	store, err := history.New(root)
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	entries := []history.Entry{
		{
			Task: "first",
			ProviderHealth: []history.ProviderHealth{
				{ProviderName: "fast", ModelSource: "http", AttemptCount: 1, PassCount: 1, EstimatedCostMicroUSD: 106},
			},
		},
		{
			Task: "second",
			ProviderHealth: []history.ProviderHealth{
				{ProviderName: "fast", ModelSource: "http", AttemptCount: 2, PassCount: 1, FailCount: 1, ErrorCount: 1, RateLimitedCount: 1, EstimatedCostMicroUSD: 44},
				{ProviderName: "cheap", ModelSource: "http", AttemptCount: 1, FailCount: 1, ErrorCount: 1, UnauthorizedCount: 1},
			},
		},
	}
	for _, entry := range entries {
		if _, err := store.Append(context.Background(), entry); err != nil {
			t.Fatalf("Append returned error: %v", err)
		}
	}

	// When
	err = Run(context.Background(), &out, []string{"--workspace", root, "--provider-health"})

	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v\n%s", err, out.String())
	}
	var body struct {
		HistoryPath    string `json:"history_path"`
		ProviderHealth []struct {
			ProviderName           string  `json:"provider_name"`
			AttemptCount           int     `json:"attempt_count"`
			PassCount              int     `json:"pass_count"`
			FailCount              int     `json:"fail_count"`
			ErrorCount             int     `json:"error_count"`
			UnauthorizedCount      int     `json:"unauthorized_count"`
			RateLimitedCount       int     `json:"rate_limited_count"`
			EstimatedCostMicroUSD  int64   `json:"estimated_cost_microusd"`
			FailureRate            float64 `json:"failure_rate"`
			CostPerAttemptMicroUSD int64   `json:"cost_per_attempt_microusd"`
			Recommendation         string  `json:"recommendation"`
		} `json:"provider_health"`
		Summary struct {
			AvoidCount   int `json:"avoid_count"`
			WatchCount   int `json:"watch_count"`
			HealthyCount int `json:"healthy_count"`
			UnknownCount int `json:"unknown_count"`
		} `json:"summary"`
	}
	if err := json.Unmarshal(out.Bytes(), &body); err != nil {
		t.Fatalf("output must be JSON: %v\n%s", err, out.String())
	}
	if body.HistoryPath == "" {
		t.Fatalf("HistoryPath is empty")
	}
	if len(body.ProviderHealth) != 2 {
		t.Fatalf("provider health length = %d, want 2: %#v", len(body.ProviderHealth), body.ProviderHealth)
	}
	if body.ProviderHealth[0].ProviderName != "cheap" || body.ProviderHealth[0].UnauthorizedCount != 1 {
		t.Fatalf("cheap health = %#v, want unauthorized failure", body.ProviderHealth[0])
	}
	if body.ProviderHealth[0].Recommendation != "avoid" {
		t.Fatalf("cheap recommendation = %q, want avoid", body.ProviderHealth[0].Recommendation)
	}
	fast := body.ProviderHealth[1]
	if fast.ProviderName != "fast" || fast.AttemptCount != 3 || fast.PassCount != 2 || fast.FailCount != 1 {
		t.Fatalf("fast health = %#v, want summed provider health", fast)
	}
	if fast.RateLimitedCount != 1 || fast.EstimatedCostMicroUSD != 150 {
		t.Fatalf("fast health totals = %#v, want rate limit and cost 150", fast)
	}
	if fast.FailureRate != 0.333333 || fast.CostPerAttemptMicroUSD != 50 {
		t.Fatalf("fast derived rates = %#v, want failure rate 0.333333 and cost per attempt 50", fast)
	}
	if fast.Recommendation != "watch" {
		t.Fatalf("fast recommendation = %q, want watch", fast.Recommendation)
	}
	if body.Summary.AvoidCount != 1 || body.Summary.WatchCount != 1 || body.Summary.HealthyCount != 0 || body.Summary.UnknownCount != 0 {
		t.Fatalf("summary = %#v, want 1 avoid, 1 watch, 0 healthy, 0 unknown", body.Summary)
	}
}
