package config

import (
	"fmt"

	"ceoharness/internal/jobpacket"
)

func validateSubagentBudget(cfg Config) error {
	if cfg.MaxSubagents < 0 || cfg.MaxSubagents > jobpacket.MaxDelegatedSubagents {
		return fmt.Errorf("max_subagents: %w", ErrInvalidConfig)
	}
	return nil
}
