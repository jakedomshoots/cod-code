package ceo

import (
	"strconv"
	"strings"

	"ceoharness/internal/checkrunner"
)

func writeCEOReviewChecks(builder *strings.Builder, checks []checkrunner.Result) {
	if len(checks) == 0 {
		builder.WriteString("- none\n")
		return
	}
	for _, check := range checks {
		builder.WriteString("- status=")
		builder.WriteString(check.Status)
		writeCEOReviewCheckMetadata(builder, check)
		if check.Stdout != "" {
			builder.WriteString(" stdout=")
			builder.WriteString(compactCEOReviewText(check.Stdout))
		}
		if check.Stderr != "" {
			builder.WriteString(" stderr=")
			builder.WriteString(compactCEOReviewText(check.Stderr))
		}
		builder.WriteString("\n")
	}
}

func writeCEOReviewCheckMetadata(builder *strings.Builder, check checkrunner.Result) {
	if len(check.Argv) > 0 {
		builder.WriteString(" argv=")
		builder.WriteString(strconv.Quote(strings.Join(check.Argv, " ")))
	}
	if check.CheckIndex > 0 {
		builder.WriteString(" index=")
		builder.WriteString(strconv.Itoa(check.CheckIndex))
	}
	if check.Attempt > 0 {
		builder.WriteString(" attempt=")
		builder.WriteString(strconv.Itoa(check.Attempt))
		if check.MaxAttempts > 0 {
			builder.WriteString("/")
			builder.WriteString(strconv.Itoa(check.MaxAttempts))
		}
	}
	builder.WriteString(" exit_code=")
	builder.WriteString(strconv.Itoa(check.ExitCode))
	if check.DurationMS > 0 {
		builder.WriteString(" duration_ms=")
		builder.WriteString(strconv.FormatInt(check.DurationMS, 10))
	}
}
