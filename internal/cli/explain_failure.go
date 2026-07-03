package cli

import (
	"context"
	"fmt"
	"io"
	"strings"
)

func runExplainFailure(ctx context.Context, out io.Writer, workspaceDir string, rawJobID string) error {
	jobID, err := resolveFailedJobID(ctx, workspaceDir, rawJobID)
	if err != nil {
		return err
	}
	loaded, err := loadCompactJobContext(ctx, workspaceDir, jobID)
	if err != nil {
		return err
	}
	if loaded.Context.Verdict != "fail" {
		return fmt.Errorf("job %s is not failed; verdict is %q", loaded.Context.JobID, loaded.Context.Verdict)
	}
	if _, err := fmt.Fprint(out, renderFailureExplanation(workspaceDir, loaded.Context)); err != nil {
		return fmt.Errorf("write failure explanation: %w", err)
	}
	return nil
}

func renderFailureExplanation(workspaceDir string, context compactJobContext) string {
	var builder strings.Builder
	builder.WriteString("Failure explanation\n")
	builder.WriteString("Job: " + context.JobID + "\n")
	builder.WriteString("Verdict: " + context.Verdict + "\n")
	builder.WriteString("Likely reason: " + failureReason(context) + "\n")
	if task := strings.TrimSpace(context.Task); task != "" {
		builder.WriteString("Task: " + trimText(oneLine(task), 140) + "\n")
	}
	builder.WriteString("Failed checks:\n")
	if len(context.FailedChecks) > 0 {
		for _, check := range context.FailedChecks {
			builder.WriteString("- " + renderFailureCheck(check) + "\n")
		}
	} else {
		builder.WriteString("- none recorded in the saved report\n")
	}
	retryable := strings.TrimSpace(context.Task) != "" && context.Verdict == "fail"
	builder.WriteString("Retryable: " + yesNo(retryable) + "\n")
	if retryable {
		builder.WriteString("Suggested retry: ceo-packet retry " + context.JobID)
		if strings.TrimSpace(workspaceDir) != "" {
			builder.WriteString(" --workspace " + workspaceDir)
		}
		builder.WriteString("\n")
	}
	builder.WriteString("Evidence path: " + reportEvidencePointer(context.JobID) + "\n")
	builder.WriteString("Report path: " + reportEvidencePointer(context.JobID) + "\n")
	return builder.String()
}

func failureReason(context compactJobContext) string {
	if len(context.FailedChecks) > 0 {
		return "one or more checks failed"
	}
	if strings.TrimSpace(context.NextAction) != "" {
		return context.NextAction
	}
	return "job did not reach a passing verdict"
}

func renderFailureCheck(check compactCheckResult) string {
	command := strings.TrimSpace(strings.Join(check.Command, " "))
	if command == "" {
		command = "check"
	}
	line := command + " [" + strings.TrimSpace(check.Status) + "]"
	if excerpt := strings.TrimSpace(check.FailureExcerpt); excerpt != "" {
		line += ": " + trimText(oneLine(excerpt), 160)
	}
	return line
}
