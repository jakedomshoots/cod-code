package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"ceoharness/internal/ceo"
	"ceoharness/internal/history"
)

func optionsWithResumeContext(ctx context.Context, opts options) (options, error) {
	jobID := strings.TrimSpace(opts.resumeJobID)
	if jobID == "" {
		if len(cleanResumeAnswers(opts.resumeAnswers)) > 0 {
			return options{}, fmt.Errorf("--answer requires --resume")
		}
		return opts, nil
	}
	if strings.TrimSpace(opts.rerunJobID) != "" {
		return options{}, fmt.Errorf("--resume cannot be combined with --rerun")
	}
	if strings.TrimSpace(opts.task) != "" {
		return options{}, fmt.Errorf("--resume cannot be combined with task text")
	}
	answers := cleanResumeAnswers(opts.resumeAnswers)
	if len(answers) == 0 {
		return options{}, fmt.Errorf("--resume requires at least one --answer")
	}
	jobID, err := resolveSavedJobID(ctx, opts.workspaceDir, jobID)
	if err != nil {
		return options{}, err
	}
	report, err := readResumeReport(ctx, opts.workspaceDir, jobID)
	if err != nil {
		return options{}, err
	}
	if report.Verdict != "needs_input" {
		return options{}, fmt.Errorf("resume job %s verdict is %q, want needs_input", jobID, report.Verdict)
	}
	task := strings.TrimSpace(report.JobPacket.Task)
	if task == "" {
		return options{}, fmt.Errorf("resume job %s task is empty", jobID)
	}
	questions := resumeQuestions(report)
	if len(questions) == 0 {
		return options{}, fmt.Errorf("resume job %s has no questions", jobID)
	}
	opts.task = task
	opts.resumeContext = &ceo.ResumeContext{
		JobID:     jobID,
		Questions: questions,
		Answers:   answers,
	}
	return opts, nil
}

func readResumeReport(ctx context.Context, workspaceDir string, jobID string) (ceo.Report, error) {
	store, err := history.New(workspaceDir)
	if err != nil {
		return ceo.Report{}, err
	}
	payload, err := store.ReadReportSnapshot(ctx, jobID)
	if err != nil {
		return ceo.Report{}, fmt.Errorf("find resume job report: %w", err)
	}
	var report ceo.Report
	if err := json.Unmarshal(payload, &report); err != nil {
		return ceo.Report{}, fmt.Errorf("decode resume job report: %w", err)
	}
	return report, nil
}

func resumeQuestions(report ceo.Report) []string {
	questions := []string{}
	seen := map[string]struct{}{}
	for _, result := range report.SubagentResults {
		for _, question := range result.Questions {
			cleanQuestion := strings.TrimSpace(question)
			if cleanQuestion == "" {
				continue
			}
			if _, ok := seen[cleanQuestion]; ok {
				continue
			}
			seen[cleanQuestion] = struct{}{}
			questions = append(questions, cleanQuestion)
		}
	}
	return questions
}

func cleanResumeAnswers(answers []string) []string {
	cleaned := make([]string, 0, len(answers))
	for _, answer := range answers {
		cleanAnswer := strings.TrimSpace(answer)
		if cleanAnswer != "" {
			cleaned = append(cleaned, cleanAnswer)
		}
	}
	return cleaned
}
