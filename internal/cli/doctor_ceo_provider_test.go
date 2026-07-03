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

func Test_Run_prints_ceo_provider_doctor_check_when_ceo_provider_is_configured(t *testing.T) {
	// Given
	var out bytes.Buffer
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer ceo-token" {
			t.Fatalf("Authorization = %q, want bearer token", got)
		}
		var body struct {
			Messages []struct {
				Content string `json:"content"`
			} `json:"messages"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		if strings.Contains(body.Messages[0].Content, "candidate_subagents") {
			_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"{\"selected_subagents\":[\"coder\"],\"summary\":\"CEO provider delegated.\"}"}}]}`))
			return
		}
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"{\"recommended_verdict\":\"pass\",\"summary\":\"CEO provider approved.\"}"}}]}`))
	}))
	defer server.Close()
	root := t.TempDir()
	configJSON := `{"providers":{"main":{"http":{"url":"` + server.URL + `","model":"ceo-model","api_key_env":"CEO_MAIN_KEY"}}},"ceo_provider":"main"}`
	if err := os.WriteFile(filepath.Join(root, ".ceo-harness.json"), []byte(configJSON), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	t.Setenv("CEO_MAIN_KEY", "ceo-token")

	// When
	err := Run(context.Background(), &out, []string{"--workspace", root, "--doctor"})

	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v\n%s", err, out.String())
	}
	var body struct {
		Status string `json:"status"`
		Checks []struct {
			Name   string `json:"name"`
			Status string `json:"status"`
			Source string `json:"source"`
			Error  string `json:"error"`
		} `json:"checks"`
	}
	if err := json.Unmarshal(out.Bytes(), &body); err != nil {
		t.Fatalf("doctor output must be JSON: %v\n%s", err, out.String())
	}
	if body.Status != "pass" {
		t.Fatalf("Status = %q, want pass", body.Status)
	}
	for _, check := range body.Checks {
		if check.Name == "ceo_provider" && check.Status == "pass" && check.Source == "workspace" && check.Error == "" {
			return
		}
	}
	t.Fatalf("Checks = %#v, want passing ceo_provider check", body.Checks)
}
