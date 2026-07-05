package cli

import (
	"context"

	"ceoharness/internal/ceo"
)

func buildRunReport(ctx context.Context, opts options) (ceo.Report, error) {
	var err error
	opts, err = optionsWithResumeContext(ctx, opts)
	if err != nil {
		return ceo.Report{}, err
	}
	opts, err = optionsWithContinueJob(ctx, opts)
	if err != nil {
		return ceo.Report{}, err
	}
	opts, err = optionsWithRerunTask(ctx, opts)
	if err != nil {
		return ceo.Report{}, err
	}
	opts, err = optionsWithPriorJobContext(ctx, opts)
	if err != nil {
		return ceo.Report{}, err
	}
	opts, err = optionsWithWorkspaceDefaults(ctx, opts)
	if err != nil {
		return ceo.Report{}, err
	}
	opts, err = optionsWithWritePolicy(opts)
	if err != nil {
		return ceo.Report{}, err
	}
	if err := requireVerificationChecks(opts); err != nil {
		return ceo.Report{}, err
	}
	ctx, cancel := contextWithJobTimeout(ctx, opts.jobTimeoutMS)
	defer cancel()
	researchSelection, err := selectResearchCommand(ctx, opts)
	if err != nil {
		return ceo.Report{}, err
	}

	runtimeResult, err := runtimeFromOptions(ctx, opts)
	if err != nil {
		return ceo.Report{}, err
	}
	report, err := runtimeResult.runtime.RunJob(ctx, ceo.JobRequest{
		Task:                            opts.task,
		WorkspaceDir:                    opts.workspaceDir,
		ArtifactRoot:                    opts.artifactRoot,
		CheckCommand:                    opts.checkCommand,
		CheckCommands:                   opts.checkCommands,
		ResearchCommand:                 researchSelection.argv,
		ToolCommandTimeoutMS:            opts.toolCommandTimeoutMS,
		BrowserPolicy:                   opts.browserPolicy,
		BrowserCommand:                  opts.browserBackendCommand,
		ComputerPolicy:                  opts.computerPolicy,
		ComputerCommand:                 opts.computerBackendCommand,
		CheckAttempts:                   opts.checkAttempts,
		CheckBackoffMS:                  opts.checkBackoffMS,
		CheckFixAttempts:                opts.checkFixAttempts,
		CEORevisionAttempts:             opts.ceoRevisionAttempts,
		MaxCEOIterations:                opts.maxCEOIterations,
		MaxSubagents:                    opts.maxSubagents,
		SubagentConcurrency:             opts.subagentConcurrency,
		MaxToolRequests:                 opts.maxToolRequests,
		Subagents:                       opts.subagents,
		SubagentAttempts:                opts.subagentAttempts,
		SubagentBackoffMS:               opts.subagentBackoffMS,
		NoProgressStop:                  opts.noProgressStop,
		ProviderCostBudgetMicroUSD:      opts.providerCostBudgetMicroUSD,
		ProviderHealthPolicy:            opts.providerHealthPolicy,
		ProviderHealthAvoidedRouteCount: runtimeResult.providerHealthAvoidance.avoidedRouteCount,
		ProviderHealthAvoidedProviders:  runtimeResult.providerHealthAvoidance.avoidedProviders,
		ProviderRouteDecisions:          runtimeResult.providerRouteDecisions,
		MaxContextBytes:                 opts.maxContextBytes,
		MaxSubagentOutputBytes:          opts.maxSubagentOutputBytes,
		Continuation:                    opts.continuation,
		WorkspaceBriefMaxFiles:          opts.workspaceBriefMaxFiles,
		WorkspaceBriefExcludes:          opts.workspaceBriefExcludes,
		Resume:                          opts.resumeContext,
		Patches:                         opts.patches,
		ScorerFailedChecks:              opts.scorerFailedChecks,
		ApprovedPreviewDigest:           opts.approvedPreviewDigest,
		DryRun:                          opts.dryRun,
		ApplyModelPatches:               opts.applyModelPatches,
		PreviewModelPatches:             opts.previewModelPatches,
		MaxModelPatches:                 opts.maxModelPatches,
	})
	if err != nil {
		return ceo.Report{}, err
	}

	return report, nil
}
