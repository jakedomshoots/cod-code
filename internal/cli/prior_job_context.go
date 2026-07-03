package cli

import (
	"context"
	"fmt"
	"strings"

	"ceoharness/internal/history"
)

const contextLineLimit = 240

func optionsWithPriorJobContext(ctx context.Context, opts options) (options, error) {
	jobID := strings.TrimSpace(opts.priorJobContextID)
	if jobID == "" {
		return opts, nil
	}
	if strings.TrimSpace(opts.resumeJobID) != "" {
		return options{}, fmt.Errorf("--with-job-context cannot be combined with --resume")
	}
	if strings.TrimSpace(opts.task) == "" {
		return options{}, fmt.Errorf("--with-job-context requires task text or --rerun")
	}
	loaded, err := loadCompactJobContext(ctx, opts.workspaceDir, jobID)
	if err != nil {
		return options{}, err
	}
	opts.task = taskWithPriorJobContext(opts.task, loaded.Context)
	return opts, nil
}

func taskWithPriorJobContext(task string, prior compactJobContext) string {
	cleanTask := taskWithoutPriorJobContext(task)
	rendered := renderPriorJobContext(prior)
	if rendered == "" {
		return cleanTask
	}
	return strings.TrimSpace(cleanTask + "\n\nprior_job_context:\n" + rendered)
}

func taskWithoutPriorJobContext(task string) string {
	cleanTask := strings.TrimSpace(task)
	if before, _, ok := strings.Cut(cleanTask, "\n\nprior_job_context:"); ok {
		return strings.TrimSpace(before)
	}
	return cleanTask
}

func renderPriorJobContext(prior compactJobContext) string {
	var builder strings.Builder
	writeContextLine(&builder, "previous_job", prior.JobID)
	writeContextLine(&builder, "previous_task", taskWithoutPriorJobContext(prior.Task))
	if prior.RunLedger != nil {
		writeContextLine(&builder, "previous_run_ledger", renderPriorRunLedger(*prior.RunLedger))
	} else {
		writeContextLine(&builder, "previous_verdict", prior.Verdict)
		writeContextLine(&builder, "previous_next_action", prior.NextAction)
	}
	if len(prior.ChangedFiles) > 0 {
		writeContextLine(&builder, "previous_changed_files", strings.Join(prior.ChangedFiles, ", "))
	}
	if len(prior.Questions) > 0 {
		writeContextLine(&builder, "previous_questions", strings.Join(prior.Questions, " | "))
	}
	for _, check := range priorFailedChecksForPrompt(prior.FailedChecks) {
		writeContextLine(&builder, "previous_failed_check", renderPriorCheck(check))
	}
	for _, subagent := range prior.Subagents {
		writeContextLine(&builder, "previous_subagent", renderPriorSubagent(subagent))
	}
	writeContextLine(&builder, "previous_ceo_review", prior.CEOReviewSummary)
	return strings.TrimSpace(builder.String())
}

func renderPriorRunLedger(ledger history.RunLedger) string {
	parts := []string{}
	addPart := func(label string, value string) {
		clean := strings.TrimSpace(value)
		if clean == "" {
			return
		}
		parts = append(parts, fmt.Sprintf("%s=%s", label, clean))
	}
	addPart("owner", ledger.Owner)
	addPart("verdict", ledger.Verdict)
	if nextAction := strings.TrimSpace(ledger.NextAction); nextAction != "" {
		parts = append(parts, fmt.Sprintf("next=%q", nextAction))
	}
	addPart("verification", ledger.VerificationStatus)
	parts = append(parts, fmt.Sprintf("changed=%d", ledger.ChangedFileCount))
	parts = append(parts, fmt.Sprintf("routes=%d", ledger.ProviderRouteCount))
	if len(ledger.ProviderRouteReasons) > 0 {
		parts = append(parts, "reasons="+strings.Join(ledger.ProviderRouteReasons, ","))
	}
	return strings.Join(parts, " ")
}

func writeContextLine(builder *strings.Builder, label string, value string) {
	clean := trimContextText(value, contextLineLimit)
	if clean == "" {
		return
	}
	builder.WriteString(label)
	builder.WriteString(": ")
	builder.WriteString(clean)
	builder.WriteByte('\n')
}

func renderPriorCheck(check compactCheckResult) string {
	command := strings.TrimSpace(strings.Join(check.Command, " "))
	if command == "" {
		command = "check"
	}
	summary := command + " [" + strings.TrimSpace(check.Status) + "]"
	if metadata := renderPriorCheckMetadata(check); metadata != "" {
		summary += " " + metadata
	}
	if excerpt := trimContextText(check.FailureExcerpt, contextLineLimit); excerpt != "" {
		summary += ": " + excerpt
	}
	return summary
}

func renderPriorCheckMetadata(check compactCheckResult) string {
	parts := []string{}
	if check.CheckIndex > 0 {
		parts = append(parts, fmt.Sprintf("index=%d", check.CheckIndex))
	}
	if check.Attempt > 0 {
		attempt := fmt.Sprintf("attempt=%d", check.Attempt)
		if check.MaxAttempts > 0 {
			attempt = fmt.Sprintf("%s/%d", attempt, check.MaxAttempts)
		}
		parts = append(parts, attempt)
	}
	if check.DurationMS > 0 {
		parts = append(parts, fmt.Sprintf("duration_ms=%d", check.DurationMS))
	}
	return strings.Join(parts, " ")
}

func priorFailedChecksForPrompt(checks []compactCheckResult) []compactCheckResult {
	if len(checks) == 0 {
		return nil
	}
	compacted := []compactCheckResult{}
	byKey := map[string]int{}
	for _, check := range checks {
		key := priorCheckKey(check)
		if key == "" {
			compacted = append(compacted, check)
			continue
		}
		index, ok := byKey[key]
		if !ok {
			byKey[key] = len(compacted)
			compacted = append(compacted, check)
			continue
		}
		compacted[index] = check
	}
	return compacted
}

func priorCheckKey(check compactCheckResult) string {
	if check.CheckIndex > 0 {
		return fmt.Sprintf("index:%d", check.CheckIndex)
	}
	command := strings.TrimSpace(strings.Join(check.Command, "\x00"))
	if command == "" {
		return ""
	}
	return "command:" + command
}

func renderPriorSubagent(result compactSubagentResult) string {
	name := strings.TrimSpace(result.Name)
	if name == "" {
		name = "subagent"
	}
	summary := name + " [" + strings.TrimSpace(result.Status) + "]"
	if text := trimContextText(result.Summary, contextLineLimit); text != "" {
		summary += ": " + text
	}
	return summary
}

func trimContextText(text string, limit int) string {
	clean := strings.Join(strings.Fields(strings.TrimSpace(text)), " ")
	if clean == "" || limit < 1 || len(clean) <= limit {
		return clean
	}
	return clean[:limit] + "..."
}
