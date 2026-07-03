package ceo

import "ceoharness/internal/history"

func buildHistoryEntry(report Report) history.Entry {
	cliPatchCount, modelPatchCount := patchSourceCounts(report.PatchAudit)
	runLedger := report.RunLedger
	runLedger.ChangedFiles = append([]string(nil), report.RunLedger.ChangedFiles...)
	runLedger.ProviderRouteReasons = append([]string(nil), report.RunLedger.ProviderRouteReasons...)
	return history.Entry{
		Task:                          report.JobPacket.Task,
		TaskKind:                      report.JobPacket.TaskProfile.Kind,
		RiskLevel:                     report.JobPacket.TaskProfile.RiskLevel,
		RiskAreas:                     append([]string(nil), report.JobPacket.TaskProfile.RiskAreas...),
		Verdict:                       report.Verdict,
		LifecycleState:                string(report.LifecycleState),
		LifecycleEvents:               lifecycleHistoryEvents(report.LifecycleEvents),
		RunLedger:                     &runLedger,
		ChangedFiles:                  append([]string(nil), report.ChangedFiles...),
		ExecutionPlanStepCount:        len(report.ExecutionPlan.Steps),
		ExecutionPlanNextAction:       report.ExecutionPlan.NextAction,
		SubagentCount:                 len(report.SubagentResults),
		ReusedSubagentCount:           report.RunManifest.ReusedSubagentCount,
		SubagentAttemptCount:          report.VerificationSummary.SubagentAttemptCount,
		SubagentRetryCount:            report.VerificationSummary.SubagentRetryCount,
		SubagentRetriedCount:          report.VerificationSummary.SubagentRetriedCount,
		SubagentRetryExhaustedCount:   report.VerificationSummary.SubagentRetryExhaustedCount,
		SubagentNoProgressStopCount:   report.VerificationSummary.SubagentNoProgressStopCount,
		CheckCount:                    len(report.CheckResults),
		PatchCount:                    len(report.PatchResults),
		CLIPatchCount:                 cliPatchCount,
		ModelPatchCount:               modelPatchCount,
		CheckFixCount:                 checkFixCount(report),
		ProviderErrorCount:            report.VerificationSummary.ProviderErrorCount,
		ProviderUnauthorizedCount:     report.VerificationSummary.ProviderUnauthorizedCount,
		ProviderRateLimitedCount:      report.VerificationSummary.ProviderRateLimitedCount,
		ProviderUnavailableCount:      report.VerificationSummary.ProviderUnavailableCount,
		ProviderEstimatedCostMicroUSD: report.VerificationSummary.ProviderEstimatedCostMicroUSD,
		ProviderCostBudgetMicroUSD:    report.VerificationSummary.ProviderCostBudgetMicroUSD,
		ProviderCostOverBudget:        report.VerificationSummary.ProviderCostOverBudget,
		ProviderHealth:                providerHealthHistoryRows(report.VerificationSummary.ProviderHealth),
	}
}

func lifecycleHistoryEvents(events []LifecycleEvent) []history.LifecycleEvent {
	if len(events) == 0 {
		return nil
	}
	copied := make([]history.LifecycleEvent, 0, len(events))
	for _, event := range events {
		copied = append(copied, history.LifecycleEvent{
			Index:         event.Index,
			State:         string(event.State),
			PreviousState: string(event.PreviousState),
			Summary:       event.Summary,
		})
	}
	return copied
}

func patchSourceCounts(entries []PatchAuditEntry) (cliPatchCount int, modelPatchCount int) {
	for _, entry := range entries {
		switch entry.Source {
		case "cli":
			cliPatchCount++
		case "model":
			modelPatchCount++
		}
	}
	return cliPatchCount, modelPatchCount
}

func checkFixCount(report Report) int {
	count := 0
	for _, result := range report.SubagentResults {
		if result.AgentName == "coder" && result.Role == checkFixRole {
			count++
		}
	}
	return count
}

func providerHealthHistoryRows(rows []ProviderHealth) []history.ProviderHealth {
	if len(rows) == 0 {
		return nil
	}
	copied := make([]history.ProviderHealth, 0, len(rows))
	for _, row := range rows {
		copied = append(copied, history.ProviderHealth{
			ProviderName:           row.ProviderName,
			ModelSource:            row.ModelSource,
			AttemptCount:           row.AttemptCount,
			PassCount:              row.PassCount,
			FailCount:              row.FailCount,
			ErrorCount:             row.ErrorCount,
			UnauthorizedCount:      row.UnauthorizedCount,
			RateLimitedCount:       row.RateLimitedCount,
			UnavailableCount:       row.UnavailableCount,
			EstimatedCostMicroUSD:  row.EstimatedCostMicroUSD,
			FailureRate:            row.FailureRate,
			CostPerAttemptMicroUSD: row.CostPerAttemptMicroUSD,
			Recommendation:         row.Recommendation,
		})
	}
	return copied
}
