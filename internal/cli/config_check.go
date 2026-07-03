package cli

import (
	"context"
	"io"
	"path/filepath"
	"strings"
	"time"

	"ceoharness/internal/adapter"
	"ceoharness/internal/config"
)

type configCheckReport struct {
	WorkspaceDir                              string           `json:"workspace_dir"`
	ConfigPath                                string           `json:"config_path"`
	ModelCommandSource                        string           `json:"model_command_source"`
	ModelCommandArgc                          int              `json:"model_command_argc"`
	ModelCommandPresent                       bool             `json:"model_command_present"`
	CEOModelCommandSource                     string           `json:"ceo_model_command_source"`
	CEOModelCommandArgc                       int              `json:"ceo_model_command_argc"`
	CEOModelCommandPresent                    bool             `json:"ceo_model_command_present"`
	CEOProvider                               string           `json:"ceo_provider,omitempty"`
	CEOProviderPresent                        bool             `json:"ceo_provider_present"`
	ResearchCommandSource                     string           `json:"research_command_source"`
	ResearchCommandArgc                       int              `json:"research_command_argc"`
	ResearchCommandPresent                    bool             `json:"research_command_present"`
	ModelCommandTimeoutMS                     int              `json:"model_command_timeout_ms"`
	ToolCommandTimeoutMS                      int              `json:"tool_command_timeout_ms"`
	JobTimeoutMS                              int              `json:"job_timeout_ms"`
	AgentModelCommandCount                    int              `json:"agent_model_command_count"`
	ProviderCount                             int              `json:"provider_count"`
	ProviderHTTPCount                         int              `json:"provider_http_count"`
	ProviderPolicyRuleCount                   int              `json:"provider_policy_rule_count"`
	ProviderFallbackProvider                  string           `json:"provider_fallback_provider,omitempty"`
	ProviderHTTPCostCount                     int              `json:"provider_http_cost_count"`
	ProviderHTTPTimeoutCount                  int              `json:"provider_http_timeout_count"`
	ProviderHTTPMaxOutputTokensCount          int              `json:"provider_http_max_output_tokens_count"`
	ProviderHTTPResponseFormatCount           int              `json:"provider_http_response_format_count"`
	ProviderEnvVarCount                       int              `json:"provider_env_var_count"`
	ProviderEnvVarPresentCount                int              `json:"provider_env_var_present_count"`
	ProviderEnvVarMissingCount                int              `json:"provider_env_var_missing_count"`
	ProviderEnvVarMissingNames                []string         `json:"provider_env_var_missing_names,omitempty"`
	ProviderCostBudgetMicroUSD                int64            `json:"provider_cost_budget_microusd"`
	ProviderHealthAvoidFailureRate            float64          `json:"provider_health_avoid_failure_rate"`
	ProviderHealthWatchFailureRate            float64          `json:"provider_health_watch_failure_rate"`
	ProviderHealthWatchCostPerAttemptMicroUSD int64            `json:"provider_health_watch_cost_per_attempt_microusd"`
	ProviderHealthAvoidedRouteCount           int              `json:"provider_health_avoided_route_count"`
	ProviderHealthAvoidedProviders            []string         `json:"provider_health_avoided_providers,omitempty"`
	ProviderSetupSteps                        []string         `json:"provider_setup_steps,omitempty"`
	AdapterCapabilities                       []adapter.Report `json:"adapter_capabilities"`
	CheckAttempts                             int              `json:"check_attempts"`
	CheckBackoffMS                            int              `json:"check_backoff_ms"`
	RequireChecks                             bool             `json:"require_checks"`
	CEORevisionAttempts                       int              `json:"ceo_revision_attempts"`
	MaxCEOIterations                          int              `json:"max_ceo_iterations"`
	MaxSubagents                              int              `json:"max_subagents"`
	DelegatedSubagentCount                    int              `json:"delegated_subagent_count"`
	SubagentConcurrency                       int              `json:"subagent_concurrency"`
	MaxToolRequests                           int              `json:"max_tool_requests"`
	SubagentAttempts                          int              `json:"subagent_attempts"`
	SubagentBackoffMS                         int              `json:"subagent_backoff_ms"`
	NoProgressStop                            int              `json:"no_progress_stop"`
	MaxContextBytes                           int              `json:"max_context_bytes"`
	MaxSubagentOutputBytes                    int              `json:"max_subagent_output_bytes"`
	MinSubagentConfidence                     float64          `json:"min_subagent_confidence"`
	WorkspaceBriefMaxFiles                    int              `json:"workspace_brief_max_files"`
	WorkspaceBriefExcludeCount                int              `json:"workspace_brief_exclude_count"`
	CheckCommandArgc                          int              `json:"check_command_argc"`
	CheckCommandPresent                       bool             `json:"check_command_present"`
	CheckCommandCount                         int              `json:"check_command_count"`
	CheckSetCount                             int              `json:"check_set_count"`
	AutoCheckSetCount                         int              `json:"auto_check_set_count"`
	DefaultCheckSet                           string           `json:"default_check_set"`
}

func runConfigCheck(ctx context.Context, out io.Writer, opts options) error {
	report, err := buildConfigCheckReport(ctx, opts)
	if err != nil {
		return err
	}
	return writeConfigCheckReport(out, report, opts.reportFormat)
}

