package config

import (
	"fmt"
	"strings"
)

type Provider struct {
	ModelCommand []string     `json:"model_command"`
	EnvVars      []string     `json:"env_vars"`
	HTTP         HTTPProvider `json:"http"`
}

func (cfg Config) AgentEnvVars() map[string][]string {
	return cfg.AgentEnvVarsFor(cfg.AgentProviders)
}

func (cfg Config) AgentEnvVarsFor(agentProviders map[string]string) map[string][]string {
	envVars := map[string][]string{}
	for agentName, providerName := range agentProviders {
		provider := cfg.Providers[providerName]
		providerVars := provider.EnvVars
		if strings.TrimSpace(provider.HTTP.APIKeyEnv) != "" {
			providerVars = append(providerVars, provider.HTTP.APIKeyEnv)
		}
		if len(providerVars) > 0 {
			envVars[agentName] = append([]string(nil), providerVars...)
		}
	}
	return envVars
}

func (cfg Config) ProviderEnvVarNames() []string {
	seen := map[string]bool{}
	names := []string{}
	for _, provider := range cfg.Providers {
		if strings.TrimSpace(provider.HTTP.APIKeyEnv) != "" {
			name := strings.TrimSpace(provider.HTTP.APIKeyEnv)
			if !seen[name] {
				seen[name] = true
				names = append(names, name)
			}
		}
		for _, rawName := range provider.EnvVars {
			name := strings.TrimSpace(rawName)
			if name == "" || seen[name] {
				continue
			}
			seen[name] = true
			names = append(names, name)
		}
	}
	return names
}

func validateProviderEnvVars(name string, envVars []string) error {
	for index, rawName := range envVars {
		envName := strings.TrimSpace(rawName)
		if !validEnvVarName(envName) {
			return fmt.Errorf("%s.env_vars[%d]: %w", name, index, ErrInvalidConfig)
		}
	}
	return nil
}

func validEnvVarName(name string) bool {
	if name == "" {
		return false
	}
	for index, char := range name {
		if index == 0 {
			if (char >= 'A' && char <= 'Z') || (char >= 'a' && char <= 'z') || char == '_' {
				continue
			}
			return false
		}
		if (char >= 'A' && char <= 'Z') || (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') || char == '_' {
			continue
		}
		return false
	}
	return true
}
