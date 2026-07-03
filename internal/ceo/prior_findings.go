package ceo

import (
	"fmt"
	"sort"
	"strings"

	"ceoharness/internal/subagent"
)

const maxPriorFindingBytes = 180

func appendPriorStageResults(prior []subagent.Result, stageResults []scheduledSubagentResult) []subagent.Result {
	ordered := append([]scheduledSubagentResult(nil), stageResults...)
	sort.Slice(ordered, func(i int, j int) bool {
		return ordered[i].index < ordered[j].index
	})
	for _, result := range ordered {
		prior = append(prior, result.result)
	}
	return prior
}

func hasNeedsInputResult(results []scheduledSubagentResult) bool {
	for _, result := range results {
		if result.result.Status == "needs_input" {
			return true
		}
	}
	return false
}

func compactSubagentResults(results []subagent.Result) []subagent.Result {
	compacted := make([]subagent.Result, 0, len(results))
	for _, result := range results {
		if result.AgentName == "" {
			continue
		}
		compacted = append(compacted, result)
	}
	return compacted
}

func renderPriorFindings(results []subagent.Result) string {
	if len(results) == 0 {
		return ""
	}
	lines := make([]string, 0, len(results))
	for _, result := range results {
		lines = append(lines, fmt.Sprintf(
			"- %s(%s): %s",
			result.AgentName,
			result.Status,
			trimPriorFinding(result.Summary),
		))
	}
	return strings.Join(lines, "\n")
}

func trimPriorFinding(summary string) string {
	clean := strings.Join(strings.Fields(summary), " ")
	if len(clean) <= maxPriorFindingBytes {
		return clean
	}
	return clean[:maxPriorFindingBytes] + "..."
}
