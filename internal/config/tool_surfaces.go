package config

import (
	"fmt"

	"ceoharness/internal/browseruse"
	"ceoharness/internal/computeruse"
)

func validateBrowserToolPolicy(cfg Config) error {
	if err := browseruse.ValidatePolicy(cfg.BrowserPolicy); err != nil {
		return fmt.Errorf("browser_policy: %w", ErrInvalidConfig)
	}
	return validateCommand("browser_command", cfg.BrowserCommand, true)
}

func validateComputerToolPolicy(cfg Config) error {
	if err := computeruse.ValidatePolicy(cfg.ComputerPolicy); err != nil {
		return fmt.Errorf("computer_policy: %w", ErrInvalidConfig)
	}
	return validateCommand("computer_command", cfg.ComputerCommand, true)
}
