package ceo

import (
	"ceoharness/internal/checkrunner"
	"ceoharness/internal/config"
	"ceoharness/internal/history"
	"ceoharness/internal/jobpacket"
	"ceoharness/internal/subagent"
	"ceoharness/internal/workspace"
)

type reportBuildInput struct {
	Packet                          jobpacket.Packet
	Delegation                      *CEODelegation
	Continuation                    *ContinuationContext
	Workspace                       *workspace.Brief
	Resume                          *ResumeContext
	Results                         []subagent.Result
	ChangedFiles                    []string
	Checks                          []checkrunner.Result
	VerificationContract            VerificationContract
	ProviderRouteDecisions          []config.ProviderRouteDecision
	Summary                         VerificationSummary
	Plan                            ExecutionPlan
	Patches                         []workspace.ReplaceTextResult
	Previews                        []workspace.ReplaceTextResult
	PreviewEvents                   []PatchPreviewEvent
	PatchAudit                      []PatchAuditEntry
	PatchApproval                   *PatchApproval
	CEOReview                       *CEOReview
	RetryHistory                    []RetryHistoryEntry
	SubagentConcurrency             int
	MaxToolRequests                 int
	MaxOutputBytes                  int
	NoProgressStop                  int
	MaxCEOIterations                int
	CEOIterationCount               int
	CEOIterationExhausted           bool
	DryRun                          bool
	ProviderHealthAvoidedRouteCount int
	ProviderHealthAvoidedProviders  []string
	Verdict                         string
}

func buildReport(input reportBuildInput) Report {
	owner := jobpacket.OwnerForPacket(input.Packet)
	lifecycle := buildLifecycle(lifecycleInput{
		Recovered:      input.Continuation != nil || input.Resume != nil,
		ResultStatuses: subagentResultStatuses(input.Results),
		CheckStatuses:  checkResultStatuses(input.Checks),
		PreviewCount:   len(input.Previews),
		AppliedCount:   len(input.Patches),
		Verdict:        input.Verdict,
	})
	return Report{
		SchemaVersion:          history.ReportSchemaVersion,
		JobPacket:              input.Packet,
		JobOwner:               owner,
		LifecycleState:         lifecycle.State,
		LifecycleEvents:        lifecycle.Events,
		VerificationContract:   input.VerificationContract,
		ProviderRouteDecisions: append([]config.ProviderRouteDecision(nil), input.ProviderRouteDecisions...),
		ContextTrace:           buildContextTrace(input.Packet, input.Workspace, input.Results),
		RunLedger: NewRunLedger(RunLedgerInput{
			Owner:                  owner,
			Verdict:                input.Verdict,
			NextAction:             input.Plan.NextAction,
			VerificationContract:   input.VerificationContract,
			ChangedFiles:           input.ChangedFiles,
			ProviderRouteDecisions: input.ProviderRouteDecisions,
		}),
		RunManifest: buildRunManifest(runManifestInput{
			Packet:                          input.Packet,
			SubagentCount:                   len(input.Results),
			ReusedSubagentCount:             reusedSubagentCount(input.Results),
			ChangedFileCount:                len(input.ChangedFiles),
			CheckAttemptCount:               len(input.Checks),
			PatchCount:                      len(input.Patches),
			SubagentConcurrency:             input.SubagentConcurrency,
			MaxToolRequests:                 input.MaxToolRequests,
			MaxOutputBytes:                  input.MaxOutputBytes,
			NoProgressStop:                  input.NoProgressStop,
			MaxCEOIterations:                input.MaxCEOIterations,
			CEOIterationCount:               input.CEOIterationCount,
			CEOIterationExhausted:           input.CEOIterationExhausted,
			DryRun:                          input.DryRun,
			ProviderHealthAvoidedRouteCount: input.ProviderHealthAvoidedRouteCount,
			ProviderHealthAvoidedProviders:  input.ProviderHealthAvoidedProviders,
			Verdict:                         input.Verdict,
		}),
		Continuation: reportContinuation(input.Continuation, input.Results),
		RunEvents: buildRunEvents(runEventsInput{
			Packet:                          input.Packet,
			Delegation:                      input.Delegation,
			WorkspaceBrief:                  input.Workspace,
			ProviderHealthAvoidedRouteCount: input.ProviderHealthAvoidedRouteCount,
			ProviderHealthAvoidedProviders:  input.ProviderHealthAvoidedProviders,
			Results:                         input.Results,
			Checks:                          input.Checks,
			PatchAudit:                      input.PatchAudit,
			PatchPreviewEvents:              input.PreviewEvents,
			PatchApproval:                   input.PatchApproval,
			CEOReview:                       input.CEOReview,
			LifecycleEvents:                 lifecycle.Events,
			Verdict:                         input.Verdict,
		}),
		WorkspaceBrief:      input.Workspace,
		Resume:              input.Resume,
		SubagentResults:     input.Results,
		ChangedFiles:        input.ChangedFiles,
		CheckResults:        input.Checks,
		VerificationSummary: input.Summary,
		ExecutionPlan:       input.Plan,
		PatchResults:        input.Patches,
		PatchPreviews:       input.Previews,
		PatchAudit:          input.PatchAudit,
		PatchApproval:       input.PatchApproval,
		CEODelegation:       input.Delegation,
		CEOReview:           input.CEOReview,
		RetryHistory:        input.RetryHistory,
		Verdict:             input.Verdict,
	}
}

func subagentResultStatuses(results []subagent.Result) []string {
	statuses := make([]string, 0, len(results))
	for _, result := range results {
		statuses = append(statuses, result.Status)
	}
	return statuses
}

func checkResultStatuses(results []checkrunner.Result) []string {
	statuses := make([]string, 0, len(results))
	for _, result := range results {
		statuses = append(statuses, result.Status)
	}
	return statuses
}
