package cli

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"

	"ceoharness/internal/config"
	"ceoharness/internal/history"
)

type providerRouteSelection struct {
	routes          map[string]string
	decisions       []config.ProviderRouteDecision
	healthAvoidance providerHealthRouteAvoidance
}

type providerHealthRouteAvoidance struct {
	avoidedRouteCount int
	avoidedProviders  []string
}

type providerHealthAvoidanceInput struct {
	routes       map[string]string
	decisions    []config.ProviderRouteDecision
	workspaceDir string
	cfg          config.Config
}

func providerRoutesForSelection(ctx context.Context, cfg config.Config, opts options) (providerRouteSelection, error) {
	selection, err := baseProviderRoutesForSelection(cfg, opts)
	if err != nil {
		return providerRouteSelection{}, err
	}
	return routesWithProviderHealthAvoidance(ctx, providerHealthAvoidanceInput{
		routes:       selection.routes,
		decisions:    append([]config.ProviderRouteDecision(nil), selection.decisions...),
		workspaceDir: opts.workspaceDir,
		cfg:          cfg,
	})
}

func baseProviderRoutesForSelection(cfg config.Config, opts options) (providerRouteSelection, error) {
	if strings.TrimSpace(opts.task) == "" {
		decisions := providerDecisionsFromAgentProviders(cfg.AgentProviders)
		return providerRouteSelection{
			routes:    routesFromProviderDecisions(decisions),
			decisions: decisions,
		}, nil
	}
	decisions, err := cfg.ProviderRouteDecisionsForTask(opts.task, opts.subagents)
	if err != nil {
		return providerRouteSelection{}, fmt.Errorf("resolve provider policy: %w", err)
	}
	return providerRouteSelection{
		routes:    routesFromProviderDecisions(decisions),
		decisions: decisions,
	}, nil
}

func routesWithProviderHealthAvoidance(ctx context.Context, input providerHealthAvoidanceInput) (providerRouteSelection, error) {
	selection := providerRouteSelection{routes: input.routes, decisions: append([]config.ProviderRouteDecision(nil), input.decisions...)}
	fallbackProvider := strings.TrimSpace(input.cfg.ProviderPolicy.FallbackProvider)
	if fallbackProvider == "" || strings.TrimSpace(input.workspaceDir) == "" {
		return selection, nil
	}
	avoided, err := avoidedProvidersFromHistory(ctx, input.workspaceDir, providerHealthPolicyFromConfig(input.cfg))
	if err != nil {
		return providerRouteSelection{}, err
	}
	if _, fallbackAvoided := avoided[fallbackProvider]; fallbackAvoided || len(avoided) == 0 {
		return selection, nil
	}
	changedProviders := map[string]struct{}{}
	for agentName, providerName := range selection.routes {
		if _, shouldAvoid := avoided[providerName]; shouldAvoid {
			selection.routes[agentName] = fallbackProvider
			updateProviderHealthDecision(selection.decisions, agentName, providerName, fallbackProvider)
			selection.healthAvoidance.avoidedRouteCount++
			changedProviders[providerName] = struct{}{}
		}
	}
	selection.healthAvoidance.avoidedProviders = sortedProviderNames(changedProviders)
	return selection, nil
}

func providerDecisionsFromAgentProviders(agentProviders map[string]string) []config.ProviderRouteDecision {
	decisions := make([]config.ProviderRouteDecision, 0, len(agentProviders))
	for agentName, providerName := range agentProviders {
		cleanAgent := strings.TrimSpace(agentName)
		cleanProvider := strings.TrimSpace(providerName)
		if cleanAgent == "" || cleanProvider == "" {
			continue
		}
		decisions = append(decisions, config.ProviderRouteDecision{
			AgentName:    cleanAgent,
			ProviderName: cleanProvider,
			Reason:       "agent_providers",
		})
	}
	sort.Slice(decisions, func(left, right int) bool {
		return decisions[left].AgentName < decisions[right].AgentName
	})
	return decisions
}

func routesFromProviderDecisions(decisions []config.ProviderRouteDecision) map[string]string {
	routes := make(map[string]string, len(decisions))
	for _, decision := range decisions {
		routes[decision.AgentName] = decision.ProviderName
	}
	return routes
}

func updateProviderHealthDecision(decisions []config.ProviderRouteDecision, agentName string, fromProvider string, toProvider string) {
	for index := range decisions {
		if decisions[index].AgentName != agentName {
			continue
		}
		decisions[index].FallbackFrom = fromProvider
		decisions[index].ProviderName = toProvider
		decisions[index].Reason = "provider_health.fallback"
		return
	}
}

func avoidedProvidersFromHistory(ctx context.Context, workspaceDir string, policy history.ProviderHealthPolicy) (map[string]struct{}, error) {
	avoided := map[string]struct{}{}
	store, err := history.New(workspaceDir)
	if err != nil {
		return nil, err
	}
	entries, err := store.ReadAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("read provider health history: %w", err)
	}
	for _, row := range history.AggregateProviderHealthWithPolicy(entries, policy) {
		if row.Recommendation == "avoid" {
			avoided[row.ProviderName] = struct{}{}
		}
	}
	return avoided, nil
}

func cloneAgentProviders(providers map[string]string) map[string]string {
	copied := make(map[string]string, len(providers))
	for agentName, providerName := range providers {
		copied[agentName] = providerName
	}
	return copied
}

func cloneProviderConfigs(providers map[string]config.Provider) map[string]config.Provider {
	copied := make(map[string]config.Provider, len(providers))
	for providerName, provider := range providers {
		copied[providerName] = config.Provider{
			ModelCommand: append([]string(nil), provider.ModelCommand...),
			EnvVars:      append([]string(nil), provider.EnvVars...),
			HTTP:         provider.HTTP,
		}
	}
	return copied
}

func sortedProviderNames(providers map[string]struct{}) []string {
	names := make([]string, 0, len(providers))
	for providerName := range providers {
		names = append(names, providerName)
	}
	sort.Strings(names)
	return names
}

func providerEnvCounts(names []string) (present int, missing int, missingNames []string) {
	for _, name := range names {
		value, ok := os.LookupEnv(name)
		if ok && value != "" {
			present++
			continue
		}
		missing++
		missingNames = append(missingNames, name)
	}
	return present, missing, missingNames
}
