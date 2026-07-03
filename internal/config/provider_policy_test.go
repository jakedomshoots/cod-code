package config

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func Test_LoadWorkspace_reads_provider_policy_when_config_exists(t *testing.T) {
	// Given
	root := t.TempDir()
	content := `{"providers":{"cheap":{"model_command":["echo","cheap"]},"premium":{"model_command":["echo","premium"]}},"provider_policy":{"default_provider":"cheap","fallback_provider":"premium","risk_providers":{"high":"premium"},"kind_providers":{"research":"premium"},"risk_area_providers":{"database":"premium"}}}`
	if err := os.WriteFile(filepath.Join(root, ".ceo-harness.json"), []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	// When
	cfg, err := LoadWorkspace(context.Background(), root)

	// Then
	if err != nil {
		t.Fatalf("LoadWorkspace returned error: %v", err)
	}
	if cfg.ProviderPolicy.DefaultProvider != "cheap" {
		t.Fatalf("DefaultProvider = %q, want cheap", cfg.ProviderPolicy.DefaultProvider)
	}
	if cfg.ProviderPolicy.FallbackProvider != "premium" {
		t.Fatalf("FallbackProvider = %q, want premium", cfg.ProviderPolicy.FallbackProvider)
	}
	if cfg.ProviderPolicy.RiskProviders["high"] != "premium" {
		t.Fatalf("RiskProviders = %#v, want high premium", cfg.ProviderPolicy.RiskProviders)
	}
	if cfg.ProviderPolicy.KindProviders["research"] != "premium" {
		t.Fatalf("KindProviders = %#v, want research premium", cfg.ProviderPolicy.KindProviders)
	}
	if cfg.ProviderPolicy.RiskAreaProviders["database"] != "premium" {
		t.Fatalf("RiskAreaProviders = %#v, want database premium", cfg.ProviderPolicy.RiskAreaProviders)
	}
}

func Test_LoadWorkspace_rejects_unknown_provider_policy_fallback(t *testing.T) {
	// Given
	root := t.TempDir()
	content := `{"providers":{"cheap":{"model_command":["echo","cheap"]}},"provider_policy":{"fallback_provider":"missing"}}`
	if err := os.WriteFile(filepath.Join(root, ".ceo-harness.json"), []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	// When
	_, err := LoadWorkspace(context.Background(), root)

	// Then
	if !errors.Is(err, ErrInvalidConfig) {
		t.Fatalf("error = %v, want ErrInvalidConfig", err)
	}
}

func Test_LoadWorkspace_rejects_unknown_provider_policy_target(t *testing.T) {
	// Given
	root := t.TempDir()
	content := `{"providers":{"cheap":{"model_command":["echo","cheap"]}},"provider_policy":{"risk_providers":{"high":"missing"}}}`
	if err := os.WriteFile(filepath.Join(root, ".ceo-harness.json"), []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	// When
	_, err := LoadWorkspace(context.Background(), root)

	// Then
	if !errors.Is(err, ErrInvalidConfig) {
		t.Fatalf("error = %v, want ErrInvalidConfig", err)
	}
}

func Test_LoadWorkspace_rejects_unknown_provider_policy_risk_area(t *testing.T) {
	// Given
	root := t.TempDir()
	content := `{"providers":{"premium":{"model_command":["echo","premium"]}},"provider_policy":{"risk_area_providers":{"unknown":"premium"}}}`
	if err := os.WriteFile(filepath.Join(root, ".ceo-harness.json"), []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	// When
	_, err := LoadWorkspace(context.Background(), root)

	// Then
	if !errors.Is(err, ErrInvalidConfig) {
		t.Fatalf("error = %v, want ErrInvalidConfig", err)
	}
}

func Test_ProviderPolicyRoutesForTask_routes_agents_by_risk_when_no_explicit_agent_route_exists(t *testing.T) {
	// Given
	cfg := Config{
		Providers: map[string]Provider{
			"cheap":   {ModelCommand: []string{"echo", "cheap"}},
			"premium": {ModelCommand: []string{"echo", "premium"}},
		},
		ProviderPolicy: ProviderPolicy{
			DefaultProvider: "cheap",
			RiskProviders:   map[string]string{"high": "premium"},
		},
	}

	// When
	routes, err := cfg.ProviderRoutesForTask("Fix auth bug", nil)

	// Then
	if err != nil {
		t.Fatalf("ProviderRoutesForTask returned error: %v", err)
	}
	for _, agentName := range []string{"coder", "security", "reviewer"} {
		if routes[agentName] != "premium" {
			t.Fatalf("route[%s] = %q, want premium", agentName, routes[agentName])
		}
	}
}

func Test_ProviderPolicyRoutesForTask_keeps_explicit_agent_provider_when_policy_matches(t *testing.T) {
	// Given
	cfg := Config{
		Providers: map[string]Provider{
			"cheap":   {ModelCommand: []string{"echo", "cheap"}},
			"premium": {ModelCommand: []string{"echo", "premium"}},
		},
		AgentProviders: map[string]string{"scanner": "cheap"},
		ProviderPolicy: ProviderPolicy{
			DefaultProvider: "cheap",
			RiskProviders:   map[string]string{"high": "premium"},
		},
	}

	// When
	routes, err := cfg.ProviderRoutesForTask("Fix auth bug", nil)

	// Then
	if err != nil {
		t.Fatalf("ProviderRoutesForTask returned error: %v", err)
	}
	if routes["scanner"] != "cheap" {
		t.Fatalf("scanner route = %q, want explicit cheap", routes["scanner"])
	}
	if routes["coder"] != "premium" || routes["reviewer"] != "premium" {
		t.Fatalf("routes = %#v, want policy premium for unset agents", routes)
	}
}

func Test_ProviderPolicyRoutesForTask_routes_matching_risk_area_specialists_only(t *testing.T) {
	// Given
	cfg := Config{
		Providers: map[string]Provider{
			"cheap":   {ModelCommand: []string{"echo", "cheap"}},
			"premium": {ModelCommand: []string{"echo", "premium"}},
		},
		ProviderPolicy: ProviderPolicy{
			DefaultProvider: "cheap",
			RiskAreaProviders: map[string]string{
				"billing":  "premium",
				"database": "premium",
			},
		},
		MaxSubagents: 7,
	}

	// When
	routes, err := cfg.ProviderRoutesForTask("Research payment database migration and deploy the fix", nil)

	// Then
	if err != nil {
		t.Fatalf("ProviderRoutesForTask returned error: %v", err)
	}
	for _, agentName := range []string{"planner", "researcher", "coder", "release", "reviewer"} {
		if routes[agentName] != "cheap" {
			t.Fatalf("route[%s] = %q, want cheap", agentName, routes[agentName])
		}
	}
	for _, agentName := range []string{"billing", "database"} {
		if routes[agentName] != "premium" {
			t.Fatalf("route[%s] = %q, want premium", agentName, routes[agentName])
		}
	}
}

func Test_ProviderRouteDecisionsForTask_explains_risk_area_policy_routes(t *testing.T) {
	// Given
	cfg := Config{
		Providers: map[string]Provider{
			"cheap":   {ModelCommand: []string{"echo", "cheap"}},
			"premium": {ModelCommand: []string{"echo", "premium"}},
		},
		ProviderPolicy: ProviderPolicy{
			DefaultProvider:   "cheap",
			RiskAreaProviders: map[string]string{"database": "premium"},
		},
	}

	// When
	decisions, err := cfg.ProviderRouteDecisionsForTask("Implement database migration fix", nil)

	// Then
	if err != nil {
		t.Fatalf("ProviderRouteDecisionsForTask returned error: %v", err)
	}
	assertProviderDecision(t, decisions, "coder", "cheap", "provider_policy.default_provider")
	assertProviderDecision(t, decisions, "database", "premium", "provider_policy.risk_area_providers.database")
	assertProviderDecision(t, decisions, "reviewer", "cheap", "provider_policy.default_provider")
}

func Test_ProviderPolicyRoutesForTask_keeps_explicit_agent_provider_when_risk_area_policy_matches(t *testing.T) {
	// Given
	cfg := Config{
		Providers: map[string]Provider{
			"cheap":   {ModelCommand: []string{"echo", "cheap"}},
			"premium": {ModelCommand: []string{"echo", "premium"}},
		},
		AgentProviders: map[string]string{"database": "cheap"},
		ProviderPolicy: ProviderPolicy{
			DefaultProvider:   "cheap",
			RiskAreaProviders: map[string]string{"database": "premium"},
		},
	}

	// When
	routes, err := cfg.ProviderRoutesForTask("Implement database migration fix", nil)

	// Then
	if err != nil {
		t.Fatalf("ProviderRoutesForTask returned error: %v", err)
	}
	if routes["database"] != "cheap" {
		t.Fatalf("database route = %q, want explicit cheap", routes["database"])
	}
	if routes["coder"] != "cheap" || routes["reviewer"] != "cheap" {
		t.Fatalf("routes = %#v, want default cheap for unset agents", routes)
	}
}

func assertProviderDecision(t *testing.T, decisions []ProviderRouteDecision, agentName string, providerName string, reason string) {
	t.Helper()
	for _, decision := range decisions {
		if decision.AgentName != agentName {
			continue
		}
		if decision.ProviderName != providerName || decision.Reason != reason {
			t.Fatalf("decision for %s = %#v, want provider %q reason %q", agentName, decision, providerName, reason)
		}
		return
	}
	t.Fatalf("missing provider decision for %s in %#v", agentName, decisions)
}
