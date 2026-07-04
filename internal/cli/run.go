package cli

import (
	"context"
	"errors"
	"io"
	"strings"

	"ceoharness/internal/ceo"
	"ceoharness/internal/eval"
)

var (
	ErrVerdictFailed     = errors.New("CEO verdict failed")
	ErrVerdictNeedsInput = errors.New("CEO verdict needs input")
)

func Run(ctx context.Context, out io.Writer, args []string) error {
	return RunWithIO(ctx, strings.NewReader(""), out, args)
}

func RunWithIO(ctx context.Context, in io.Reader, out io.Writer, args []string) error {
	if len(args) > 0 && args[0] == "eval" {
		return eval.RunCLI(ctx, out, out, args[1:])
	}
	if len(args) > 0 && args[0] == "gauntlet" {
		if verbHelpRequested(args[1:]) {
			return runHelp(out)
		}
		return eval.RunCLI(ctx, out, out, gauntletEvalArgs(args[1:]))
	}
	opts, err := parseArgs(args)
	if err != nil {
		return err
	}
	if opts.showHelp {
		return runHelp(out)
	}
	if opts.showAdvancedHelp {
		return runAdvancedHelp(out)
	}
	if opts.showVersion {
		return runVersion(out)
	}
	if err := validateOptions(opts); err != nil {
		return err
	}
	if opts.quickstartDir != "" {
		return runQuickstart(ctx, out, opts)
	}
	if opts.startDir != "" {
		return runStart(ctx, out, opts)
	}
	if opts.initDemoRepoDir != "" {
		return runInitDemoRepo(ctx, out, opts)
	}
	if opts.providerWizardPreset != "" {
		return runProviderWizard(ctx, out, opts)
	}
	if opts.doctorProviderName != "" {
		return runNamedProviderDoctor(ctx, out, opts)
	}
	if opts.showProviderHealth {
		return runProviderHealthRollup(ctx, out, historyQuery{
			workspaceDir:   opts.workspaceDir,
			verdict:        opts.historyVerdict,
			task:           opts.historyTask,
			limit:          opts.historyLimit,
			summaryOnly:    opts.historySummaryOnly,
			since:          opts.historySince,
			until:          opts.historyUntil,
			provider:       opts.providerFilter,
			recommendation: opts.recommendationFilter,
			topProviders:   opts.topProviders,
		})
	}
	if opts.showProductionStatus {
		return runProductionStatus(out, opts)
	}
	if opts.showReviewQueue {
		return runReviewQueue(ctx, out, reviewQueueRequestFromOptions(opts))
	}
	if opts.showInbox {
		return runInbox(ctx, out, opts)
	}
	if opts.showHistory {
		return runHistory(ctx, out, historyQuery{
			workspaceDir: opts.workspaceDir,
			verdict:      opts.historyVerdict,
			task:         opts.historyTask,
			limit:        opts.historyLimit,
			summaryOnly:  opts.historySummaryOnly,
			since:        opts.historySince,
			until:        opts.historyUntil,
		})
	}
	if opts.jobID != "" {
		return runJobLookup(ctx, out, opts.workspaceDir, opts.jobID)
	}
	if opts.jobContextID != "" {
		return runJobContextLookup(ctx, out, opts.workspaceDir, opts.jobContextID, opts.reportFormat)
	}
	if opts.contextTraceID != "" {
		return runContextTraceLookup(ctx, out, opts.workspaceDir, opts.contextTraceID, opts.reportFormat)
	}
	if opts.jobReportID != "" {
		return runJobReportLookup(ctx, out, opts.workspaceDir, opts.jobReportID)
	}
	if opts.explainFailureJobID != "" {
		return runExplainFailure(ctx, out, opts.workspaceDir, opts.explainFailureJobID)
	}
	if opts.jobEventsID != "" {
		return runJobEventsLookup(ctx, out, opts.workspaceDir, opts.jobEventsID)
	}
	if opts.judgeJobID != "" {
		return runHumanJudgment(ctx, out, opts)
	}
	if opts.initConfig {
		return runConfigInit(ctx, out, opts)
	}
	if opts.showConfigCompletions {
		return runConfigCompletions(out, opts)
	}
	if opts.showConfigExplain {
		return runConfigExplain(out, opts)
	}
	if opts.showConfigDoctor {
		return runConfigDoctor(ctx, out, opts)
	}
	if opts.showConfigCheck {
		return runConfigCheck(ctx, out, opts)
	}
	if opts.showDoctor {
		return runDoctor(ctx, out, opts)
	}
	if opts.planOnly {
		return runPlanOnly(ctx, out, opts)
	}
	if opts.showDemo {
		return runDemo(ctx, out, opts)
	}
	if opts.showTUI {
		return runTUI(ctx, in, out, opts)
	}
	if opts.rollbackReportPath != "" {
		return runRollbackReport(ctx, out, opts)
	}
	if opts.interactive {
		opts, err = optionsWithInteractiveFormat(opts)
		if err != nil {
			return err
		}
		return runInteractive(ctx, in, out, opts)
	}
	return runOnce(ctx, out, opts)
}

func runOnce(ctx context.Context, out io.Writer, opts options) error {
	report, err := buildRunReport(ctx, opts)
	if err != nil {
		return err
	}
	if err := writeRunReport(out, reportOutputRequest{
		Report:       report,
		Format:       opts.reportFormat,
		WorkspaceDir: opts.workspaceDir,
	}); err != nil {
		return err
	}
	return verdictError(report)
}

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
