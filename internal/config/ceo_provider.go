package config

import (
	"fmt"
	"strings"
)

func (cfg Config) validateCEOProvider() error {
	providerName := strings.TrimSpace(cfg.CEOProvider)
	if providerName == "" {
		return nil
	}
	if _, ok := cfg.Providers[providerName]; !ok {
		return fmt.Errorf("ceo_provider: %w", ErrInvalidConfig)
	}
	return nil
}
