package config

import "fmt"

func (cfg Config) HasToolPolicy() bool {
	return cfg.MaxToolRequests > 0
}

func validateToolPolicy(cfg Config) error {
	if cfg.MaxToolRequests < 0 {
		return fmt.Errorf("max_tool_requests: %w", ErrInvalidConfig)
	}
	return nil
}
