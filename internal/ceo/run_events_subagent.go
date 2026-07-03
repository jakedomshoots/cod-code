package ceo

import (
	"fmt"

	"ceoharness/internal/subagent"
)

func (b *runEventBuilder) addSubagentEvents(result subagent.Result) {
	b.add(RunEvent{
		Kind:                   "subagent",
		Status:                 result.Status,
		AgentName:              result.AgentName,
		Stage:                  result.Stage,
		Source:                 subagentEventSource(result),
		ProviderName:           result.ProviderName,
		ProviderFallbackFrom:   result.ProviderFallbackFrom,
		ProviderFallbackReason: result.ProviderFallbackReason,
		Summary:                initialSubagentSummary(result),
	})
	for _, request := range result.ToolRequests {
		b.add(RunEvent{
			Kind:      "tool_request",
			Status:    "requested",
			AgentName: result.AgentName,
			Stage:     result.Stage,
			Action:    request.Action,
			Path:      request.Path,
			Query:     request.Query,
		})
	}
	for _, toolResult := range result.ToolResults {
		b.add(RunEvent{
			Kind:      "tool_result",
			Status:    toolResult.Status,
			AgentName: result.AgentName,
			Stage:     result.Stage,
			Action:    toolResult.Action,
			Path:      toolResult.Path,
			Query:     toolResult.Query,
			Summary:   toolResultSummary(toolResult),
		})
	}
	if result.ToolFeedbackPasses > 0 {
		b.add(RunEvent{
			Kind:      "tool_feedback",
			Status:    result.Status,
			AgentName: result.AgentName,
			Stage:     result.Stage,
			Summary:   result.Summary,
		})
	}
}

func toolResultSummary(result subagent.ToolResult) string {
	if result.Error != "" {
		return result.Error
	}
	if result.Output != "" {
		return result.Output
	}
	if result.MatchCount > 0 {
		return fmt.Sprintf("%d match(es)", result.MatchCount)
	}
	return ""
}

func initialSubagentSummary(result subagent.Result) string {
	if result.InitialSummary != "" {
		return result.InitialSummary
	}
	return result.Summary
}

func subagentEventSource(result subagent.Result) string {
	if result.Reused {
		return "history"
	}
	return ""
}
