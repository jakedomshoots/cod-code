package config

import "fmt"

func (cfg Config) HasCostPolicy() bool {
	return cfg.ProviderCostBudgetMicroUSD > 0
}

func validateCostPolicy(cfg Config) error {
	if cfg.ProviderCostBudgetMicroUSD < 0 {
		return fmt.Errorf("provider_cost_budget_microusd: %w", ErrInvalidConfig)
	}
	return nil
}
