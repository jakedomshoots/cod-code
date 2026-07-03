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

func Test_Run_reports_structured_http_provider_error_fields_when_provider_fails(t *testing.T) {
	// Given
	var out bytes.Buffer
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		if _, err := w.Write([]byte(`{"error":"bad key"}`)); err != nil {
			t.Fatalf("write unauthorized response: %v", err)
		}
	}))
	defer server.Close()
	root := t.TempDir()
	configJSON := `{"providers":{"fast":{"http":{"url":"` + server.URL + `","model":"fast-model"}}},"agent_providers":{"coder":"fast"}}`
	if err := os.WriteFile(filepath.Join(root, ".ceo-harness.json"), []byte(configJSON), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	// When
	err := Run(context.Background(), &out, []string{"--workspace", root, "Fix", "an", "auth", "failure"})

	// Then
	if !errors.Is(err, ErrVerdictFailed) {
		t.Fatalf("Run error = %v, want failed verdict", err)
	}
	var body struct {
		SubagentResults []struct {
			AgentName          string `json:"agent_name"`
			Status             string `json:"status"`
			ProviderErrorKind  string `json:"provider_error_kind"`
			ProviderHTTPStatus int    `json:"provider_http_status"`
			AttemptRecords     []struct {
				Status             string `json:"status"`
				ProviderErrorKind  string `json:"provider_error_kind"`
				ProviderHTTPStatus int    `json:"provider_http_status"`
			} `json:"attempt_records"`
		} `json:"subagent_results"`
		VerificationSummary struct {
			ProviderErrorCount        int `json:"provider_error_count"`
			ProviderUnauthorizedCount int `json:"provider_unauthorized_count"`
			ProviderHealth            []struct {
				ProviderName      string `json:"provider_name"`
				FailCount         int    `json:"fail_count"`
				ErrorCount        int    `json:"error_count"`
				UnauthorizedCount int    `json:"unauthorized_count"`
			} `json:"provider_health"`
		} `json:"verification_summary"`
		Verdict string `json:"verdict"`
	}
	if err := json.Unmarshal(out.Bytes(), &body); err != nil {
		t.Fatalf("output must be JSON: %v\n%s", err, out.String())
	}
	if body.Verdict != "fail" {
		t.Fatalf("Verdict = %q, want fail", body.Verdict)
	}
	var coder struct {
		AgentName          string `json:"agent_name"`
		Status             string `json:"status"`
		ProviderErrorKind  string `json:"provider_error_kind"`
		ProviderHTTPStatus int    `json:"provider_http_status"`
		AttemptRecords     []struct {
			Status             string `json:"status"`
			ProviderErrorKind  string `json:"provider_error_kind"`
			ProviderHTTPStatus int    `json:"provider_http_status"`
		} `json:"attempt_records"`
	}
	for _, result := range body.SubagentResults {
		if result.AgentName == "coder" {
			coder = result
			break
		}
	}
	if coder.AgentName != "coder" {
		t.Fatalf("subagent results = %#v, want coder result", body.SubagentResults)
	}
	if coder.Status != "fail" {
		t.Fatalf("coder result = %#v, want failed coder", coder)
	}
	if coder.ProviderErrorKind != "unauthorized" || coder.ProviderHTTPStatus != http.StatusUnauthorized {
		t.Fatalf("coder provider error = %q/%d, want unauthorized/401", coder.ProviderErrorKind, coder.ProviderHTTPStatus)
	}
	if len(coder.AttemptRecords) != 1 {
		t.Fatalf("coder attempt records = %#v, want one failed attempt", coder.AttemptRecords)
	}
	if coder.AttemptRecords[0].ProviderErrorKind != "unauthorized" || coder.AttemptRecords[0].ProviderHTTPStatus != http.StatusUnauthorized {
		t.Fatalf("attempt provider error = %#v, want unauthorized/401", coder.AttemptRecords[0])
	}
	if body.VerificationSummary.ProviderErrorCount != 1 || body.VerificationSummary.ProviderUnauthorizedCount != 1 {
		t.Fatalf("provider summary = %#v, want one unauthorized provider error", body.VerificationSummary)
	}
	if len(body.VerificationSummary.ProviderHealth) != 1 {
		t.Fatalf("provider health = %#v, want one provider row", body.VerificationSummary.ProviderHealth)
	}
	health := body.VerificationSummary.ProviderHealth[0]
	if health.ProviderName != "fast" || health.FailCount != 1 || health.ErrorCount != 1 || health.UnauthorizedCount != 1 {
		t.Fatalf("provider health = %#v, want fast unauthorized failure", health)
	}

	historyOut := bytes.Buffer{}
	if err := Run(context.Background(), &historyOut, []string{"--workspace", root, "--history"}); err != nil {
		t.Fatalf("Run history returned error: %v", err)
	}
	var historyBody struct {
		History []struct {
			ProviderErrorCount        int `json:"provider_error_count"`
			ProviderUnauthorizedCount int `json:"provider_unauthorized_count"`
			ProviderHealth            []struct {
				ProviderName      string `json:"provider_name"`
				FailCount         int    `json:"fail_count"`
				ErrorCount        int    `json:"error_count"`
				UnauthorizedCount int    `json:"unauthorized_count"`
			} `json:"provider_health"`
		} `json:"history"`
	}
	if err := json.Unmarshal(historyOut.Bytes(), &historyBody); err != nil {
		t.Fatalf("history output must be JSON: %v\n%s", err, historyOut.String())
	}
	if len(historyBody.History) != 1 {
		t.Fatalf("history length = %d, want 1", len(historyBody.History))
	}
	if historyBody.History[0].ProviderErrorCount != 1 || historyBody.History[0].ProviderUnauthorizedCount != 1 {
		t.Fatalf("history provider counters = %#v, want one unauthorized provider error", historyBody.History[0])
	}
	if len(historyBody.History[0].ProviderHealth) != 1 {
		t.Fatalf("history provider health = %#v, want one provider row", historyBody.History[0].ProviderHealth)
	}
	historyHealth := historyBody.History[0].ProviderHealth[0]
	if historyHealth.ProviderName != "fast" || historyHealth.FailCount != 1 || historyHealth.ErrorCount != 1 || historyHealth.UnauthorizedCount != 1 {
		t.Fatalf("history provider health = %#v, want fast unauthorized failure", historyHealth)
	}
}
