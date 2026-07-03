package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"ceoharness/internal/config"
)

func Test_Run_writes_risk_and_kind_provider_policy_when_init_config_policy_flags_are_supplied(t *testing.T) {
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
		"cheap-model",
		"--default-provider",
		"cheap",
		"--http-provider",
		"premium",
		"--http-preset",
		"openai",
		"--http-model",
		"premium-model",
		"--risk-provider",
		"high=premium",
		"--kind-provider",
		"research=premium",
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
	if cfg.ProviderPolicy.RiskProviders["high"] != "premium" {
		t.Fatalf("risk providers = %#v, want high premium", cfg.ProviderPolicy.RiskProviders)
	}
	if cfg.ProviderPolicy.KindProviders["research"] != "premium" {
		t.Fatalf("kind providers = %#v, want research premium", cfg.ProviderPolicy.KindProviders)
	}
	var body struct {
		ProviderPolicyRuleCount int `json:"provider_policy_rule_count"`
	}
	if err := json.Unmarshal(out.Bytes(), &body); err != nil {
		t.Fatalf("output must be JSON: %v\n%s", err, out.String())
	}
	if body.ProviderPolicyRuleCount != 3 {
		t.Fatalf("ProviderPolicyRuleCount = %d, want 3", body.ProviderPolicyRuleCount)
	}
}
