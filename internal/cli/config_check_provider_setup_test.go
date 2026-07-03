package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func Test_Run_config_check_reports_provider_setup_steps(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	content := `{"providers":{"fast":{"http":{"url":"http://127.0.0.1:8080/v1/chat/completions","model":"fast-model","api_key_env":"CEO_MISSING_TOKEN"}}},"agent_providers":{"scanner":"fast"}}`
	if err := os.WriteFile(filepath.Join(root, ".ceo-harness.json"), []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	t.Setenv("CEO_MISSING_TOKEN", "")

	// When
	err := Run(context.Background(), &out, []string{"--workspace", root, "--config-check"})

	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	var body struct {
		ProviderSetupSteps []string `json:"provider_setup_steps"`
	}
	if jsonErr := json.Unmarshal(out.Bytes(), &body); jsonErr != nil {
		t.Fatalf("output must be JSON: %v\n%s", jsonErr, out.String())
	}
	want := []string{
		"export CEO_MISSING_TOKEN=...",
		`ceo-packet --workspace "` + root + `" --doctor-provider "fast" --format text`,
		`ceo-packet --workspace "` + root + `" --plan-only "Smoke provider routing"`,
	}
	for _, step := range want {
		if !containsString(body.ProviderSetupSteps, step) {
			t.Fatalf("provider_setup_steps = %#v, want %q", body.ProviderSetupSteps, step)
		}
	}
}

func Test_Run_config_check_prints_provider_setup_text(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	content := `{"providers":{"fast":{"model_command":["echo","ok"],"env_vars":["CEO_PROVIDER_TOKEN"]}},"agent_providers":{"scanner":"fast"}}`
	if err := os.WriteFile(filepath.Join(root, ".ceo-harness.json"), []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	t.Setenv("CEO_PROVIDER_TOKEN", "")

	// When
	err := Run(context.Background(), &out, []string{"--workspace", root, "--config-check", "--format", "text"})

	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v\n%s", err, out.String())
	}
	text := out.String()
	for _, want := range []string{
		"Config: " + filepath.Join(root, ".ceo-harness.json"),
		"Providers: 1",
		"Provider env: 0/1 set",
		"export CEO_PROVIDER_TOKEN=...",
		`--doctor-provider "fast" --format text`,
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("config-check text missing %q:\n%s", want, text)
		}
	}
	if strings.Contains(text, "secret") {
		t.Fatalf("config-check text leaked secret-like value:\n%s", text)
	}
}

func Test_Run_config_check_rejects_secret_like_provider_env_var_without_printing_setup_text(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	secretLikeName := "sk-secret-as-env-name"
	content := `{"providers":{"fast":{"model_command":["echo","ok"],"env_vars":["` + secretLikeName + `"]}},"agent_providers":{"scanner":"fast"}}`
	if err := os.WriteFile(filepath.Join(root, ".ceo-harness.json"), []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	// When
	err := Run(context.Background(), &out, []string{"--workspace", root, "--config-check", "--format", "text"})

	// Then
	if err == nil {
		t.Fatal("expected invalid config error")
	}
	if strings.Contains(out.String(), secretLikeName) || strings.Contains(err.Error(), secretLikeName) {
		t.Fatalf("config-check leaked secret-like env name; output=%q error=%v", out.String(), err)
	}
}

func Test_Run_config_check_rejects_events_format(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()

	// When
	err := Run(context.Background(), &out, []string{"--workspace", root, "--config-check", "--format", "events"})

	// Then
	if err == nil {
		t.Fatal("expected events format error")
	}
	if !strings.Contains(err.Error(), "only available for run reports") {
		t.Fatalf("error = %q, want run reports guidance", err.Error())
	}
}
