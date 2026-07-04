package cli

import (
	"fmt"
	"strings"
)

type httpProviderPreset struct {
	URL             string
	APIKeyEnv       string
	DisableThinking bool
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
	case "kimi", "kimi-code", "kimicode":
		return httpProviderPreset{
			URL:       "https://api.kimi.com/coding/v1/chat/completions",
			APIKeyEnv: "KIMI_CODE_API_KEY",
		}, nil
	case "moonshot":
		return httpProviderPreset{
			URL:       "https://api.moonshot.ai/v1/chat/completions",
			APIKeyEnv: "MOONSHOT_API_KEY",
		}, nil
	case "minimax":
		return httpProviderPreset{
			URL:             "https://api.minimax.io/v1/chat/completions",
			APIKeyEnv:       "MINIMAX_API_KEY",
			DisableThinking: true,
		}, nil
	default:
		return httpProviderPreset{}, fmt.Errorf("unknown --http-preset %q", name)
	}
}
