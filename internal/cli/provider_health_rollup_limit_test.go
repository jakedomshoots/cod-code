package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"ceoharness/internal/history"
)

func Test_Run_limits_provider_health_rollup_when_top_providers_flag_is_supplied(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	store, err := history.New(root)
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	entries := []history.Entry{
		{
			Task: "first",
			ProviderHealth: []history.ProviderHealth{
				{ProviderName: "zeta", ModelSource: "http", AttemptCount: 1, FailCount: 1, ErrorCount: 1},
				{ProviderName: "alpha", ModelSource: "http", AttemptCount: 3, PassCount: 2, FailCount: 1, ErrorCount: 1},
				{ProviderName: "beta", ModelSource: "http", AttemptCount: 1, PassCount: 1},
			},
		},
	}
	for _, entry := range entries {
		if _, err := store.Append(context.Background(), entry); err != nil {
			t.Fatalf("Append returned error: %v", err)
		}
	}

	// When
	err = Run(context.Background(), &out, []string{"--workspace", root, "--provider-health", "--top-providers", "1"})
	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v\n%s", err, out.String())
	}
	var body struct {
		TopProviders   int `json:"top_providers,omitempty"`
		ProviderHealth []struct {
			ProviderName   string `json:"provider_name"`
			Recommendation string `json:"recommendation"`
		} `json:"provider_health"`
		Summary struct {
			AvoidCount   int `json:"avoid_count"`
			WatchCount   int `json:"watch_count"`
			HealthyCount int `json:"healthy_count"`
		} `json:"summary"`
	}
	if err := json.Unmarshal(out.Bytes(), &body); err != nil {
		t.Fatalf("output must be JSON: %v\n%s", err, out.String())
	}
	if body.TopProviders != 1 {
		t.Fatalf("TopProviders = %d, want 1", body.TopProviders)
	}
	if len(body.ProviderHealth) != 1 {
		t.Fatalf("provider health length = %d, want 1: %#v", len(body.ProviderHealth), body.ProviderHealth)
	}
	if body.ProviderHealth[0].ProviderName != "zeta" || body.ProviderHealth[0].Recommendation != "avoid" {
		t.Fatalf("provider health = %#v, want zeta avoid", body.ProviderHealth[0])
	}
	if body.Summary.AvoidCount != 1 || body.Summary.WatchCount != 0 || body.Summary.HealthyCount != 0 {
		t.Fatalf("summary = %#v, want only one avoid row counted", body.Summary)
	}
}
