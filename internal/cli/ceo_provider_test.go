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
	"strings"
	"testing"

	"ceoharness/internal/config"
)

func Test_Run_uses_ceo_http_provider_from_workspace_config_to_veto_passing_run(t *testing.T) {
	// Given
	var out bytes.Buffer
	server := newCEOProviderTestServer(t)
	defer server.Close()
	root := t.TempDir()
	configJSON := `{"providers":{"main":{"http":{"url":"` + server.URL + `","model":"ceo-model","api_key_env":"CEO_MAIN_KEY"}}},"ceo_provider":"main"}`
	if err := os.WriteFile(filepath.Join(root, ".ceo-harness.json"), []byte(configJSON), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	t.Setenv("CEO_MAIN_KEY", "ceo-token")

	// When
	err := Run(context.Background(), &out, []string{"--workspace", root, "Fix", "a", "failing", "test"})

	// Then
	if !errors.Is(err, ErrVerdictFailed) {
		t.Fatalf("Run error = %v, want CEO veto failure", err)
	}
	var body struct {
		CEODelegation struct {
			ModelSource  string `json:"model_source"`
			ProviderName string `json:"provider_name"`
			Summary      string `json:"summary"`
		} `json:"ceo_delegation"`
		CEOReview struct {
			ModelSource  string `json:"model_source"`
			ProviderName string `json:"provider_name"`
			Summary      string `json:"summary"`
		} `json:"ceo_review"`
		Verdict string `json:"verdict"`
	}
	if jsonErr := json.Unmarshal(out.Bytes(), &body); jsonErr != nil {
		t.Fatalf("output must be JSON: %v\n%s", jsonErr, out.String())
	}
	if body.CEOReview.Summary != "CEO provider veto" || body.Verdict != "fail" {
		t.Fatalf("CEO review = %#v verdict = %q, want provider veto", body.CEOReview, body.Verdict)
	}
	if body.CEOReview.ModelSource != "http" || body.CEOReview.ProviderName != "main" {
		t.Fatalf("CEO route = source %q provider %q, want http main", body.CEOReview.ModelSource, body.CEOReview.ProviderName)
	}
	if body.CEODelegation.ModelSource != "http" || body.CEODelegation.ProviderName != "main" {
		t.Fatalf("CEO delegation route = source %q provider %q, want http main", body.CEODelegation.ModelSource, body.CEODelegation.ProviderName)
	}
}

func Test_Run_writes_ceo_provider_when_init_config_ceo_provider_flag_is_supplied(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	args := []string{
		"--workspace",
		root,
		"--init-config",
		"--http-provider",
		"main",
		"--http-preset",
		"openai",
		"--http-model",
		"gpt-5",
		"--ceo-provider",
		"main",
	}

	// When
	err := Run(context.Background(), &out, args)

	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	cfg, loadErr := config.LoadWorkspace(context.Background(), root)
	if loadErr != nil {
		t.Fatalf("LoadWorkspace returned error: %v", loadErr)
	}
	if cfg.CEOProvider != "main" {
		t.Fatalf("CEOProvider = %q, want main", cfg.CEOProvider)
	}
}

func Test_Run_quickstart_ceo_provider_overrides_example_ceo_adapter_when_running_workspace(t *testing.T) {
	// Given
	var out bytes.Buffer
	server := newCEOProviderTestServer(t)
	defer server.Close()
	root := t.TempDir()
	t.Setenv("CEO_MAIN_KEY", "ceo-token")

	// When
	err := Run(context.Background(), &out, []string{
		"--quickstart",
		root,
		"--http-provider",
		"main",
		"--http-url",
		server.URL,
		"--http-model",
		"ceo-model",
		"--http-api-key-env",
		"CEO_MAIN_KEY",
		"--ceo-provider",
		"main",
	})

	// Then
	if err != nil {
		t.Fatalf("quickstart returned error: %v\n%s", err, out.String())
	}
	out.Reset()

	err = Run(context.Background(), &out, []string{"--workspace", root, "Fix", "a", "failing", "test"})
	if !errors.Is(err, ErrVerdictFailed) {
		t.Fatalf("Run error = %v, want CEO provider veto", err)
	}
	var body struct {
		CEOReview struct {
			Summary string `json:"summary"`
		} `json:"ceo_review"`
	}
	if jsonErr := json.Unmarshal(out.Bytes(), &body); jsonErr != nil {
		t.Fatalf("output must be JSON: %v\n%s", jsonErr, out.String())
	}
	if body.CEOReview.Summary != "CEO provider veto" {
		t.Fatalf("CEO summary = %q, want provider veto", body.CEOReview.Summary)
	}
}

func newCEOProviderTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
			_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"{\"selected_subagents\":[\"scanner\",\"coder\",\"reviewer\"],\"summary\":\"CEO provider picked the crew.\"}"}}]}`))
			return
		}
		if strings.Contains(body.Messages[0].Content, "Doctor CEO model command check") {
			_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"{\"recommended_verdict\":\"pass\",\"summary\":\"CEO provider doctor passed\"}"}}]}`))
			return
		}
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"{\"recommended_verdict\":\"fail\",\"summary\":\"CEO provider veto\"}"}}]}`))
	}))
}
