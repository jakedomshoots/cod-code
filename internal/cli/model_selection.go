package cli

import (
	"context"

	"ceoharness/internal/config"
)

type modelCommandSelection struct {
	argv                                      []string
	agentArgv                                 map[string][]string
	agentEnvVars                              map[string][]string
	agentHTTPProviders                        map[string]config.HTTPProvider
	agentProviderNames                        map[string]string
	providerRouteDecisions                    []config.ProviderRouteDecision
	providerConfigs                           map[string]config.Provider
	fallbackArgv                              []string
	fallbackEnvVars                           []string
	fallbackHTTPProvider                      config.HTTPProvider
	fallbackProviderName                      string
	checkArgv                                 [][]string
	providerCount                             int
	providerHTTPCount                         int
	providerPolicyRuleCount                   int
	providerControls                          providerControlCounts
	providerEnvVars                           int
	providerEnvSet                            int
	providerEnvMiss                           int
	providerEnvMissingNames                   []string
	checkAttempts                             int
	checkBackoffMS                            int
	requireChecks                             bool
	ceoRevisionAttempts                       int
	maxCEOIterations                          int
	maxSubagents                              int
	subagentConcurrency                       int
	maxToolRequests                           int
	delegatedSubagentCount                    int
	subagentAttempts                          int
	subagentBackoffMS                         int
	noProgressStop                            int
	maxContextBytes                           int
	maxSubagentOutputBytes                    int
	minSubagentConfidence                     float64
	workspaceBriefMaxFiles                    int
	workspaceBriefExcludeCount                int
	providerCostBudgetMicroUSD                int64
	providerHealthAvoidFailureRate            float64
	providerHealthWatchFailureRate            float64
	providerHealthWatchCostPerAttemptMicroUSD int64
	providerHealthAvoidedRouteCount           int
	providerHealthAvoidedProviders            []string
	checkSetCount                             int
	autoCheckCount                            int
	defaultCheckSet                           string
	modelCommandTimeoutMS                     int
	toolCommandTimeoutMS                      int
	jobTimeoutMS                              int
	source                                    string
}

type commandSelection struct {
	argv         []string
	timeoutMS    int
	source       string
	providerName string
	provider     config.Provider
}

