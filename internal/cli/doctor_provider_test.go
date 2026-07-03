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

func Test_Run_prints_provider_doctor_check_when_provider_is_configured(t *testing.T) {
	// Given
	var out bytes.Buffer
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer provider-token" {
			t.Fatalf("Authorization = %q, want bearer token", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"doctor provider passed"}}]}`))
	}))
	defer server.Close()
	root := t.TempDir()
	configJSON := `{"providers":{"main":{"http":{"url":"` + server.URL + `","model":"worker-model","api_key_env":"CEO_PROVIDER_KEY"}}},"provider_policy":{"default_provider":"main"}}`
	if err := os.WriteFile(filepath.Join(root, ".ceo-harness.json"), []byte(configJSON), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	t.Setenv("CEO_PROVIDER_KEY", "provider-token")

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
		if check.Name == "provider.main" && check.Status == "pass" && check.Source == "workspace" && check.Error == "" {
			return
		}
	}
	t.Fatalf("Checks = %#v, want passing provider.main check", body.Checks)
}

func Test_Run_prints_named_provider_doctor_check_when_doctor_provider_is_supplied(t *testing.T) {
	// Given
	var out bytes.Buffer
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer provider-token" {
			t.Fatalf("Authorization = %q, want bearer token", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"named provider passed"}}]}`))
	}))
	defer server.Close()
	root := t.TempDir()
	configJSON := `{"providers":{"main":{"http":{"url":"` + server.URL + `","model":"worker-model","api_key_env":"CEO_PROVIDER_KEY"}}},"provider_policy":{"default_provider":"main"}}`
	if err := os.WriteFile(filepath.Join(root, ".ceo-harness.json"), []byte(configJSON), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	t.Setenv("CEO_PROVIDER_KEY", "provider-token")

	// When
	err := Run(context.Background(), &out, []string{"--workspace", root, "--doctor-provider", "main"})

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
	if body.Status != "pass" || len(body.Checks) != 1 {
		t.Fatalf("doctor status/checks = %q/%#v, want one passing check", body.Status, body.Checks)
	}
	if body.Checks[0].Name != "provider.main" || body.Checks[0].Status != "pass" || body.Checks[0].Source != "workspace" || body.Checks[0].Error != "" {
		t.Fatalf("check = %#v, want passing provider.main", body.Checks[0])
	}
}

func Test_Run_fails_named_provider_doctor_check_when_provider_is_unknown(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	configJSON := `{"providers":{"main":{"model_command":["echo","ok"]}},"provider_policy":{"default_provider":"main"}}`
	if err := os.WriteFile(filepath.Join(root, ".ceo-harness.json"), []byte(configJSON), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	// When
	err := Run(context.Background(), &out, []string{"--workspace", root, "--doctor-provider", "missing"})

	// Then
	if err == nil {
		t.Fatal("expected missing provider error")
	}
	var body struct {
		Status string `json:"status"`
		Checks []struct {
			Name   string `json:"name"`
			Status string `json:"status"`
			Error  string `json:"error"`
		} `json:"checks"`
	}
	if jsonErr := json.Unmarshal(out.Bytes(), &body); jsonErr != nil {
		t.Fatalf("doctor output must be JSON: %v\n%s", jsonErr, out.String())
	}
	if body.Status != "fail" || len(body.Checks) != 1 {
		t.Fatalf("doctor status/checks = %q/%#v, want one failed check", body.Status, body.Checks)
	}
	if body.Checks[0].Name != "provider.missing" || body.Checks[0].Status != "fail" || body.Checks[0].Error == "" {
		t.Fatalf("check = %#v, want failed provider.missing with error", body.Checks[0])
	}
}
