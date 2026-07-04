package ceo

import (
	"context"
	"fmt"

	"ceoharness/internal/checkrunner"
	"ceoharness/internal/jobpacket"
)

func (r Runtime) RunJob(ctx context.Context, req JobRequest) (Report, error) {
	if req.ApplyModelPatches && req.PreviewModelPatches && !req.DryRun {
		return Report{}, fmt.Errorf("choose either model patch preview or model patch application")
	}
	resume := normalizeResumeContext(req.Resume)
	packet, err := jobpacket.BuildWithOptions(jobpacket.BuildOptions{
		Task:            taskWithResumeContext(req.Task, resume),
		Subagents:       req.Subagents,
		MaxSubagents:    req.MaxSubagents,
		MaxContextBytes: req.MaxContextBytes,
	})
	if err != nil {
		return Report{}, err
	}
	packet, ceoDelegation, err := r.runCEODelegation(ctx, packet, shouldRunCEODelegation(req.Continuation))
	if err != nil {
		return Report{}, err
	}
	if ceoDelegation == nil {
		ceoDelegation = savedDelegationFromContinuation(req.Continuation)
	}
	space, hasWorkspace, err := openWorkspace(req.WorkspaceDir)
	if err != nil {
		return Report{}, err
	}
	artifactStore, err := openRuntimeArtifactStore(req, space, hasWorkspace)
	if err != nil {
		return Report{}, err
	}
	workspaceBrief, err := buildWorkspaceBrief(ctx, space, hasWorkspace, req.WorkspaceBriefMaxFiles, req.WorkspaceBriefExcludes)
	if err != nil {
		return Report{}, err
	}
	noProgressStop := normalizeNoProgressStop(req.NoProgressStop)
	iterationBudget := newCEOIterationBudget(req.MaxCEOIterations)

	results, err := r.runSubagents(ctx, subagentsRunInput{
		Packet:         packet,
		ToolState:      toolRequestState{Request: req, Space: space, HasWorkspace: hasWorkspace},
		Attempts:       req.SubagentAttempts,
		BackoffMS:      req.SubagentBackoffMS,
		NoProgressStop: noProgressStop,
		Concurrency:    req.SubagentConcurrency,
		WorkspaceBrief: renderWorkspaceBrief(workspaceBrief),
		MaxOutputBytes: req.MaxSubagentOutputBytes,
		Continuation:   req.Continuation,
	})
	if err != nil {
		if cancelErr := lifecycleCancellationError(ctx, err); cancelErr != nil {
			return persistCanceledLifecycleReport(ctx, canceledLifecycleReportInput{
				Request:         req,
				Packet:          packet,
				Delegation:      ceoDelegation,
				Workspace:       workspaceBrief,
				Resume:          resume,
				NoProgressStop:  noProgressStop,
				IterationBudget: iterationBudget,
				HasWorkspace:    hasWorkspace,
				ArtifactStore:   artifactStore,
				CanceledAt:      LifecycleDelegated,
				Err:             cancelErr,
			})
		}
		return Report{}, err
	}
	var changedFiles []string
	var checkResults []checkrunner.Result
	var retryHistory []RetryHistoryEntry
	cancelReport := func(err error, stage LifecycleState) (Report, error, bool) {
		return persistCanceledLifecycleReportForError(ctx, err, canceledLifecycleReportInput{
			Request:         req,
			Packet:          packet,
			Delegation:      ceoDelegation,
			Workspace:       workspaceBrief,
			Resume:          resume,
			Results:         results,
			ChangedFiles:    changedFiles,
			Checks:          checkResults,
			NoProgressStop:  noProgressStop,
			IterationBudget: iterationBudget,
			HasWorkspace:    hasWorkspace,
			ArtifactStore:   artifactStore,
			CanceledAt:      stage,
		})
	}
	artifacts, err := buildRuntimeArtifacts(ctx, runtimeArtifactsInput{
		Request:       req,
		Space:         space,
		HasWorkspace:  hasWorkspace,
		ArtifactStore: artifactStore,
		Results:       results,
	})
	if err != nil {
		if report, cancelErr, ok := cancelReport(err, LifecycleDelegated); ok {
			return report, cancelErr
		}
		return Report{}, err
	}
	changedFiles = artifacts.ChangedFiles
	patchResults := artifacts.PatchResults
	patchPreviews := artifacts.PatchPreviews
	patchPreviewEvents := artifacts.PreviewEvents
	patchAudit := artifacts.PatchAudit
	patchApproval := artifacts.PatchApproval

	checkResults, err = r.runChecks(ctx, req)
	if err != nil {
		if report, cancelErr, ok := cancelReport(err, LifecycleChecking); ok {
			return report, cancelErr
		}
		return Report{}, err
	}
	checkFixState, err := r.runCheckFixLoop(ctx, checkFixState{
		Packet:          packet,
		Request:         req,
		Space:           space,
		HasWorkspace:    hasWorkspace,
		ArtifactStore:   artifactStore,
		Results:         results,
		ChangedFiles:    changedFiles,
		CheckResults:    checkResults,
		PatchResults:    patchResults,
		PatchAudit:      patchAudit,
		WorkspaceBrief:  renderWorkspaceBrief(workspaceBrief),
		IterationBudget: iterationBudget,
	})
	if err != nil {
		if report, cancelErr, ok := cancelReport(err, LifecycleChecking); ok {
			return report, cancelErr
		}
		return Report{}, err
	}
	results = checkFixState.Results
	changedFiles = checkFixState.ChangedFiles
	checkResults = checkFixState.CheckResults
	patchResults = checkFixState.PatchResults
	patchAudit = checkFixState.PatchAudit
	retryHistory = checkFixState.RetryHistory

	summary := summarizeVerificationWithPolicy(results, checkResults, req.ProviderHealthPolicy)
	summary = applyProviderCostBudget(summary, req.ProviderCostBudgetMicroUSD)
	guardVerdict := verdict(results, checkResults, summary)
	ceoReview, err := r.runCEOReview(ctx, ceoReviewInput{
		Packet:        packet,
		Results:       results,
		Checks:        checkResults,
		ChangedFiles:  changedFiles,
		PatchResults:  patchResults,
		PatchPreviews: patchPreviews,
		GuardVerdict:  guardVerdict,
	})
	if err != nil {
		if report, cancelErr, ok := cancelReport(err, LifecycleReviewing); ok {
			return report, cancelErr
		}
		return Report{}, err
	}
	revisionState, err := r.runCEORevisionLoop(ctx, ceoRevisionState{
		Packet:          packet,
		Request:         req,
		Space:           space,
		HasWorkspace:    hasWorkspace,
		ArtifactStore:   artifactStore,
		Results:         results,
		ChangedFiles:    changedFiles,
		CheckResults:    checkResults,
		PatchResults:    patchResults,
		PatchPreviews:   patchPreviews,
		PatchAudit:      patchAudit,
		WorkspaceBrief:  renderWorkspaceBrief(workspaceBrief),
		Summary:         summary,
		GuardVerdict:    guardVerdict,
		CEOReview:       ceoReview,
		IterationBudget: iterationBudget,
		RetryHistory:    retryHistory,
	})
	if err != nil {
		if report, cancelErr, ok := cancelReport(err, LifecycleReviewing); ok {
			return report, cancelErr
		}
		return Report{}, err
	}
	results = revisionState.Results
	changedFiles = revisionState.ChangedFiles
	checkResults = revisionState.CheckResults
	patchResults = revisionState.PatchResults
	patchAudit = revisionState.PatchAudit
	retryHistory = revisionState.RetryHistory
	summary = revisionState.Summary
	guardVerdict = revisionState.GuardVerdict
	ceoReview = revisionState.CEOReview
	finalVerdict := applyCEOReviewVerdict(guardVerdict, ceoReview)
	executionPlan := buildExecutionPlan(executionPlanInput{
		Packet:  packet,
		Results: results,
		Checks:  checkResults,
		Verdict: finalVerdict,
	})
	if hasWorkspace && !req.DryRun {
		written, err := writeExecutionPlanArtifact(ctx, artifactStore, executionPlan)
		if err != nil {
			return Report{}, err
		}
		changedFiles = append(changedFiles, written...)
		requiredEvidence, err := writeRequiredEvidenceArtifacts(ctx, space, packet, changedFiles, checkResults)
		if err != nil {
			return Report{}, err
		}
		changedFiles = append(changedFiles, requiredEvidence...)
	}
	verificationContract := NewVerificationContract(checkCommands(req), checkResults)
	report := buildReport(reportBuildInput{
		Packet:                          packet,
		Delegation:                      ceoDelegation,
		Continuation:                    req.Continuation,
		Workspace:                       workspaceBrief,
		Resume:                          resume,
		Results:                         results,
		ChangedFiles:                    changedFiles,
		Checks:                          checkResults,
		VerificationContract:            verificationContract,
		ProviderRouteDecisions:          req.ProviderRouteDecisions,
		Summary:                         summary,
		Plan:                            executionPlan,
		Patches:                         patchResults,
		Previews:                        patchPreviews,
		PreviewEvents:                   patchPreviewEvents,
		PatchAudit:                      patchAudit,
		PatchApproval:                   patchApproval,
		CEOReview:                       ceoReview,
		RetryHistory:                    retryHistory,
		SubagentConcurrency:             req.SubagentConcurrency,
		MaxToolRequests:                 req.MaxToolRequests,
		MaxOutputBytes:                  req.MaxSubagentOutputBytes,
		NoProgressStop:                  noProgressStop,
		MaxCEOIterations:                iterationBudget.max,
		CEOIterationCount:               iterationBudget.used,
		CEOIterationExhausted:           iterationBudget.exhausted,
		DryRun:                          req.DryRun,
		ProviderHealthAvoidedRouteCount: req.ProviderHealthAvoidedRouteCount,
		ProviderHealthAvoidedProviders:  req.ProviderHealthAvoidedProviders,
		Verdict:                         finalVerdict,
	})

	return persistRuntimeReport(ctx, artifactStore, req.DryRun, report)
}
