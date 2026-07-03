package ceo

import (
	"math"
	"sort"
	"strings"

	"ceoharness/internal/history"
	"ceoharness/internal/model"
	"ceoharness/internal/subagent"
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

func countProviderErrors(summary *VerificationSummary, result subagent.Result) {
	if len(result.AttemptRecords) > 0 {
		for _, record := range result.AttemptRecords {
			addProviderError(summary, record.ProviderErrorKind, record.ProviderHTTPStatus)
		}
		return
	}
	addProviderError(summary, result.ProviderErrorKind, result.ProviderHTTPStatus)
}

func addProviderHealth(summary *VerificationSummary, result subagent.Result) {
	providerName := strings.TrimSpace(result.ProviderName)
	if providerName == "" {
		return
	}
	index := providerHealthIndex(summary.ProviderHealth, providerName)
	if index == -1 {
		summary.ProviderHealth = append(summary.ProviderHealth, ProviderHealth{
			ProviderName: providerName,
			ModelSource:  result.ModelSource,
		})
		index = len(summary.ProviderHealth) - 1
	}
	health := &summary.ProviderHealth[index]
	if health.ModelSource == "" {
		health.ModelSource = result.ModelSource
	}
	health.EstimatedCostMicroUSD += result.ProviderEstimatedCostMicroUSD
	if len(result.AttemptRecords) > 0 {
		for _, record := range result.AttemptRecords {
			addProviderAttempt(health, record.Status, record.ProviderErrorKind, record.ProviderHTTPStatus)
		}
		return
	}
	addProviderAttempt(health, result.Status, result.ProviderErrorKind, result.ProviderHTTPStatus)
}

func addProviderError(summary *VerificationSummary, kind string, httpStatus int) {
	if kind == "" && httpStatus == 0 {
		return
	}
	summary.ProviderErrorCount++
	if kind != "" {
		if summary.ProviderErrorKindCounts == nil {
			summary.ProviderErrorKindCounts = map[string]int{}
		}
		summary.ProviderErrorKindCounts[kind]++
	}
	switch kind {
	case string(model.HTTPErrorKindUnauthorized):
		summary.ProviderUnauthorizedCount++
	case string(model.HTTPErrorKindRateLimited):
		summary.ProviderRateLimitedCount++
	case string(model.HTTPErrorKindUnavailable):
		summary.ProviderUnavailableCount++
	}
}

func addProviderAttempt(health *ProviderHealth, status string, kind string, httpStatus int) {
	health.AttemptCount++
	switch status {
	case "pass":
		health.PassCount++
	case "fail":
		health.FailCount++
	default:
		if kind != "" || httpStatus != 0 {
			health.FailCount++
		}
	}
	addProviderHealthError(health, kind, httpStatus)
}

func addProviderHealthError(health *ProviderHealth, kind string, httpStatus int) {
	if kind == "" && httpStatus == 0 {
		return
	}
	health.ErrorCount++
	switch kind {
	case string(model.HTTPErrorKindUnauthorized):
		health.UnauthorizedCount++
	case string(model.HTTPErrorKindRateLimited):
		health.RateLimitedCount++
	case string(model.HTTPErrorKindUnavailable):
		health.UnavailableCount++
	}
}

func providerHealthIndex(rows []ProviderHealth, providerName string) int {
	for index, row := range rows {
		if row.ProviderName == providerName {
			return index
		}
	}
	return -1
}

func sortProviderHealth(summary *VerificationSummary, policy history.ProviderHealthPolicy) {
	for index := range summary.ProviderHealth {
		applyProviderHealthRates(&summary.ProviderHealth[index], policy)
	}
	sort.Slice(summary.ProviderHealth, func(left int, right int) bool {
		leftRow := summary.ProviderHealth[left]
		rightRow := summary.ProviderHealth[right]
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

func applyProviderHealthRates(health *ProviderHealth, policy history.ProviderHealthPolicy) {
	if health.AttemptCount < 1 {
		health.FailureRate = 0
		health.CostPerAttemptMicroUSD = 0
		health.Recommendation = providerHealthRecommendation(*health, policy)
		return
	}
	health.FailureRate = roundProviderRate(float64(health.FailCount) / float64(health.AttemptCount))
	health.CostPerAttemptMicroUSD = health.EstimatedCostMicroUSD / int64(health.AttemptCount)
	health.Recommendation = providerHealthRecommendation(*health, policy)
}

func roundProviderRate(value float64) float64 {
	return math.Round(value*1_000_000) / 1_000_000
}

func providerHealthRecommendation(health ProviderHealth, policy history.ProviderHealthPolicy) string {
	if health.AttemptCount < 1 {
		return "unknown"
	}
	policy = policy.WithDefaults()
	if health.FailureRate >= policy.AvoidFailureRate {
		return "avoid"
	}
	if health.ErrorCount > 0 {
		return "watch"
	}
	if policy.WatchFailureRate > 0 && health.FailureRate >= policy.WatchFailureRate {
		return "watch"
	}
	if policy.WatchFailureRate == 0 && health.FailureRate > 0 {
		return "watch"
	}
	if policy.WatchCostPerAttemptMicroUSD > 0 && health.CostPerAttemptMicroUSD >= policy.WatchCostPerAttemptMicroUSD {
		return "watch"
	}
	return "healthy"
}
