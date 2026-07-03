package cli

import (
	"context"
	"io"
	"strings"

	"ceoharness/internal/ceo"
	"ceoharness/internal/config"
	"ceoharness/internal/jobpacket"
)

type planOnlyReport struct {
	Mode                            string                         `json:"mode"`
	WorkspaceDir                    string                         `json:"workspace_dir,omitempty"`
	JobOwner                        string                         `json:"job_owner"`
	JobPacket                       jobpacket.Packet               `json:"job_packet"`
	ProviderRoutes                  map[string]string              `json:"provider_routes,omitempty"`
	ProviderRouteDecisions          []config.ProviderRouteDecision `json:"provider_route_decisions,omitempty"`
	RunLedger                       ceo.RunLedger                  `json:"run_ledger"`
	Continuation                    *planOnlyContinuation          `json:"continuation,omitempty"`
	ProviderHealthAvoidedRouteCount int                            `json:"provider_health_avoided_route_count,omitempty"`
	ProviderHealthAvoidedProviders  []string                       `json:"provider_health_avoided_providers,omitempty"`
	CheckCommandCount               int                            `json:"check_command_count"`
	VerificationContract            ceo.VerificationContract       `json:"verification_contract"`
	RequireChecks                   bool                           `json:"require_checks,omitempty"`
	ModelCommandSource              string                         `json:"model_command_source"`
	CEOModelCommandSource           string                         `json:"ceo_model_command_source"`
	CEOProvider                     string                         `json:"ceo_provider,omitempty"`
	CEOProviderPresent              bool                           `json:"ceo_provider_present"`
	ResearchCommandSource           string                         `json:"research_command_source"`
	MaxCEOIterations                int                            `json:"max_ceo_iterations,omitempty"`
	SubagentConcurrency             int                            `json:"subagent_concurrency,omitempty"`
	MaxSubagents                    int                            `json:"max_subagents,omitempty"`
	MaxToolRequests                 int                            `json:"max_tool_requests,omitempty"`
	NoProgressStop                  int                            `json:"no_progress_stop,omitempty"`
	MaxSubagentOutputBytes          int                            `json:"max_subagent_output_bytes,omitempty"`
	MinSubagentConfidence           float64                        `json:"min_subagent_confidence,omitempty"`
	WorkspaceBriefMaxFiles          int                            `json:"workspace_brief_max_files,omitempty"`
	ProviderCostBudgetMicroUSD      int64                          `json:"provider_cost_budget_microusd,omitempty"`
}

type planOnlyContinuation struct {
	JobID                  string `json:"job_id"`
	UseSavedDelegation     bool   `json:"use_saved_delegation"`
	PlannedSubagentCount   int    `json:"planned_subagent_count,omitempty"`
	ReusableSubagentCount  int    `json:"reusable_subagent_count,omitempty"`
	SavedDelegationPresent bool   `json:"saved_delegation_present"`
}

