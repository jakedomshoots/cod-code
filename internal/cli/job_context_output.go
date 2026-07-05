package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type jobContextLookupReport struct {
	HistoryPath string            `json:"history_path"`
	Source      string            `json:"source"`
	JobContext  compactJobContext `json:"job_context"`
}

func runJobContextLookup(ctx context.Context, out io.Writer, workspaceDir string, jobID string, format reportFormat) error {
	loaded, err := loadCompactJobContext(ctx, workspaceDir, jobID)
	if err != nil {
		return err
	}
	report := jobContextLookupReport{
		HistoryPath: loaded.HistoryPath,
		Source:      loaded.Source,
		JobContext:  loaded.Context,
	}
	return writeJobContextLookupReport(out, report, format)
}

func writeJobContextLookupReport(out io.Writer, report jobContextLookupReport, format reportFormat) error {
	switch format {
	case "", reportFormatJSON:
		encoder := json.NewEncoder(out)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(report); err != nil {
			return fmt.Errorf("write job context: %w", err)
		}
		return nil
	case reportFormatText:
		if _, err := io.WriteString(out, renderJobContextText(report)); err != nil {
			return fmt.Errorf("write text job context: %w", err)
		}
		return nil
	case reportFormatEvents:
		return fmt.Errorf("--format events is only available for run reports")
	default:
		return fmt.Errorf(reportFormatGuidance)
	}
}

func renderJobContextText(report jobContextLookupReport) string {
	context := report.JobContext
	var builder strings.Builder
	builder.WriteString("Job context: ")
	builder.WriteString(context.JobID)
	builder.WriteString("\n")
	writeJobContextLine(&builder, "Source", report.Source)
	writeJobContextLine(&builder, "Verdict", context.Verdict)
	writeJobContextLine(&builder, "Task", displayTask(context.Task))
	writeJobContextLine(&builder, "Profile", jobContextProfile(context))
	writeJobContextLine(&builder, "Next", context.NextAction)
	for _, question := range context.Questions {
		writeJobContextLine(&builder, "Question", question)
	}
	if len(context.ChangedFiles) > 0 {
		writeJobContextLine(&builder, "Changed", strings.Join(context.ChangedFiles, ", "))
	}
	for _, check := range context.FailedChecks {
		writeJobContextLine(&builder, "Failed check", renderJobContextCheck(check))
	}
	for _, result := range context.Subagents {
		writeJobContextLine(&builder, "Subagent", renderJobContextSubagent(result))
	}
	writeJobContextLine(&builder, "CEO", context.CEOReviewSummary)
	writeJobContextLine(&builder, "Resume", context.SuggestedCommand)
	return builder.String()
}

func writeJobContextLine(builder *strings.Builder, label string, value string) {
	clean := trimText(oneLine(value), 180)
	if clean == "" {
		return
	}
	builder.WriteString(label)
	builder.WriteString(": ")
	builder.WriteString(clean)
	builder.WriteString("\n")
}

func jobContextProfile(context compactJobContext) string {
	if context.TaskKind == "" && context.RiskLevel == "" {
		return ""
	}
	if context.TaskKind == "" {
		return context.RiskLevel
	}
	if context.RiskLevel == "" {
		return context.TaskKind
	}
	return context.TaskKind + "/" + context.RiskLevel
}

func renderJobContextCheck(check compactCheckResult) string {
	line := strings.Join(check.Command, " ")
	if check.ExitCode != 0 {
		line += fmt.Sprintf(" exit=%d", check.ExitCode)
	}
	if check.FailureExcerpt != "" {
		line += " " + trimText(oneLine(check.FailureExcerpt), 80)
	}
	return line
}

func renderJobContextSubagent(result compactSubagentResult) string {
	line := result.Name
	if result.Status != "" {
		line += " [" + result.Status + "]"
	}
	if result.Summary != "" {
		line += " " + result.Summary
	}
	return line
}

func renderJobContextCommand(args []string) string {
	if len(args) == 0 {
		return ""
	}
	parts := []string{"cod"}
	for _, arg := range args {
		if strings.ContainsAny(arg, " \t\n\"'<>") {
			parts = append(parts, strconv.Quote(arg))
			continue
		}
		parts = append(parts, arg)
	}
	return strings.Join(parts, " ")
}
