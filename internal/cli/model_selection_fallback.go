package cli

import (
	"strings"

	"ceoharness/internal/config"
)

type providerFallbackSelection struct {
	argv         []string
	envVars      []string
	httpProvider config.HTTPProvider
	providerName string
}

func providerFallbackRoute(cfg config.Config) providerFallbackSelection {
	providerName := strings.TrimSpace(cfg.ProviderPolicy.FallbackProvider)
	if providerName == "" {
		return providerFallbackSelection{}
	}
	provider := cfg.Providers[providerName]
	envVars := append([]string(nil), provider.EnvVars...)
	if strings.TrimSpace(provider.HTTP.APIKeyEnv) != "" {
		envVars = append(envVars, strings.TrimSpace(provider.HTTP.APIKeyEnv))
	}
	return providerFallbackSelection{
		argv:         append([]string(nil), provider.ModelCommand...),
		envVars:      envVars,
		httpProvider: provider.HTTP,
		providerName: providerName,
	}
}