func buildConfigCheckReport(ctx context.Context, opts options) (configCheckReport, error) {
	selection, err := selectModelCommand(ctx, opts)
	if err != nil {
		return configCheckReport{}, err
	}
	ceoSelection, err := selectCEOModelCommand(ctx, opts)
	if err != nil {
		return configCheckReport{}, err
	}
	researchSelection, err := selectResearchCommand(ctx, opts)
	if err != nil {
		return configCheckReport{}, err
	}
	adapterTimeout := time.Duration(selection.toolCommandTimeoutMS) * time.Millisecond
	return configCheckReport{
		WorkspaceDir:                              opts.workspaceDir,
		ConfigPath:                                workspaceConfigPath(opts.workspaceDir),
		ModelCommandSource:                        commandSource(selection.source, selection.argv),
		ModelCommandArgc:                          len(selection.argv),
		ModelCommandPresent:                       len(selection.argv) > 0,
		CEOModelCommandSource:                     ceoSelection.source,
		CEOModelCommandArgc:                       len(ceoSelection.argv),
		CEOModelCommandPresent:                    len(ceoSelection.argv) > 0,
		CEOProvider:                               ceoSelection.providerName,
		CEOProviderPresent:                        ceoSelection.providerName != "",
		ResearchCommandSource:                     researchSelection.source,
		ResearchCommandArgc:                       len(researchSelection.argv),
		ResearchCommandPresent:                    len(researchSelection.argv) > 0,
		ModelCommandTimeoutMS:                     selection.modelCommandTimeoutMS,
		ToolCommandTimeoutMS:                      selection.toolCommandTimeoutMS,
		JobTimeoutMS:                              selection.jobTimeoutMS,
		AgentModelCommandCount:                    len(selection.agentArgv),
		ProviderCount:                             selection.providerCount,
		ProviderHTTPCount:                         selection.providerHTTPCount,
		ProviderPolicyRuleCount:                   selection.providerPolicyRuleCount,
		ProviderFallbackProvider:                  selection.fallbackProviderName,
		ProviderHTTPCostCount:                     selection.providerControls.Cost,
		ProviderHTTPTimeoutCount:                  selection.providerControls.Timeout,
		ProviderHTTPMaxOutputTokensCount:          selection.providerControls.MaxOutputTokens,
		ProviderHTTPResponseFormatCount:           selection.providerControls.ResponseFormat,
		ProviderEnvVarCount:                       selection.providerEnvVars,
		ProviderEnvVarPresentCount:                selection.providerEnvSet,
		ProviderEnvVarMissingCount:                selection.providerEnvMiss,
		ProviderEnvVarMissingNames:                append([]string(nil), selection.providerEnvMissingNames...),
		ProviderCostBudgetMicroUSD:                selection.providerCostBudgetMicroUSD,
		ProviderHealthAvoidFailureRate:            selection.providerHealthAvoidFailureRate,
		ProviderHealthWatchFailureRate:            selection.providerHealthWatchFailureRate,
		ProviderHealthWatchCostPerAttemptMicroUSD: selection.providerHealthWatchCostPerAttemptMicroUSD,
		ProviderHealthAvoidedRouteCount:           selection.providerHealthAvoidedRouteCount,
		ProviderHealthAvoidedProviders:            append([]string(nil), selection.providerHealthAvoidedProviders...),
		ProviderSetupSteps:                        providerSetupSteps(selection, opts.workspaceDir),
		AdapterCapabilities:                       adapter.DoctorAll(ctx, adapter.DoctorOptions{Timeout: adapterTimeout}),
		CheckAttempts:                             selection.checkAttempts,
		CheckBackoffMS:                            selection.checkBackoffMS,
		RequireChecks:                             selection.requireChecks,
		CEORevisionAttempts:                       selection.ceoRevisionAttempts,
		MaxCEOIterations:                          selection.maxCEOIterations,
		MaxSubagents:                              selection.maxSubagents,
		DelegatedSubagentCount:                    selection.delegatedSubagentCount,
		SubagentConcurrency:                       selection.subagentConcurrency,
		MaxToolRequests:                           selection.maxToolRequests,
		SubagentAttempts:                          selection.subagentAttempts,
		SubagentBackoffMS:                         selection.subagentBackoffMS,
		NoProgressStop:                            selection.noProgressStop,
		MaxContextBytes:                           selection.maxContextBytes,
		MaxSubagentOutputBytes:                    selection.maxSubagentOutputBytes,
		MinSubagentConfidence:                     selection.minSubagentConfidence,
		WorkspaceBriefMaxFiles:                    selection.workspaceBriefMaxFiles,
		WorkspaceBriefExcludeCount:                selection.workspaceBriefExcludeCount,
		CheckCommandArgc:                          firstCommandArgc(selection.checkArgv),
		CheckCommandPresent:                       len(selection.checkArgv) > 0,
		CheckCommandCount:                         len(selection.checkArgv),
		CheckSetCount:                             selection.checkSetCount,
		AutoCheckSetCount:                         selection.autoCheckCount,
		DefaultCheckSet:                           selection.defaultCheckSet,
	}, nil
}

func firstCommandArgc(commands [][]string) int {
	if len(commands) == 0 {
		return 0
	}
	return len(commands[0])
}

func commandSource(source string, argv []string) string {
	if len(argv) == 0 {
		return "default"
	}
	return source
}

func workspaceConfigPath(workspaceDir string) string {
	root := strings.TrimSpace(workspaceDir)
	if root == "" {
		return ""
	}
	return filepath.Join(root, config.WorkspaceConfigName)
}
