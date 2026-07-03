package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"ceoharness/internal/history"
)

func Test_Run_prints_provider_health_policy_when_config_check_runs(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	configJSON := `{"provider_health_avoid_failure_rate":0.9,"provider_health_watch_failure_rate":0.5,"provider_health_watch_cost_per_attempt_microusd":50}`
	if err := os.WriteFile(filepath.Join(root, ".ceo-harness.json"), []byte(configJSON), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	// When
	err := Run(context.Background(), &out, []string{"--workspace", root, "--config-check"})
	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v\n%s", err, out.String())
	}
	var body struct {
		ProviderHealthAvoidFailureRate            float64 `json:"provider_health_avoid_failure_rate"`
		ProviderHealthWatchFailureRate            float64 `json:"provider_health_watch_failure_rate"`
		ProviderHealthWatchCostPerAttemptMicroUSD int64   `json:"provider_health_watch_cost_per_attempt_microusd"`
	}
	if err := json.Unmarshal(out.Bytes(), &body); err != nil {
		t.Fatalf("output must be JSON: %v\n%s", err, out.String())
	}
	if body.ProviderHealthAvoidFailureRate != 0.9 {
		t.Fatalf("ProviderHealthAvoidFailureRate = %f, want 0.9", body.ProviderHealthAvoidFailureRate)
	}
	if body.ProviderHealthWatchFailureRate != 0.5 {
		t.Fatalf("ProviderHealthWatchFailureRate = %f, want 0.5", body.ProviderHealthWatchFailureRate)
	}
	if body.ProviderHealthWatchCostPerAttemptMicroUSD != 50 {
		t.Fatalf("ProviderHealthWatchCostPerAttemptMicroUSD = %d, want 50", body.ProviderHealthWatchCostPerAttemptMicroUSD)
	}
}

func Test_Run_prints_provider_health_route_avoidance_when_config_check_runs(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	configJSON := `{"providers":{"cheap":{"model_command":["echo","cheap"]},"premium":{"model_command":["echo","premium"]}},"agent_providers":{"planner":"cheap","coder":"cheap"},"provider_policy":{"fallback_provider":"premium"}}`
	if err := os.WriteFile(filepath.Join(root, ".ceo-harness.json"), []byte(configJSON), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	store, err := history.New(root)
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	_, err = store.Append(context.Background(), history.Entry{
		Task:    "bad cheap run",
		Verdict: "fail",
		ProviderHealth: []history.ProviderHealth{
			{ProviderName: "cheap", ModelSource: "command", AttemptCount: 1, FailCount: 1, ErrorCount: 1},
		},
	})
	if err != nil {
		t.Fatalf("Append returned error: %v", err)
	}

	// When
	err = Run(context.Background(), &out, []string{"--workspace", root, "--config-check"})
	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v\n%s", err, out.String())
	}
	var body struct {
		ProviderHealthAvoidedRouteCount int      `json:"provider_health_avoided_route_count"`
		ProviderHealthAvoidedProviders  []string `json:"provider_health_avoided_providers"`
	}
	if err := json.Unmarshal(out.Bytes(), &body); err != nil {
		t.Fatalf("output must be JSON: %v\n%s", err, out.String())
	}
	if body.ProviderHealthAvoidedRouteCount != 2 {
		t.Fatalf("ProviderHealthAvoidedRouteCount = %d, want 2", body.ProviderHealthAvoidedRouteCount)
	}
	if len(body.ProviderHealthAvoidedProviders) != 1 || body.ProviderHealthAvoidedProviders[0] != "cheap" {
		t.Fatalf("ProviderHealthAvoidedProviders = %#v, want [cheap]", body.ProviderHealthAvoidedProviders)
	}
}
