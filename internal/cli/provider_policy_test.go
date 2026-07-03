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

func Test_Run_routes_high_risk_task_to_provider_policy_target(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	configJSON := `{"providers":{"cheap":{"model_command":["sh","-c","cat >/dev/null; printf cheap-policy"]},"premium":{"model_command":["sh","-c","cat >/dev/null; printf premium-policy"]}},"provider_policy":{"default_provider":"cheap","risk_providers":{"high":"premium"}}}`
	if err := os.WriteFile(filepath.Join(root, ".ceo-harness.json"), []byte(configJSON), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	// When
	err := Run(context.Background(), &out, []string{"--workspace", root, "Fix", "auth", "bug"})

	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	var body struct {
		ProviderRouteDecisions []struct {
			AgentName    string `json:"agent_name"`
			ProviderName string `json:"provider_name"`
			Reason       string `json:"reason"`
		} `json:"provider_route_decisions"`
		SubagentResults []struct {
			AgentName    string `json:"agent_name"`
			ProviderName string `json:"provider_name"`
			Summary      string `json:"summary"`
		} `json:"subagent_results"`
	}
	if err := json.Unmarshal(out.Bytes(), &body); err != nil {
		t.Fatalf("output must be JSON: %v\n%s", err, out.String())
	}
	if len(body.SubagentResults) != 3 {
		t.Fatalf("subagent count = %d, want 3", len(body.SubagentResults))
	}
	for _, result := range body.SubagentResults {
		if result.ProviderName != "premium" || !strings.Contains(result.Summary, "premium-policy") {
			t.Fatalf("result = %#v, want premium policy route", result)
		}
	}
}

func Test_Run_routes_risk_area_specialist_to_provider_policy_target(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	configJSON := `{"providers":{"cheap":{"model_command":["sh","-c","cat >/dev/null; printf cheap-policy"]},"premium":{"model_command":["sh","-c","cat >/dev/null; printf premium-policy"]}},"provider_policy":{"default_provider":"cheap","risk_area_providers":{"database":"premium"}}}`
	if err := os.WriteFile(filepath.Join(root, ".ceo-harness.json"), []byte(configJSON), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	// When
	err := Run(context.Background(), &out, []string{"--workspace", root, "Implement", "database", "migration", "fix"})

	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	var body struct {
		RunLedger struct {
			Owner                string   `json:"owner"`
			Verdict              string   `json:"verdict"`
			NextAction           string   `json:"next_action"`
			VerificationStatus   string   `json:"verification_status"`
			ProviderRouteCount   int      `json:"provider_route_count"`
			ProviderRouteReasons []string `json:"provider_route_reasons"`
		} `json:"run_ledger"`
		ProviderRouteDecisions []struct {
			AgentName    string `json:"agent_name"`
			ProviderName string `json:"provider_name"`
			Reason       string `json:"reason"`
		} `json:"provider_route_decisions"`
		SubagentResults []struct {
			AgentName    string `json:"agent_name"`
			ProviderName string `json:"provider_name"`
			Summary      string `json:"summary"`
		} `json:"subagent_results"`
	}
	if err := json.Unmarshal(out.Bytes(), &body); err != nil {
		t.Fatalf("output must be JSON: %v\n%s", err, out.String())
	}
	if len(body.SubagentResults) != 3 {
		t.Fatalf("subagent count = %d, want 3", len(body.SubagentResults))
	}
	for _, result := range body.SubagentResults {
		if result.AgentName == "database" {
			if result.ProviderName != "premium" || !strings.Contains(result.Summary, "premium-policy") {
				t.Fatalf("database result = %#v, want premium policy route", result)
			}
			continue
		}
		if result.ProviderName != "cheap" || !strings.Contains(result.Summary, "cheap-policy") {
			t.Fatalf("result = %#v, want cheap default route", result)
		}
	}
	assertRouteDecision(t, body.ProviderRouteDecisions, "coder", "cheap", "provider_policy.default_provider")
	assertRouteDecision(t, body.ProviderRouteDecisions, "database", "premium", "provider_policy.risk_area_providers.database")
	assertRouteDecision(t, body.ProviderRouteDecisions, "reviewer", "cheap", "provider_policy.default_provider")
	if body.RunLedger.Owner != "coder" || body.RunLedger.Verdict != "pass" || body.RunLedger.NextAction != "accept" {
		t.Fatalf("RunLedger = %+v, want coder pass accept", body.RunLedger)
	}
	if body.RunLedger.VerificationStatus != "unverified" {
		t.Fatalf("RunLedger.VerificationStatus = %q, want unverified", body.RunLedger.VerificationStatus)
	}
	if body.RunLedger.ProviderRouteCount != 3 {
		t.Fatalf("RunLedger.ProviderRouteCount = %d, want 3", body.RunLedger.ProviderRouteCount)
	}
	if !containsString(body.RunLedger.ProviderRouteReasons, "provider_policy.default_provider") {
		t.Fatalf("RunLedger.ProviderRouteReasons = %+v, want default provider reason", body.RunLedger.ProviderRouteReasons)
	}
	if !containsString(body.RunLedger.ProviderRouteReasons, "provider_policy.risk_area_providers.database") {
		t.Fatalf("RunLedger.ProviderRouteReasons = %+v, want database risk-area reason", body.RunLedger.ProviderRouteReasons)
	}
}

func Test_Run_prints_provider_route_decisions_in_plan_only_preview(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	configJSON := `{"providers":{"cheap":{"model_command":["sh","-c","cat >/dev/null; printf cheap-policy"]},"premium":{"model_command":["sh","-c","cat >/dev/null; printf premium-policy"]}},"provider_policy":{"default_provider":"cheap","risk_area_providers":{"database":"premium"}}}`
	if err := os.WriteFile(filepath.Join(root, ".ceo-harness.json"), []byte(configJSON), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	// When
	err := Run(context.Background(), &out, []string{"--workspace", root, "--plan-only", "Implement", "database", "migration", "fix"})

	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	var body struct {
		RunLedger struct {
			Owner                string   `json:"owner"`
			Verdict              string   `json:"verdict"`
			NextAction           string   `json:"next_action"`
			VerificationStatus   string   `json:"verification_status"`
			ProviderRouteReasons []string `json:"provider_route_reasons"`
		} `json:"run_ledger"`
		ProviderRouteDecisions []struct {
			AgentName    string `json:"agent_name"`
			ProviderName string `json:"provider_name"`
			Reason       string `json:"reason"`
		} `json:"provider_route_decisions"`
	}
	if err := json.Unmarshal(out.Bytes(), &body); err != nil {
		t.Fatalf("output must be JSON: %v\n%s", err, out.String())
	}
	assertRouteDecision(t, body.ProviderRouteDecisions, "database", "premium", "provider_policy.risk_area_providers.database")
	if body.RunLedger.Owner != "coder" || body.RunLedger.Verdict != "pending" || body.RunLedger.NextAction != "run" {
		t.Fatalf("RunLedger = %+v, want coder pending run", body.RunLedger)
	}
	if body.RunLedger.VerificationStatus != "unverified" {
		t.Fatalf("RunLedger.VerificationStatus = %q, want unverified", body.RunLedger.VerificationStatus)
	}
	if !containsString(body.RunLedger.ProviderRouteReasons, "provider_policy.risk_area_providers.database") {
		t.Fatalf("RunLedger.ProviderRouteReasons = %+v, want database risk-area reason", body.RunLedger.ProviderRouteReasons)
	}
}

func Test_Run_config_check_reports_provider_policy_rules_when_config_check_flag_is_supplied(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	configJSON := `{"providers":{"cheap":{"model_command":["echo","cheap"]},"premium":{"model_command":["echo","premium"]}},"provider_policy":{"default_provider":"cheap","risk_providers":{"high":"premium"},"risk_area_providers":{"database":"premium"}}}`
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
		ProviderPolicyRuleCount int `json:"provider_policy_rule_count"`
	}
	if err := json.Unmarshal(out.Bytes(), &body); err != nil {
		t.Fatalf("output must be JSON: %v\n%s", err, out.String())
	}
	if body.ProviderPolicyRuleCount != 3 {
		t.Fatalf("ProviderPolicyRuleCount = %d, want 3", body.ProviderPolicyRuleCount)
	}
}

func assertRouteDecision(t *testing.T, decisions []struct {
	AgentName    string `json:"agent_name"`
	ProviderName string `json:"provider_name"`
	Reason       string `json:"reason"`
}, agentName string, providerName string, reason string) {
	t.Helper()
	for _, decision := range decisions {
		if decision.AgentName != agentName {
			continue
		}
		if decision.ProviderName != providerName || decision.Reason != reason {
			t.Fatalf("decision for %s = %#v, want provider %q reason %q", agentName, decision, providerName, reason)
		}
		return
	}
	t.Fatalf("missing provider decision for %s in %#v", agentName, decisions)
}

func containsString(items []string, want string) bool {
	for _, item := range items {
		if item == want {
			return true
		}
	}
	return false
}
