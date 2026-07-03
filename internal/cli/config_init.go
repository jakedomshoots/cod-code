package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"ceoharness/internal/config"
)

type configInitReport struct {
	ConfigPath                                string  `json:"config_path"`
	Created                                   bool    `json:"created"`
	ExampleAdapters                           bool    `json:"example_adapters"`
	Adapter                                   string  `json:"adapter,omitempty"`
	ModelCommandArgc                          int     `json:"model_command_argc"`
	CEOModelCommandArgc                       int     `json:"ceo_model_command_argc"`
	CEOProvider                               string  `json:"ceo_provider,omitempty"`
	ResearchCommandArgc                       int     `json:"research_command_argc"`
	ModelCommandTimeoutMS                     int     `json:"model_command_timeout_ms"`
	ToolCommandTimeoutMS                      int     `json:"tool_command_timeout_ms"`
	JobTimeoutMS                              int     `json:"job_timeout_ms"`
	CheckCommandArgc                          int     `json:"check_command_argc"`
	RequireChecks                             bool    `json:"require_checks"`
	CheckAttempts                             int     `json:"check_attempts"`
	CheckBackoffMS                            int     `json:"check_backoff_ms"`
	CEORevisionAttempts                       int     `json:"ceo_revision_attempts"`
	MaxCEOIterations                          int     `json:"max_ceo_iterations"`
	MaxSubagents                              int     `json:"max_subagents"`
	SubagentConcurrency                       int     `json:"subagent_concurrency"`
	MaxToolRequests                           int     `json:"max_tool_requests"`
	SubagentAttempts                          int     `json:"subagent_attempts"`
	SubagentBackoffMS                         int     `json:"subagent_backoff_ms"`
	NoProgressStop                            int     `json:"no_progress_stop"`
	MaxContextBytes                           int     `json:"max_context_bytes"`
	MaxSubagentOutputBytes                    int     `json:"max_subagent_output_bytes"`
	MinSubagentConfidence                     float64 `json:"min_subagent_confidence"`
	WorkspaceBriefMaxFiles                    int     `json:"workspace_brief_max_files"`
	WorkspaceBriefExcludeCount                int     `json:"workspace_brief_exclude_count"`
	ProviderHealthAvoidFailureRate            float64 `json:"provider_health_avoid_failure_rate"`
	ProviderHealthWatchFailureRate            float64 `json:"provider_health_watch_failure_rate"`
	ProviderHealthWatchCostPerAttemptMicroUSD int64   `json:"provider_health_watch_cost_per_attempt_microusd"`
	HTTPProviderCount                         int     `json:"http_provider_count"`
	AgentProviderCount                        int     `json:"agent_provider_count"`
	ProviderPolicyRuleCount                   int     `json:"provider_policy_rule_count"`
	WritePolicy                               string  `json:"write_policy,omitempty"`
}

func runConfigInit(ctx context.Context, out io.Writer, opts options) error {
	report, err := buildConfigInitReport(ctx, opts)
	if err != nil {
		return err
	}
	encoder := json.NewEncoder(out)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(report); err != nil {
		return fmt.Errorf("write config init report: %w", err)
	}
	return nil
}

