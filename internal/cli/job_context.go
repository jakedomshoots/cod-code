package cli

import (
	"encoding/json"
	"fmt"
	"strings"

	"ceoharness/internal/ceo"
	"ceoharness/internal/history"
	"ceoharness/internal/subagent"
)

const failureExcerptLimit = 500

type compactJobContext struct {
	JobID            string                  `json:"job_id"`
	CreatedAt        string                  `json:"created_at,omitempty"`
	Task             string                  `json:"task"`
	TaskKind         string                  `json:"task_kind,omitempty"`
	RiskLevel        string                  `json:"risk_level,omitempty"`
	Verdict          string                  `json:"verdict"`
	NextAction       string                  `json:"next_action,omitempty"`
	RunLedger        *history.RunLedger      `json:"run_ledger,omitempty"`
	ResumeArgs       []string                `json:"resume_args,omitempty"`
	SuggestedCommand string                  `json:"suggested_command,omitempty"`
	Questions        []string                `json:"questions,omitempty"`
	ChangedFiles     []string                `json:"changed_files,omitempty"`
	Counts           jobContextCounts        `json:"counts"`
	Subagents        []compactSubagentResult `json:"subagents,omitempty"`
	FailedChecks     []compactCheckResult    `json:"failed_checks,omitempty"`
	CEOReviewSummary string                  `json:"ceo_review_summary,omitempty"`
}

type jobContextCounts struct {
	Subagents    int `json:"subagents"`
	Checks       int `json:"checks"`
	FailedChecks int `json:"failed_checks"`
	Patches      int `json:"patches"`
}

type compactSubagentResult struct {
	Name    string `json:"name"`
	Role    string `json:"role,omitempty"`
	Status  string `json:"status"`
	Summary string `json:"summary,omitempty"`
}

type compactCheckResult struct {
	Command        []string `json:"command"`
	Status         string   `json:"status"`
	ExitCode       int      `json:"exit_code"`
	CheckIndex     int      `json:"check_index,omitempty"`
	Attempt        int      `json:"attempt,omitempty"`
	MaxAttempts    int      `json:"max_attempts,omitempty"`
	DurationMS     int64    `json:"duration_ms,omitempty"`
	FailureExcerpt string   `json:"failure_excerpt,omitempty"`
}

func decodeJobContextReport(payload []byte) (ceo.Report, error) {
	var report ceo.Report
	if err := json.Unmarshal(payload, &report); err != nil {
		return ceo.Report{}, fmt.Errorf("decode job context report: %w", err)
	}
	return report, nil
}

func contextFromHistoryEntry(entry history.Entry, workspaceDir string) compactJobContext {
	packet := compactJobContext{
		JobID:        entry.ID,
		CreatedAt:    entry.CreatedAt,
		Task:         entry.Task,
		TaskKind:     entry.TaskKind,
		RiskLevel:    entry.RiskLevel,
		Verdict:      entry.Verdict,
		NextAction:   entry.ExecutionPlanNextAction,
		RunLedger:    cloneContextRunLedger(entry.RunLedger),
		ChangedFiles: append([]string(nil), entry.ChangedFiles...),
		Counts: jobContextCounts{
			Subagents: entry.SubagentCount,
			Checks:    entry.CheckCount,
			Patches:   entry.PatchCount,
		},
	}
	setContextResumeCommand(&packet, workspaceDir, entry.ID, entry.Verdict)
	return packet
}

