package ceo

import (
	"ceoharness/internal/jobpacket"
	"ceoharness/internal/subagent"
)

func failedSubagentResult(packet jobpacket.Packet, agent jobpacket.Subagent, attempts int, durationMS int64, attemptErrors []string, attemptRecords []subagent.AttemptRecord, providerError providerErrorFields) subagent.Result {
	contextBytes := len(packet.Task)
	if agent.Assignment != "" {
		contextBytes += len(agent.Assignment)
	}
	maxContextBytes := contextBudgetForAgent(packet, agent)
	contextTruncated := false
	if maxContextBytes > 0 && contextBytes > maxContextBytes {
		contextBytes = maxContextBytes
		contextTruncated = true
	}
	return subagent.Result{
		AgentName:            agent.Name,
		Role:                 agent.Role,
		Assignment:           agent.Assignment,
		Stage:                stageForAgent(agent),
		AllowedActions:       jobpacket.ActionStrings(agent.AllowedActions),
		Status:               "fail",
		Attempts:             attempts,
		DurationMS:           durationMS,
		ModelSource:          providerError.modelSource,
		ProviderName:         providerError.providerName,
		ProviderErrorKind:    providerError.kind,
		ProviderHTTPStatus:   providerError.httpStatus,
		ProviderRetryAfterMS: providerError.retryAfterMS,
		ContextReceived:      packet.ContextPolicy.Mode,
		ContextBytes:         contextBytes,
		ContextTruncated:     contextTruncated,
		Summary:              "subagent retries exhausted",
		AttemptErrors:        append([]string(nil), attemptErrors...),
		AttemptRecords:       append([]subagent.AttemptRecord(nil), attemptRecords...),
		Evidence:             []string{"subagent retries exhausted"},
	}
}