func buildConfigInitReport(ctx context.Context, opts options) (configInitReport, error) {
	command := opts.modelCommand
	ceoCommand := opts.ceoModelCommand
	researchCommand := opts.researchCommand
	if opts.initExampleAdapters {
		adapters, err := exampleAdapterCommands()
		if err != nil {
			return configInitReport{}, err
		}
		if len(command) == 0 {
			command = adapters.ModelCommand
		}
		if len(ceoCommand) == 0 {
			ceoCommand = adapters.CEOModelCommand
		}
		if len(researchCommand) == 0 {
			researchCommand = adapters.ResearchCommand
		}
	}
	adapterName := strings.TrimSpace(opts.adapterName)
	if adapterName != "" && len(command) == 0 {
		adapterCommand, err := commandForExternalAdapter(adapterName)
		if err != nil {
			return configInitReport{}, err
		}
		command = adapterCommand
	}
	if len(command) == 0 {
		var err error
		command, err = modelCommandFromEnv()
		if err != nil {
			return configInitReport{}, err
		}
	}
	if len(ceoCommand) == 0 {
		var err error
		ceoCommand, err = ceoModelCommandFromEnv()
		if err != nil {
			return configInitReport{}, err
		}
	}
	if len(researchCommand) == 0 {
		var err error
		researchCommand, err = researchCommandFromEnv()
		if err != nil {
			return configInitReport{}, err
		}
	}
	providers, agentProviders, err := initHTTPProviders(opts)
	if err != nil {
		return configInitReport{}, err
	}
	path, err := config.CreateWorkspace(ctx, opts.workspaceDir, config.Config{
		ModelCommand:          command,
		CEOModelCommand:       ceoCommand,
		ResearchCommand:       researchCommand,
		ModelCommandTimeoutMS: opts.modelCommandTimeoutMS,
		ToolCommandTimeoutMS:  opts.toolCommandTimeoutMS,
		JobTimeoutMS:          opts.jobTimeoutMS,
		Providers:             providers,
		CEOProvider:           strings.TrimSpace(opts.ceoProvider),
		AgentProviders:        agentProviders,
		ProviderPolicy: config.ProviderPolicy{
			DefaultProvider:   strings.TrimSpace(opts.defaultProvider),
			FallbackProvider:  strings.TrimSpace(opts.fallbackProvider),
			RiskProviders:     cloneStringMap(opts.riskProviders),
			KindProviders:     cloneStringMap(opts.kindProviders),
			RiskAreaProviders: cloneStringMap(opts.riskAreaProviders),
		},
		CheckCommand:                              opts.checkCommand,
		RequireChecks:                             opts.requireChecks,
		CheckAttempts:                             opts.checkAttempts,
		CheckBackoffMS:                            opts.checkBackoffMS,
		CEORevisionAttempts:                       opts.ceoRevisionAttempts,
		MaxCEOIterations:                          opts.maxCEOIterations,
		MaxSubagents:                              opts.maxSubagents,
		SubagentConcurrency:                       opts.subagentConcurrency,
		MaxToolRequests:                           opts.maxToolRequests,
		SubagentAttempts:                          opts.subagentAttempts,
		SubagentBackoffMS:                         opts.subagentBackoffMS,
		NoProgressStop:                            opts.noProgressStop,
		MaxContextBytes:                           opts.maxContextBytes,
		MaxSubagentOutputBytes:                    opts.maxSubagentOutputBytes,
		MinSubagentConfidence:                     opts.minSubagentConfidence,
		WorkspaceBriefMaxFiles:                    opts.workspaceBriefMaxFiles,
		WorkspaceBriefExcludes:                    append([]string(nil), opts.workspaceBriefExcludes...),
		WritePolicy:                               strings.TrimSpace(opts.writePolicy),
		ProviderHealthAvoidFailureRate:            opts.providerHealthAvoidRate,
		ProviderHealthWatchFailureRate:            opts.providerHealthWatchRate,
		ProviderHealthWatchCostPerAttemptMicroUSD: opts.providerHealthWatchCostPerAttemptMicroUSD,
	})
	if err != nil {
		return configInitReport{}, err
	}
	return configInitReport{
		ConfigPath:                     path,
		Created:                        true,
		ExampleAdapters:                opts.initExampleAdapters,
		Adapter:                        adapterName,
		ModelCommandArgc:               len(command),
		CEOModelCommandArgc:            len(ceoCommand),
		CEOProvider:                    strings.TrimSpace(opts.ceoProvider),
		ResearchCommandArgc:            len(researchCommand),
		ModelCommandTimeoutMS:          opts.modelCommandTimeoutMS,
		ToolCommandTimeoutMS:           opts.toolCommandTimeoutMS,
		JobTimeoutMS:                   opts.jobTimeoutMS,
		CheckCommandArgc:               len(opts.checkCommand),
		RequireChecks:                  opts.requireChecks,
		CheckAttempts:                  opts.checkAttempts,
		CheckBackoffMS:                 opts.checkBackoffMS,
		CEORevisionAttempts:            opts.ceoRevisionAttempts,
		MaxCEOIterations:               opts.maxCEOIterations,
		MaxSubagents:                   opts.maxSubagents,
		SubagentConcurrency:            opts.subagentConcurrency,
		MaxToolRequests:                opts.maxToolRequests,
		SubagentAttempts:               opts.subagentAttempts,
		SubagentBackoffMS:              opts.subagentBackoffMS,
		NoProgressStop:                 opts.noProgressStop,
		MaxContextBytes:                opts.maxContextBytes,
		MaxSubagentOutputBytes:         opts.maxSubagentOutputBytes,
		MinSubagentConfidence:          opts.minSubagentConfidence,
		WorkspaceBriefMaxFiles:         opts.workspaceBriefMaxFiles,
		WorkspaceBriefExcludeCount:     len(opts.workspaceBriefExcludes),
		WritePolicy:                    strings.TrimSpace(opts.writePolicy),
		ProviderHealthAvoidFailureRate: opts.providerHealthAvoidRate,
		ProviderHealthWatchFailureRate: opts.providerHealthWatchRate,
		ProviderHealthWatchCostPerAttemptMicroUSD: opts.providerHealthWatchCostPerAttemptMicroUSD,
		HTTPProviderCount:                         len(providers),
		AgentProviderCount:                        len(agentProviders),
		ProviderPolicyRuleCount:                   config.ProviderPolicy{DefaultProvider: opts.defaultProvider, FallbackProvider: opts.fallbackProvider, RiskProviders: opts.riskProviders, KindProviders: opts.kindProviders, RiskAreaProviders: opts.riskAreaProviders}.RuleCount(),
	}, nil
}

func cloneStringMap(values map[string]string) map[string]string {
	if len(values) == 0 {
		return nil
	}
	copied := make(map[string]string, len(values))
	for key, value := range values {
		copied[key] = value
	}
	return copied
}
