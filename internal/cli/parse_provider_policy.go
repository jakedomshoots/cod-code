package cli

import (
	"fmt"
	"strconv"
	"strings"
)

func parseProviderPolicyFlag(args []string, index int, opts *options) (bool, int, error) {
	switch args[index] {
	case "--default-provider":
		value, err := parseNextValue(args, index, "--default-provider requires a provider name")
		if err != nil {
			return true, index, err
		}
		opts.defaultProvider = value
	case "--fallback-provider":
		value, err := parseNextValue(args, index, "--fallback-provider requires a provider name")
		if err != nil {
			return true, index, err
		}
		opts.fallbackProvider = value
	case "--ceo-provider":
		value, err := parseNextValue(args, index, "--ceo-provider requires a provider name")
		if err != nil {
			return true, index, err
		}
		opts.ceoProvider = value
	case "--risk-area-provider":
		value, err := parseNextValue(args, index, "--risk-area-provider requires area=provider")
		if err != nil {
			return true, index, err
		}
		area, provider, err := parsePolicyMapEntry(value, "--risk-area-provider", "area")
		if err != nil {
			return true, index, err
		}
		if opts.riskAreaProviders == nil {
			opts.riskAreaProviders = map[string]string{}
		}
		if _, exists := opts.riskAreaProviders[area]; exists {
			return true, index, fmt.Errorf("--risk-area-provider duplicate area %q", area)
		}
		opts.riskAreaProviders[area] = provider
	case "--risk-provider":
		value, err := parseNextValue(args, index, "--risk-provider requires risk=provider")
		if err != nil {
			return true, index, err
		}
		risk, provider, err := parsePolicyMapEntry(value, "--risk-provider", "risk")
		if err != nil {
			return true, index, err
		}
		if opts.riskProviders == nil {
			opts.riskProviders = map[string]string{}
		}
		if _, exists := opts.riskProviders[risk]; exists {
			return true, index, fmt.Errorf("--risk-provider duplicate risk %q", risk)
		}
		opts.riskProviders[risk] = provider
	case "--kind-provider":
		value, err := parseNextValue(args, index, "--kind-provider requires kind=provider")
		if err != nil {
			return true, index, err
		}
		kind, provider, err := parsePolicyMapEntry(value, "--kind-provider", "kind")
		if err != nil {
			return true, index, err
		}
		if opts.kindProviders == nil {
			opts.kindProviders = map[string]string{}
		}
		if _, exists := opts.kindProviders[kind]; exists {
			return true, index, fmt.Errorf("--kind-provider duplicate kind %q", kind)
		}
		opts.kindProviders[kind] = provider
	case "--min-subagent-confidence":
		value, err := parseFailureRateFlag(args, index, "--min-subagent-confidence")
		if err != nil {
			return true, index, err
		}
		opts.minSubagentConfidence = value
	case "--provider-health-avoid-failure-rate":
		value, err := parseFailureRateFlag(args, index, "--provider-health-avoid-failure-rate")
		if err != nil {
			return true, index, err
		}
		opts.providerHealthAvoidRate = value
	case "--provider-health-watch-failure-rate":
		value, err := parseFailureRateFlag(args, index, "--provider-health-watch-failure-rate")
		if err != nil {
			return true, index, err
		}
		opts.providerHealthWatchRate = value
	case "--provider-health-watch-cost-per-attempt-microusd":
		raw, err := parseNextValue(args, index, "--provider-health-watch-cost-per-attempt-microusd requires a number")
		if err != nil {
			return true, index, err
		}
		value, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || value < 0 {
			return true, index, fmt.Errorf("--provider-health-watch-cost-per-attempt-microusd must be a non-negative integer")
		}
		opts.providerHealthWatchCostPerAttemptMicroUSD = value
	default:
		return false, index, nil
	}
	return true, index + 1, nil
}

func parsePolicyMapEntry(raw string, flag string, keyName string) (string, string, error) {
	key, value, ok := strings.Cut(raw, "=")
	key = strings.TrimSpace(key)
	value = strings.TrimSpace(value)
	if !ok || key == "" || value == "" {
		return "", "", fmt.Errorf("%s must be %s=provider", flag, keyName)
	}
	return key, value, nil
}

func parseFailureRateFlag(args []string, index int, flag string) (float64, error) {
	value, err := parseNonNegativeFloatFlag(args, index, flag)
	if err != nil {
		return 0, err
	}
	if value > 1 {
		return 0, fmt.Errorf("%s must be between 0 and 1", flag)
	}
	return value, nil
}
