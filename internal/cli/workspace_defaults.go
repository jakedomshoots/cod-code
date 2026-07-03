package cli

import (
	"context"
	"fmt"
	"strings"

	"ceoharness/internal/config"
	"ceoharness/internal/jobpacket"
)

func optionsWithWorkspaceDefaults(ctx context.Context, opts options) (options, error) {
	if strings.TrimSpace(opts.workspaceDir) == "" {
		if strings.TrimSpace(opts.checkSet) != "" {
			return options{}, fmt.Errorf("--check-set requires --workspace")
		}
		return opts, nil
	}
	cfg, err := config.LoadWorkspace(ctx, opts.workspaceDir)
	if err != nil {
		return options{}, err
	}
	opts = optionsWithRetryPolicyDefaults(opts, cfg)
	opts = optionsWithToolPolicyDefaults(opts, cfg)
	opts = optionsWithSubagentBudgetDefaults(opts, cfg)
	opts = optionsWithSubagentOutputPolicyDefaults(opts, cfg)
	opts = optionsWithConfidencePolicyDefaults(opts, cfg)
	opts = optionsWithCostPolicyDefaults(opts, cfg)
	opts = optionsWithContextPolicyDefaults(opts, cfg)
	opts = optionsWithWritePolicyDefaults(opts, cfg)
	opts = optionsWithWorkspaceBriefDefaults(opts, cfg)
	opts = optionsWithVerificationPolicyDefaults(opts, cfg)
	opts = optionsWithModelCommandDefaults(opts, cfg)
	opts = optionsWithToolCommandDefaults(opts, cfg)
	opts = optionsWithJobPolicyDefaults(opts, cfg)
	if len(opts.subagents) == 0 && len(cfg.Subagents) > 0 {
		opts.subagents = append([]jobpacket.Subagent(nil), cfg.Subagents...)
	}
	opts.providerHealthPolicy = providerHealthPolicyFromConfig(cfg)
	if len(opts.checkCommand) > 0 || len(opts.checkCommands) > 0 {
		return opts, nil
	}
	if setName := strings.TrimSpace(opts.checkSet); setName != "" {
		commands, ok := cfg.CheckCommandsForSet(setName)
		if !ok {
			return options{}, fmt.Errorf("check set %q: %w", setName, config.ErrInvalidConfig)
		}
		opts.checkCommands = commands
		return opts, nil
	}
	if commands, ok := cfg.AutoCheckCommandsForTask(opts.task); ok {
		opts.checkCommands = commands
		return opts, nil
	}
	if commands := cfg.CheckCommandList(); len(commands) > 0 {
		opts.checkCommands = commands
		return opts, nil
	}
	if cfg.DefaultCheckSet != "" {
		commands, ok := cfg.CheckCommandsForSet(cfg.DefaultCheckSet)
		if !ok {
			return options{}, fmt.Errorf("default check set %q: %w", cfg.DefaultCheckSet, config.ErrInvalidConfig)
		}
		opts.checkCommands = commands
	}
	return opts, nil
}

func optionsWithWritePolicyDefaults(opts options, cfg config.Config) options {
	if strings.TrimSpace(opts.writePolicy) == "" {
		opts.writePolicy = cfg.WritePolicy
	}
	return opts
}

func optionsWithVerificationPolicyDefaults(opts options, cfg config.Config) options {
	opts.requireChecks = opts.requireChecks || cfg.RequireChecks
	return opts
}

func optionsWithRetryPolicyDefaults(opts options, cfg config.Config) options {
	if opts.checkAttempts == 0 {
		opts.checkAttempts = cfg.CheckAttempts
	}
	if opts.checkBackoffMS == 0 {
		opts.checkBackoffMS = cfg.CheckBackoffMS
	}
	if opts.ceoRevisionAttempts == 0 {
		opts.ceoRevisionAttempts = cfg.CEORevisionAttempts
	}
	if opts.subagentConcurrency == 0 {
		opts.subagentConcurrency = cfg.SubagentConcurrency
	}
	if opts.subagentAttempts == 0 {
		opts.subagentAttempts = cfg.SubagentAttempts
	}
	if opts.subagentBackoffMS == 0 {
		opts.subagentBackoffMS = cfg.SubagentBackoffMS
	}
	if opts.noProgressStop == 0 {
		opts.noProgressStop = cfg.NoProgressStop
	}
	return opts
}

func optionsWithToolPolicyDefaults(opts options, cfg config.Config) options {
	if opts.maxToolRequests == 0 {
		opts.maxToolRequests = cfg.MaxToolRequests
	}
	return opts
}

func optionsWithSubagentBudgetDefaults(opts options, cfg config.Config) options {
	if opts.maxSubagents == 0 {
		opts.maxSubagents = cfg.MaxSubagents
	}
	return opts
}

func optionsWithSubagentOutputPolicyDefaults(opts options, cfg config.Config) options {
	if opts.maxSubagentOutputBytes == 0 {
		opts.maxSubagentOutputBytes = cfg.MaxSubagentOutputBytes
	}
	return opts
}

func optionsWithConfidencePolicyDefaults(opts options, cfg config.Config) options {
	if opts.minSubagentConfidence == 0 {
		opts.minSubagentConfidence = cfg.MinSubagentConfidence
	}
	return opts
}

func optionsWithCostPolicyDefaults(opts options, cfg config.Config) options {
	if opts.providerCostBudgetMicroUSD == 0 {
		opts.providerCostBudgetMicroUSD = cfg.ProviderCostBudgetMicroUSD
	}
	return opts
}

func optionsWithContextPolicyDefaults(opts options, cfg config.Config) options {
	if opts.maxContextBytes == 0 {
		opts.maxContextBytes = cfg.MaxContextBytes
	}
	return opts
}

func optionsWithWorkspaceBriefDefaults(opts options, cfg config.Config) options {
	if opts.workspaceBriefMaxFiles == 0 {
		opts.workspaceBriefMaxFiles = cfg.WorkspaceBriefMaxFiles
	}
	if len(cfg.WorkspaceBriefExcludes) > 0 {
		opts.workspaceBriefExcludes = append(append([]string(nil), cfg.WorkspaceBriefExcludes...), opts.workspaceBriefExcludes...)
	}
	return opts
}

func optionsWithModelCommandDefaults(opts options, cfg config.Config) options {
	if opts.modelCommandTimeoutMS == 0 {
		opts.modelCommandTimeoutMS = cfg.ModelCommandTimeoutMS
	}
	return opts
}

func optionsWithToolCommandDefaults(opts options, cfg config.Config) options {
	if opts.toolCommandTimeoutMS == 0 {
		opts.toolCommandTimeoutMS = cfg.ToolCommandTimeoutMS
	}
	return opts
}

func optionsWithJobPolicyDefaults(opts options, cfg config.Config) options {
	if opts.jobTimeoutMS == 0 {
		opts.jobTimeoutMS = cfg.JobTimeoutMS
	}
	if opts.maxCEOIterations == 0 {
		opts.maxCEOIterations = cfg.MaxCEOIterations
	}
	return opts
}
