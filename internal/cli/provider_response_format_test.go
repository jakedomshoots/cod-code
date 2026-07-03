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

func Test_Run_sends_provider_response_format_to_http_provider(t *testing.T) {
	// Given
	var out bytes.Buffer
	var gotFormat string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			ResponseFormat struct {
				Type string `json:"type"`
			} `json:"response_format"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		gotFormat = body.ResponseFormat.Type
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(`{"model":"served-model","choices":[{"message":{"content":"{\"status\":\"ok\"}"}}]}`)); err != nil {
			t.Fatalf("write response: %v", err)
		}
	}))
	defer server.Close()
	root := t.TempDir()
	configJSON := `{"providers":{"fast":{"http":{"url":"` + server.URL + `","model":"configured-model","response_format":"json_object"}}},"agent_providers":{"scanner":"fast"}}`
	if err := os.WriteFile(filepath.Join(root, ".ceo-harness.json"), []byte(configJSON), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	// When
	err := Run(context.Background(), &out, []string{"--workspace", root, "Fix", "response", "format", "smoke"})

	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if gotFormat != "json_object" {
		t.Fatalf("response_format.type = %q, want json_object", gotFormat)
	}
}

func Test_Run_rejects_unstructured_provider_output_when_response_format_requires_json(t *testing.T) {
	// Given
	var out bytes.Buffer
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(`{"model":"served-model","choices":[{"message":{"content":"loose provider prose"}}]}`)); err != nil {
			t.Fatalf("write response: %v", err)
		}
	}))
	defer server.Close()
	root := t.TempDir()
	configJSON := `{"providers":{"fast":{"http":{"url":"` + server.URL + `","model":"configured-model","response_format":"json_object"}}},"agent_providers":{"scanner":"fast"}}`
	if err := os.WriteFile(filepath.Join(root, ".ceo-harness.json"), []byte(configJSON), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	// When
	err := Run(context.Background(), &out, []string{"--subagent-attempts", "1", "--workspace", root, "Fix", "response", "format", "smoke"})

	// Then
	if !errors.Is(err, ErrVerdictFailed) {
		t.Fatalf("Run error = %v, want ErrVerdictFailed", err)
	}
	var body struct {
		SubagentResults []struct {
			Status            string `json:"status"`
			ProviderErrorKind string `json:"provider_error_kind"`
		} `json:"subagent_results"`
	}
	if jsonErr := json.Unmarshal(out.Bytes(), &body); jsonErr != nil {
		t.Fatalf("output must be JSON: %v\n%s", jsonErr, out.String())
	}
	if len(body.SubagentResults) == 0 || body.SubagentResults[0].Status != "fail" {
		t.Fatalf("subagent results = %#v, want failed provider result", body.SubagentResults)
	}
	if body.SubagentResults[0].ProviderErrorKind != "model_output_invalid" {
		t.Fatalf("provider error kind = %q, want model_output_invalid", body.SubagentResults[0].ProviderErrorKind)
	}
}
