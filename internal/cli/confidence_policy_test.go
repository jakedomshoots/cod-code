package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"ceoharness/internal/config"
)

func Test_Run_falls_back_when_primary_provider_confidence_is_low(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	configJSON := `{"min_subagent_confidence":0.6,"providers":{"cheap":{"model_command":["sh","-c","cat >/dev/null; printf '{\"summary\":\"cheap unsure\",\"confidence\":0.2}'"]},"premium":{"model_command":["sh","-c","cat >/dev/null; printf '{\"summary\":\"premium sure\",\"confidence\":0.9}'"]}},"provider_policy":{"default_provider":"cheap","fallback_provider":"premium"}}`
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
			Confidence             *float64 `json:"confidence"`
			Summary                string   `json:"summary"`
		} `json:"subagent_results"`
	}
	if jsonErr := json.Unmarshal(out.Bytes(), &body); jsonErr != nil {
		t.Fatalf("output must be JSON: %v\n%s", jsonErr, out.String())
	}
	for _, result := range body.SubagentResults {
		if result.ProviderName != "premium" || result.ProviderFallbackFrom != "cheap" || result.ProviderFallbackReason != "low_confidence" {
			t.Fatalf("result = %+v, want premium low-confidence fallback", result)
		}
		if result.Confidence == nil || *result.Confidence != 0.9 || result.Summary != "premium sure" {
			t.Fatalf("result = %+v, want fallback confidence and summary", result)
		}
	}
}

func Test_Run_uses_min_subagent_confidence_flag_over_workspace_config(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	configJSON := `{"min_subagent_confidence":0.1,"providers":{"cheap":{"model_command":["sh","-c","cat >/dev/null; printf '{\"summary\":\"cheap unsure\",\"confidence\":0.2}'"]},"premium":{"model_command":["sh","-c","cat >/dev/null; printf '{\"summary\":\"premium sure\",\"confidence\":0.9}'"]}},"provider_policy":{"default_provider":"cheap","fallback_provider":"premium"}}`
	if err := os.WriteFile(filepath.Join(root, ".ceo-harness.json"), []byte(configJSON), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	// When
	err := Run(context.Background(), &out, []string{
		"--workspace",
		root,
		"--min-subagent-confidence",
		"0.6",
		"Fix",
		"a",
		"bug",
	})
	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v\n%s", err, out.String())
	}
	var body struct {
		SubagentResults []struct {
			ProviderName           string `json:"provider_name"`
			ProviderFallbackReason string `json:"provider_fallback_reason"`
		} `json:"subagent_results"`
	}
	if jsonErr := json.Unmarshal(out.Bytes(), &body); jsonErr != nil {
		t.Fatalf("output must be JSON: %v\n%s", jsonErr, out.String())
	}
	for _, result := range body.SubagentResults {
		if result.ProviderName != "premium" || result.ProviderFallbackReason != "low_confidence" {
			t.Fatalf("result = %+v, want flag-driven premium fallback", result)
		}
	}
}

func Test_Run_writes_min_subagent_confidence_when_init_config_flag_is_supplied(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()

	// When
	err := Run(context.Background(), &out, []string{
		"--workspace",
		root,
		"--init-config",
		"--min-subagent-confidence",
		"0.7",
	})
	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	cfg, loadErr := config.LoadWorkspace(context.Background(), root)
	if loadErr != nil {
		t.Fatalf("LoadWorkspace returned error: %v", loadErr)
	}
	if cfg.MinSubagentConfidence != 0.7 {
		t.Fatalf("MinSubagentConfidence = %v, want 0.7", cfg.MinSubagentConfidence)
	}
	var body struct {
		MinSubagentConfidence float64 `json:"min_subagent_confidence"`
	}
	if jsonErr := json.Unmarshal(out.Bytes(), &body); jsonErr != nil {
		t.Fatalf("output must be JSON: %v\n%s", jsonErr, out.String())
	}
	if body.MinSubagentConfidence != 0.7 {
		t.Fatalf("MinSubagentConfidence report = %v, want 0.7", body.MinSubagentConfidence)
	}
}

func Test_Run_prints_min_subagent_confidence_config_check(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, ".ceo-harness.json"), []byte(`{"min_subagent_confidence":0.8}`), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	// When
	err := Run(context.Background(), &out, []string{"--workspace", root, "--config-check"})
	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	var body struct {
		MinSubagentConfidence float64 `json:"min_subagent_confidence"`
	}
	if jsonErr := json.Unmarshal(out.Bytes(), &body); jsonErr != nil {
		t.Fatalf("output must be JSON: %v\n%s", jsonErr, out.String())
	}
	if body.MinSubagentConfidence != 0.8 {
		t.Fatalf("MinSubagentConfidence = %v, want 0.8", body.MinSubagentConfidence)
	}
}
