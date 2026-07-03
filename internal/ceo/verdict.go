package ceo

import (
	"ceoharness/internal/checkrunner"
	"ceoharness/internal/subagent"
)

func verdict(results []subagent.Result, checks []checkrunner.Result, summary VerificationSummary) string {
	if summary.ProviderCostOverBudget {
		return "fail"
	}
	for _, result := range results {
		if result.Status == "needs_input" {
			return "needs_input"
		}
		if result.Status != "pass" {
			return "fail"
		}
	}
	if len(checks) > 0 && checks[len(checks)-1].Status != "pass" {
		return "fail"
	}
	return "pass"
}
