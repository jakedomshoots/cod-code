package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func Test_Run_fails_verdict_when_provider_cost_exceeds_workspace_budget(t *testing.T) {
	// Given
	var out bytes.Buffer
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(`{"model":"served-model","choices":[{"message":{"content":"scanner budget response"}}],"usage":{"prompt_tokens":21,"completion_tokens":8,"total_tokens":29}}`)); err != nil {
			t.Fatalf("write response: %v", err)
		}
	}))
	defer server.Close()
	root := t.TempDir()
	configJSON := `{"provider_cost_budget_microusd":50,"providers":{"fast":{"http":{"url":"` + server.URL + `","model":"configured-model","input_cost_per_million_tokens":2,"output_cost_per_million_tokens":8}}},"agent_providers":{"scanner":"fast"}}`
	if err := os.WriteFile(filepath.Join(root, ".ceo-harness.json"), []byte(configJSON), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	// When
	err := Run(context.Background(), &out, []string{"--workspace", root, "Fix", "budget", "smoke"})

	// Then
	if !errors.Is(err, ErrVerdictFailed) {
		t.Fatalf("Run error = %v, want failed verdict", err)
	}
	var body struct {
		VerificationSummary struct {
			ProviderEstimatedCostMicroUSD int64 `json:"provider_estimated_cost_microusd"`
			ProviderCostBudgetMicroUSD    int64 `json:"provider_cost_budget_microusd"`
			ProviderCostOverBudget        bool  `json:"provider_cost_over_budget"`
		} `json:"verification_summary"`
		Verdict string `json:"verdict"`
	}
	if err := json.Unmarshal(out.Bytes(), &body); err != nil {
		t.Fatalf("output must be JSON: %v\n%s", err, out.String())
	}
	if body.Verdict != "fail" {
		t.Fatalf("Verdict = %q, want fail", body.Verdict)
	}
	summary := body.VerificationSummary
	if summary.ProviderEstimatedCostMicroUSD != 106 || summary.ProviderCostBudgetMicroUSD != 50 || !summary.ProviderCostOverBudget {
		t.Fatalf("provider budget summary = %#v, want cost 106 over budget 50", summary)
	}

	historyOut := bytes.Buffer{}
	if err := Run(context.Background(), &historyOut, []string{"--workspace", root, "--history"}); err != nil {
		t.Fatalf("Run history returned error: %v", err)
	}
	var historyBody struct {
		History []struct {
			ProviderCostBudgetMicroUSD int64 `json:"provider_cost_budget_microusd"`
			ProviderCostOverBudget     bool  `json:"provider_cost_over_budget"`
		} `json:"history"`
	}
	if err := json.Unmarshal(historyOut.Bytes(), &historyBody); err != nil {
		t.Fatalf("history output must be JSON: %v\n%s", err, historyOut.String())
	}
	if len(historyBody.History) != 1 {
		t.Fatalf("history length = %d, want 1", len(historyBody.History))
	}
	if historyBody.History[0].ProviderCostBudgetMicroUSD != 50 || !historyBody.History[0].ProviderCostOverBudget {
		t.Fatalf("history provider budget = %#v, want over budget 50", historyBody.History[0])
	}
}

func Test_Run_prints_provider_cost_budget_when_config_check_runs(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	configJSON := `{"provider_cost_budget_microusd":50}`
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
		ProviderCostBudgetMicroUSD int64 `json:"provider_cost_budget_microusd"`
	}
	if err := json.Unmarshal(out.Bytes(), &body); err != nil {
		t.Fatalf("output must be JSON: %v\n%s", err, out.String())
	}
	if body.ProviderCostBudgetMicroUSD != 50 {
		t.Fatalf("ProviderCostBudgetMicroUSD = %d, want 50", body.ProviderCostBudgetMicroUSD)
	}
}
