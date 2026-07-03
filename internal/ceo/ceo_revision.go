package ceo

import (
	"context"
	"fmt"
	"strings"
	"time"

	"ceoharness/internal/checkrunner"
	"ceoharness/internal/jobpacket"
	"ceoharness/internal/subagent"
	"ceoharness/internal/workspace"
)

const ceoRevisionRole = "revise work after CEO review feedback"

type ceoRevisionState struct {
	Packet          jobpacket.Packet
	Request         JobRequest
	Space           workspace.Workspace
	HasWorkspace    bool
	ArtifactStore   runtimeArtifactStore
	Results         []subagent.Result
	ChangedFiles    []string
	CheckResults    []checkrunner.Result
	PatchResults    []workspace.ReplaceTextResult
	PatchPreviews   []workspace.ReplaceTextResult
	PatchAudit      []PatchAuditEntry
	WorkspaceBrief  string
	Summary         VerificationSummary
	GuardVerdict    string
	CEOReview       *CEOReview
	IterationBudget *ceoIterationBudget
	RetryHistory    []RetryHistoryEntry
}

type ceoRevisionRequest struct {
	Packet         jobpacket.Packet
	Request        JobRequest
	Review         CEOReview
	Attempt        int
	WorkspaceBrief string
	FailedChecks   []RepairFailureDetail
}

func (r Runtime) runCEORevisionLoop(ctx context.Context, state ceoRevisionState) (ceoRevisionState, error) {
	if state.Request.CEORevisionAttempts < 1 || !shouldRunCEORevision(state.GuardVerdict, state.CEOReview) {
		return state, nil
	}
	if !state.HasWorkspace {
		return state, fmt.Errorf("workspace is required for CEO revision attempts")
	}
	if !state.Request.ApplyModelPatches {
		return state, fmt.Errorf("CEO revision attempts require model patch application")
	}

	noProgress := newNoProgressTracker(state.Request.NoProgressStop)
	for attempt := 1; attempt <= state.Request.CEORevisionAttempts && shouldRunCEORevision(state.GuardVerdict, state.CEOReview); attempt++ {
		if !consumeCEOIterationBudget(state.IterationBudget) {
			return state, nil
		}
		failedChecks := ceoReviewFailureDetail(*state.CEOReview)
		prompt := buildCEORevisionTask(state.Packet.Task, *state.CEOReview, attempt, failedChecks)
		history := RetryHistoryEntry{
			Kind:             "ceo_revision",
			Attempt:          attempt,
			Status:           "fail",
			FailedChecks:     failedChecks,
			CorrectivePrompt: prompt,
		}
		result, err := r.runCoderCEORevision(ctx, ceoRevisionRequest{
			Packet:         state.Packet,
			Request:        state.Request,
			Review:         *state.CEOReview,
			Attempt:        attempt,
			WorkspaceBrief: state.WorkspaceBrief,
			FailedChecks:   failedChecks,
		})
		if err != nil {
			return state, err
		}
		result = compactSubagentOutput(result, state.Request.MaxSubagentOutputBytes)
		state.Results = append(state.Results, result)
		path, err := writeSubagentEvidenceFile(ctx, state.ArtifactStore.Space, fmt.Sprintf("coder-ceo-revision-%d", attempt), result)
		if err != nil {
			return state, err
		}
		if changedPath, ok := state.ArtifactStore.changedPath(path); ok {
			state.ChangedFiles = append(state.ChangedFiles, changedPath)
		}
		patches, err := coderPatchProposals(result)
		if err != nil {
			history.ModelPatchStatus = "invalid"
			history.Reason = err.Error()
			state.RetryHistory = append(state.RetryHistory, history)
			if noProgress.observe("invalid:" + err.Error()) {
				markLastRetryNoProgress(state.Results, state.RetryHistory)
				break
			}
			continue
		}
		if len(patches) == 0 {
			history.ModelPatchStatus = "empty"
			history.Reason = "model returned no patch proposals"
			state.RetryHistory = append(state.RetryHistory, history)
			if noProgress.observe("empty") {
				markLastRetryNoProgress(state.Results, state.RetryHistory)
				break
			}
			continue
		}
		if patchesAreNoOp(patches) {
			history.ModelPatchStatus = "no_op"
			history.Reason = "model patch does not change content"
			state.RetryHistory = append(state.RetryHistory, history)
			if noProgress.observe("no_op:" + patchSignature(patches)) {
				markLastRetryNoProgress(state.Results, state.RetryHistory)
				break
			}
			continue
		}
		failedPatchSignature := failedPatchNoProgressSignature(patches)
		if noProgress.hasObserved(failedPatchSignature) && noProgress.observe(failedPatchSignature) {
			history.ModelPatchStatus = "identical"
			history.Reason = "model repeated an identical failed patch"
			state.RetryHistory = append(state.RetryHistory, history)
			markLastRetryNoProgress(state.Results, state.RetryHistory)
			break
		}
		if err := enforceModelPatchLimit(patches, state.Request.MaxModelPatches); err != nil {
			history.ModelPatchStatus = "invalid"
			history.Reason = err.Error()
			state.RetryHistory = append(state.RetryHistory, history)
			continue
		}
		applied, err := applyPatchRequests(ctx, state.Space, patches)
		if err != nil {
			history.ModelPatchStatus = "apply_failed"
			history.Reason = err.Error()
			state.RetryHistory = append(state.RetryHistory, history)
			if noProgress.observe(applyFailureNoProgressSignature(patches, err)) {
				markLastRetryNoProgress(state.Results, state.RetryHistory)
				break
			}
			continue
		}
		history.ModelPatchStatus = "applied"
		history.ChangedFiles = changedFilesFromPatchResults(applied)
		state.PatchResults = append(state.PatchResults, applied...)
		state.PatchAudit = append(state.PatchAudit, patchAuditEntries(applied, "model", "coder")...)
		state.ChangedFiles = appendChangedPatchFiles(state.ChangedFiles, applied)
		nextChecks, err := r.runChecks(ctx, state.Request)
		if err != nil {
			return state, err
		}
		state.CheckResults = append(state.CheckResults, nextChecks...)
		state.Summary = summarizeVerificationWithPolicy(state.Results, state.CheckResults, state.Request.ProviderHealthPolicy)
		state.Summary = applyProviderCostBudget(state.Summary, state.Request.ProviderCostBudgetMicroUSD)
		state.GuardVerdict = verdict(state.Results, state.CheckResults, state.Summary)
		if state.GuardVerdict != "pass" {
			history.FinalVerdict = state.GuardVerdict
			state.RetryHistory = append(state.RetryHistory, history)
			return state, nil
		}
		review, err := r.runCEOReview(ctx, ceoReviewInput{
			Packet:        state.Packet,
			Results:       state.Results,
			Checks:        state.CheckResults,
			ChangedFiles:  state.ChangedFiles,
			PatchResults:  state.PatchResults,
			PatchPreviews: state.PatchPreviews,
			GuardVerdict:  state.GuardVerdict,
		})
		if err != nil {
			return state, err
		}
		state.CEOReview = review
		if state.CEOReview == nil {
			history.Status = "pass"
			history.FinalVerdict = "pass"
			state.RetryHistory = append(state.RetryHistory, history)
			return state, nil
		}
		if !shouldRunCEORevision(state.GuardVerdict, state.CEOReview) {
			history.Status = "pass"
			history.FinalVerdict = "pass"
		} else {
			history.FinalVerdict = "fail"
		}
		state.RetryHistory = append(state.RetryHistory, history)
		if history.FinalVerdict != "pass" && noProgress.observe(failedPatchSignature) {
			markLastRetryNoProgress(state.Results, state.RetryHistory)
			break
		}
	}
	return state, nil
}

