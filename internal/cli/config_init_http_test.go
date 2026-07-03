package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"ceoharness/internal/config"
)

func Test_Run_writes_http_provider_when_init_config_http_provider_flags_are_supplied(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	args := []string{
		"--workspace",
		root,
		"--init-config",
		"--http-provider",
		"fast",
		"--http-url",
		"http://127.0.0.1:8080/v1/chat/completions",
		"--http-model",
		"fast-model",
		"--http-api-key-env",
		"CEO_FAST_KEY",
		"--http-agent",
		"scanner",
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
	provider := cfg.Providers["fast"].HTTP
	if provider.URL != "http://127.0.0.1:8080/v1/chat/completions" || provider.Model != "fast-model" || provider.APIKeyEnv != "CEO_FAST_KEY" {
		t.Fatalf("http provider = %#v, want init-config provider", provider)
	}
	if cfg.AgentProviders["scanner"] != "fast" {
		t.Fatalf("agent provider = %#v, want scanner -> fast", cfg.AgentProviders)
	}
	var body struct {
		HTTPProviderCount  int `json:"http_provider_count"`
		AgentProviderCount int `json:"agent_provider_count"`
	}
	if jsonErr := json.Unmarshal(out.Bytes(), &body); jsonErr != nil {
		t.Fatalf("output must be JSON: %v\n%s", jsonErr, out.String())
	}
	if body.HTTPProviderCount != 1 || body.AgentProviderCount != 1 {
		t.Fatalf("init report provider counts = %#v, want one provider and one assignment", body)
	}
}

func Test_Run_writes_http_provider_controls_when_init_config_http_control_flags_are_supplied(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	args := []string{
		"--workspace",
		root,
		"--init-config",
		"--http-provider",
		"fast",
		"--http-url",
		"http://127.0.0.1:8080/v1/chat/completions",
		"--http-model",
		"fast-model",
		"--http-agent",
		"scanner",
		"--http-timeout-ms",
		"2500",
		"--http-max-output-tokens",
		"64",
		"--http-response-format",
		"json_object",
		"--http-input-cost-per-million",
		"2.5",
		"--http-output-cost-per-million",
		"8.5",
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
	provider := cfg.Providers["fast"].HTTP
	if provider.TimeoutMS != 2500 || provider.MaxOutputTokens != 64 || provider.ResponseFormat != "json_object" {
		t.Fatalf("http provider controls = %#v, want timeout/output/response format", provider)
	}
	if provider.InputCostPerMillionTokens != 2.5 || provider.OutputCostPerMillionTokens != 8.5 {
		t.Fatalf("http provider costs = %#v, want configured prices", provider)
	}
}

func Test_Run_rejects_partial_init_config_http_provider_flags_without_writing_config(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	args := []string{
		"--workspace",
		root,
		"--init-config",
		"--http-provider",
		"fast",
		"--http-url",
		"http://127.0.0.1:8080/v1/chat/completions",
	}

	// When
	err := Run(context.Background(), &out, args)

	// Then
	if err == nil || !strings.Contains(err.Error(), "required together") {
		t.Fatalf("error = %v, want required-together validation", err)
	}
	if _, statErr := os.Stat(filepath.Join(root, ".ceo-harness.json")); !os.IsNotExist(statErr) {
		t.Fatalf("workspace config stat error = %v, want file not created", statErr)
	}
}

func Test_Run_rejects_negative_init_config_http_control_flag(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	args := []string{
		"--workspace",
		root,
		"--init-config",
		"--http-provider",
		"fast",
		"--http-url",
		"http://127.0.0.1:8080/v1/chat/completions",
		"--http-model",
		"fast-model",
		"--http-agent",
		"scanner",
		"--http-timeout-ms",
		"-1",
	}

	// When
	err := Run(context.Background(), &out, args)

	// Then
	if err == nil || !strings.Contains(err.Error(), "non-negative integer") {
		t.Fatalf("error = %v, want non-negative timeout validation", err)
	}
}
