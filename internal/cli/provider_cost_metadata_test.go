package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func Test_Run_prints_provider_cost_estimate_when_http_provider_has_prices(t *testing.T) {
	// Given
	var out bytes.Buffer
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(`{"model":"served-model","choices":[{"message":{"content":"scanner cost response"}}],"usage":{"prompt_tokens":21,"completion_tokens":8,"total_tokens":29}}`)); err != nil {
			t.Fatalf("write response: %v", err)
		}
	}))
	defer server.Close()
	root := t.TempDir()
	configJSON := `{"providers":{"fast":{"http":{"url":"` + server.URL + `","model":"configured-model","input_cost_per_million_tokens":2,"output_cost_per_million_tokens":8}}},"agent_providers":{"scanner":"fast"}}`
	if err := os.WriteFile(filepath.Join(root, ".ceo-harness.json"), []byte(configJSON), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	// When
	err := Run(context.Background(), &out, []string{"--workspace", root, "Fix", "cost", "smoke"})
	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	var body struct {
		SubagentResults []struct {
			ProviderEstimatedCostMicroUSD int64 `json:"provider_estimated_cost_microusd"`
		} `json:"subagent_results"`
		VerificationSummary struct {
			ProviderEstimatedCostMicroUSD int64 `json:"provider_estimated_cost_microusd"`
		} `json:"verification_summary"`
	}
	if jsonErr := json.Unmarshal(out.Bytes(), &body); jsonErr != nil {
		t.Fatalf("output must be JSON: %v\n%s", jsonErr, out.String())
	}
	if body.SubagentResults[0].ProviderEstimatedCostMicroUSD != 106 {
		t.Fatalf("provider estimated cost = %d microusd, want 106", body.SubagentResults[0].ProviderEstimatedCostMicroUSD)
	}
	if body.VerificationSummary.ProviderEstimatedCostMicroUSD != 106 {
		t.Fatalf("summary provider estimated cost = %d microusd, want 106", body.VerificationSummary.ProviderEstimatedCostMicroUSD)
	}

	historyOut := bytes.Buffer{}
	if err := Run(context.Background(), &historyOut, []string{"--workspace", root, "--history"}); err != nil {
		t.Fatalf("Run history returned error: %v", err)
	}
	var historyBody struct {
		History []struct {
			ProviderEstimatedCostMicroUSD int64 `json:"provider_estimated_cost_microusd"`
		} `json:"history"`
	}
	if jsonErr := json.Unmarshal(historyOut.Bytes(), &historyBody); jsonErr != nil {
		t.Fatalf("history output must be JSON: %v\n%s", jsonErr, historyOut.String())
	}
	if len(historyBody.History) != 1 {
		t.Fatalf("history length = %d, want 1", len(historyBody.History))
	}
	if historyBody.History[0].ProviderEstimatedCostMicroUSD != 106 {
		t.Fatalf("history provider estimated cost = %d microusd, want 106", historyBody.History[0].ProviderEstimatedCostMicroUSD)
	}
}
