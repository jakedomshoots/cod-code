package cli

import (
	"fmt"
	"strings"

	"ceoharness/internal/config"
)

func initHTTPProviders(opts options) (map[string]config.Provider, map[string]string, error) {
	if len(opts.httpInits) == 0 {
		return nil, nil, nil
	}
	providers := map[string]config.Provider{}
	agentProviders := map[string]string{}
	for _, init := range opts.httpInits {
		name, provider, agent, err := buildHTTPProviderFromInit(opts, init)
		if err != nil {
			return nil, nil, err
		}
		if _, exists := providers[name]; exists {
			return nil, nil, fmt.Errorf("duplicate --http-provider %q", name)
		}
		providers[name] = provider
		if agent == "" {
			continue
		}
		if _, exists := agentProviders[agent]; exists {
			return nil, nil, fmt.Errorf("duplicate --http-agent %q", agent)
		}
		agentProviders[agent] = name
	}
	return providers, agentProviders, nil
}

func buildHTTPProviderFromInit(opts options, init httpInitOptions) (string, config.Provider, string, error) {
	name := strings.TrimSpace(init.providerName)
	presetName := strings.TrimSpace(init.presetName)
	url := strings.TrimSpace(init.providerURL)
	model := strings.TrimSpace(init.providerModel)
	agent := strings.TrimSpace(init.agent)
	apiKeyEnv := strings.TrimSpace(init.apiKeyEnv)
	responseFormat := strings.TrimSpace(init.responseFormat)
	preset, err := resolveHTTPProviderPreset(presetName)
	if err != nil {
		return "", config.Provider{}, "", err
	}
	if url == "" {
		url = preset.URL
	}
	if apiKeyEnv == "" {
		apiKeyEnv = preset.APIKeyEnv
	}
	if name == "" || url == "" || model == "" {
		return "", config.Provider{}, "", fmt.Errorf("--http-provider, --http-url or --http-preset, and --http-model are required together")
	}
	if agent == "" && !httpProviderUsedByPolicy(opts, name) {
		return "", config.Provider{}, "", fmt.Errorf("--http-agent or a matching --default-provider/--fallback-provider is required with --http-provider")
	}
	return name, config.Provider{
		HTTP: config.HTTPProvider{
			URL:                        url,
			Model:                      model,
			APIKeyEnv:                  apiKeyEnv,
			InputCostPerMillionTokens:  init.inputCostPerMillion,
			OutputCostPerMillionTokens: init.outputCostPerMillion,
			TimeoutMS:                  init.timeoutMS,
			MaxOutputTokens:            init.maxOutputTokens,
			ResponseFormat:             responseFormat,
		},
	}, agent, nil
}

func httpProviderUsedByPolicy(opts options, providerName string) bool {
	name := strings.TrimSpace(providerName)
	if strings.TrimSpace(opts.defaultProvider) == name ||
		strings.TrimSpace(opts.fallbackProvider) == name {
		return true
	}
	if strings.TrimSpace(opts.ceoProvider) == name {
		return true
	}
	for _, policyProvider := range opts.riskProviders {
		if strings.TrimSpace(policyProvider) == name {
			return true
		}
	}
	for _, policyProvider := range opts.kindProviders {
		if strings.TrimSpace(policyProvider) == name {
			return true
		}
	}
	for _, policyProvider := range opts.riskAreaProviders {
		if strings.TrimSpace(policyProvider) == name {
			return true
		}
	}
	return false
}
