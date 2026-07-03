package config

import "fmt"

func (cfg Config) HasRetryPolicy() bool {
	return cfg.CheckAttempts > 0 ||
		cfg.CheckBackoffMS > 0 ||
		cfg.CEORevisionAttempts > 0 ||
		cfg.SubagentConcurrency > 0 ||
		cfg.SubagentAttempts > 0 ||
		cfg.SubagentBackoffMS > 0 ||
		cfg.NoProgressStop > 0
}

func validateRetryPolicy(cfg Config) error {
	if cfg.CheckAttempts < 0 {
		return fmt.Errorf("check_attempts: %w", ErrInvalidConfig)
	}
	if cfg.CheckBackoffMS < 0 {
		return fmt.Errorf("check_backoff_ms: %w", ErrInvalidConfig)
	}
	if cfg.CEORevisionAttempts < 0 {
		return fmt.Errorf("ceo_revision_attempts: %w", ErrInvalidConfig)
	}
	if cfg.SubagentConcurrency < 0 {
		return fmt.Errorf("subagent_concurrency: %w", ErrInvalidConfig)
	}
	if cfg.SubagentAttempts < 0 {
		return fmt.Errorf("subagent_attempts: %w", ErrInvalidConfig)
	}
	if cfg.SubagentBackoffMS < 0 {
		return fmt.Errorf("subagent_backoff_ms: %w", ErrInvalidConfig)
	}
	if cfg.NoProgressStop < 0 {
		return fmt.Errorf("no_progress_stop: %w", ErrInvalidConfig)
	}
	return nil
}
