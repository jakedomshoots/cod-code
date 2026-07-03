package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func Test_Run_uses_max_context_bytes_flag(t *testing.T) {
	// Given
	var out bytes.Buffer
	args := []string{"--max-context-bytes", "10", "Fix", "a", "failing", "test"}

	// When
	err := Run(context.Background(), &out, args)
	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	assertReportContextBudget(t, out.Bytes(), 10)
}

func Test_Run_uses_workspace_max_context_bytes_default(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, ".ceo-harness.json"), []byte(`{"max_context_bytes":12}`), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	// When
	err := Run(context.Background(), &out, []string{"--workspace", root, "Fix", "a", "failing", "test"})
	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	assertReportContextBudget(t, out.Bytes(), 12)
}

func Test_Run_prints_max_context_bytes_config_check(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, ".ceo-harness.json"), []byte(`{"max_context_bytes":2048}`), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	// When
	err := Run(context.Background(), &out, []string{"--workspace", root, "--config-check"})
	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	var body struct {
		MaxContextBytes int `json:"max_context_bytes"`
	}
	if jsonErr := json.Unmarshal(out.Bytes(), &body); jsonErr != nil {
		t.Fatalf("output must be JSON: %v\n%s", jsonErr, out.String())
	}
	if body.MaxContextBytes != 2048 {
		t.Fatalf("MaxContextBytes = %d, want 2048", body.MaxContextBytes)
	}
}

func Test_Run_prints_workspace_brief_exclude_count_config_check(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, ".ceo-harness.json"), []byte(`{"workspace_brief_excludes":["generated","*.lock"]}`), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	// When
	err := Run(context.Background(), &out, []string{"--workspace", root, "--config-check"})

	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	var body struct {
		WorkspaceBriefExcludeCount int `json:"workspace_brief_exclude_count"`
	}
	if jsonErr := json.Unmarshal(out.Bytes(), &body); jsonErr != nil {
		t.Fatalf("output must be JSON: %v\n%s", jsonErr, out.String())
	}
	if body.WorkspaceBriefExcludeCount != 2 {
		t.Fatalf("WorkspaceBriefExcludeCount = %d, want 2", body.WorkspaceBriefExcludeCount)
	}
}

func assertReportContextBudget(t *testing.T, output []byte, want int) {
	t.Helper()
	var body struct {
		JobPacket struct {
			ContextPolicy struct {
				MaxBytes int `json:"max_bytes"`
			} `json:"context_policy"`
		} `json:"job_packet"`
		SubagentResults []struct {
			ContextBytes int `json:"context_bytes"`
		} `json:"subagent_results"`
	}
	if jsonErr := json.Unmarshal(output, &body); jsonErr != nil {
		t.Fatalf("output must be JSON: %v\n%s", jsonErr, string(output))
	}
	if body.JobPacket.ContextPolicy.MaxBytes != want {
		t.Fatalf("MaxBytes = %d, want %d", body.JobPacket.ContextPolicy.MaxBytes, want)
	}
	for index, result := range body.SubagentResults {
		if result.ContextBytes != want {
			t.Fatalf("SubagentResults[%d].ContextBytes = %d, want %d", index, result.ContextBytes, want)
		}
	}
}
