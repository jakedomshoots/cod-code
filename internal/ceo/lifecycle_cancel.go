package ceo

import (
	"context"
	"errors"
	"fmt"

	"ceoharness/internal/checkrunner"
	"ceoharness/internal/jobpacket"
	"ceoharness/internal/subagent"
	"ceoharness/internal/workspace"
)

type canceledLifecycleReportInput struct {
	Request         JobRequest
	Packet          jobpacket.Packet
	Delegation      *CEODelegation
	Workspace       *workspace.Brief
	Resume          *ResumeContext
	Results         []subagent.Result
	ChangedFiles    []string
	Checks          []checkrunner.Result
	NoProgressStop  int
	IterationBudget *ceoIterationBudget
	HasWorkspace    bool
	ArtifactStore   runtimeArtifactStore
	CanceledAt      LifecycleState
	Err             error
}

func persistCanceledLifecycleReport(ctx context.Context, input canceledLifecycleReportInput) (Report, error) {
	cancelErr := input.Err
	if cancelErr == nil {
		cancelErr = context.Canceled
	}
	report := buildCanceledLifecycleReport(input, cancelErr)
	if input.HasWorkspace && !input.Request.DryRun {
		persisted, err := persistRuntimeReport(context.WithoutCancel(ctx), input.ArtifactStore, input.Request.DryRun, report)
		if err != nil {
			return report, errors.Join(cancelErr, fmt.Errorf("persist canceled lifecycle report: %w", err))
		}
		report = persisted
	}
	return report, cancelErr
}

func persistCanceledLifecycleReportForError(ctx context.Context, err error, input canceledLifecycleReportInput) (Report, error, bool) {
	cancelErr := lifecycleCancellationError(ctx, err)
	if cancelErr == nil {
		return Report{}, err, false
	}
	input.Err = cancelErr
	report, persistErr := persistCanceledLifecycleReport(ctx, input)
	return report, persistErr, true
}

func buildCanceledLifecycleReport(input canceledLifecycleReportInput, cancelErr error) Report {
	summary := summarizeVerificationWithPolicy(input.Results, input.Checks, input.Request.ProviderHealthPolicy)
	summary = applyProviderCostBudget(summary, input.Request.ProviderCostBudgetMicroUSD)
	plan := buildExecutionPlan(executionPlanInput{
		Packet:  input.Packet,
		Results: input.Results,
		Checks:  input.Checks,
		Verdict: "canceled",
	})
	verificationContract := NewVerificationContract(checkCommands(input.Request), input.Checks)
	report := buildReport(reportBuildInput{
		Packet:                          input.Packet,
		Delegation:                      input.Delegation,
		Continuation:                    input.Request.Continuation,
		Workspace:                       input.Workspace,
		Resume:                          input.Resume,
		Results:                         input.Results,
		ChangedFiles:                    input.ChangedFiles,
		Checks:                          input.Checks,
		VerificationContract:            verificationContract,
		ProviderRouteDecisions:          input.Request.ProviderRouteDecisions,
		Summary:                         summary,
		Plan:                            plan,
		SubagentConcurrency:             input.Request.SubagentConcurrency,
		MaxToolRequests:                 input.Request.MaxToolRequests,
		MaxOutputBytes:                  input.Request.MaxSubagentOutputBytes,
		NoProgressStop:                  input.NoProgressStop,
		MaxCEOIterations:                canceledLifecycleMaxIterations(input.IterationBudget),
		CEOIterationCount:               canceledLifecycleIterationCount(input.IterationBudget),
		CEOIterationExhausted:           canceledLifecycleIterationExhausted(input.IterationBudget),
		DryRun:                          input.Request.DryRun,
		ProviderHealthAvoidedRouteCount: input.Request.ProviderHealthAvoidedRouteCount,
		ProviderHealthAvoidedProviders:  input.Request.ProviderHealthAvoidedProviders,
		Verdict:                         "canceled",
	})
	lifecycle := buildLifecycle(lifecycleInput{
		Recovered:      input.Request.Continuation != nil || input.Resume != nil,
		Canceled:       true,
		CanceledAt:     input.CanceledAt,
		CancelReason:   cancelErr.Error(),
		ResultStatuses: subagentResultStatuses(input.Results),
		CheckStatuses:  checkResultStatuses(input.Checks),
		Verdict:        "canceled",
	})
	report.LifecycleState = lifecycle.State
	report.LifecycleEvents = lifecycle.Events
	report.RunEvents = buildRunEvents(runEventsInput{
		Packet:                          input.Packet,
		Delegation:                      input.Delegation,
		WorkspaceBrief:                  input.Workspace,
		ProviderHealthAvoidedRouteCount: input.Request.ProviderHealthAvoidedRouteCount,
		ProviderHealthAvoidedProviders:  input.Request.ProviderHealthAvoidedProviders,
		Results:                         input.Results,
		Checks:                          input.Checks,
		LifecycleEvents:                 lifecycle.Events,
		Verdict:                         "canceled",
	})
	return report
}

func canceledLifecycleMaxIterations(budget *ceoIterationBudget) int {
	if budget == nil {
		return DefaultMaxCEOIterations
	}
	return budget.max
}

func canceledLifecycleIterationCount(budget *ceoIterationBudget) int {
	if budget == nil {
		return 0
	}
	return budget.used
}

func canceledLifecycleIterationExhausted(budget *ceoIterationBudget) bool {
	if budget == nil {
		return false
	}
	return budget.exhausted
}

func lifecycleCancellationError(ctx context.Context, err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return err
	}
	if ctxErr := ctx.Err(); ctxErr != nil {
		return ctxErr
	}
	return nil
}
