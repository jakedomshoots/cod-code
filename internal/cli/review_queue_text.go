package cli

import (
	"fmt"
	"strings"
)

func writeReviewContextText(builder *strings.Builder, context *compactJobContext) {
	if context == nil {
		return
	}
	writeReviewDetailLine(builder, "Action", context.NextAction)
	for _, question := range context.Questions {
		writeReviewDetailLine(builder, "Question", question)
	}
	if len(context.ChangedFiles) > 0 {
		writeReviewDetailLine(builder, "Changed", strings.Join(context.ChangedFiles, ", "))
	}
	for _, check := range context.FailedChecks {
		line := strings.Join(check.Command, " ")
		if check.ExitCode != 0 {
			line += fmt.Sprintf(" exit=%d", check.ExitCode)
		}
		writeReviewDetailLine(builder, "Failed check", line)
	}
	writeReviewDetailLine(builder, "CEO", context.CEOReviewSummary)
}

func writeReviewDetailLine(builder *strings.Builder, label string, value string) {
	clean := trimText(oneLine(value), 120)
	if clean == "" {
		return
	}
	builder.WriteString("  ")
	builder.WriteString(label)
	builder.WriteString(": ")
	builder.WriteString(clean)
	builder.WriteString("\n")
}
