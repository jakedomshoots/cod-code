package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"ceoharness/internal/config"
)

func Test_Run_writes_provider_health_policy_when_init_config_policy_flags_are_supplied(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	args := []string{
		"--workspace",
		root,
		"--init-config",
		"--provider-health-avoid-failure-rate",
		"0.9",
		"--provider-health-watch-failure-rate",
		"0.5",
		"--provider-health-watch-cost-per-attempt-microusd",
		"50",
	}

	// When
	err := Run(context.Background(), &out, args)
	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	cfg, loadErr := config.LoadWorkspace(context.Background(), root)
	if loadErr != nil {
		t.Fatalf("LoadWorkspace returned error: %v", loadErr)
	}
	if cfg.ProviderHealthAvoidFailureRate != 0.9 ||
		cfg.ProviderHealthWatchFailureRate != 0.5 ||
		cfg.ProviderHealthWatchCostPerAttemptMicroUSD != 50 {
		t.Fatalf("provider health policy = %#v, want init-config thresholds", cfg)
	}
	var body struct {
		ProviderHealthAvoidFailureRate            float64 `json:"provider_health_avoid_failure_rate"`
		ProviderHealthWatchFailureRate            float64 `json:"provider_health_watch_failure_rate"`
		ProviderHealthWatchCostPerAttemptMicroUSD int64   `json:"provider_health_watch_cost_per_attempt_microusd"`
	}
	if jsonErr := json.Unmarshal(out.Bytes(), &body); jsonErr != nil {
		t.Fatalf("output must be JSON: %v\n%s", jsonErr, out.String())
	}
	if body.ProviderHealthAvoidFailureRate != 0.9 ||
		body.ProviderHealthWatchFailureRate != 0.5 ||
		body.ProviderHealthWatchCostPerAttemptMicroUSD != 50 {
		t.Fatalf("init report provider health policy = %#v, want thresholds", body)
	}
}

func Test_Run_rejects_invalid_provider_health_policy_when_init_config_policy_flags_are_supplied(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	args := []string{
		"--workspace",
		root,
		"--init-config",
		"--provider-health-avoid-failure-rate",
		"0.4",
		"--provider-health-watch-failure-rate",
		"0.8",
	}

	// When
	err := Run(context.Background(), &out, args)

	// Then
	if err == nil || !errors.Is(err, config.ErrInvalidConfig) {
		t.Fatalf("error = %v, want invalid config", err)
	}
	if _, statErr := os.Stat(filepath.Join(root, ".ceo-harness.json")); !os.IsNotExist(statErr) {
		t.Fatalf("workspace config stat error = %v, want file not created", statErr)
	}
}