func selectModelCommand(ctx context.Context, opts options) (modelCommandSelection, error) {
	if len(opts.modelCommand) > 0 {
		return modelCommandSelection{argv: opts.modelCommand, requireChecks: opts.requireChecks, maxCEOIterations: opts.maxCEOIterations, maxSubagents: opts.maxSubagents, maxToolRequests: opts.maxToolRequests, noProgressStop: opts.noProgressStop, maxSubagentOutputBytes: opts.maxSubagentOutputBytes, minSubagentConfidence: opts.minSubagentConfidence, modelCommandTimeoutMS: opts.modelCommandTimeoutMS, toolCommandTimeoutMS: opts.toolCommandTimeoutMS, jobTimeoutMS: opts.jobTimeoutMS, source: "flag"}, nil
	}
	command, err := modelCommandFromEnv()
	if err != nil {
		return modelCommandSelection{}, err
	}
	if len(command) > 0 {
		return modelCommandSelection{argv: command, requireChecks: opts.requireChecks, maxCEOIterations: opts.maxCEOIterations, maxSubagents: opts.maxSubagents, maxToolRequests: opts.maxToolRequests, noProgressStop: opts.noProgressStop, maxSubagentOutputBytes: opts.maxSubagentOutputBytes, minSubagentConfidence: opts.minSubagentConfidence, modelCommandTimeoutMS: opts.modelCommandTimeoutMS, toolCommandTimeoutMS: opts.toolCommandTimeoutMS, jobTimeoutMS: opts.jobTimeoutMS, source: "env"}, nil
	}
	cfg, err := config.LoadWorkspace(ctx, opts.workspaceDir)
	if err != nil {
		return modelCommandSelection{}, err
	}
	agentProviderSelection, err := providerRoutesForSelection(ctx, cfg, opts)
	if err != nil {
		return modelCommandSelection{}, err
	}
	agentProviderNames := agentProviderSelection.routes
	agentCommands := cfg.AgentCommandsFor(agentProviderNames)
	agentEnvVars := cfg.AgentEnvVarsFor(agentProviderNames)
	agentHTTPProviders := cfg.AgentHTTPProvidersFor(agentProviderNames)
	fallbackRoute := providerFallbackRoute(cfg)
	providerEnvNames := cfg.ProviderEnvVarNames()
	providerEnvSet, providerEnvMiss, providerEnvMissingNames := providerEnvCounts(providerEnvNames)
	checkCommands := cfg.CheckCommandList()
	if len(checkCommands) == 0 && cfg.DefaultCheckSet != "" {
		checkCommands, _ = cfg.CheckCommandsForSet(cfg.DefaultCheckSet)
	}
	requireChecks := cfg.RequireChecks || opts.requireChecks
	minSubagentConfidence := cfg.MinSubagentConfidence
	if opts.minSubagentConfidence > 0 {
		minSubagentConfidence = opts.minSubagentConfidence
	}
	if len(cfg.ModelCommand) > 0 || cfg.ModelCommandTimeoutMS > 0 || cfg.ToolCommandTimeoutMS > 0 || cfg.HasJobPolicy() || len(cfg.Subagents) > 0 || len(agentCommands) > 0 || len(agentHTTPProviders) > 0 || len(cfg.Providers) > 0 || len(checkCommands) > 0 || len(cfg.CheckSets) > 0 || len(cfg.AutoCheckSets) > 0 || cfg.HasRetryPolicy() || cfg.HasVerificationPolicy() || cfg.HasToolPolicy() || cfg.HasSubagentBudget() || cfg.HasSubagentOutputPolicy() || cfg.HasConfidencePolicy() || cfg.HasCostPolicy() || cfg.HasContextPolicy() || cfg.HasProviderHealthPolicy() || cfg.HasProviderPolicy() || cfg.HasBrowserToolPolicy() || cfg.HasComputerToolPolicy() || cfg.HasExtensions() {
		return modelCommandSelection{
			argv:                           cfg.ModelCommand,
			agentArgv:                      agentCommands,
			agentEnvVars:                   agentEnvVars,
			agentHTTPProviders:             agentHTTPProviders,
			agentProviderNames:             agentProviderNames,
			providerRouteDecisions:         append([]config.ProviderRouteDecision(nil), agentProviderSelection.decisions...),
			providerConfigs:                cloneProviderConfigs(cfg.Providers),
			fallbackArgv:                   fallbackRoute.argv,
			fallbackEnvVars:                fallbackRoute.envVars,
			fallbackHTTPProvider:           fallbackRoute.httpProvider,
			fallbackProviderName:           fallbackRoute.providerName,
			checkArgv:                      checkCommands,
			providerCount:                  len(cfg.Providers),
			providerHTTPCount:              countHTTPProviders(cfg.Providers),
			providerPolicyRuleCount:        cfg.ProviderPolicy.RuleCount(),
			providerControls:               countHTTPProviderControls(cfg.Providers),
			providerEnvVars:                len(providerEnvNames),
			providerEnvSet:                 providerEnvSet,
			providerEnvMiss:                providerEnvMiss,
			providerEnvMissingNames:        providerEnvMissingNames,
			checkAttempts:                  cfg.CheckAttempts,
			checkBackoffMS:                 cfg.CheckBackoffMS,
			requireChecks:                  requireChecks,
			ceoRevisionAttempts:            cfg.CEORevisionAttempts,
			maxCEOIterations:               cfg.MaxCEOIterations,
			maxSubagents:                   cfg.MaxSubagents,
			subagentConcurrency:            cfg.SubagentConcurrency,
			maxToolRequests:                cfg.MaxToolRequests,
			delegatedSubagentCount:         len(cfg.Subagents),
			subagentAttempts:               cfg.SubagentAttempts,
			subagentBackoffMS:              cfg.SubagentBackoffMS,
			noProgressStop:                 cfg.NoProgressStop,
			maxContextBytes:                cfg.MaxContextBytes,
			maxSubagentOutputBytes:         cfg.MaxSubagentOutputBytes,
			minSubagentConfidence:          minSubagentConfidence,
			workspaceBriefMaxFiles:         cfg.WorkspaceBriefMaxFiles,
			workspaceBriefExcludeCount:     len(cfg.WorkspaceBriefExcludes),
			providerCostBudgetMicroUSD:     cfg.ProviderCostBudgetMicroUSD,
			providerHealthAvoidFailureRate: cfg.ProviderHealthAvoidFailureRate,
			providerHealthWatchFailureRate: cfg.ProviderHealthWatchFailureRate,
			providerHealthWatchCostPerAttemptMicroUSD: cfg.ProviderHealthWatchCostPerAttemptMicroUSD,
			providerHealthAvoidedRouteCount:           agentProviderSelection.healthAvoidance.avoidedRouteCount,
			providerHealthAvoidedProviders:            append([]string(nil), agentProviderSelection.healthAvoidance.avoidedProviders...),
			checkSetCount:                             len(cfg.CheckSets),
			autoCheckCount:                            len(cfg.AutoCheckSets),
			defaultCheckSet:                           cfg.DefaultCheckSet,
			modelCommandTimeoutMS:                     cfg.ModelCommandTimeoutMS,
			toolCommandTimeoutMS:                      cfg.ToolCommandTimeoutMS,
			jobTimeoutMS:                              cfg.JobTimeoutMS,
			source:                                    "workspace",
		}, nil
	}
	return modelCommandSelection{requireChecks: opts.requireChecks, maxCEOIterations: opts.maxCEOIterations, maxSubagents: opts.maxSubagents, maxToolRequests: opts.maxToolRequests, noProgressStop: opts.noProgressStop, maxSubagentOutputBytes: opts.maxSubagentOutputBytes, minSubagentConfidence: opts.minSubagentConfidence, modelCommandTimeoutMS: opts.modelCommandTimeoutMS, toolCommandTimeoutMS: opts.toolCommandTimeoutMS, jobTimeoutMS: opts.jobTimeoutMS, source: "default"}, nil
}
