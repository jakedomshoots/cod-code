package ceo

import (
	"testing"

	"ceoharness/internal/history"
	"ceoharness/internal/subagent"
)

func Test_summarizeVerification_assigns_provider_health_recommendations(t *testing.T) {
	// Given
	results := []subagent.Result{
		{
			Status:       "fail",
			ModelSource:  "command",
			ProviderName: "zeta",
			AttemptRecords: []subagent.AttemptRecord{
				{Status: "fail", Error: "provider failed"},
				{Status: "pass"},
			},
		},
		{
			Status:       "pass",
			ModelSource:  "command",
			ProviderName: "alpha",
			AttemptRecords: []subagent.AttemptRecord{
				{Status: "fail", Error: "provider failed"},
				{Status: "pass"},
				{Status: "pass"},
			},
		},
		{
			Status:       "pass",
			ModelSource:  "command",
			ProviderName: "beta",
		},
	}

	// When
	summary := summarizeVerification(results, nil)

	// Then
	if len(summary.ProviderHealth) != 3 {
		t.Fatalf("ProviderHealth length = %d, want 3: %#v", len(summary.ProviderHealth), summary.ProviderHealth)
	}
	if summary.ProviderHealth[0].Recommendation != "avoid" {
		t.Fatalf("zeta recommendation = %q, want avoid", summary.ProviderHealth[0].Recommendation)
	}
	if summary.ProviderHealth[1].Recommendation != "watch" {
		t.Fatalf("alpha recommendation = %q, want watch", summary.ProviderHealth[1].Recommendation)
	}
	if summary.ProviderHealth[2].Recommendation != "healthy" {
		t.Fatalf("beta recommendation = %q, want healthy", summary.ProviderHealth[2].Recommendation)
	}
}

func Test_summarizeVerificationWithPolicy_uses_provider_health_thresholds(t *testing.T) {
	// Given
	results := []subagent.Result{
		{
			Status:       "pass",
			ModelSource:  "command",
			ProviderName: "zeta",
			AttemptRecords: []subagent.AttemptRecord{
				{Status: "fail"},
				{Status: "pass"},
			},
		},
		{
			Status:       "pass",
			ModelSource:  "command",
			ProviderName: "alpha",
			AttemptRecords: []subagent.AttemptRecord{
				{Status: "fail"},
				{Status: "pass"},
				{Status: "pass"},
			},
		},
	}
	policy := history.ProviderHealthPolicy{
		AvoidFailureRate: 0.9,
		WatchFailureRate: 0.5,
	}

	// When
	summary := summarizeVerificationWithPolicy(results, nil, policy)

	// Then
	if len(summary.ProviderHealth) != 2 {
		t.Fatalf("ProviderHealth length = %d, want 2: %#v", len(summary.ProviderHealth), summary.ProviderHealth)
	}
	if summary.ProviderHealth[0].ProviderName != "zeta" || summary.ProviderHealth[0].Recommendation != "watch" {
		t.Fatalf("zeta health = %#v, want watch under configured policy", summary.ProviderHealth[0])
	}
	if summary.ProviderHealth[1].ProviderName != "alpha" || summary.ProviderHealth[1].Recommendation != "healthy" {
		t.Fatalf("alpha health = %#v, want healthy under configured policy", summary.ProviderHealth[1])
	}
}

func Test_summarizeVerificationWithPolicy_marks_costly_provider_as_watch(t *testing.T) {
	// Given
	results := []subagent.Result{
		{
			Status:                        "pass",
			ModelSource:                   "command",
			ProviderName:                  "premium",
			ProviderEstimatedCostMicroUSD: 100,
		},
		{
			Status:                        "pass",
			ModelSource:                   "command",
			ProviderName:                  "cheap",
			ProviderEstimatedCostMicroUSD: 10,
		},
	}
	policy := history.ProviderHealthPolicy{
		WatchCostPerAttemptMicroUSD: 50,
	}

	// When
	summary := summarizeVerificationWithPolicy(results, nil, policy)

	// Then
	if len(summary.ProviderHealth) != 2 {
		t.Fatalf("ProviderHealth length = %d, want 2: %#v", len(summary.ProviderHealth), summary.ProviderHealth)
	}
	if summary.ProviderHealth[0].ProviderName != "premium" || summary.ProviderHealth[0].Recommendation != "watch" {
		t.Fatalf("premium health = %#v, want watch under cost policy", summary.ProviderHealth[0])
	}
	if summary.ProviderHealth[1].ProviderName != "cheap" || summary.ProviderHealth[1].Recommendation != "healthy" {
		t.Fatalf("cheap health = %#v, want healthy under cost policy", summary.ProviderHealth[1])
	}
}