func runPlanOnly(ctx context.Context, out io.Writer, opts options) error {
	opts, err := optionsWithPlanContext(ctx, opts)
	if err != nil {
		return err
	}
	packet, err := jobpacket.BuildWithOptions(jobpacket.BuildOptions{
		Task:            opts.task,
		Subagents:       opts.subagents,
		MaxSubagents:    opts.maxSubagents,
		MaxContextBytes: opts.maxContextBytes,
	})
	if err != nil {
		return err
	}
	modelSelection, err := selectModelCommand(ctx, opts)
	if err != nil {
		return err
	}
	ceoSelection, err := selectCEOModelCommand(ctx, opts)
	if err != nil {
		return err
	}
	researchSelection, err := selectResearchCommand(ctx, opts)
	if err != nil {
		return err
	}
	owner := jobpacket.OwnerForPacket(packet)
	checkCommands := planCheckCommands(opts)
	verificationContract := ceo.NewPendingVerificationContract(checkCommands)
	report := planOnlyReport{
		Mode:                   "plan_only",
		WorkspaceDir:           opts.workspaceDir,
		JobOwner:               owner,
		JobPacket:              packet,
		ProviderRoutes:         cloneAgentProviders(modelSelection.agentProviderNames),
		ProviderRouteDecisions: append([]config.ProviderRouteDecision(nil), modelSelection.providerRouteDecisions...),
		RunLedger: ceo.NewRunLedger(ceo.RunLedgerInput{
			Owner:                  owner,
			Verdict:                "pending",
			NextAction:             "run",
			VerificationContract:   verificationContract,
			ProviderRouteDecisions: modelSelection.providerRouteDecisions,
		}),
		Continuation:                    buildPlanOnlyContinuation(opts, packet),
		ProviderHealthAvoidedRouteCount: modelSelection.providerHealthAvoidedRouteCount,
		ProviderHealthAvoidedProviders:  append([]string(nil), modelSelection.providerHealthAvoidedProviders...),
		CheckCommandCount:               len(checkCommands),
		VerificationContract:            verificationContract,
		RequireChecks:                   opts.requireChecks,
		ModelCommandSource:              commandSource(modelSelection.source, modelSelection.argv),
		CEOModelCommandSource:           ceoSelection.source,
		CEOProvider:                     ceoSelection.providerName,
		CEOProviderPresent:              ceoSelection.providerName != "",
		ResearchCommandSource:           researchSelection.source,
		MaxCEOIterations:                opts.maxCEOIterations,
		SubagentConcurrency:             opts.subagentConcurrency,
		MaxSubagents:                    opts.maxSubagents,
		MaxToolRequests:                 opts.maxToolRequests,
		NoProgressStop:                  opts.noProgressStop,
		MaxSubagentOutputBytes:          opts.maxSubagentOutputBytes,
		MinSubagentConfidence:           opts.minSubagentConfidence,
		WorkspaceBriefMaxFiles:          opts.workspaceBriefMaxFiles,
		ProviderCostBudgetMicroUSD:      opts.providerCostBudgetMicroUSD,
	}
	return writePlanOnlyReport(out, report, opts.reportFormat)
}

func buildPlanOnlyContinuation(opts options, packet jobpacket.Packet) *planOnlyContinuation {
	if opts.continuation == nil {
		return nil
	}
	jobID := strings.TrimSpace(opts.continuation.JobID)
	if jobID == "" {
		return nil
	}
	return &planOnlyContinuation{
		JobID:                  jobID,
		UseSavedDelegation:     opts.continuation.UseSavedDelegation,
		PlannedSubagentCount:   len(packet.Subagents),
		ReusableSubagentCount:  countReusablePlanOnlySubagents(packet, opts.continuation),
		SavedDelegationPresent: opts.continuation.SavedDelegation != nil,
	}
}

func countReusablePlanOnlySubagents(packet jobpacket.Packet, continuation *ceo.ContinuationContext) int {
	if continuation == nil {
		return 0
	}
	count := 0
	for _, agent := range packet.Subagents {
		for _, result := range continuation.ReusableResults {
			if ceo.CanReuseSubagentResult(agent, result) {
				count++
				break
			}
		}
	}
	return count
}

func optionsWithPlanContext(ctx context.Context, opts options) (options, error) {
	var err error
	opts, err = optionsWithResumeContext(ctx, opts)
	if err != nil {
		return options{}, err
	}
	opts, err = optionsWithContinueJob(ctx, opts)
	if err != nil {
		return options{}, err
	}
	opts, err = optionsWithRerunTask(ctx, opts)
	if err != nil {
		return options{}, err
	}
	opts, err = optionsWithPriorJobContext(ctx, opts)
	if err != nil {
		return options{}, err
	}
	opts, err = optionsWithWorkspaceDefaults(ctx, opts)
	if err != nil {
		return options{}, err
	}
	return opts, requireVerificationChecks(opts)
}

func planCheckCommandCount(opts options) int {
	return len(planCheckCommands(opts))
}

func planCheckCommands(opts options) [][]string {
	if len(opts.checkCommands) > 0 {
		return clonePlanCheckCommands(opts.checkCommands)
	}
	if len(opts.checkCommand) == 0 {
		return nil
	}
	return [][]string{append([]string(nil), opts.checkCommand...)}
}

func clonePlanCheckCommands(commands [][]string) [][]string {
	copied := make([][]string, 0, len(commands))
	for _, command := range commands {
		copied = append(copied, append([]string(nil), command...))
	}
	return copied
}
