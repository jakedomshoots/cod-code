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
	"testing"
)

func Test_Run_uses_http_provider_profile_when_workspace_config_assigns_agent_provider(t *testing.T) {
	// Given
	var out bytes.Buffer
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer test-token" {
			t.Fatalf("Authorization = %q, want bearer token", got)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(`{"choices":[{"message":{"content":"scanner http response"}}]}`)); err != nil {
			t.Fatalf("write response: %v", err)
		}
	}))
	defer server.Close()
	root := t.TempDir()
	configJSON := `{"providers":{"fast":{"http":{"url":"` + server.URL + `","model":"fast-model","api_key_env":"CEO_FAST_KEY"}}},"agent_providers":{"scanner":"fast"}}`
	if err := os.WriteFile(filepath.Join(root, ".ceo-harness.json"), []byte(configJSON), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	t.Setenv("CEO_FAST_KEY", "test-token")

	// When
	err := Run(context.Background(), &out, []string{"--workspace", root, "Fix", "a", "failing", "test"})

	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	var body struct {
		SubagentResults []struct {
			AgentName    string `json:"agent_name"`
			ModelSource  string `json:"model_source"`
			ProviderName string `json:"provider_name"`
			Summary      string `json:"summary"`
		} `json:"subagent_results"`
	}
	if err := json.Unmarshal(out.Bytes(), &body); err != nil {
		t.Fatalf("output must be JSON: %v\n%s", err, out.String())
	}
	if !strings.Contains(body.SubagentResults[0].Summary, "scanner http response") {
		t.Fatalf("scanner summary = %q, want http response", body.SubagentResults[0].Summary)
	}
	if body.SubagentResults[0].ModelSource != "http" || body.SubagentResults[0].ProviderName != "fast" {
		t.Fatalf("scanner metadata = source %q provider %q, want http fast", body.SubagentResults[0].ModelSource, body.SubagentResults[0].ProviderName)
	}
	if body.SubagentResults[1].Summary != "local deterministic model response" {
		t.Fatalf("coder summary = %q, want default local response", body.SubagentResults[1].Summary)
	}
	if body.SubagentResults[1].ModelSource != "local" || body.SubagentResults[1].ProviderName != "" {
		t.Fatalf("coder metadata = source %q provider %q, want local without provider", body.SubagentResults[1].ModelSource, body.SubagentResults[1].ProviderName)
	}
}
