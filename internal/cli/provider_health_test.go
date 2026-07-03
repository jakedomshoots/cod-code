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

func Test_Run_prints_provider_health_and_history_when_http_provider_has_prices(t *testing.T) {
	// Given
	var out bytes.Buffer
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(`{"model":"served-model","choices":[{"message":{"content":"scanner health response"}}],"usage":{"prompt_tokens":21,"completion_tokens":8,"total_tokens":29}}`)); err != nil {
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
	err := Run(context.Background(), &out, []string{"--workspace", root, "Fix", "provider", "health"})

	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	var body struct {
		VerificationSummary struct {
			ProviderHealth []struct {
				ProviderName          string `json:"provider_name"`
				ModelSource           string `json:"model_source"`
				AttemptCount          int    `json:"attempt_count"`
				PassCount             int    `json:"pass_count"`
				EstimatedCostMicroUSD int64  `json:"estimated_cost_microusd"`
			} `json:"provider_health"`
		} `json:"verification_summary"`
	}
	if err := json.Unmarshal(out.Bytes(), &body); err != nil {
		t.Fatalf("output must be JSON: %v\n%s", err, out.String())
	}
	if len(body.VerificationSummary.ProviderHealth) != 1 {
		t.Fatalf("provider health = %#v, want one provider row", body.VerificationSummary.ProviderHealth)
	}
	health := body.VerificationSummary.ProviderHealth[0]
	if health.ProviderName != "fast" || health.ModelSource != "http" {
		t.Fatalf("provider health route = %#v, want fast http", health)
	}
	if health.AttemptCount != 1 || health.PassCount != 1 || health.EstimatedCostMicroUSD != 106 {
		t.Fatalf("provider health = %#v, want one pass with cost 106", health)
	}

	historyOut := bytes.Buffer{}
	if err := Run(context.Background(), &historyOut, []string{"--workspace", root, "--history"}); err != nil {
		t.Fatalf("Run history returned error: %v", err)
	}
	var historyBody struct {
		History []struct {
			ProviderHealth []struct {
				ProviderName          string `json:"provider_name"`
				PassCount             int    `json:"pass_count"`
				EstimatedCostMicroUSD int64  `json:"estimated_cost_microusd"`
			} `json:"provider_health"`
		} `json:"history"`
	}
	if err := json.Unmarshal(historyOut.Bytes(), &historyBody); err != nil {
		t.Fatalf("history output must be JSON: %v\n%s", err, historyOut.String())
	}
	if len(historyBody.History) != 1 || len(historyBody.History[0].ProviderHealth) != 1 {
		t.Fatalf("history provider health = %#v, want one stored provider row", historyBody.History)
	}
	historyHealth := historyBody.History[0].ProviderHealth[0]
	if historyHealth.ProviderName != "fast" || historyHealth.PassCount != 1 || historyHealth.EstimatedCostMicroUSD != 106 {
		t.Fatalf("history provider health = %#v, want fast pass with cost 106", historyHealth)
	}
}
