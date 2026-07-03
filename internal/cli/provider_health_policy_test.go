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

func Test_Run_provider_health_rollup_uses_workspace_policy_thresholds(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	configJSON := `{"provider_health_avoid_failure_rate":0.9,"provider_health_watch_failure_rate":0.5}`
	if err := os.WriteFile(filepath.Join(root, ".ceo-harness.json"), []byte(configJSON), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	store, err := history.New(root)
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	entry := history.Entry{
		Task: "policy",
		ProviderHealth: []history.ProviderHealth{
			{ProviderName: "zeta", ModelSource: "command", AttemptCount: 2, PassCount: 1, FailCount: 1},
		},
	}
	if _, err := store.Append(context.Background(), entry); err != nil {
		t.Fatalf("Append returned error: %v", err)
	}

	// When
	err = Run(context.Background(), &out, []string{"--workspace", root, "--provider-health"})
	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v\n%s", err, out.String())
	}
	var body struct {
		ProviderHealth []struct {
			ProviderName   string `json:"provider_name"`
			Recommendation string `json:"recommendation"`
		} `json:"provider_health"`
		Summary struct {
			AvoidCount int `json:"avoid_count"`
			WatchCount int `json:"watch_count"`
		} `json:"summary"`
	}
	if err := json.Unmarshal(out.Bytes(), &body); err != nil {
		t.Fatalf("output must be JSON: %v\n%s", err, out.String())
	}
	if len(body.ProviderHealth) != 1 {
		t.Fatalf("provider health length = %d, want 1: %#v", len(body.ProviderHealth), body.ProviderHealth)
	}
	if body.ProviderHealth[0].ProviderName != "zeta" || body.ProviderHealth[0].Recommendation != "watch" {
		t.Fatalf("provider health = %#v, want zeta watch", body.ProviderHealth[0])
	}
	if body.Summary.AvoidCount != 0 || body.Summary.WatchCount != 1 {
		t.Fatalf("summary = %#v, want 0 avoid and 1 watch", body.Summary)
	}
}

func Test_Run_provider_health_rollup_uses_workspace_cost_policy_threshold(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	configJSON := `{"provider_health_watch_cost_per_attempt_microusd":50}`
	if err := os.WriteFile(filepath.Join(root, ".ceo-harness.json"), []byte(configJSON), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	store, err := history.New(root)
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	entry := history.Entry{
		Task: "cost policy",
		ProviderHealth: []history.ProviderHealth{
			{ProviderName: "premium", ModelSource: "command", AttemptCount: 2, PassCount: 2, EstimatedCostMicroUSD: 200},
		},
	}
	if _, err := store.Append(context.Background(), entry); err != nil {
		t.Fatalf("Append returned error: %v", err)
	}

	// When
	err = Run(context.Background(), &out, []string{"--workspace", root, "--provider-health"})
	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v\n%s", err, out.String())
	}
	var body struct {
		ProviderHealth []struct {
			ProviderName   string `json:"provider_name"`
			Recommendation string `json:"recommendation"`
		} `json:"provider_health"`
	}
	if err := json.Unmarshal(out.Bytes(), &body); err != nil {
		t.Fatalf("output must be JSON: %v\n%s", err, out.String())
	}
	if len(body.ProviderHealth) != 1 {
		t.Fatalf("provider health length = %d, want 1: %#v", len(body.ProviderHealth), body.ProviderHealth)
	}
	if body.ProviderHealth[0].ProviderName != "premium" || body.ProviderHealth[0].Recommendation != "watch" {
		t.Fatalf("provider health = %#v, want premium watch", body.ProviderHealth[0])
	}
}
