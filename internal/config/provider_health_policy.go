package config

import "fmt"

func (cfg Config) HasProviderHealthPolicy() bool {
	return cfg.ProviderHealthAvoidFailureRate > 0 ||
		cfg.ProviderHealthWatchFailureRate > 0 ||
		cfg.ProviderHealthWatchCostPerAttemptMicroUSD > 0
}

func validateProviderHealthPolicy(cfg Config) error {
	if cfg.ProviderHealthAvoidFailureRate < 0 || cfg.ProviderHealthAvoidFailureRate > 1 {
		return fmt.Errorf("provider_health_avoid_failure_rate: %w", ErrInvalidConfig)
	}
	if cfg.ProviderHealthWatchFailureRate < 0 || cfg.ProviderHealthWatchFailureRate > 1 {
		return fmt.Errorf("provider_health_watch_failure_rate: %w", ErrInvalidConfig)
	}
	avoid := cfg.ProviderHealthAvoidFailureRate
	if avoid == 0 {
		avoid = 0.5
	}
	if cfg.ProviderHealthWatchFailureRate > avoid {
		return fmt.Errorf("provider_health_watch_failure_rate: %w", ErrInvalidConfig)
	}
	if cfg.ProviderHealthWatchCostPerAttemptMicroUSD < 0 {
		return fmt.Errorf("provider_health_watch_cost_per_attempt_microusd: %w", ErrInvalidConfig)
	}
	return nil
}
