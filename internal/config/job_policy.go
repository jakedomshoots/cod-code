package config

import "fmt"

func (cfg Config) HasJobPolicy() bool {
	return cfg.JobTimeoutMS > 0 ||
		cfg.MaxCEOIterations > 0
}

func validateJobPolicy(cfg Config) error {
	if cfg.JobTimeoutMS < 0 {
		return fmt.Errorf("job_timeout_ms: %w", ErrInvalidConfig)
	}
	if cfg.MaxCEOIterations < 0 {
		return fmt.Errorf("max_ceo_iterations: %w", ErrInvalidConfig)
	}
	return nil
}
