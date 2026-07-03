package config

import "fmt"

func (cfg Config) HasConfidencePolicy() bool {
	return cfg.MinSubagentConfidence > 0
}

func validateConfidencePolicy(cfg Config) error {
	if cfg.MinSubagentConfidence < 0 || cfg.MinSubagentConfidence > 1 {
		return fmt.Errorf("min_subagent_confidence: %w", ErrInvalidConfig)
	}
	return nil
}
