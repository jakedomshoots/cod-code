package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func Test_Run_config_check_reports_provider_env_counts_without_secret_values(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	content := `{"providers":{"fast":{"model_command":["echo","ok"],"env_vars":["CEO_PROVIDER_TOKEN"]}},"agent_providers":{"scanner":"fast"}}`
	if err := os.WriteFile(filepath.Join(root, ".ceo-harness.json"), []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	t.Setenv("CEO_PROVIDER_TOKEN", "secret-value")

	// When
	err := Run(context.Background(), &out, []string{"--workspace", root, "--config-check"})
	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	var body struct {
		ProviderEnvVarCount        int `json:"provider_env_var_count"`
		ProviderEnvVarPresentCount int `json:"provider_env_var_present_count"`
		ProviderEnvVarMissingCount int `json:"provider_env_var_missing_count"`
	}
	if jsonErr := json.Unmarshal(out.Bytes(), &body); jsonErr != nil {
		t.Fatalf("output must be JSON: %v\n%s", jsonErr, out.String())
	}
	if body.ProviderEnvVarCount != 1 || body.ProviderEnvVarPresentCount != 1 || body.ProviderEnvVarMissingCount != 0 {
		t.Fatalf("provider env counts = %#v, want one present env binding", body)
	}
	if strings.Contains(out.String(), "secret-value") {
		t.Fatalf("config-check leaked secret value: %s", out.String())
	}
}

func Test_Run_config_check_reports_missing_provider_env_names_without_secret_values(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	content := `{"providers":{"fast":{"model_command":["echo","ok"],"env_vars":["CEO_PROVIDER_TOKEN","CEO_MISSING_TOKEN"]}},"agent_providers":{"scanner":"fast"}}`
	if err := os.WriteFile(filepath.Join(root, ".ceo-harness.json"), []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	t.Setenv("CEO_PROVIDER_TOKEN", "secret-value")
	t.Setenv("CEO_MISSING_TOKEN", "")

	// When
	err := Run(context.Background(), &out, []string{"--workspace", root, "--config-check"})
	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	var body struct {
		ProviderEnvVarMissingCount int      `json:"provider_env_var_missing_count"`
		ProviderEnvVarMissingNames []string `json:"provider_env_var_missing_names"`
	}
	if jsonErr := json.Unmarshal(out.Bytes(), &body); jsonErr != nil {
		t.Fatalf("output must be JSON: %v\n%s", jsonErr, out.String())
	}
	if body.ProviderEnvVarMissingCount != 1 {
		t.Fatalf("ProviderEnvVarMissingCount = %d, want 1", body.ProviderEnvVarMissingCount)
	}
	if len(body.ProviderEnvVarMissingNames) != 1 || body.ProviderEnvVarMissingNames[0] != "CEO_MISSING_TOKEN" {
		t.Fatalf("ProviderEnvVarMissingNames = %#v, want missing env var name", body.ProviderEnvVarMissingNames)
	}
	if strings.Contains(out.String(), "secret-value") {
		t.Fatalf("config-check leaked secret value: %s", out.String())
	}
}

func Test_Run_returns_error_when_provider_env_var_is_missing(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	content := `{"providers":{"fast":{"model_command":["echo","ok"],"env_vars":["CEO_PROVIDER_TOKEN"]}},"agent_providers":{"scanner":"fast"}}`
	if err := os.WriteFile(filepath.Join(root, ".ceo-harness.json"), []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	t.Setenv("CEO_PROVIDER_TOKEN", "")

	// When
	err := Run(context.Background(), &out, []string{"--workspace", root, "Fix", "a", "failing", "test"})

	// Then
	if err == nil {
		t.Fatal("expected missing provider env var error")
	}
	if errors.Is(err, ErrVerdictFailed) {
		t.Fatalf("error = %v, want config/runtime setup error", err)
	}
	if !strings.Contains(err.Error(), "CEO_PROVIDER_TOKEN") {
		t.Fatalf("error = %v, want missing env var name", err)
	}
}