func contextFromReport(entry history.Entry, report ceo.Report, workspaceDir string) compactJobContext {
	packet := contextFromHistoryEntry(entry, workspaceDir)
	if strings.TrimSpace(report.JobID) != "" {
		packet.JobID = report.JobID
		setContextResumeCommand(&packet, workspaceDir, report.JobID, report.Verdict)
	}
	if strings.TrimSpace(report.JobPacket.Task) != "" {
		packet.Task = report.JobPacket.Task
	}
	if strings.TrimSpace(report.JobPacket.TaskProfile.Kind) != "" {
		packet.TaskKind = report.JobPacket.TaskProfile.Kind
	}
	if strings.TrimSpace(report.JobPacket.TaskProfile.RiskLevel) != "" {
		packet.RiskLevel = report.JobPacket.TaskProfile.RiskLevel
	}
	if strings.TrimSpace(report.Verdict) != "" {
		packet.Verdict = report.Verdict
		setContextResumeCommand(&packet, workspaceDir, packet.JobID, report.Verdict)
	}
	if strings.TrimSpace(report.ExecutionPlan.NextAction) != "" {
		packet.NextAction = report.ExecutionPlan.NextAction
	}
	if hasRunLedger(report.RunLedger) {
		packet.RunLedger = cloneContextRunLedger(&report.RunLedger)
	}
	if len(report.ChangedFiles) > 0 {
		packet.ChangedFiles = append([]string(nil), report.ChangedFiles...)
	}
	packet.Questions = collectQuestions(report.Resume, report.SubagentResults)
	packet.Subagents = compactSubagents(report.SubagentResults)
	packet.FailedChecks = compactFailedChecks(report.CheckResults)
	packet.Counts = jobContextCounts{
		Subagents:    len(report.SubagentResults),
		Checks:       len(report.CheckResults),
		FailedChecks: len(packet.FailedChecks),
		Patches:      len(report.PatchResults),
	}
	if report.CEOReview != nil {
		packet.CEOReviewSummary = strings.TrimSpace(report.CEOReview.Summary)
	}
	return packet
}

func cloneContextRunLedger(ledger *history.RunLedger) *history.RunLedger {
	if ledger == nil {
		return nil
	}
	cloned := *ledger
	cloned.ChangedFiles = append([]string(nil), ledger.ChangedFiles...)
	cloned.ProviderRouteReasons = append([]string(nil), ledger.ProviderRouteReasons...)
	return &cloned
}

func hasRunLedger(ledger history.RunLedger) bool {
	return strings.TrimSpace(ledger.Owner) != "" ||
		strings.TrimSpace(ledger.Verdict) != "" ||
		strings.TrimSpace(ledger.NextAction) != "" ||
		strings.TrimSpace(ledger.VerificationStatus) != "" ||
		ledger.ChangedFileCount > 0 ||
		ledger.ProviderRouteCount > 0
}

func resumeArgs(workspaceDir string, jobID string, verdict string) []string {
	if verdict != "needs_input" || strings.TrimSpace(jobID) == "" {
		return nil
	}
	return []string{"--workspace", workspaceDir, "--resume", jobID, "--answer", "<answer>", "--"}
}

func setContextResumeCommand(context *compactJobContext, workspaceDir string, jobID string, verdict string) {
	context.ResumeArgs = resumeArgs(workspaceDir, jobID, verdict)
	context.SuggestedCommand = renderJobContextCommand(context.ResumeArgs)
}

func collectQuestions(resume *ceo.ResumeContext, results []subagent.Result) []string {
	questions := []string{}
	seen := map[string]struct{}{}
	addQuestion := func(question string) {
		clean := strings.TrimSpace(question)
		if clean == "" {
			return
		}
		if _, ok := seen[clean]; ok {
			return
		}
		seen[clean] = struct{}{}
		questions = append(questions, clean)
	}
	if resume != nil {
		for _, question := range resume.Questions {
			addQuestion(question)
		}
	}
	for _, result := range results {
		for _, question := range result.Questions {
			addQuestion(question)
		}
	}
	return questions
}

func compactSubagents(results []subagent.Result) []compactSubagentResult {
	if len(results) == 0 {
		return nil
	}
	compacted := make([]compactSubagentResult, 0, len(results))
	for _, result := range results {
		compacted = append(compacted, compactSubagentResult{
			Name:    result.AgentName,
			Role:    result.Role,
			Status:  result.Status,
			Summary: trimContextText(result.Summary, contextLineLimit),
		})
	}
	return compacted
}
