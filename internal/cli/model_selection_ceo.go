package cli

import (
	"context"

	"ceoharness/internal/config"
)

func selectCEOModelCommand(ctx context.Context, opts options) (commandSelection, error) {
	if len(opts.ceoModelCommand) > 0 {
		return commandSelection{argv: opts.ceoModelCommand, timeoutMS: opts.modelCommandTimeoutMS, source: "flag"}, nil
	}
	command, err := ceoModelCommandFromEnv()
	if err != nil {
		return commandSelection{}, err
	}
	if len(command) > 0 {
		return commandSelection{argv: command, timeoutMS: opts.modelCommandTimeoutMS, source: "env"}, nil
	}
	cfg, err := config.LoadWorkspace(ctx, opts.workspaceDir)
	if err != nil {
		return commandSelection{}, err
	}
	if cfg.CEOProvider != "" {
		return commandSelection{
			timeoutMS:    cfg.ModelCommandTimeoutMS,
			source:       "workspace",
			providerName: cfg.CEOProvider,
			provider:     cfg.Providers[cfg.CEOProvider],
		}, nil
	}
	if len(cfg.CEOModelCommand) > 0 {
		return commandSelection{argv: cfg.CEOModelCommand, timeoutMS: cfg.ModelCommandTimeoutMS, source: "workspace"}, nil
	}
	return commandSelection{timeoutMS: opts.modelCommandTimeoutMS, source: "default"}, nil
}
