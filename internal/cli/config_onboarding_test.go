package cli

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func Test_Run_config_completions_writes_shell_files_when_requested(t *testing.T) {
	tests := []struct {
		name     string
		shell    string
		fileName string
		want     string
	}{
		{name: "zsh", shell: "zsh", fileName: "_ceo-packet", want: "#compdef ceo-packet"},
		{name: "bash", shell: "bash", fileName: "ceo-packet.bash", want: "complete -F _ceo_packet ceo-packet"},
		{name: "fish", shell: "fish", fileName: "ceo-packet.fish", want: "complete -c ceo-packet"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			var out bytes.Buffer
			outputDir := t.TempDir()

			// When
			err := Run(context.Background(), &out, []string{"config", "completions", "--shell", tt.shell, "--output-dir", outputDir})
			// Then
			if err != nil {
				t.Fatalf("Run returned error: %v\n%s", err, out.String())
			}
			path := filepath.Join(outputDir, tt.fileName)
			content, readErr := os.ReadFile(path)
			if readErr != nil {
				t.Fatalf("read completion file: %v", readErr)
			}
			if !strings.Contains(string(content), tt.want) || !strings.Contains(string(content), "ceo-packet") {
				t.Fatalf("%s completion content missing expected command:\n%s", tt.shell, string(content))
			}
			for _, command := range []string{"start", "run", "gauntlet", "doctor", "inbox", "status", "resume", "retry", "rollback", "explain-failure"} {
				if !strings.Contains(string(content), command) {
					t.Fatalf("%s completion content missing primary command %q:\n%s", tt.shell, command, string(content))
				}
			}
			for _, want := range []string{"production-actions", "action-state", "ready missing_env empty_env setup_blocked waiting", "release_proof provider_proof competitor_setup comparison final_readiness", "openai openrouter kimi-code moonshot minimax kimi codex"} {
				if !strings.Contains(string(content), want) {
					t.Fatalf("%s completion content missing production action completion %q:\n%s", tt.shell, want, string(content))
				}
			}
			if !strings.Contains(out.String(), path) {
				t.Fatalf("output = %q, want generated path %q", out.String(), path)
			}
		})
	}
}

func Test_Run_config_completions_rejects_unknown_shell(t *testing.T) {
	// Given
	var out bytes.Buffer
	outputDir := t.TempDir()

	// When
	err := Run(context.Background(), &out, []string{"config", "completions", "--shell", "powershell", "--output-dir", outputDir})

	// Then
	if err == nil {
		t.Fatal("expected unknown shell error")
	}
	if !strings.Contains(err.Error(), "zsh, bash, or fish") {
		t.Fatalf("error = %q, want shell guidance", err.Error())
	}
}

func Test_Run_config_doctor_prints_compact_health_when_alias_is_supplied(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	content := `{"providers":{"fast":{"model_command":["echo","ok"],"env_vars":["OPENAI_API_KEY"]}},"agent_providers":{"scanner":"fast"},"check_command":["go","test","./..."],"require_checks":true}`
	if err := os.WriteFile(filepath.Join(root, ".ceo-harness.json"), []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	t.Setenv("OPENAI_API_KEY", "")

	// When
	err := Run(context.Background(), &out, []string{"config", "doctor", "--workspace", root, "--format", "text"})
	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v\n%s", err, out.String())
	}
	text := out.String()
	for _, want := range []string{
		"Config doctor: needs setup",
		"Providers: 1",
		"Checks: configured=true required=true",
		"export OPENAI_API_KEY=...",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("config doctor text missing %q:\n%s", want, text)
		}
	}
}

func Test_Run_config_explain_prints_compact_first_run_checklist(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()

	// When
	err := Run(context.Background(), &out, []string{"config", "explain", "--workspace", root, "--format", "text"})
	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v\n%s", err, out.String())
	}
	text := out.String()
	for _, want := range []string{
		"first-run checklist",
		"provider wizard",
		"dogfood",
		"write policy",
		"first task",
	} {
		if !strings.Contains(strings.ToLower(text), want) {
			t.Fatalf("config explain text missing %q:\n%s", want, text)
		}
	}
	if lines := strings.Count(strings.TrimSpace(text), "\n") + 1; lines > 8 {
		t.Fatalf("config explain lines = %d, want compact checklist:\n%s", lines, text)
	}
}

func Test_Run_config_doctor_redacts_secret_values_when_api_key_is_missing(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	secret := "sk-task9-secret-value"
	content := `{"providers":{"main":{"http":{"url":"http://127.0.0.1:8080/v1/chat/completions","model":"gpt-5","api_key_env":"OPENAI_API_KEY"}}},"agent_providers":{"scanner":"main"}}`
	if err := os.WriteFile(filepath.Join(root, ".ceo-harness.json"), []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	t.Setenv("OPENAI_API_KEY", "")
	t.Setenv("UNRELATED_SECRET", secret)

	// When
	err := Run(context.Background(), &out, []string{"config", "doctor", "--workspace", root, "--format", "text"})
	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v\n%s", err, out.String())
	}
	text := out.String()
	if !strings.Contains(text, "export OPENAI_API_KEY=...") {
		t.Fatalf("config doctor text missing export guidance:\n%s", text)
	}
	if strings.Contains(text, secret) {
		t.Fatalf("config doctor leaked secret value:\n%s", text)
	}
}

func Test_Run_config_doctor_rejects_secret_like_http_provider_api_key_env_without_printing_it(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	secretLikeName := "sk-secret-as-env-name"
	content := `{"providers":{"main":{"http":{"url":"http://127.0.0.1:8080/v1/chat/completions","model":"gpt-5","api_key_env":"` + secretLikeName + `"}}},"agent_providers":{"scanner":"main"}}`
	if err := os.WriteFile(filepath.Join(root, ".ceo-harness.json"), []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	// When
	err := Run(context.Background(), &out, []string{"config", "doctor", "--workspace", root, "--format", "text"})

	// Then
	if err == nil {
		t.Fatal("expected invalid config error")
	}
	if strings.Contains(out.String(), secretLikeName) || strings.Contains(err.Error(), secretLikeName) {
		t.Fatalf("config doctor leaked secret-like env name; output=%q error=%v", out.String(), err)
	}
}

func Test_Run_config_doctor_rejects_secret_like_provider_env_var_without_printing_it(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	secretLikeName := "sk-secret-as-env-name"
	content := `{"providers":{"fast":{"model_command":["echo","ok"],"env_vars":["` + secretLikeName + `"]}},"agent_providers":{"scanner":"fast"}}`
	if err := os.WriteFile(filepath.Join(root, ".ceo-harness.json"), []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	// When
	err := Run(context.Background(), &out, []string{"config", "doctor", "--workspace", root, "--format", "text"})

	// Then
	if err == nil {
		t.Fatal("expected invalid config error")
	}
	if strings.Contains(out.String(), secretLikeName) || strings.Contains(err.Error(), secretLikeName) {
		t.Fatalf("config doctor leaked secret-like env name; output=%q error=%v", out.String(), err)
	}
}
