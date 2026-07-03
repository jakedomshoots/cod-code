package ceo

import (
	"strings"

	"ceoharness/internal/jobpacket"
	"ceoharness/internal/subagent"
)

func reusableStageResults(stage subagentStage, continuation *ContinuationContext) ([]scheduledSubagentResult, subagentStage) {
	results := []scheduledSubagentResult{}
	remaining := subagentStage{index: stage.index}
	for _, item := range stage.agents {
		result, ok := reusableSubagentResult(item.agent, continuation)
		if !ok {
			remaining.agents = append(remaining.agents, item)
			continue
		}
		result.Stage = stage.index
		result.Reused = true
		results = append(results, scheduledSubagentResult{
			index:  item.index,
			result: result,
		})
	}
	return results, remaining
}

func reusableSubagentResult(agent jobpacket.Subagent, continuation *ContinuationContext) (subagent.Result, bool) {
	if continuation == nil {
		return subagent.Result{}, false
	}
	for _, result := range continuation.ReusableResults {
		if CanReuseSubagentResult(agent, result) {
			return result, true
		}
	}
	return subagent.Result{}, false
}

// CanReuseSubagentResult reports whether a saved result can stand in for a planned subagent.
func CanReuseSubagentResult(agent jobpacket.Subagent, result subagent.Result) bool {
	return strings.TrimSpace(result.Status) == "pass" &&
		strings.TrimSpace(result.AgentName) == agent.Name &&
		strings.TrimSpace(result.Role) == agent.Role &&
		strings.TrimSpace(result.Assignment) == strings.TrimSpace(agent.Assignment) &&
		len(result.PatchProposals) == 0 &&
		sameActionSet(result.AllowedActions, jobpacket.ActionStrings(agent.AllowedActions))
}

func sameActionSet(left []string, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	seen := map[string]struct{}{}
	for _, value := range left {
		seen[strings.TrimSpace(value)] = struct{}{}
	}
	for _, value := range right {
		if _, ok := seen[strings.TrimSpace(value)]; !ok {
			return false
		}
	}
	return true
}

func reusedSubagentCount(results []subagent.Result) int {
	count := 0
	for _, result := range results {
		if result.Reused {
			count++
		}
	}
	return count
}

func reportContinuation(input *ContinuationContext, results []subagent.Result) *ContinuationContext {
	if input == nil || strings.TrimSpace(input.JobID) == "" {
		return nil
	}
	return &ContinuationContext{
		JobID:               strings.TrimSpace(input.JobID),
		ReusedSubagentCount: reusedSubagentCount(results),
	}
}

func shouldRunCEODelegation(continuation *ContinuationContext) bool {
	return continuation == nil || !continuation.UseSavedDelegation
}

func savedDelegationFromContinuation(continuation *ContinuationContext) *CEODelegation {
	if continuation == nil || !continuation.UseSavedDelegation || continuation.SavedDelegation == nil {
		return nil
	}
	saved := *continuation.SavedDelegation
	saved.Source = "history"
	saved.SelectedSubagents = append([]string(nil), continuation.SavedDelegation.SelectedSubagents...)
	saved.NewSubagents = append([]jobpacket.Subagent(nil), continuation.SavedDelegation.NewSubagents...)
	saved.Assignments = cloneDelegationAssignments(continuation.SavedDelegation.Assignments)
	saved.PromptBytes = 0
	return &saved
}

func cloneDelegationAssignments(assignments map[string]string) map[string]string {
	if len(assignments) == 0 {
		return nil
	}
	cloned := make(map[string]string, len(assignments))
	for key, value := range assignments {
		cloned[key] = value
	}
	return cloned
}
