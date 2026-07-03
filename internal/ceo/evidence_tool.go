package ceo

import (
	"fmt"
	"strings"

	"ceoharness/internal/subagent"
)

const maxVerifyEvidenceOutputBytes = 1600

func runVerifyEvidenceTool(state toolRequestState, result subagent.ToolResult) subagent.ToolResult {
	output := renderPriorEvidence(state.PriorResults)
	if output == "" {
		result.Status = "skipped"
		result.Error = "prior evidence is not available"
		return result
	}
	result.Output, result.Truncated = compactOutputText(output, maxVerifyEvidenceOutputBytes)
	result.Status = "pass"
	result.Bytes = len(result.Output)
	return result
}

func renderPriorEvidence(results []subagent.Result) string {
	lines := []string{}
	for _, result := range results {
		evidence := normalizedEvidence(result.Evidence)
		if result.AgentName == "" || len(evidence) == 0 {
			continue
		}
		lines = append(lines, fmt.Sprintf("%s(%s): %s", result.AgentName, result.Status, strings.Join(evidence, "; ")))
	}
	return strings.Join(lines, "\n")
}

func normalizedEvidence(evidence []string) []string {
	normalized := make([]string, 0, len(evidence))
	for _, item := range evidence {
		clean := strings.Join(strings.Fields(item), " ")
		if clean == "" {
			continue
		}
		normalized = append(normalized, clean)
	}
	return normalized
}
