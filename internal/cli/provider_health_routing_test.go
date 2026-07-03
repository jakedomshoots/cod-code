package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"ceoharness/internal/history"
)

func Test_Run_routes_away_from_avoided_provider_to_fallback(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	configJSON := `{"providers":{"cheap":{"model_command":["sh","-c","cat >/dev/null; printf cheap-route"]},"premium":{"model_command":["sh","-c","cat >/dev/null; printf premium-route"]}},"provider_policy":{"default_provider":"cheap","fallback_provider":"premium"}}`
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
	err = Run(context.Background(), &out, []string{"--workspace", root, "Plan", "roadmap"})
	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v\n%s", err, out.String())
	}
	var body struct {
		SubagentResults []struct {
			ProviderName string `json:"provider_name"`
			Summary      string `json:"summary"`
		} `json:"subagent_results"`
	}
	if err := json.Unmarshal(out.Bytes(), &body); err != nil {
		t.Fatalf("output must be JSON: %v\n%s", err, out.String())
	}
	if len(body.SubagentResults) == 0 {
		t.Fatal("expected subagent results")
	}
	for _, result := range body.SubagentResults {
		if result.ProviderName != "premium" || !strings.Contains(result.Summary, "premium-route") {
			t.Fatalf("result = %#v, want health-routed premium fallback", result)
		}
	}
}

func Test_Run_keeps_route_when_fallback_provider_is_also_avoided(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	configJSON := `{"providers":{"cheap":{"model_command":["sh","-c","cat >/dev/null; printf cheap-route"]},"premium":{"model_command":["sh","-c","cat >/dev/null; printf premium-route"]}},"provider_policy":{"default_provider":"cheap","fallback_provider":"premium"}}`
	if err := os.WriteFile(filepath.Join(root, ".ceo-harness.json"), []byte(configJSON), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	store, err := history.New(root)
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	_, err = store.Append(context.Background(), history.Entry{
		Task:    "bad routes",
		Verdict: "fail",
		ProviderHealth: []history.ProviderHealth{
			{ProviderName: "cheap", ModelSource: "command", AttemptCount: 1, FailCount: 1, ErrorCount: 1},
			{ProviderName: "premium", ModelSource: "command", AttemptCount: 1, FailCount: 1, ErrorCount: 1},
		},
	})
	if err != nil {
		t.Fatalf("Append returned error: %v", err)
	}

	// When
	err = Run(context.Background(), &out, []string{"--workspace", root, "Plan", "roadmap"})
	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v\n%s", err, out.String())
	}
	var body struct {
		SubagentResults []struct {
			ProviderName string `json:"provider_name"`
			Summary      string `json:"summary"`
		} `json:"subagent_results"`
	}
	if err := json.Unmarshal(out.Bytes(), &body); err != nil {
		t.Fatalf("output must be JSON: %v\n%s", err, out.String())
	}
	for _, result := range body.SubagentResults {
		if result.ProviderName != "cheap" || !strings.Contains(result.Summary, "cheap-route") {
			t.Fatalf("result = %#v, want original cheap route when fallback is avoided", result)
		}
	}
}

func Test_Run_reports_provider_health_route_avoidance_in_run_manifest(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	configJSON := `{"providers":{"cheap":{"model_command":["sh","-c","cat >/dev/null; printf cheap-route"]},"premium":{"model_command":["sh","-c","cat >/dev/null; printf premium-route"]}},"provider_policy":{"default_provider":"cheap","fallback_provider":"premium"}}`
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
	err = Run(context.Background(), &out, []string{"--workspace", root, "Plan", "roadmap"})
	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v\n%s", err, out.String())
	}
	var body struct {
		RunManifest struct {
			ProviderHealthAvoidedRouteCount int      `json:"provider_health_avoided_route_count"`
			ProviderHealthAvoidedProviders  []string `json:"provider_health_avoided_providers"`
		} `json:"run_manifest"`
		RunEvents []struct {
			Kind          string   `json:"kind"`
			Status        string   `json:"status"`
			RouteCount    int      `json:"route_count"`
			ProviderNames []string `json:"provider_names"`
		} `json:"run_events"`
		SubagentResults []struct {
			ProviderName string `json:"provider_name"`
		} `json:"subagent_results"`
		ProviderRouteDecisions []struct {
			ProviderName string `json:"provider_name"`
			Reason       string `json:"reason"`
			FallbackFrom string `json:"fallback_from"`
		} `json:"provider_route_decisions"`
	}
	if err := json.Unmarshal(out.Bytes(), &body); err != nil {
		t.Fatalf("output must be JSON: %v\n%s", err, out.String())
	}
	if body.RunManifest.ProviderHealthAvoidedRouteCount != len(body.SubagentResults) {
		t.Fatalf("ProviderHealthAvoidedRouteCount = %d, want %d", body.RunManifest.ProviderHealthAvoidedRouteCount, len(body.SubagentResults))
	}
	if len(body.RunManifest.ProviderHealthAvoidedProviders) != 1 || body.RunManifest.ProviderHealthAvoidedProviders[0] != "cheap" {
		t.Fatalf("ProviderHealthAvoidedProviders = %#v, want [cheap]", body.RunManifest.ProviderHealthAvoidedProviders)
	}
	foundRouteEvent := false
	for _, event := range body.RunEvents {
		if event.Kind != "provider_health_route" {
			continue
		}
		foundRouteEvent = true
		if event.Status != "rerouted" || event.RouteCount != len(body.SubagentResults) {
			t.Fatalf("route event = %#v, want rerouted count %d", event, len(body.SubagentResults))
		}
		if len(event.ProviderNames) != 1 || event.ProviderNames[0] != "cheap" {
			t.Fatalf("route event providers = %#v, want [cheap]", event.ProviderNames)
		}
	}
	if !foundRouteEvent {
		t.Fatalf("run events = %#v, want provider_health_route event", body.RunEvents)
	}
	for _, decision := range body.ProviderRouteDecisions {
		if decision.ProviderName != "premium" || decision.Reason != "provider_health.fallback" || decision.FallbackFrom != "cheap" {
			t.Fatalf("route decision = %#v, want premium health fallback from cheap", decision)
		}
	}
}
