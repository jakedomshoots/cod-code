package cli

import (
	"context"
	"fmt"
	"os"

	"ceoharness/internal/ceo"
	"ceoharness/internal/config"
	"ceoharness/internal/model"
	"ceoharness/internal/subagent"
)

const (
	modelCommandEnv    = "CEO_MODEL_COMMAND_JSON"
	ceoModelCommandEnv = "CEO_REVIEW_MODEL_COMMAND_JSON"
	researchCommandEnv = "CEO_RESEARCH_COMMAND_JSON"
)

type runtimeFromOptionsResult struct {
	runtime                 ceo.Runtime
	providerHealthAvoidance providerHealthRouteAvoidance
	providerRouteDecisions  []config.ProviderRouteDecision
}

func runtimeFromOptions(ctx context.Context, opts options) (runtimeFromOptionsResult, error) {
	selection, err := selectModelCommand(ctx, opts)
	if err != nil {
		return runtimeFromOptionsResult{}, err
	}
	result := runtimeFromOptionsResult{
		providerRouteDecisions: append([]config.ProviderRouteDecision(nil), selection.providerRouteDecisions...),
		providerHealthAvoidance: providerHealthRouteAvoidance{
			avoidedRouteCount: selection.providerHealthAvoidedRouteCount,
			avoidedProviders:  append([]string(nil), selection.providerHealthAvoidedProviders...),
		},
	}
	ceoSelection, err := selectCEOModelCommand(ctx, opts)
	if err != nil {
		return runtimeFromOptionsResult{}, err
	}
	ceoReviewer, err := ceoReviewerFromSelection(ceoSelection)
	if err != nil {
		return runtimeFromOptionsResult{}, err
	}
	ceoReviewerRoute := ceoReviewerRouteFromSelection(ceoSelection)
	providerClients, providerMetadata := providerClientsFromSelection(selection)
	if len(selection.argv) == 0 && len(selection.agentArgv) == 0 && len(selection.agentHTTPProviders) == 0 && len(providerClients) == 0 {
		if ceoReviewer != nil {
			result.runtime = ceo.NewRuntimeWithCEOReviewerAndRoute(ceoReviewer, ceoReviewerRoute)
			return result, nil
		}
		result.runtime = ceo.NewRuntime()
		return result, nil
	}
	defaultClient := model.Client(model.NewStaticClient())
	defaultMetadata := subagent.RouteMetadata{Source: "local"}
	if len(selection.argv) > 0 {
		defaultClient, err = newModelCommandClient(selection.argv, selection.modelCommandTimeoutMS)
		if err != nil {
			return runtimeFromOptionsResult{}, err
		}
		defaultMetadata = subagent.RouteMetadata{Source: "command"}
	}
	agentClients := map[string]model.Client{}
	agentMetadata := map[string]subagent.RouteMetadata{}
	for agentName, argv := range selection.agentArgv {
		env, err := providerEnvPairs(selection.agentEnvVars[agentName])
		if err != nil {
			return runtimeFromOptionsResult{}, err
		}
		client, err := newModelCommandClientWithEnv(argv, env, selection.modelCommandTimeoutMS)
		if err != nil {
			return runtimeFromOptionsResult{}, fmt.Errorf("create %s model command client: %w", agentName, err)
		}
		agentClients[agentName] = client
		agentMetadata[agentName] = subagent.RouteMetadata{
			Source:       "command",
			ProviderName: selection.agentProviderNames[agentName],
		}
	}
	for agentName, provider := range selection.agentHTTPProviders {
		if _, ok := agentClients[agentName]; ok {
			continue
		}
		apiKey, err := providerAPIKey(provider.APIKeyEnv)
		if err != nil {
			return runtimeFromOptionsResult{}, err
		}
		client, err := newHTTPProviderClient(provider, apiKey)
		if err != nil {
			return runtimeFromOptionsResult{}, fmt.Errorf("create %s http provider client: %w", agentName, err)
		}
		agentClients[agentName] = client
		agentMetadata[agentName] = subagent.RouteMetadata{
			Source:       "http",
			ProviderName: selection.agentProviderNames[agentName],
		}
	}
	fallbackClient, fallbackMetadata, err := fallbackClientFromSelection(selection)
	if err != nil {
		return runtimeFromOptionsResult{}, err
	}
	result.runtime = ceo.NewRuntimeWithSubagentRunnerAndCEOReviewerRoute(subagent.NewRoutingRunnerWithConfig(subagent.RoutingConfig{
		DefaultClient:    defaultClient,
		DefaultMetadata:  defaultMetadata,
		Clients:          agentClients,
		Metadata:         agentMetadata,
		ProviderClients:  providerClients,
		ProviderMetadata: providerMetadata,
		FallbackClient:   fallbackClient,
		FallbackMetadata: fallbackMetadata,
		MinConfidence:    selection.minSubagentConfidence,
	}), ceoReviewer, ceoReviewerRoute)
	return result, nil
}

