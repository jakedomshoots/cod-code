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

func Test_Run_sends_provider_max_output_tokens_to_http_provider(t *testing.T) {
	// Given
	var out bytes.Buffer
	var gotMaxTokens int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			MaxTokens int `json:"max_tokens"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		gotMaxTokens = body.MaxTokens
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(`{"model":"served-model","choices":[{"message":{"content":"scanner capped response"}}]}`)); err != nil {
			t.Fatalf("write response: %v", err)
		}
	}))
	defer server.Close()
	root := t.TempDir()
	configJSON := `{"providers":{"fast":{"http":{"url":"` + server.URL + `","model":"configured-model","max_output_tokens":64}}},"agent_providers":{"coder":"fast"}}`
	if err := os.WriteFile(filepath.Join(root, ".ceo-harness.json"), []byte(configJSON), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	// When
	err := Run(context.Background(), &out, []string{"--workspace", root, "Fix", "token", "cap", "smoke"})

	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if gotMaxTokens != 64 {
		t.Fatalf("max_tokens = %d, want 64", gotMaxTokens)
	}
}
