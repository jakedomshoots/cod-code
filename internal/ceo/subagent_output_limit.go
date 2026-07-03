package ceo

import "ceoharness/internal/subagent"

const truncatedOutputMarker = "\n[truncated]"

func compactScheduledSubagentOutputs(results []scheduledSubagentResult, maxBytes int) []scheduledSubagentResult {
	if maxBytes <= 0 {
		return results
	}
	out := append([]scheduledSubagentResult(nil), results...)
	for index := range out {
		out[index].result = compactSubagentOutput(out[index].result, maxBytes)
	}
	return out
}

func compactSubagentOutput(result subagent.Result, maxBytes int) subagent.Result {
	if maxBytes <= 0 {
		return result
	}
	truncated := false
	result.Summary, truncated = compactOutputText(result.Summary, maxBytes)
	result.Evidence, truncated = compactOutputList(result.Evidence, maxBytes, truncated)
	result.Questions, truncated = compactOutputList(result.Questions, maxBytes, truncated)
	result.AttemptErrors, truncated = compactOutputList(result.AttemptErrors, maxBytes, truncated)
	result.OutputTruncated = result.OutputTruncated || truncated
	return result
}

func compactOutputList(values []string, maxBytes int, alreadyTruncated bool) ([]string, bool) {
	if len(values) == 0 {
		return values, alreadyTruncated
	}
	out := append([]string(nil), values...)
	truncated := alreadyTruncated
	for index, value := range out {
		next, wasTruncated := compactOutputText(value, maxBytes)
		out[index] = next
		truncated = truncated || wasTruncated
	}
	return out, truncated
}

func compactOutputText(text string, maxBytes int) (string, bool) {
	if maxBytes <= 0 || len(text) <= maxBytes {
		return text, false
	}
	return truncateUTF8Bytes(text, maxBytes) + truncatedOutputMarker, true
}

func truncateUTF8Bytes(text string, maxBytes int) string {
	if len(text) <= maxBytes {
		return text
	}
	cut := 0
	for index := range text {
		if index > maxBytes {
			break
		}
		cut = index
	}
	if cut == 0 {
		return text[:maxBytes]
	}
	return text[:cut]
}