func fallbackClientFromSelection(selection modelCommandSelection) (model.Client, subagent.RouteMetadata, error) {
	if selection.fallbackProviderName == "" {
		return nil, subagent.RouteMetadata{}, nil
	}
	metadata := subagent.RouteMetadata{
		ProviderName: selection.fallbackProviderName,
	}
	if len(selection.fallbackArgv) > 0 {
		env, err := providerEnvPairs(selection.fallbackEnvVars)
		if err != nil {
			return nil, subagent.RouteMetadata{}, err
		}
		client, err := newModelCommandClientWithEnv(selection.fallbackArgv, env, selection.modelCommandTimeoutMS)
		if err != nil {
			return nil, subagent.RouteMetadata{}, fmt.Errorf("create fallback model command client: %w", err)
		}
		metadata.Source = "command"
		return client, metadata, nil
	}
	if !selection.fallbackHTTPProvider.IsZero() {
		apiKey, err := providerAPIKey(selection.fallbackHTTPProvider.APIKeyEnv)
		if err != nil {
			return nil, subagent.RouteMetadata{}, err
		}
		client, err := newHTTPProviderClient(selection.fallbackHTTPProvider, apiKey)
		if err != nil {
			return nil, subagent.RouteMetadata{}, fmt.Errorf("create fallback http provider client: %w", err)
		}
		metadata.Source = "http"
		return client, metadata, nil
	}
	return nil, subagent.RouteMetadata{}, nil
}

func ceoReviewerFromSelection(selection commandSelection) (model.Client, error) {
	if len(selection.argv) == 0 {
		if selection.providerName == "" {
			return nil, nil
		}
		client, err := clientForProvider(selection.provider, selection.timeoutMS)
		if err != nil {
			return nil, fmt.Errorf("create CEO provider client: %w", err)
		}
		return client, nil
	}
	client, err := newModelCommandClient(selection.argv, selection.timeoutMS)
	if err != nil {
		return nil, fmt.Errorf("create CEO model command client: %w", err)
	}
	return client, nil
}

func ceoReviewerRouteFromSelection(selection commandSelection) subagent.RouteMetadata {
	if selection.providerName != "" {
		return subagent.RouteMetadata{
			Source:       providerRouteSource(selection.provider),
			ProviderName: selection.providerName,
		}
	}
	if len(selection.argv) > 0 {
		return subagent.RouteMetadata{Source: "command"}
	}
	return subagent.RouteMetadata{}
}

func newModelCommandClient(argv []string, timeoutMS int) (model.Client, error) {
	return newModelCommandClientWithEnv(argv, nil, timeoutMS)
}

func newModelCommandClientWithEnv(argv []string, env []string, timeoutMS int) (model.Client, error) {
	client, err := model.NewCommandClient(model.CommandSpec{
		Argv:      argv,
		Env:       env,
		TimeoutMS: timeoutMS,
	})
	if err != nil {
		return nil, fmt.Errorf("create model command client: %w", err)
	}
	return client, nil
}

func providerEnvPairs(names []string) ([]string, error) {
	pairs := []string{}
	for _, name := range names {
		value, ok := os.LookupEnv(name)
		if !ok || value == "" {
			return nil, fmt.Errorf("provider env var %s is required", name)
		}
		pairs = append(pairs, name+"="+value)
	}
	return pairs, nil
}