func (r Runtime) runCoderCEORevision(ctx context.Context, req ceoRevisionRequest) (subagent.Result, error) {
	packet := req.Packet
	packet.Task = buildCEORevisionTask(req.Packet.Task, req.Review, req.Attempt, req.FailedChecks)
	agent := jobpacket.Subagent{
		Name:           "coder",
		Role:           ceoRevisionRole,
		AllowedActions: jobpacket.DefaultActionsForAgent("coder"),
	}
	attempts := req.Request.SubagentAttempts
	if attempts < 1 {
		attempts = 1
	}
	backoff := time.Duration(req.Request.SubagentBackoffMS) * time.Millisecond
	return r.runSubagentWithAttempts(ctx, subagentRunInput{
		Packet:         packet,
		Agent:          agent,
		Attempts:       attempts,
		Backoff:        backoff,
		NoProgressStop: req.Request.NoProgressStop,
		WorkspaceBrief: req.WorkspaceBrief,
	})
}

func buildCEORevisionTask(task string, review CEOReview, attempt int, failureDetails ...[]RepairFailureDetail) string {
	details := []RepairFailureDetail(nil)
	if len(failureDetails) > 0 {
		details = failureDetails[0]
	}
	scorerDetails := renderRepairFailureDetails(details)
	if scorerDetails != "" {
		scorerDetails = "\n" + scorerDetails
	}
	return fmt.Sprintf(
		"%s\n\nCEO review failed. Revision attempt %d.\nCEO feedback: %s%s\nReturn only coder patch JSON.",
		task,
		attempt,
		trimCEORevisionFeedback(review.Summary),
		scorerDetails,
	)
}

func shouldRunCEORevision(guardVerdict string, review *CEOReview) bool {
	return guardVerdict == "pass" && review != nil && review.RecommendedVerdict == "fail"
}

func trimCEORevisionFeedback(feedback string) string {
	const maxCEORevisionFeedbackBytes = 800
	clean := strings.TrimSpace(feedback)
	if len(clean) <= maxCEORevisionFeedbackBytes {
		return clean
	}
	return clean[:maxCEORevisionFeedbackBytes] + "\n[truncated]"
}
