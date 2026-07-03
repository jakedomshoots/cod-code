package cli

import (
	"context"

	"ceoharness/internal/config"
	"ceoharness/internal/history"
)

func providerHealthPolicyFromConfig(cfg config.Config) history.ProviderHealthPolicy {
	return history.ProviderHealthPolicy{
		AvoidFailureRate:            cfg.ProviderHealthAvoidFailureRate,
		WatchFailureRate:            cfg.ProviderHealthWatchFailureRate,
		WatchCostPerAttemptMicroUSD: cfg.ProviderHealthWatchCostPerAttemptMicroUSD,
	}
}

func providerHealthPolicyForWorkspace(ctx context.Context, workspaceDir string) (history.ProviderHealthPolicy, error) {
	cfg, err := config.LoadWorkspace(ctx, workspaceDir)
	if err != nil {
		return history.ProviderHealthPolicy{}, err
	}
	return providerHealthPolicyFromConfig(cfg), nil
}
