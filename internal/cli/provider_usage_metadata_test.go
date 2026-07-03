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

func Test_Run_prints_provider_usage_metadata_when_http_provider_returns_it(t *testing.T) {
	// Given
	var out bytes.Buffer
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(`{"model":"served-model","choices":[{"message":{"content":"scanner usage response"}}],"usage":{"prompt_tokens":13,"completion_tokens":5,"total_tokens":18}}`)); err != nil {
			t.Fatalf("write response: %v", err)
		}
	}))
	defer server.Close()
	root := t.TempDir()
	configJSON := `{"providers":{"fast":{"http":{"url":"` + server.URL + `","model":"configured-model"}}},"agent_providers":{"scanner":"fast"}}`
	if err := os.WriteFile(filepath.Join(root, ".ceo-harness.json"), []byte(configJSON), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	// When
	err := Run(context.Background(), &out, []string{"--workspace", root, "Fix", "usage", "smoke"})
	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	var body struct {
		SubagentResults []struct {
			ProviderPromptTokens     int `json:"provider_prompt_tokens"`
			ProviderCompletionTokens int `json:"provider_completion_tokens"`
			ProviderTotalTokens      int `json:"provider_total_tokens"`
		} `json:"subagent_results"`
	}
	if jsonErr := json.Unmarshal(out.Bytes(), &body); jsonErr != nil {
		t.Fatalf("output must be JSON: %v\n%s", jsonErr, out.String())
	}
	got := body.SubagentResults[0]
	if got.ProviderPromptTokens != 13 || got.ProviderCompletionTokens != 5 || got.ProviderTotalTokens != 18 {
		t.Fatalf("provider usage = prompt %d completion %d total %d, want 13/5/18", got.ProviderPromptTokens, got.ProviderCompletionTokens, got.ProviderTotalTokens)
	}
}
