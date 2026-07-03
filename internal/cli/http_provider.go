package cli

import (
	"fmt"
	"os"
	"strings"

	"ceoharness/internal/config"
	"ceoharness/internal/model"
)

func providerAPIKey(name string) (string, error) {
	trimmedName := strings.TrimSpace(name)
	if trimmedName == "" {
		return "", nil
	}
	value, ok := os.LookupEnv(trimmedName)
	if !ok || value == "" {
		return "", fmt.Errorf("provider env var %s is required", trimmedName)
	}
	return value, nil
}

func newHTTPProviderClient(provider config.HTTPProvider, apiKey string) (model.Client, error) {
	client, err := model.NewHTTPClient(model.HTTPConfig{
		URL:                        provider.URL,
		Model:                      provider.Model,
		APIKey:                     apiKey,
		InputCostPerMillionTokens:  provider.InputCostPerMillionTokens,
		OutputCostPerMillionTokens: provider.OutputCostPerMillionTokens,
		TimeoutMS:                  provider.TimeoutMS,
		MaxOutputTokens:            provider.MaxOutputTokens,
		ResponseFormat:             provider.ResponseFormat,
	})
	if err != nil {
		return nil, fmt.Errorf("create http model client: %w", err)
	}
	return client, nil
}

func countHTTPProviders(providers map[string]config.Provider) int {
	count := 0
	for _, provider := range providers {
		if !provider.HTTP.IsZero() {
			count++
		}
	}
	return count
}

type providerControlCounts struct {
	Cost            int
	Timeout         int
	MaxOutputTokens int
	ResponseFormat  int
}

func countHTTPProviderControls(providers map[string]config.Provider) providerControlCounts {
	var counts providerControlCounts
	for _, provider := range providers {
		http := provider.HTTP
		if http.IsZero() {
			continue
		}
		if http.InputCostPerMillionTokens > 0 || http.OutputCostPerMillionTokens > 0 {
			counts.Cost++
		}
		if http.TimeoutMS > 0 {
			counts.Timeout++
		}
		if http.MaxOutputTokens > 0 {
			counts.MaxOutputTokens++
		}
		if http.ResponseFormat != "" {
			counts.ResponseFormat++
		}
	}
	return counts
}
