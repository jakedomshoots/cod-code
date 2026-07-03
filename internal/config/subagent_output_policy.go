package config

import "fmt"

func (cfg Config) HasSubagentOutputPolicy() bool {
	return cfg.MaxSubagentOutputBytes > 0
}

func validateSubagentOutputPolicy(cfg Config) error {
	if cfg.MaxSubagentOutputBytes < 0 {
		return fmt.Errorf("max_subagent_output_bytes: %w", ErrInvalidConfig)
	}
	return nil
}
