package config

import (
	"fmt"
	"strings"

	"ceoharness/internal/jobpacket"
)

func validateSubagents(subagents []jobpacket.Subagent) error {
	if len(subagents) == 0 {
		return nil
	}
	if len(subagents) > jobpacket.MaxDelegatedSubagents {
		return fmt.Errorf("subagents count: %w", ErrInvalidConfig)
	}
	seen := map[string]struct{}{}
	for index, subagent := range subagents {
		name := strings.TrimSpace(subagent.Name)
		role := strings.TrimSpace(subagent.Role)
		if name == "" || role == "" {
			return fmt.Errorf("subagents[%d]: %w", index, ErrInvalidConfig)
		}
		if _, ok := seen[name]; ok {
			return fmt.Errorf("subagents[%d]: %w", index, ErrInvalidConfig)
		}
		if _, ok := jobpacket.NormalizeActions(subagent.AllowedActions); !ok {
			return fmt.Errorf("subagents[%d] allowed_actions: %w", index, ErrInvalidConfig)
		}
		seen[name] = struct{}{}
	}
	return nil
}

func validateSubagentProviders(cfg Config) error {
	for index, subagent := range cfg.Subagents {
		providerName := strings.TrimSpace(subagent.ProviderName)
		if providerName == "" {
			continue
		}
		if _, ok := cfg.Providers[providerName]; !ok {
			return fmt.Errorf("subagents[%d].provider: %w", index, ErrInvalidConfig)
		}
	}
	return nil
}
