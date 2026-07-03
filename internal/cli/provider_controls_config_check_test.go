package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func Test_Run_config_check_reports_http_provider_control_counts(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	content := `{"providers":{"fast":{"http":{"url":"http://127.0.0.1:8080/v1/chat/completions","model":"fast-model","input_cost_per_million_tokens":2,"output_cost_per_million_tokens":8,"timeout_ms":2500,"max_output_tokens":64,"response_format":"json_object"}}},"agent_providers":{"scanner":"fast"}}`
	if err := os.WriteFile(filepath.Join(root, ".ceo-harness.json"), []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	// When
	err := Run(context.Background(), &out, []string{"--workspace", root, "--config-check"})

	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	var body struct {
		ProviderHTTPCostCount            int `json:"provider_http_cost_count"`
		ProviderHTTPTimeoutCount         int `json:"provider_http_timeout_count"`
		ProviderHTTPMaxOutputTokensCount int `json:"provider_http_max_output_tokens_count"`
		ProviderHTTPResponseFormatCount  int `json:"provider_http_response_format_count"`
	}
	if jsonErr := json.Unmarshal(out.Bytes(), &body); jsonErr != nil {
		t.Fatalf("output must be JSON: %v\n%s", jsonErr, out.String())
	}
	if body.ProviderHTTPCostCount != 1 || body.ProviderHTTPTimeoutCount != 1 || body.ProviderHTTPMaxOutputTokensCount != 1 || body.ProviderHTTPResponseFormatCount != 1 {
		t.Fatalf("provider control counts = %#v, want one of each", body)
	}
}
