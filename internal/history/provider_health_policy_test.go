package history

import "testing"

func Test_AggregateProviderHealthWithPolicy_uses_configured_thresholds(t *testing.T) {
	// Given
	entries := []Entry{
		{
			Task: "policy",
			ProviderHealth: []ProviderHealth{
				{ProviderName: "zeta", AttemptCount: 2, PassCount: 1, FailCount: 1},
				{ProviderName: "alpha", AttemptCount: 10, PassCount: 9, FailCount: 1},
			},
		},
	}
	policy := ProviderHealthPolicy{
		AvoidFailureRate: 0.9,
		WatchFailureRate: 0.5,
	}

	// When
	health := AggregateProviderHealthWithPolicy(entries, policy)

	// Then
	if len(health) != 2 {
		t.Fatalf("provider health length = %d, want 2: %#v", len(health), health)
	}
	if health[0].ProviderName != "zeta" || health[0].Recommendation != "watch" {
		t.Fatalf("zeta health = %#v, want watch under configured policy", health[0])
	}
	if health[1].ProviderName != "alpha" || health[1].Recommendation != "healthy" {
		t.Fatalf("alpha health = %#v, want healthy under configured policy", health[1])
	}
}

func Test_AggregateProviderHealthWithPolicy_marks_costly_provider_as_watch(t *testing.T) {
	// Given
	entries := []Entry{
		{
			Task: "cost policy",
			ProviderHealth: []ProviderHealth{
				{ProviderName: "premium", AttemptCount: 2, PassCount: 2, EstimatedCostMicroUSD: 200},
				{ProviderName: "cheap", AttemptCount: 2, PassCount: 2, EstimatedCostMicroUSD: 20},
			},
		},
	}
	policy := ProviderHealthPolicy{
		WatchCostPerAttemptMicroUSD: 50,
	}

	// When
	health := AggregateProviderHealthWithPolicy(entries, policy)

	// Then
	if len(health) != 2 {
		t.Fatalf("provider health length = %d, want 2: %#v", len(health), health)
	}
	if health[0].ProviderName != "premium" || health[0].Recommendation != "watch" {
		t.Fatalf("premium health = %#v, want watch under cost policy", health[0])
	}
	if health[1].ProviderName != "cheap" || health[1].Recommendation != "healthy" {
		t.Fatalf("cheap health = %#v, want healthy under cost policy", health[1])
	}
}
