package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
)

func Test_Run_retries_rate_limited_http_provider_once_without_subagent_attempts_flag(t *testing.T) {
	// Given
	var out bytes.Buffer
	var requests int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := atomic.AddInt32(&requests, 1)
		if count == 1 {
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(http.StatusTooManyRequests)
			if _, err := w.Write([]byte(`{"error":"slow down"}`)); err != nil {
				t.Fatalf("write rate limit response: %v", err)
			}
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(`{"choices":[{"message":{"content":"scanner retried response"}}]}`)); err != nil {
			t.Fatalf("write success response: %v", err)
		}
	}))
	defer server.Close()
	root := t.TempDir()
	configJSON := `{"providers":{"fast":{"http":{"url":"` + server.URL + `","model":"fast-model"}}},"agent_providers":{"scanner":"fast"}}`
	if err := os.WriteFile(filepath.Join(root, ".ceo-harness.json"), []byte(configJSON), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	// When
	err := Run(context.Background(), &out, []string{"--workspace", root, "Fix", "a", "rate", "limit"})

	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v\n%s", err, out.String())
	}
	var body struct {
		SubagentResults []struct {
			Attempts       int    `json:"attempts"`
			Summary        string `json:"summary"`
			AttemptRecords []struct {
				Status               string `json:"status"`
				Error                string `json:"error"`
				ProviderRetryAfterMS int64  `json:"provider_retry_after_ms"`
			} `json:"attempt_records"`
		} `json:"subagent_results"`
		Verdict string `json:"verdict"`
	}
	if err := json.Unmarshal(out.Bytes(), &body); err != nil {
		t.Fatalf("output must be JSON: %v\n%s", err, out.String())
	}
	if body.Verdict != "pass" {
		t.Fatalf("Verdict = %q, want pass", body.Verdict)
	}
	scanner := body.SubagentResults[0]
	if scanner.Attempts != 2 {
		t.Fatalf("scanner attempts = %d, want one automatic retry", scanner.Attempts)
	}
	if !strings.Contains(scanner.Summary, "scanner retried response") {
		t.Fatalf("scanner summary = %q, want retried HTTP response", scanner.Summary)
	}
	if len(scanner.AttemptRecords) != 2 || scanner.AttemptRecords[0].Status != "fail" || !strings.Contains(scanner.AttemptRecords[0].Error, "rate limited") {
		t.Fatalf("scanner attempt records = %#v, want rate limit then pass", scanner.AttemptRecords)
	}
	if scanner.AttemptRecords[0].ProviderRetryAfterMS != 1000 {
		t.Fatalf("provider retry after = %d, want 1000", scanner.AttemptRecords[0].ProviderRetryAfterMS)
	}
	if atomic.LoadInt32(&requests) != 2 {
		t.Fatalf("provider requests = %d, want 2", requests)
	}
}
