package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
)

func Test_Run_with_example_command_model_script(t *testing.T) {
	// Given
	var out bytes.Buffer
	script := filepath.Join("..", "..", "examples", "command-model.sh")
	args := []string{
		"--subagent-attempts",
		"1",
		"--model-command",
		"sh",
		script,
		"--",
		"Inspect",
		"the",
		"workspace",
	}

	// When
	err := Run(context.Background(), &out, args)

	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	var body struct {
		SubagentResults []struct {
			Summary string `json:"summary"`
		} `json:"subagent_results"`
		VerificationSummary struct {
			ProviderErrorCount int `json:"provider_error_count"`
		} `json:"verification_summary"`
	}
	if err := json.Unmarshal(out.Bytes(), &body); err != nil {
		t.Fatalf("output must be JSON: %v\n%s", err, out.String())
	}
	if len(body.SubagentResults) == 0 || !strings.Contains(body.SubagentResults[0].Summary, "example command model handled planner") {
		t.Fatalf("subagent results = %#v, want planner-specific example summary", body.SubagentResults)
	}
	if body.VerificationSummary.ProviderErrorCount != 0 {
		t.Fatalf("ProviderErrorCount = %d, want 0", body.VerificationSummary.ProviderErrorCount)
	}
}
