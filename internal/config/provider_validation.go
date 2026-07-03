package config

import (
	"fmt"
	"strings"
)

func (cfg Config) validateProviders() error {
	for providerName, provider := range cfg.Providers {
		if strings.TrimSpace(providerName) == "" {
			return fmt.Errorf("providers key: %w", ErrInvalidConfig)
		}
		if err := validateProvider(fmt.Sprintf("providers[%s]", providerName), provider); err != nil {
			return err
		}
	}
	return nil
}

func validateProvider(name string, provider Provider) error {
	hasCommand := len(provider.ModelCommand) > 0
	hasHTTP := !provider.HTTP.IsZero()
	if hasCommand == hasHTTP {
		return fmt.Errorf("%s backend: %w", name, ErrInvalidConfig)
	}
	if hasCommand {
		if err := validateCommand(name+".model_command", provider.ModelCommand, false); err != nil {
			return err
		}
	}
	if hasHTTP {
		if err := validateHTTPProvider(name, provider.HTTP); err != nil {
			return err
		}
	}
	return validateProviderEnvVars(name, provider.EnvVars)
}
