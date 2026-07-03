package config

import (
	"fmt"
	"strings"

	"ceoharness/internal/jobpacket"
)

type ProviderPolicy struct {
	DefaultProvider   string            `json:"default_provider"`
	FallbackProvider  string            `json:"fallback_provider"`
	RiskProviders     map[string]string `json:"risk_providers"`
	KindProviders     map[string]string `json:"kind_providers"`
	RiskAreaProviders map[string]string `json:"risk_area_providers"`
}

func (cfg Config) HasProviderPolicy() bool {
	return cfg.ProviderPolicy.RuleCount() > 0
}

func (policy ProviderPolicy) RuleCount() int {
	count := len(policy.RiskProviders) + len(policy.KindProviders) + len(policy.RiskAreaProviders)
	if strings.TrimSpace(policy.DefaultProvider) != "" {
		count++
	}
	if strings.TrimSpace(policy.FallbackProvider) != "" {
		count++
	}
	return count
}

func (cfg Config) ProviderRoutesForTask(task string, subagents []jobpacket.Subagent) (map[string]string, error) {
	decisions, err := cfg.ProviderRouteDecisionsForTask(task, subagents)
	if err != nil {
		return nil, err
	}
	return routesFromDecisions(decisions), nil
}

func (policy ProviderPolicy) ProviderForProfile(profile jobpacket.TaskProfile) string {
	providerName, _ := policy.ProviderForProfileDecision(profile)
	return providerName
}

func (policy ProviderPolicy) ProviderForProfileDecision(profile jobpacket.TaskProfile) (string, string) {
	if providerName := strings.TrimSpace(policy.RiskProviders[profile.RiskLevel]); providerName != "" {
		return providerName, "provider_policy.risk_providers." + profile.RiskLevel
	}
	if providerName := strings.TrimSpace(policy.KindProviders[profile.Kind]); providerName != "" {
		return providerName, "provider_policy.kind_providers." + profile.Kind
	}
	if providerName := strings.TrimSpace(policy.DefaultProvider); providerName != "" {
		return providerName, "provider_policy.default_provider"
	}
	return "", ""
}

func (cfg Config) validateProviderPolicy() error {
	if err := cfg.validatePolicyProvider("provider_policy.default_provider", cfg.ProviderPolicy.DefaultProvider); err != nil {
		return err
	}
	if err := cfg.validatePolicyProvider("provider_policy.fallback_provider", cfg.ProviderPolicy.FallbackProvider); err != nil {
		return err
	}
	for riskLevel, providerName := range cfg.ProviderPolicy.RiskProviders {
		if !validRiskLevel(riskLevel) {
			return fmt.Errorf("provider_policy.risk_providers[%s]: %w", riskLevel, ErrInvalidConfig)
		}
		if err := cfg.validatePolicyProvider("provider_policy.risk_providers["+riskLevel+"]", providerName); err != nil {
			return err
		}
	}
	for taskKind, providerName := range cfg.ProviderPolicy.KindProviders {
		if !validTaskKind(taskKind) {
			return fmt.Errorf("provider_policy.kind_providers[%s]: %w", taskKind, ErrInvalidConfig)
		}
		if err := cfg.validatePolicyProvider("provider_policy.kind_providers["+taskKind+"]", providerName); err != nil {
			return err
		}
	}
	for riskArea, providerName := range cfg.ProviderPolicy.RiskAreaProviders {
		if !validRiskArea(riskArea) {
			return fmt.Errorf("provider_policy.risk_area_providers[%s]: %w", riskArea, ErrInvalidConfig)
		}
		if err := cfg.validatePolicyProvider("provider_policy.risk_area_providers["+riskArea+"]", providerName); err != nil {
			return err
		}
	}
	return nil
}

func (cfg Config) validatePolicyProvider(name string, providerName string) error {
	cleanName := strings.TrimSpace(providerName)
	if cleanName == "" {
		return nil
	}
	if _, ok := cfg.Providers[cleanName]; !ok {
		return fmt.Errorf("%s: %w", name, ErrInvalidConfig)
	}
	return nil
}

func validRiskLevel(level string) bool {
	switch strings.TrimSpace(level) {
	case "low", "medium", "high":
		return true
	default:
		return false
	}
}

func validTaskKind(kind string) bool {
	switch strings.TrimSpace(kind) {
	case "coding", "planning", "research", "mixed":
		return true
	default:
		return false
	}
}

func validRiskArea(area string) bool {
	switch strings.TrimSpace(area) {
	case "billing", "database", "release", "security":
		return true
	default:
		return false
	}
}
