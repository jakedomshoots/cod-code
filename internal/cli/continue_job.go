package cli

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"ceoharness/internal/ceo"
	"ceoharness/internal/history"
	"ceoharness/internal/jobpacket"
)

func optionsWithContinueJob(ctx context.Context, opts options) (options, error) {
	jobID := strings.TrimSpace(opts.continueJobID)
	if jobID == "" {
		return opts, nil
	}
	if strings.TrimSpace(opts.resumeJobID) != "" {
		return options{}, fmt.Errorf("--continue-job cannot be combined with --resume")
	}
	if strings.TrimSpace(opts.rerunJobID) != "" {
		return options{}, fmt.Errorf("--continue-job cannot be combined with --rerun")
	}
	if strings.TrimSpace(opts.priorJobContextID) != "" {
		return options{}, fmt.Errorf("--continue-job cannot be combined with --with-job-context")
	}
	if strings.TrimSpace(opts.task) != "" {
		return options{}, fmt.Errorf("--continue-job cannot be combined with task text")
	}
	jobID, err := resolveSavedJobID(ctx, opts.workspaceDir, jobID)
	if err != nil {
		return options{}, err
	}
	if err := ensureContinueAllowedByHumanJudgment(ctx, opts.workspaceDir, jobID); err != nil {
		return options{}, err
	}
	report, err := readResumeReport(ctx, opts.workspaceDir, jobID)
	if err != nil {
		return options{}, fmt.Errorf("find continue job report: %w", err)
	}
	task := strings.TrimSpace(report.JobPacket.Task)
	if task == "" {
		return options{}, fmt.Errorf("continue job %s task is empty", jobID)
	}
	opts.task = task
	if len(report.JobPacket.Subagents) > 0 {
		opts.subagents = append([]jobpacket.Subagent(nil), report.JobPacket.Subagents...)
		opts.maxSubagents = len(opts.subagents)
	}
	opts.continuation = &ceo.ContinuationContext{
		JobID:              jobID,
		ReusableResults:    report.SubagentResults,
		UseSavedDelegation: len(report.JobPacket.Subagents) > 0,
		SavedDelegation:    report.CEODelegation,
	}
	return opts, nil
}

func ensureContinueAllowedByHumanJudgment(ctx context.Context, workspaceDir string, jobID string) error {
	store, err := history.New(workspaceDir)
	if err != nil {
		return err
	}
	judgment, err := store.ReadHumanJudgment(ctx, jobID)
	if err != nil {
		if errors.Is(err, history.ErrEntryNotFound) {
			return nil
		}
		return fmt.Errorf("find human judgment: %w", err)
	}
	if judgment.Verdict == "reject" {
		return fmt.Errorf("human judgment rejected job %s; inspect with --job %s before continuing", jobID, jobID)
	}
	return nil
}
