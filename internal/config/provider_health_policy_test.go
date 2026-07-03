package config

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func Test_LoadWorkspace_reads_provider_health_policy_when_config_exists(t *testing.T) {
	// Given
	root := t.TempDir()
	content := `{"provider_health_avoid_failure_rate":0.9,"provider_health_watch_failure_rate":0.5,"provider_health_watch_cost_per_attempt_microusd":50}`
	if err := os.WriteFile(filepath.Join(root, ".ceo-harness.json"), []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	// When
	cfg, err := LoadWorkspace(context.Background(), root)

	// Then
	if err != nil {
		t.Fatalf("LoadWorkspace returned error: %v", err)
	}
	if cfg.ProviderHealthAvoidFailureRate != 0.9 {
		t.Fatalf("ProviderHealthAvoidFailureRate = %f, want 0.9", cfg.ProviderHealthAvoidFailureRate)
	}
	if cfg.ProviderHealthWatchFailureRate != 0.5 {
		t.Fatalf("ProviderHealthWatchFailureRate = %f, want 0.5", cfg.ProviderHealthWatchFailureRate)
	}
	if cfg.ProviderHealthWatchCostPerAttemptMicroUSD != 50 {
		t.Fatalf("ProviderHealthWatchCostPerAttemptMicroUSD = %d, want 50", cfg.ProviderHealthWatchCostPerAttemptMicroUSD)
	}
}

func Test_LoadWorkspace_rejects_invalid_provider_health_policy(t *testing.T) {
	// Given
	root := t.TempDir()
	content := `{"provider_health_avoid_failure_rate":0.25,"provider_health_watch_failure_rate":0.5}`
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

func Test_LoadWorkspace_rejects_negative_provider_health_cost_policy(t *testing.T) {
	// Given
	root := t.TempDir()
	content := `{"provider_health_watch_cost_per_attempt_microusd":-1}`
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
