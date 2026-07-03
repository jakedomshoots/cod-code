package ceo

import (
	"testing"

	"ceoharness/internal/subagent"
)

func Test_summarizeVerification_totals_provider_estimated_cost(t *testing.T) {
	// Given
	results := []subagent.Result{
		{Status: "pass", ProviderEstimatedCostMicroUSD: 106},
		{Status: "pass", ProviderEstimatedCostMicroUSD: 44},
		{Status: "pass"},
	}

	// When
	summary := summarizeVerification(results, nil)

	// Then
	if summary.ProviderEstimatedCostMicroUSD != 150 {
		t.Fatalf("ProviderEstimatedCostMicroUSD = %d, want 150", summary.ProviderEstimatedCostMicroUSD)
	}
}

func Test_summarizeVerification_groups_provider_health_by_name(t *testing.T) {
	// Given
	results := []subagent.Result{
		{
			Status:                        "pass",
			ModelSource:                   "http",
			ProviderName:                  "fast",
			ProviderEstimatedCostMicroUSD: 106,
			AttemptRecords: []subagent.AttemptRecord{
				{Status: "fail", ProviderErrorKind: "rate_limited", ProviderHTTPStatus: 429},
				{Status: "pass"},
			},
		},
		{
			Status:                        "pass",
			ModelSource:                   "http",
			ProviderName:                  "fast",
			ProviderEstimatedCostMicroUSD: 44,
		},
		{
			Status:             "fail",
			ModelSource:        "http",
			ProviderName:       "cheap",
			ProviderErrorKind:  "unauthorized",
			ProviderHTTPStatus: 401,
		},
		{Status: "pass", ModelSource: "local"},
	}

	// When
	summary := summarizeVerification(results, nil)

	// Then
	if len(summary.ProviderHealth) != 2 {
		t.Fatalf("ProviderHealth length = %d, want 2: %#v", len(summary.ProviderHealth), summary.ProviderHealth)
	}
	cheap := summary.ProviderHealth[0]
	if cheap.ProviderName != "cheap" || cheap.ModelSource != "http" {
		t.Fatalf("first provider health = %#v, want cheap http", cheap)
	}
	if cheap.AttemptCount != 1 || cheap.FailCount != 1 || cheap.UnauthorizedCount != 1 {
		t.Fatalf("cheap health = %#v, want one unauthorized failure", cheap)
	}
	fast := summary.ProviderHealth[1]
	if fast.ProviderName != "fast" || fast.ModelSource != "http" {
		t.Fatalf("second provider health = %#v, want fast http", fast)
	}
	if fast.AttemptCount != 3 || fast.PassCount != 2 || fast.FailCount != 1 {
		t.Fatalf("fast attempts = %#v, want 3 attempts with 2 pass and 1 fail", fast)
	}
	if fast.RateLimitedCount != 1 || fast.EstimatedCostMicroUSD != 150 {
		t.Fatalf("fast health = %#v, want one rate limit and cost 150", fast)
	}
	if fast.FailureRate != 0.333333 || fast.CostPerAttemptMicroUSD != 50 {
		t.Fatalf("fast derived rates = %#v, want failure rate 0.333333 and cost per attempt 50", fast)
	}
}

func Test_summarizeVerification_sorts_provider_health_by_worst_first(t *testing.T) {
	// Given
	results := []subagent.Result{
		{
			Status:                        "pass",
			ModelSource:                   "http",
			ProviderName:                  "alpha",
			ProviderEstimatedCostMicroUSD: 300,
			AttemptRecords: []subagent.AttemptRecord{
				{Status: "fail", ProviderErrorKind: "rate_limited", ProviderHTTPStatus: 429},
				{Status: "pass"},
				{Status: "pass"},
			},
		},
		{
			Status:             "fail",
			ModelSource:        "http",
			ProviderName:       "zeta",
			ProviderErrorKind:  "unauthorized",
			ProviderHTTPStatus: 401,
		},
		{
			Status:                        "pass",
			ModelSource:                   "http",
			ProviderName:                  "beta",
			ProviderEstimatedCostMicroUSD: 999,
		},
	}

	// When
	summary := summarizeVerification(results, nil)

	// Then
	if len(summary.ProviderHealth) != 3 {
		t.Fatalf("ProviderHealth length = %d, want 3: %#v", len(summary.ProviderHealth), summary.ProviderHealth)
	}
	if summary.ProviderHealth[0].ProviderName != "zeta" || summary.ProviderHealth[1].ProviderName != "alpha" || summary.ProviderHealth[2].ProviderName != "beta" {
		t.Fatalf("provider order = %q, %q, %q; want zeta, alpha, beta", summary.ProviderHealth[0].ProviderName, summary.ProviderHealth[1].ProviderName, summary.ProviderHealth[2].ProviderName)
	}
	if summary.ProviderHealth[0].FailureRate != 1 || summary.ProviderHealth[1].FailureRate != 0.333333 || summary.ProviderHealth[2].FailureRate != 0 {
		t.Fatalf("failure rates = %f, %f, %f; want 1, 0.333333, 0", summary.ProviderHealth[0].FailureRate, summary.ProviderHealth[1].FailureRate, summary.ProviderHealth[2].FailureRate)
	}
}
