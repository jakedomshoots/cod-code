package config

import (
	"errors"
	"fmt"
	"strings"

	"ceoharness/internal/jobpacket"
)

const WorkspaceConfigName = ".ceo-harness.json"

var (
	ErrConfigExists  = errors.New("workspace config already exists")
	ErrInvalidConfig = errors.New("invalid workspace config")
)

type Config struct {
	ModelCommand                              []string              `json:"model_command"`
	CEOModelCommand                           []string              `json:"ceo_model_command"`
	ResearchCommand                           []string              `json:"research_command"`
	ModelCommandTimeoutMS                     int                   `json:"model_command_timeout_ms"`
	ToolCommandTimeoutMS                      int                   `json:"tool_command_timeout_ms"`
	JobTimeoutMS                              int                   `json:"job_timeout_ms"`
	Subagents                                 []jobpacket.Subagent  `json:"subagents"`
	AgentModelCommands                        map[string][]string   `json:"agent_model_commands"`
	Providers                                 map[string]Provider   `json:"providers"`
	CEOProvider                               string                `json:"ceo_provider"`
	AgentProviders                            map[string]string     `json:"agent_providers"`
	ProviderPolicy                            ProviderPolicy        `json:"provider_policy"`
	CheckCommand                              []string              `json:"check_command"`
	CheckCommands                             [][]string            `json:"check_commands"`
	CheckSets                                 map[string][][]string `json:"check_sets"`
	DefaultCheckSet                           string                `json:"default_check_set"`
	AutoCheckSets                             []AutoCheckSet        `json:"auto_check_sets"`
	RequireChecks                             bool                  `json:"require_checks,omitempty"`
	CheckAttempts                             int                   `json:"check_attempts"`
	CheckBackoffMS                            int                   `json:"check_backoff_ms"`
	CEORevisionAttempts                       int                   `json:"ceo_revision_attempts"`
	MaxCEOIterations                          int                   `json:"max_ceo_iterations"`
	MaxSubagents                              int                   `json:"max_subagents"`
	SubagentConcurrency                       int                   `json:"subagent_concurrency"`
	MaxToolRequests                           int                   `json:"max_tool_requests"`
	SubagentAttempts                          int                   `json:"subagent_attempts"`
	SubagentBackoffMS                         int                   `json:"subagent_backoff_ms"`
	NoProgressStop                            int                   `json:"no_progress_stop"`
	MaxContextBytes                           int                   `json:"max_context_bytes"`
	MaxSubagentOutputBytes                    int                   `json:"max_subagent_output_bytes"`
	MinSubagentConfidence                     float64               `json:"min_subagent_confidence"`
	WorkspaceBriefMaxFiles                    int                   `json:"workspace_brief_max_files"`
	WorkspaceBriefExcludes                    []string              `json:"workspace_brief_excludes"`
	WritePolicy                               string                `json:"write_policy,omitempty"`
	ProviderCostBudgetMicroUSD                int64                 `json:"provider_cost_budget_microusd"`
	ProviderHealthAvoidFailureRate            float64               `json:"provider_health_avoid_failure_rate"`
	ProviderHealthWatchFailureRate            float64               `json:"provider_health_watch_failure_rate"`
	ProviderHealthWatchCostPerAttemptMicroUSD int64                 `json:"provider_health_watch_cost_per_attempt_microusd"`
}

func (cfg Config) Validate() error {
	if err := validateCommand("model_command", cfg.ModelCommand, true); err != nil {
		return err
	}
	if err := validateCommand("ceo_model_command", cfg.CEOModelCommand, true); err != nil {
		return err
	}
	if err := validateCommand("research_command", cfg.ResearchCommand, true); err != nil {
		return err
	}
	if cfg.ModelCommandTimeoutMS < 0 {
		return fmt.Errorf("model_command_timeout_ms: %w", ErrInvalidConfig)
	}
	if cfg.ToolCommandTimeoutMS < 0 {
		return fmt.Errorf("tool_command_timeout_ms: %w", ErrInvalidConfig)
	}
	if err := validateJobPolicy(cfg); err != nil {
		return err
	}
	if err := validateSubagents(cfg.Subagents); err != nil {
		return err
	}
	if err := validateCommand("check_command", cfg.CheckCommand, true); err != nil {
		return err
	}
	for index, command := range cfg.CheckCommands {
		if err := validateCommand(fmt.Sprintf("check_commands[%d]", index), command, false); err != nil {
			return err
		}
	}
	for setName, commands := range cfg.CheckSets {
		if strings.TrimSpace(setName) == "" || len(commands) == 0 {
			return fmt.Errorf("check_sets key: %w", ErrInvalidConfig)
		}
		for index, command := range commands {
			if err := validateCommand(fmt.Sprintf("check_sets[%s][%d]", setName, index), command, false); err != nil {
				return err
			}
		}
	}
	if strings.TrimSpace(cfg.DefaultCheckSet) != "" {
		if _, ok := cfg.CheckSets[cfg.DefaultCheckSet]; !ok {
			return fmt.Errorf("default_check_set: %w", ErrInvalidConfig)
		}
	}
	if err := cfg.validateAutoCheckSets(); err != nil {
		return err
	}
	if err := validateRetryPolicy(cfg); err != nil {
		return err
	}
	if err := validateSubagentBudget(cfg); err != nil {
		return err
	}
	if err := validateToolPolicy(cfg); err != nil {
		return err
	}
	if err := validateSubagentOutputPolicy(cfg); err != nil {
		return err
	}
	if err := validateConfidencePolicy(cfg); err != nil {
		return err
	}
	if err := validateContextPolicy(cfg); err != nil {
		return err
	}
	if err := validateWritePolicy(cfg); err != nil {
		return err
	}
	if err := validateCostPolicy(cfg); err != nil {
		return err
	}
	if err := validateProviderHealthPolicy(cfg); err != nil {
		return err
	}
	for agentName, command := range cfg.AgentModelCommands {
		if strings.TrimSpace(agentName) == "" {
			return fmt.Errorf("agent_model_commands key: %w", ErrInvalidConfig)
		}
		if err := validateCommand(fmt.Sprintf("agent_model_commands[%s]", agentName), command, false); err != nil {
			return err
		}
	}
	if err := cfg.validateProviders(); err != nil {
		return err
	}
	if err := cfg.validateCEOProvider(); err != nil {
		return err
	}
	if err := validateSubagentProviders(cfg); err != nil {
		return err
	}
	if err := cfg.validateProviderPolicy(); err != nil {
		return err
	}
	for agentName, providerName := range cfg.AgentProviders {
		if strings.TrimSpace(agentName) == "" || strings.TrimSpace(providerName) == "" {
			return fmt.Errorf("agent_providers entry: %w", ErrInvalidConfig)
		}
		if _, ok := cfg.Providers[providerName]; !ok {
			return fmt.Errorf("agent_providers[%s]: %w", agentName, ErrInvalidConfig)
		}
	}
	return nil
}

func (cfg Config) HasSubagentBudget() bool {
	return cfg.MaxSubagents > 0
}

func (cfg Config) HasVerificationPolicy() bool {
	return cfg.RequireChecks
}
