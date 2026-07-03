package cli

import (
	"context"
	"fmt"
	"strings"

	"ceoharness/internal/config"
	"ceoharness/internal/model"
	"ceoharness/internal/subagent"
)

type lazyProviderClient struct {
	name      string
	provider  config.Provider
	timeoutMS int
}

func (c lazyProviderClient) Complete(ctx context.Context, req model.Request) (model.Response, error) {
	if err := ctx.Err(); err != nil {
		return model.Response{}, err
	}
	client, err := clientForProvider(c.provider, c.timeoutMS)
	if err != nil {
		return model.Response{}, fmt.Errorf("create provider %s client: %w", c.name, err)
	}
	return client.Complete(ctx, req)
}

func providerClientsFromSelection(selection modelCommandSelection) (map[string]model.Client, map[string]subagent.RouteMetadata) {
	clients := make(map[string]model.Client, len(selection.providerConfigs))
	metadata := make(map[string]subagent.RouteMetadata, len(selection.providerConfigs))
	for rawName, provider := range selection.providerConfigs {
		name := strings.TrimSpace(rawName)
		if name == "" {
			continue
		}
		clients[name] = lazyProviderClient{
			name:      name,
			provider:  provider,
			timeoutMS: selection.modelCommandTimeoutMS,
		}
		metadata[name] = subagent.RouteMetadata{
			Source:       providerRouteSource(provider),
			ProviderName: name,
		}
	}
	return clients, metadata
}

func clientForProvider(provider config.Provider, timeoutMS int) (model.Client, error) {
	if len(provider.ModelCommand) > 0 {
		env, err := providerEnvPairs(providerEnvVars(provider))
		if err != nil {
			return nil, err
		}
		return newModelCommandClientWithEnv(provider.ModelCommand, env, timeoutMS)
	}
	if !provider.HTTP.IsZero() {
		apiKey, err := providerAPIKey(provider.HTTP.APIKeyEnv)
		if err != nil {
			return nil, err
		}
		return newHTTPProviderClient(provider.HTTP, apiKey)
	}
	return nil, fmt.Errorf("provider backend: %w", config.ErrInvalidConfig)
}

func providerEnvVars(provider config.Provider) []string {
	names := append([]string(nil), provider.EnvVars...)
	if strings.TrimSpace(provider.HTTP.APIKeyEnv) != "" {
		names = append(names, provider.HTTP.APIKeyEnv)
	}
	return names
}

func providerRouteSource(provider config.Provider) string {
	if len(provider.ModelCommand) > 0 {
		return "command"
	}
	if !provider.HTTP.IsZero() {
		return "http"
	}
	return ""
}
