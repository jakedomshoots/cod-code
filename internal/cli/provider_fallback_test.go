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

func Test_Run_falls_back_to_provider_policy_fallback_when_primary_provider_fails(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	configJSON := `{"providers":{"cheap":{"model_command":["sh","-c","cat >/dev/null; echo cheap-down >&2; exit 42"]},"premium":{"model_command":["sh","-c","cat >/dev/null; printf premium-fallback"]}},"provider_policy":{"default_provider":"cheap","fallback_provider":"premium"}}`
	if err := os.WriteFile(filepath.Join(root, ".ceo-harness.json"), []byte(configJSON), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	// When
	err := Run(context.Background(), &out, []string{"--workspace", root, "Fix", "a", "bug"})
	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v\n%s", err, out.String())
	}
	var body struct {
		SubagentResults []struct {
			ProviderName           string   `json:"provider_name"`
			ProviderFallbackFrom   string   `json:"provider_fallback_from"`
			ProviderFallbackReason string   `json:"provider_fallback_reason"`
			Summary                string   `json:"summary"`
			AttemptErrors          []string `json:"attempt_errors"`
		} `json:"subagent_results"`
		RunEvents []struct {
			Kind                   string `json:"kind"`
			AgentName              string `json:"agent_name"`
			ProviderName           string `json:"provider_name"`
			ProviderFallbackFrom   string `json:"provider_fallback_from"`
			ProviderFallbackReason string `json:"provider_fallback_reason"`
		} `json:"run_events"`
	}
	if jsonErr := json.Unmarshal(out.Bytes(), &body); jsonErr != nil {
		t.Fatalf("output must be JSON: %v\n%s", jsonErr, out.String())
	}
	for _, result := range body.SubagentResults {
		if result.ProviderName != "premium" || result.ProviderFallbackFrom != "cheap" {
			t.Fatalf("result = %+v, want premium fallback from cheap", result)
		}
		if result.ProviderFallbackReason != "command_failed" {
			t.Fatalf("fallback reason = %q, want command_failed", result.ProviderFallbackReason)
		}
		if result.Summary != "premium-fallback" {
			t.Fatalf("summary = %q, want premium fallback", result.Summary)
		}
		if len(result.AttemptErrors) != 1 || !strings.Contains(result.AttemptErrors[0], "cheap-down") {
			t.Fatalf("AttemptErrors = %#v, want cheap failure", result.AttemptErrors)
		}
	}
	foundFallbackEvent := false
	for _, event := range body.RunEvents {
		if event.Kind != "subagent" {
			continue
		}
		if event.ProviderName == "premium" && event.ProviderFallbackFrom == "cheap" && event.ProviderFallbackReason == "command_failed" {
			foundFallbackEvent = true
		}
	}
	if !foundFallbackEvent {
		t.Fatalf("run events = %#v, want subagent fallback event", body.RunEvents)
	}
}

func Test_Run_config_check_reports_provider_policy_fallback(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	configJSON := `{"providers":{"cheap":{"model_command":["echo","cheap"]},"premium":{"model_command":["echo","premium"]}},"provider_policy":{"default_provider":"cheap","fallback_provider":"premium"}}`
	if err := os.WriteFile(filepath.Join(root, ".ceo-harness.json"), []byte(configJSON), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	// When
	err := Run(context.Background(), &out, []string{"--workspace", root, "--config-check"})
	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	var body struct {
		ProviderFallbackProvider string `json:"provider_fallback_provider"`
	}
	if jsonErr := json.Unmarshal(out.Bytes(), &body); jsonErr != nil {
		t.Fatalf("output must be JSON: %v\n%s", jsonErr, out.String())
	}
	if body.ProviderFallbackProvider != "premium" {
		t.Fatalf("ProviderFallbackProvider = %q, want premium", body.ProviderFallbackProvider)
	}
}
