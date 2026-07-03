package config

import (
	"sort"
	"strings"

	"ceoharness/internal/jobpacket"
)

type ProviderRouteDecision struct {
	AgentName    string `json:"agent_name"`
	ProviderName string `json:"provider_name"`
	Reason       string `json:"reason"`
	FallbackFrom string `json:"fallback_from,omitempty"`
}

func (cfg Config) ProviderRouteDecisionsForTask(task string, subagents []jobpacket.Subagent) ([]ProviderRouteDecision, error) {
	packet, err := jobpacket.BuildWithOptions(jobpacket.BuildOptions{
		Task:         task,
		Subagents:    subagents,
		MaxSubagents: cfg.MaxSubagents,
	})
	if err != nil {
		return nil, err
	}
	decisions := decisionsFromAgentProviders(cfg.AgentProviders)
	applySubagentProviderDecisions(decisions, packet.Subagents)
	fixedRoutes := decisionNames(decisions)
	if cfg.HasProviderPolicy() {
		applyProfileProviderDecisions(profileProviderDecisionInput{
			Decisions:   decisions,
			FixedRoutes: fixedRoutes,
			Policy:      cfg.ProviderPolicy,
			Subagents:   packet.Subagents,
			Profile:     packet.TaskProfile,
		})
		applyRiskAreaProviderDecisions(riskAreaProviderDecisionInput{
			Decisions:   decisions,
			FixedRoutes: fixedRoutes,
			Policy:      cfg.ProviderPolicy,
			Subagents:   packet.Subagents,
			RiskAreas:   packet.TaskProfile.RiskAreas,
		})
	}
	return orderedDecisions(decisions, packet.Subagents), nil
}

func decisionsFromAgentProviders(agentProviders map[string]string) map[string]ProviderRouteDecision {
	decisions := make(map[string]ProviderRouteDecision, len(agentProviders))
	for agentName, providerName := range agentProviders {
		cleanAgent := strings.TrimSpace(agentName)
		cleanProvider := strings.TrimSpace(providerName)
		if cleanAgent == "" || cleanProvider == "" {
			continue
		}
		decisions[cleanAgent] = ProviderRouteDecision{
			AgentName:    cleanAgent,
			ProviderName: cleanProvider,
			Reason:       "agent_providers",
		}
	}
	return decisions
}

func applySubagentProviderDecisions(decisions map[string]ProviderRouteDecision, subagents []jobpacket.Subagent) {
	for _, agent := range subagents {
		providerName := strings.TrimSpace(agent.ProviderName)
		if providerName == "" {
			continue
		}
		decisions[agent.Name] = ProviderRouteDecision{
			AgentName:    agent.Name,
			ProviderName: providerName,
			Reason:       "subagent.provider",
		}
	}
}

type profileProviderDecisionInput struct {
	Decisions   map[string]ProviderRouteDecision
	FixedRoutes map[string]struct{}
	Policy      ProviderPolicy
	Subagents   []jobpacket.Subagent
	Profile     jobpacket.TaskProfile
}

func applyProfileProviderDecisions(input profileProviderDecisionInput) {
	providerName, reason := input.Policy.ProviderForProfileDecision(input.Profile)
	if providerName == "" {
		return
	}
	for _, agent := range input.Subagents {
		if _, fixed := input.FixedRoutes[agent.Name]; fixed {
			continue
		}
		input.Decisions[agent.Name] = ProviderRouteDecision{
			AgentName:    agent.Name,
			ProviderName: providerName,
			Reason:       reason,
		}
	}
}

type riskAreaProviderDecisionInput struct {
	Decisions   map[string]ProviderRouteDecision
	FixedRoutes map[string]struct{}
	Policy      ProviderPolicy
	Subagents   []jobpacket.Subagent
	RiskAreas   []string
}

func applyRiskAreaProviderDecisions(input riskAreaProviderDecisionInput) {
	areas := riskAreaSet(input.RiskAreas)
	for _, agent := range input.Subagents {
		if _, fixed := input.FixedRoutes[agent.Name]; fixed {
			continue
		}
		if _, matched := areas[agent.Name]; !matched {
			continue
		}
		if providerName := strings.TrimSpace(input.Policy.RiskAreaProviders[agent.Name]); providerName != "" {
			input.Decisions[agent.Name] = ProviderRouteDecision{
				AgentName:    agent.Name,
				ProviderName: providerName,
				Reason:       "provider_policy.risk_area_providers." + agent.Name,
			}
		}
	}
}

func orderedDecisions(decisions map[string]ProviderRouteDecision, subagents []jobpacket.Subagent) []ProviderRouteDecision {
	ordered := make([]ProviderRouteDecision, 0, len(decisions))
	seen := map[string]struct{}{}
	for _, agent := range subagents {
		if decision, ok := decisions[agent.Name]; ok {
			ordered = append(ordered, decision)
			seen[agent.Name] = struct{}{}
		}
	}
	extraNames := make([]string, 0, len(decisions))
	for agentName := range decisions {
		if _, ok := seen[agentName]; ok {
			continue
		}
		extraNames = append(extraNames, agentName)
	}
	sort.Strings(extraNames)
	for _, agentName := range extraNames {
		ordered = append(ordered, decisions[agentName])
	}
	return ordered
}

func routesFromDecisions(decisions []ProviderRouteDecision) map[string]string {
	routes := make(map[string]string, len(decisions))
	for _, decision := range decisions {
		routes[decision.AgentName] = decision.ProviderName
	}
	return routes
}

func decisionNames(decisions map[string]ProviderRouteDecision) map[string]struct{} {
	names := make(map[string]struct{}, len(decisions))
	for agentName := range decisions {
		names[agentName] = struct{}{}
	}
	return names
}

func riskAreaSet(riskAreas []string) map[string]struct{} {
	areas := make(map[string]struct{}, len(riskAreas))
	for _, riskArea := range riskAreas {
		areas[strings.TrimSpace(riskArea)] = struct{}{}
	}
	return areas
}
