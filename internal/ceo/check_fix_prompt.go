package ceo

import (
	"strconv"
	"strings"

	"ceoharness/internal/checkrunner"
)

func renderCheckFixMetadata(check checkrunner.Result) string {
	var builder strings.Builder
	builder.WriteString("Status: ")
	builder.WriteString(check.Status)
	if check.CheckIndex > 0 {
		builder.WriteString("\nCheck index: ")
		builder.WriteString(strconv.Itoa(check.CheckIndex))
	}
	if check.Attempt > 0 {
		builder.WriteString("\nCheck attempt: ")
		builder.WriteString(strconv.Itoa(check.Attempt))
		if check.MaxAttempts > 0 {
			builder.WriteString("/")
			builder.WriteString(strconv.Itoa(check.MaxAttempts))
		}
	}
	if check.DurationMS > 0 {
		builder.WriteString("\nDuration ms: ")
		builder.WriteString(strconv.FormatInt(check.DurationMS, 10))
	}
	builder.WriteString("\nCommand: ")
	builder.WriteString(strings.Join(check.Argv, " "))
	builder.WriteString("\nExit code: ")
	builder.WriteString(strconv.Itoa(check.ExitCode))
	return builder.String()
}

func renderRepairFailureDetails(details []RepairFailureDetail) string {
	if len(details) == 0 {
		return ""
	}
	var builder strings.Builder
	builder.WriteString("Failed scorer checks:")
	for _, detail := range details {
		builder.WriteString("\n- ")
		builder.WriteString(detail.Name)
		builder.WriteString(" status=")
		builder.WriteString(detail.Status)
		if detail.Message != "" {
			builder.WriteString(" message=")
			builder.WriteString(detail.Message)
		}
		if detail.Evidence != "" {
			builder.WriteString(" evidence=")
			builder.WriteString(detail.Evidence)
		}
	}
	return builder.String()
}
