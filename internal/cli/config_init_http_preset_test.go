package cli

import (
	"bytes"
	"context"
	"testing"

	"ceoharness/internal/config"
)

func Test_Run_writes_http_provider_from_preset_when_init_config_http_preset_is_supplied(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	args := []string{
		"--workspace",
		root,
		"--init-config",
		"--http-provider",
		"fast",
		"--http-preset",
		"openrouter",
		"--http-model",
		"~openai/gpt-latest",
		"--http-agent",
		"planner",
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
	if provider.URL != "https://openrouter.ai/api/v1/chat/completions" || provider.APIKeyEnv != "OPENROUTER_API_KEY" {
		t.Fatalf("preset http provider = %#v, want OpenRouter URL and env", provider)
	}
	if provider.Model != "~openai/gpt-latest" {
		t.Fatalf("preset model = %q, want requested model", provider.Model)
	}
	if cfg.AgentProviders["planner"] != "fast" {
		t.Fatalf("agent provider = %#v, want planner -> fast", cfg.AgentProviders)
	}
}

func Test_Run_writes_default_provider_policy_when_init_config_default_provider_is_supplied(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	args := []string{
		"--workspace",
		root,
		"--init-config",
		"--http-provider",
		"fast",
		"--http-preset",
		"openrouter",
		"--http-model",
		"~openai/gpt-latest",
		"--default-provider",
		"fast",
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
	if cfg.ProviderPolicy.DefaultProvider != "fast" {
		t.Fatalf("default provider = %q, want fast", cfg.ProviderPolicy.DefaultProvider)
	}
	if len(cfg.AgentProviders) != 0 {
		t.Fatalf("agent providers = %#v, want policy routing without fixed agent route", cfg.AgentProviders)
	}
}

func Test_Run_writes_multiple_http_providers_when_init_config_http_provider_repeats(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	args := []string{
		"--workspace",
		root,
		"--init-config",
		"--http-provider",
		"cheap",
		"--http-preset",
		"openrouter",
		"--http-model",
		"~openai/gpt-latest",
		"--default-provider",
		"cheap",
		"--http-provider",
		"premium",
		"--http-preset",
		"openai",
		"--http-model",
		"gpt-5.5",
		"--fallback-provider",
		"premium",
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
	if cfg.ProviderPolicy.DefaultProvider != "cheap" || cfg.ProviderPolicy.FallbackProvider != "premium" {
		t.Fatalf("provider policy = %#v, want cheap default and premium fallback", cfg.ProviderPolicy)
	}
	if cfg.Providers["cheap"].HTTP.APIKeyEnv != "OPENROUTER_API_KEY" {
		t.Fatalf("cheap provider = %#v, want OpenRouter preset", cfg.Providers["cheap"])
	}
	if cfg.Providers["premium"].HTTP.APIKeyEnv != "OPENAI_API_KEY" {
		t.Fatalf("premium provider = %#v, want OpenAI preset", cfg.Providers["premium"])
	}
}

func Test_Run_writes_risk_area_provider_policy_when_init_config_risk_area_provider_is_supplied(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	args := []string{
		"--workspace",
		root,
		"--init-config",
		"--http-provider",
		"cheap",
		"--http-preset",
		"openrouter",
		"--http-model",
		"~openai/gpt-latest",
		"--default-provider",
		"cheap",
		"--http-provider",
		"premium",
		"--http-preset",
		"openai",
		"--http-model",
		"gpt-5.5",
		"--risk-area-provider",
		"database=premium",
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
	if cfg.ProviderPolicy.RiskAreaProviders["database"] != "premium" {
		t.Fatalf("risk area providers = %#v, want database premium", cfg.ProviderPolicy.RiskAreaProviders)
	}
}
