package cli

import (
	"strings"

	"ceoharness/internal/checkrunner"
)

func compactFailedChecks(results []checkrunner.Result) []compactCheckResult {
	failed := []compactCheckResult{}
	for _, result := range results {
		if result.Status == "pass" {
			continue
		}
		failed = append(failed, compactCheckResult{
			Command:        append([]string(nil), result.Argv...),
			Status:         result.Status,
			ExitCode:       result.ExitCode,
			CheckIndex:     result.CheckIndex,
			Attempt:        result.Attempt,
			MaxAttempts:    result.MaxAttempts,
			DurationMS:     result.DurationMS,
			FailureExcerpt: compactFailureExcerpt(result),
		})
	}
	return failed
}

func compactFailureExcerpt(result checkrunner.Result) string {
	text := strings.TrimSpace(result.Stderr)
	if text == "" {
		text = strings.TrimSpace(result.Stdout)
	}
	if len(text) <= failureExcerptLimit {
		return text
	}
	return text[:failureExcerptLimit] + "..."
}
