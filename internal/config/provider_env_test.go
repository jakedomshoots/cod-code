package config

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func Test_LoadWorkspace_reads_provider_env_vars_when_config_exists(t *testing.T) {
	// Given
	root := t.TempDir()
	content := `{"providers":{"fast":{"model_command":["echo","ok"],"env_vars":["CEO_PROVIDER_TOKEN"]}},"agent_providers":{"scanner":"fast"}}`
	if err := os.WriteFile(filepath.Join(root, ".ceo-harness.json"), []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	// When
	cfg, err := LoadWorkspace(context.Background(), root)
	// Then
	if err != nil {
		t.Fatalf("LoadWorkspace returned error: %v", err)
	}
	envVars := cfg.AgentEnvVars()
	if len(envVars["scanner"]) != 1 || envVars["scanner"][0] != "CEO_PROVIDER_TOKEN" {
		t.Fatalf("scanner env vars = %#v, want provider token binding", envVars["scanner"])
	}
}

func Test_LoadWorkspace_rejects_empty_provider_env_var(t *testing.T) {
	// Given
	root := t.TempDir()
	content := `{"providers":{"fast":{"model_command":["echo","ok"],"env_vars":[""]}}}`
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

func Test_LoadWorkspace_rejects_secret_like_provider_env_var(t *testing.T) {
	// Given
	root := t.TempDir()
	content := `{"providers":{"fast":{"model_command":["echo","ok"],"env_vars":["sk-secret-as-env-name"]}}}`
	if err := os.WriteFile(filepath.Join(root, ".ceo-harness.json"), []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	// When
	_, err := LoadWorkspace(context.Background(), root)

	// Then
	if !errors.Is(err, ErrInvalidConfig) {
		t.Fatalf("error = %v, want ErrInvalidConfig", err)
	}
	if strings.Contains(err.Error(), "sk-secret-as-env-name") {
		t.Fatalf("error leaked secret-like env name: %v", err)
	}
}
