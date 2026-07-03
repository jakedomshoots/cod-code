package cli

import (
	"fmt"
	"strings"
)

type httpProviderPreset struct {
	URL       string
	APIKeyEnv string
}

func resolveHTTPProviderPreset(name string) (httpProviderPreset, error) {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "":
		return httpProviderPreset{}, nil
	case "openai":
		return httpProviderPreset{
			URL:       "https://api.openai.com/v1/chat/completions",
			APIKeyEnv: "OPENAI_API_KEY",
		}, nil
	case "openrouter":
		return httpProviderPreset{
			URL:       "https://openrouter.ai/api/v1/chat/completions",
			APIKeyEnv: "OPENROUTER_API_KEY",
		}, nil
	case "kimi", "moonshot":
		return httpProviderPreset{
			URL:       "https://api.moonshot.ai/v1/chat/completions",
			APIKeyEnv: "MOONSHOT_API_KEY",
		}, nil
	default:
		return httpProviderPreset{}, fmt.Errorf("unknown --http-preset %q", name)
	}
}
