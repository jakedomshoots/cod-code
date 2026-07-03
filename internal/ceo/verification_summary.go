package ceo

import (
	"ceoharness/internal/checkrunner"
	"ceoharness/internal/history"
	"ceoharness/internal/subagent"
)

type VerificationSummary struct {
	SubagentPassCount             int              `json:"subagent_pass_count"`
	SubagentFailCount             int              `json:"subagent_fail_count"`
	SubagentAttemptCount          int              `json:"subagent_attempt_count"`
	SubagentRetryCount            int              `json:"subagent_retry_count"`
	SubagentRetriedCount          int              `json:"subagent_retried_count"`
	SubagentRetryExhaustedCount   int              `json:"subagent_retry_exhausted_count"`
	SubagentNoProgressStopCount   int              `json:"subagent_no_progress_stop_count,omitempty"`
	CheckAttemptCount             int              `json:"check_attempt_count"`
	CheckPassCount                int              `json:"check_pass_count"`
	CheckFailCount                int              `json:"check_fail_count"`
	CheckTotalDurationMS          int64            `json:"check_total_duration_ms"`
	ProviderErrorCount            int              `json:"provider_error_count"`
	ProviderUnauthorizedCount     int              `json:"provider_unauthorized_count"`
	ProviderRateLimitedCount      int              `json:"provider_rate_limited_count"`
	ProviderUnavailableCount      int              `json:"provider_unavailable_count"`
	ProviderErrorKindCounts       map[string]int   `json:"provider_error_kind_counts,omitempty"`
	ProviderEstimatedCostMicroUSD int64            `json:"provider_estimated_cost_microusd"`
	ProviderCostBudgetMicroUSD    int64            `json:"provider_cost_budget_microusd,omitempty"`
	ProviderCostOverBudget        bool             `json:"provider_cost_over_budget,omitempty"`
	ProviderHealth                []ProviderHealth `json:"provider_health,omitempty"`
}

func summarizeVerification(results []subagent.Result, checks []checkrunner.Result) VerificationSummary {
	return summarizeVerificationWithPolicy(results, checks, history.ProviderHealthPolicy{})
}

func summarizeVerificationWithPolicy(results []subagent.Result, checks []checkrunner.Result, policy history.ProviderHealthPolicy) VerificationSummary {
	summary := VerificationSummary{
		CheckAttemptCount: len(checks),
	}
	for _, result := range results {
		if result.Status == "pass" {
			summary.SubagentPassCount++
		} else {
			summary.SubagentFailCount++
		}
		if result.NoProgressStopped {
			summary.SubagentNoProgressStopCount++
		}
		countSubagentRetryPressure(&summary, result)
		summary.ProviderEstimatedCostMicroUSD += result.ProviderEstimatedCostMicroUSD
		countProviderErrors(&summary, result)
		addProviderHealth(&summary, result)
	}
	sortProviderHealth(&summary, policy)
	for _, check := range checks {
		if check.Status == "pass" {
			summary.CheckPassCount++
		} else {
			summary.CheckFailCount++
		}
		summary.CheckTotalDurationMS += check.DurationMS
	}
	return summary
}

func countSubagentRetryPressure(summary *VerificationSummary, result subagent.Result) {
	attempts := result.Attempts
	if len(result.AttemptRecords) > attempts {
		attempts = len(result.AttemptRecords)
	}
	if attempts < 1 {
		attempts = 1
	}
	summary.SubagentAttemptCount += attempts
	retries := attempts - 1
	if retries < 1 {
		return
	}
	summary.SubagentRetryCount += retries
	summary.SubagentRetriedCount++
	if result.Status != "pass" {
		summary.SubagentRetryExhaustedCount++
	}
}

func applyProviderCostBudget(summary VerificationSummary, budgetMicroUSD int64) VerificationSummary {
	if budgetMicroUSD <= 0 {
		return summary
	}
	summary.ProviderCostBudgetMicroUSD = budgetMicroUSD
	summary.ProviderCostOverBudget = summary.ProviderEstimatedCostMicroUSD > budgetMicroUSD
	return summary
}
