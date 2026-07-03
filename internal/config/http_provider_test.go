package config

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func Test_LoadWorkspace_resolves_agent_http_provider_when_provider_exists(t *testing.T) {
	// Given
	root := t.TempDir()
	content := `{"providers":{"fast":{"http":{"url":"http://127.0.0.1:8080/v1/chat/completions","model":"fast-model","api_key_env":"CEO_FAST_KEY","input_cost_per_million_tokens":2,"output_cost_per_million_tokens":8,"timeout_ms":2500,"max_output_tokens":64,"response_format":"json_object"}}},"agent_providers":{"scanner":"fast"}}`
	if err := os.WriteFile(filepath.Join(root, ".ceo-harness.json"), []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	// When
	cfg, err := LoadWorkspace(context.Background(), root)

	// Then
	if err != nil {
		t.Fatalf("LoadWorkspace returned error: %v", err)
	}
	providers := cfg.AgentHTTPProviders()
	got := providers["scanner"]
	if got.URL != "http://127.0.0.1:8080/v1/chat/completions" || got.Model != "fast-model" || got.APIKeyEnv != "CEO_FAST_KEY" || got.InputCostPerMillionTokens != 2 || got.OutputCostPerMillionTokens != 8 || got.TimeoutMS != 2500 || got.MaxOutputTokens != 64 || got.ResponseFormat != "json_object" {
		t.Fatalf("scanner http provider = %#v, want configured provider", got)
	}
}

func Test_LoadWorkspace_rejects_secret_like_http_provider_api_key_env(t *testing.T) {
	// Given
	root := t.TempDir()
	content := `{"providers":{"fast":{"http":{"url":"http://127.0.0.1:8080/v1/chat/completions","model":"fast-model","api_key_env":"sk-secret-as-env-name"}}}}`
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

func Test_LoadWorkspace_rejects_negative_http_provider_cost(t *testing.T) {
	// Given
	root := t.TempDir()
	content := `{"providers":{"fast":{"http":{"url":"http://127.0.0.1:8080/v1/chat/completions","model":"fast-model","input_cost_per_million_tokens":-1}}}}`
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

func Test_LoadWorkspace_rejects_negative_http_provider_timeout(t *testing.T) {
	// Given
	root := t.TempDir()
	content := `{"providers":{"fast":{"http":{"url":"http://127.0.0.1:8080/v1/chat/completions","model":"fast-model","timeout_ms":-1}}}}`
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

func Test_LoadWorkspace_rejects_negative_http_provider_max_output_tokens(t *testing.T) {
	// Given
	root := t.TempDir()
	content := `{"providers":{"fast":{"http":{"url":"http://127.0.0.1:8080/v1/chat/completions","model":"fast-model","max_output_tokens":-1}}}}`
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

func Test_LoadWorkspace_rejects_invalid_http_provider_response_format(t *testing.T) {
	// Given
	root := t.TempDir()
	content := `{"providers":{"fast":{"http":{"url":"http://127.0.0.1:8080/v1/chat/completions","model":"fast-model","response_format":"xml"}}}}`
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

func Test_LoadWorkspace_rejects_provider_with_multiple_backends(t *testing.T) {
	// Given
	root := t.TempDir()
	content := `{"providers":{"fast":{"model_command":["echo","ok"],"http":{"url":"http://127.0.0.1:8080/v1/chat/completions","model":"fast-model"}}}}`
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
