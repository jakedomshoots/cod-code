package history

import (
	"math"
	"sort"
	"strings"
)

type ProviderHealth struct {
	ProviderName           string  `json:"provider_name"`
	ModelSource            string  `json:"model_source,omitempty"`
	AttemptCount           int     `json:"attempt_count"`
	PassCount              int     `json:"pass_count"`
	FailCount              int     `json:"fail_count"`
	ErrorCount             int     `json:"error_count"`
	UnauthorizedCount      int     `json:"unauthorized_count"`
	RateLimitedCount       int     `json:"rate_limited_count"`
	UnavailableCount       int     `json:"unavailable_count"`
	EstimatedCostMicroUSD  int64   `json:"estimated_cost_microusd,omitempty"`
	FailureRate            float64 `json:"failure_rate"`
	CostPerAttemptMicroUSD int64   `json:"cost_per_attempt_microusd,omitempty"`
	Recommendation         string  `json:"recommendation"`
}

type ProviderHealthSummary struct {
	ProviderCount         int   `json:"provider_count"`
	AttemptCount          int   `json:"attempt_count"`
	PassCount             int   `json:"pass_count"`
	FailCount             int   `json:"fail_count"`
	ErrorCount            int   `json:"error_count"`
	EstimatedCostMicroUSD int64 `json:"estimated_cost_microusd"`
	AvoidCount            int   `json:"avoid_count"`
	WatchCount            int   `json:"watch_count"`
	HealthyCount          int   `json:"healthy_count"`
	UnknownCount          int   `json:"unknown_count"`
}

type ProviderHealthPolicy struct {
	AvoidFailureRate            float64
	WatchFailureRate            float64
	WatchCostPerAttemptMicroUSD int64
}

func AggregateProviderHealth(entries []Entry) []ProviderHealth {
	return AggregateProviderHealthWithPolicy(entries, ProviderHealthPolicy{})
}

func AggregateProviderHealthWithPolicy(entries []Entry, policy ProviderHealthPolicy) []ProviderHealth {
	rows := []ProviderHealth{}
	for _, entry := range entries {
		for _, health := range entry.ProviderHealth {
			providerName := strings.TrimSpace(health.ProviderName)
			if providerName == "" {
				continue
			}
			index := aggregateProviderIndex(rows, providerName)
			if index == -1 {
				rows = append(rows, ProviderHealth{
					ProviderName: providerName,
					ModelSource:  health.ModelSource,
				})
				index = len(rows) - 1
			}
			row := &rows[index]
			if row.ModelSource == "" {
				row.ModelSource = health.ModelSource
			}
			row.AttemptCount += health.AttemptCount
			row.PassCount += health.PassCount
			row.FailCount += health.FailCount
			row.ErrorCount += health.ErrorCount
			row.UnauthorizedCount += health.UnauthorizedCount
			row.RateLimitedCount += health.RateLimitedCount
			row.UnavailableCount += health.UnavailableCount
			row.EstimatedCostMicroUSD += health.EstimatedCostMicroUSD
		}
	}
	for index := range rows {
		applyProviderHealthRates(&rows[index], policy)
	}
	sortProviderHealthRows(rows)
	return rows
}

func SummarizeProviderHealth(rows []ProviderHealth) ProviderHealthSummary {
	summary := ProviderHealthSummary{}
	for _, row := range rows {
		summary.ProviderCount++
		summary.AttemptCount += row.AttemptCount
		summary.PassCount += row.PassCount
		summary.FailCount += row.FailCount
		summary.ErrorCount += row.ErrorCount
		summary.EstimatedCostMicroUSD += row.EstimatedCostMicroUSD
		switch row.Recommendation {
		case "avoid":
			summary.AvoidCount++
		case "watch":
			summary.WatchCount++
		case "healthy":
			summary.HealthyCount++
		default:
			summary.UnknownCount++
		}
	}
	return summary
}

func sortProviderHealthRows(rows []ProviderHealth) {
	sort.Slice(rows, func(left int, right int) bool {
		leftRow := rows[left]
		rightRow := rows[right]
		if leftRow.FailureRate != rightRow.FailureRate {
			return leftRow.FailureRate > rightRow.FailureRate
		}
		if leftRow.ErrorCount != rightRow.ErrorCount {
			return leftRow.ErrorCount > rightRow.ErrorCount
		}
		if leftRow.EstimatedCostMicroUSD != rightRow.EstimatedCostMicroUSD {
			return leftRow.EstimatedCostMicroUSD > rightRow.EstimatedCostMicroUSD
		}
		return leftRow.ProviderName < rightRow.ProviderName
	})
}

func aggregateProviderIndex(rows []ProviderHealth, providerName string) int {
	for index, row := range rows {
		if row.ProviderName == providerName {
			return index
		}
	}
	return -1
}

func applyProviderHealthRates(row *ProviderHealth, policy ProviderHealthPolicy) {
	if row.AttemptCount < 1 {
		row.FailureRate = 0
		row.CostPerAttemptMicroUSD = 0
		row.Recommendation = providerHealthRecommendation(*row, policy)
		return
	}
	row.FailureRate = roundProviderRate(float64(row.FailCount) / float64(row.AttemptCount))
	row.CostPerAttemptMicroUSD = row.EstimatedCostMicroUSD / int64(row.AttemptCount)
	row.Recommendation = providerHealthRecommendation(*row, policy)
}

func roundProviderRate(value float64) float64 {
	return math.Round(value*1_000_000) / 1_000_000
}

func providerHealthRecommendation(row ProviderHealth, policy ProviderHealthPolicy) string {
	if row.AttemptCount < 1 {
		return "unknown"
	}
	policy = policy.WithDefaults()
	if row.FailureRate >= policy.AvoidFailureRate {
		return "avoid"
	}
	if row.ErrorCount > 0 {
		return "watch"
	}
	if policy.WatchFailureRate > 0 && row.FailureRate >= policy.WatchFailureRate {
		return "watch"
	}
	if policy.WatchFailureRate == 0 && row.FailureRate > 0 {
		return "watch"
	}
	if policy.WatchCostPerAttemptMicroUSD > 0 && row.CostPerAttemptMicroUSD >= policy.WatchCostPerAttemptMicroUSD {
		return "watch"
	}
	return "healthy"
}

func (policy ProviderHealthPolicy) WithDefaults() ProviderHealthPolicy {
	if policy.AvoidFailureRate == 0 {
		policy.AvoidFailureRate = 0.5
	}
	return policy
}
