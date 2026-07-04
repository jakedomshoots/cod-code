package config

import (
	"fmt"
	"net/url"
	"strings"
)

type HTTPProvider struct {
	URL                        string  `json:"url"`
	Model                      string  `json:"model"`
	APIKeyEnv                  string  `json:"api_key_env"`
	InputCostPerMillionTokens  float64 `json:"input_cost_per_million_tokens"`
	OutputCostPerMillionTokens float64 `json:"output_cost_per_million_tokens"`
	TimeoutMS                  int     `json:"timeout_ms"`
	MaxOutputTokens            int     `json:"max_output_tokens"`
	ResponseFormat             string  `json:"response_format"`
	DisableThinking            bool    `json:"disable_thinking,omitempty"`
}

func (p HTTPProvider) IsZero() bool {
	return strings.TrimSpace(p.URL) == "" &&
		strings.TrimSpace(p.Model) == "" &&
		strings.TrimSpace(p.APIKeyEnv) == "" &&
		p.InputCostPerMillionTokens == 0 &&
		p.OutputCostPerMillionTokens == 0 &&
		p.TimeoutMS == 0 &&
		p.MaxOutputTokens == 0 &&
		strings.TrimSpace(p.ResponseFormat) == "" &&
		!p.DisableThinking
}

func (cfg Config) AgentHTTPProviders() map[string]HTTPProvider {
	return cfg.AgentHTTPProvidersFor(cfg.AgentProviders)
}

func (cfg Config) AgentHTTPProvidersFor(agentProviders map[string]string) map[string]HTTPProvider {
	providers := map[string]HTTPProvider{}
	for agentName, providerName := range agentProviders {
		provider := cfg.Providers[providerName]
		if !provider.HTTP.IsZero() {
			providers[agentName] = provider.HTTP
		}
	}
	return providers
}

func validateHTTPProvider(name string, provider HTTPProvider) error {
	if strings.TrimSpace(provider.URL) == "" {
		return fmt.Errorf("%s.http.url: %w", name, ErrInvalidConfig)
	}
	if strings.TrimSpace(provider.Model) == "" {
		return fmt.Errorf("%s.http.model: %w", name, ErrInvalidConfig)
	}
	if _, err := url.ParseRequestURI(provider.URL); err != nil {
		return fmt.Errorf("%s.http.url: %w", name, ErrInvalidConfig)
	}
	apiKeyEnv := strings.TrimSpace(provider.APIKeyEnv)
	if apiKeyEnv != "" && !validEnvVarName(apiKeyEnv) {
		return fmt.Errorf("%s.http.api_key_env: %w", name, ErrInvalidConfig)
	}
	if provider.InputCostPerMillionTokens < 0 {
		return fmt.Errorf("%s.http.input_cost_per_million_tokens: %w", name, ErrInvalidConfig)
	}
	if provider.OutputCostPerMillionTokens < 0 {
		return fmt.Errorf("%s.http.output_cost_per_million_tokens: %w", name, ErrInvalidConfig)
	}
	if provider.TimeoutMS < 0 {
		return fmt.Errorf("%s.http.timeout_ms: %w", name, ErrInvalidConfig)
	}
	if provider.MaxOutputTokens < 0 {
		return fmt.Errorf("%s.http.max_output_tokens: %w", name, ErrInvalidConfig)
	}
	if !validHTTPResponseFormat(provider.ResponseFormat) {
		return fmt.Errorf("%s.http.response_format: %w", name, ErrInvalidConfig)
	}
	return nil
}

func validHTTPResponseFormat(value string) bool {
	switch strings.TrimSpace(value) {
	case "", "text", "json_object":
		return true
	default:
		return false
	}
}
