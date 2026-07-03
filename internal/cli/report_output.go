package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"ceoharness/internal/ceo"
)

type reportFormat string

const (
	reportFormatJSON   reportFormat = "json"
	reportFormatText   reportFormat = "text"
	reportFormatEvents reportFormat = "events"

	reportFormatGuidance = "--format must be json, text, or events"
)

type reportOutputRequest struct {
	Report       ceo.Report
	Format       reportFormat
	WorkspaceDir string
}

func parseReportFormat(raw string) (reportFormat, error) {
	switch strings.TrimSpace(raw) {
	case "", string(reportFormatJSON):
		return reportFormatJSON, nil
	case string(reportFormatText):
		return reportFormatText, nil
	case string(reportFormatEvents):
		return reportFormatEvents, nil
	default:
		return "", errors.New(reportFormatGuidance)
	}
}

func writeRunReport(out io.Writer, req reportOutputRequest) error {
	switch req.Format {
	case "", reportFormatJSON:
		encoder := json.NewEncoder(out)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(req.Report); err != nil {
			return fmt.Errorf("write CEO report: %w", err)
		}
		return nil
	case reportFormatText:
		if _, err := io.WriteString(out, renderTextReport(req)); err != nil {
			return fmt.Errorf("write text CEO report: %w", err)
		}
		return nil
	case reportFormatEvents:
		if err := writeRunEvents(out, req.Report.RunEvents); err != nil {
			return fmt.Errorf("write events CEO report: %w", err)
		}
		return nil
	default:
		return errors.New(reportFormatGuidance)
	}
}

func writeRunEvents(out io.Writer, events []ceo.RunEvent) error {
	encoder := json.NewEncoder(out)
	for _, event := range events {
		if err := encoder.Encode(event); err != nil {
			return err
		}
	}
	return nil
}

func renderTextReport(req reportOutputRequest) string {
	report := req.Report
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("CEO verdict: %s\n", report.Verdict))
	if report.JobID != "" {
		builder.WriteString(fmt.Sprintf("Job: %s\n", report.JobID))
	}
	builder.WriteString(fmt.Sprintf("Task: %s\n", displayTask(report.JobPacket.Task)))
	if strings.TrimSpace(report.JobOwner) != "" {
		builder.WriteString(fmt.Sprintf("Owner: %s\n", report.JobOwner))
	}
	builder.WriteString(fmt.Sprintf("Next: %s\n", report.ExecutionPlan.NextAction))
	writeTextCEODelegation(&builder, report)
	writeTextRunLedger(&builder, report)
	writeTextVerificationContract(&builder, report)
	writeTextPatchApproval(&builder, report)
	builder.WriteString("\nSubagents:\n")
	for _, result := range report.SubagentResults {
		builder.WriteString(fmt.Sprintf("- %s [%s]", result.AgentName, result.Status))
		if result.Stage > 0 {
			builder.WriteString(fmt.Sprintf(" stage %d", result.Stage))
		}
		if summary := oneLine(result.Summary); summary != "" {
			builder.WriteString(": " + trimText(summary, 110))
		}
		builder.WriteString("\n")
	}
	writeTextCheckSummary(&builder, report)
	writeTextChangedFiles(&builder, report)
	writeTextQuestions(&builder, req)
	return builder.String()
}

func oneLine(text string) string {
	return strings.Join(strings.Fields(text), " ")
}

func displayTask(task string) string {
	baseTask := task
	for _, marker := range []string{"\n\nresume_context:", "\n\nprior_job_context:"} {
		before, _, found := strings.Cut(baseTask, marker)
		if found {
			baseTask = before
		}
	}
	return trimText(oneLine(baseTask), 140)
}

func pluralize(word string, count int) string {
	if count == 1 {
		return word
	}
	return word + "s"
}

func trimText(text string, maxBytes int) string {
	if maxBytes < 1 || len(text) <= maxBytes {
		return text
	}
	end := 0
	for index := range text {
		if index > maxBytes {
			break
		}
		end = index
	}
	if end == 0 {
		return ""
	}
	return text[:end] + "..."
}
