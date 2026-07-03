package cli

import (
	"context"

	"ceoharness/internal/config"
)

func selectResearchCommand(ctx context.Context, opts options) (commandSelection, error) {
	if len(opts.researchCommand) > 0 {
		return commandSelection{argv: opts.researchCommand, timeoutMS: opts.toolCommandTimeoutMS, source: "flag"}, nil
	}
	command, err := researchCommandFromEnv()
	if err != nil {
		return commandSelection{}, err
	}
	if len(command) > 0 {
		return commandSelection{argv: command, timeoutMS: opts.toolCommandTimeoutMS, source: "env"}, nil
	}
	cfg, err := config.LoadWorkspace(ctx, opts.workspaceDir)
	if err != nil {
		return commandSelection{}, err
	}
	if len(cfg.ResearchCommand) > 0 {
		return commandSelection{argv: append([]string(nil), cfg.ResearchCommand...), timeoutMS: cfg.ToolCommandTimeoutMS, source: "workspace"}, nil
	}
	return commandSelection{timeoutMS: opts.toolCommandTimeoutMS, source: "default"}, nil
}
