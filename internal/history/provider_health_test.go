package history

import "testing"

func Test_AggregateProviderHealth_sums_provider_rows_by_name(t *testing.T) {
	// Given
	entries := []Entry{
		{
			Task: "first",
			ProviderHealth: []ProviderHealth{
				{ProviderName: "fast", ModelSource: "http", AttemptCount: 1, PassCount: 1, EstimatedCostMicroUSD: 106},
			},
		},
		{
			Task: "second",
			ProviderHealth: []ProviderHealth{
				{ProviderName: "cheap", ModelSource: "http", AttemptCount: 1, FailCount: 1, ErrorCount: 1, UnauthorizedCount: 1},
				{ProviderName: "fast", ModelSource: "http", AttemptCount: 2, PassCount: 1, FailCount: 1, ErrorCount: 1, RateLimitedCount: 1, EstimatedCostMicroUSD: 44},
			},
		},
		{Task: "legacy"},
	}

	// When
	health := AggregateProviderHealth(entries)

	// Then
	if len(health) != 2 {
		t.Fatalf("provider health length = %d, want 2: %#v", len(health), health)
	}
	if health[0].ProviderName != "cheap" || health[0].FailCount != 1 || health[0].UnauthorizedCount != 1 {
		t.Fatalf("cheap health = %#v, want one unauthorized failure", health[0])
	}
	if health[1].ProviderName != "fast" || health[1].AttemptCount != 3 || health[1].PassCount != 2 || health[1].FailCount != 1 {
		t.Fatalf("fast health counts = %#v, want summed attempts", health[1])
	}
	if health[1].RateLimitedCount != 1 || health[1].EstimatedCostMicroUSD != 150 {
		t.Fatalf("fast health totals = %#v, want rate limit and cost 150", health[1])
	}
	if health[1].FailureRate != 0.333333 || health[1].CostPerAttemptMicroUSD != 50 {
		t.Fatalf("fast derived rates = %#v, want failure rate 0.333333 and cost per attempt 50", health[1])
	}
}

func Test_AggregateProviderHealth_sorts_worst_provider_first(t *testing.T) {
	// Given
	entries := []Entry{
		{
			Task: "health",
			ProviderHealth: []ProviderHealth{
				{ProviderName: "alpha", ModelSource: "http", AttemptCount: 3, PassCount: 2, FailCount: 1, ErrorCount: 1, EstimatedCostMicroUSD: 300},
				{ProviderName: "zeta", ModelSource: "http", AttemptCount: 1, FailCount: 1, ErrorCount: 1},
				{ProviderName: "beta", ModelSource: "http", AttemptCount: 1, PassCount: 1, EstimatedCostMicroUSD: 999},
			},
		},
	}

	// When
	health := AggregateProviderHealth(entries)

	// Then
	if len(health) != 3 {
		t.Fatalf("provider health length = %d, want 3: %#v", len(health), health)
	}
	if health[0].ProviderName != "zeta" || health[1].ProviderName != "alpha" || health[2].ProviderName != "beta" {
		t.Fatalf("provider order = %q, %q, %q; want zeta, alpha, beta", health[0].ProviderName, health[1].ProviderName, health[2].ProviderName)
	}
	if health[0].FailureRate != 1 || health[1].FailureRate != 0.333333 || health[2].FailureRate != 0 {
		t.Fatalf("failure rates = %f, %f, %f; want 1, 0.333333, 0", health[0].FailureRate, health[1].FailureRate, health[2].FailureRate)
	}
}

func Test_AggregateProviderHealth_assigns_recommendation_labels(t *testing.T) {
	// Given
	entries := []Entry{
		{
			Task: "recommendations",
			ProviderHealth: []ProviderHealth{
				{ProviderName: "zeta", AttemptCount: 2, PassCount: 1, FailCount: 1, ErrorCount: 1},
				{ProviderName: "alpha", AttemptCount: 3, PassCount: 2, FailCount: 1, ErrorCount: 1},
				{ProviderName: "beta", AttemptCount: 1, PassCount: 1},
			},
		},
	}

	// When
	health := AggregateProviderHealth(entries)

	// Then
	if len(health) != 3 {
		t.Fatalf("provider health length = %d, want 3: %#v", len(health), health)
	}
	if health[0].Recommendation != "avoid" {
		t.Fatalf("zeta recommendation = %q, want avoid", health[0].Recommendation)
	}
	if health[1].Recommendation != "watch" {
		t.Fatalf("alpha recommendation = %q, want watch", health[1].Recommendation)
	}
	if health[2].Recommendation != "healthy" {
		t.Fatalf("beta recommendation = %q, want healthy", health[2].Recommendation)
	}
}

func Test_SummarizeProviderHealth_counts_recommendation_labels(t *testing.T) {
	// Given
	rows := []ProviderHealth{
		{ProviderName: "avoid", Recommendation: "avoid", AttemptCount: 2, PassCount: 1, FailCount: 1, ErrorCount: 1, EstimatedCostMicroUSD: 100},
		{ProviderName: "watch", Recommendation: "watch", AttemptCount: 3, PassCount: 2, FailCount: 1, ErrorCount: 1, EstimatedCostMicroUSD: 150},
		{ProviderName: "healthy", Recommendation: "healthy", AttemptCount: 1, PassCount: 1, EstimatedCostMicroUSD: 50},
		{ProviderName: "unknown", Recommendation: "unknown"},
		{ProviderName: "missing", AttemptCount: 2, FailCount: 2, ErrorCount: 2, EstimatedCostMicroUSD: 20},
	}

	// When
	summary := SummarizeProviderHealth(rows)

	// Then
	if summary.AvoidCount != 1 {
		t.Fatalf("AvoidCount = %d, want 1", summary.AvoidCount)
	}
	if summary.WatchCount != 1 {
		t.Fatalf("WatchCount = %d, want 1", summary.WatchCount)
	}
	if summary.HealthyCount != 1 {
		t.Fatalf("HealthyCount = %d, want 1", summary.HealthyCount)
	}
	if summary.UnknownCount != 2 {
		t.Fatalf("UnknownCount = %d, want 2", summary.UnknownCount)
	}
	if summary.ProviderCount != 5 ||
		summary.AttemptCount != 8 ||
		summary.PassCount != 4 ||
		summary.FailCount != 4 ||
		summary.ErrorCount != 4 ||
		summary.EstimatedCostMicroUSD != 320 {
		t.Fatalf("summary totals = %#v, want provider health totals", summary)
	}
}
