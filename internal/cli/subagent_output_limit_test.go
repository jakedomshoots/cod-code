package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"ceoharness/internal/config"
)

func Test_Run_caps_subagent_output_with_flag(t *testing.T) {
	// Given
	var out bytes.Buffer
	t.Setenv("GO_WANT_CLI_SUBAGENT_OUTPUT_MODEL", "1")

	// When
	err := Run(context.Background(), &out, []string{
		"--model-command",
		os.Args[0],
		"-test.run=Test_HelperProcess_cli_subagent_output_model",
		"--",
		"--max-subagent-output-bytes",
		"32",
		"Fix",
		"a",
		"bug",
	})
	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	var body struct {
		RunManifest struct {
			MaxSubagentOutputBytes int `json:"max_subagent_output_bytes"`
		} `json:"run_manifest"`
		SubagentResults []struct {
			Summary         string   `json:"summary"`
			Evidence        []string `json:"evidence"`
			Questions       []string `json:"questions"`
			OutputTruncated bool     `json:"output_truncated"`
		} `json:"subagent_results"`
	}
	if jsonErr := json.Unmarshal(out.Bytes(), &body); jsonErr != nil {
		t.Fatalf("output must be JSON: %v\n%s", jsonErr, out.String())
	}
	if body.RunManifest.MaxSubagentOutputBytes != 32 {
		t.Fatalf("MaxSubagentOutputBytes = %d, want 32", body.RunManifest.MaxSubagentOutputBytes)
	}
	result := body.SubagentResults[0]
	if !result.OutputTruncated {
		t.Fatal("OutputTruncated = false, want true")
	}
	if !strings.Contains(result.Summary, "[truncated]") || strings.Contains(result.Summary, "RAW_TAIL") {
		t.Fatalf("summary = %q, want capped without raw tail", result.Summary)
	}
	if len(result.Evidence) == 0 || !strings.Contains(result.Evidence[0], "[truncated]") {
		t.Fatalf("evidence = %#v, want capped marker", result.Evidence)
	}
	if len(result.Questions) == 0 || !strings.Contains(result.Questions[0], "[truncated]") {
		t.Fatalf("questions = %#v, want capped marker", result.Questions)
	}
}

func Test_Run_uses_workspace_max_subagent_output_default(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	t.Setenv("GO_WANT_CLI_SUBAGENT_OUTPUT_MODEL", "1")
	configBody := `{"max_subagent_output_bytes":32,"model_command":["` + os.Args[0] + `","-test.run=Test_HelperProcess_cli_subagent_output_model"]}`
	if err := os.WriteFile(filepath.Join(root, ".ceo-harness.json"), []byte(configBody), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	// When
	err := Run(context.Background(), &out, []string{"--workspace", root, "Fix", "a", "bug"})
	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	assertSubagentOutputLimit(t, out.Bytes(), 32)
}

func Test_Run_writes_max_subagent_output_when_init_config_flag_is_supplied(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()

	// When
	err := Run(context.Background(), &out, []string{
		"--workspace",
		root,
		"--init-config",
		"--max-subagent-output-bytes",
		"64",
	})
	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	cfg, loadErr := config.LoadWorkspace(context.Background(), root)
	if loadErr != nil {
		t.Fatalf("LoadWorkspace returned error: %v", loadErr)
	}
	if cfg.MaxSubagentOutputBytes != 64 {
		t.Fatalf("MaxSubagentOutputBytes = %d, want 64", cfg.MaxSubagentOutputBytes)
	}
	var body struct {
		MaxSubagentOutputBytes int `json:"max_subagent_output_bytes"`
	}
	if jsonErr := json.Unmarshal(out.Bytes(), &body); jsonErr != nil {
		t.Fatalf("output must be JSON: %v\n%s", jsonErr, out.String())
	}
	if body.MaxSubagentOutputBytes != 64 {
		t.Fatalf("MaxSubagentOutputBytes report = %d, want 64", body.MaxSubagentOutputBytes)
	}
}

func Test_Run_prints_max_subagent_output_config_check(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, ".ceo-harness.json"), []byte(`{"max_subagent_output_bytes":96}`), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	// When
	err := Run(context.Background(), &out, []string{"--workspace", root, "--config-check"})
	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	var body struct {
		MaxSubagentOutputBytes int `json:"max_subagent_output_bytes"`
	}
	if jsonErr := json.Unmarshal(out.Bytes(), &body); jsonErr != nil {
		t.Fatalf("output must be JSON: %v\n%s", jsonErr, out.String())
	}
	if body.MaxSubagentOutputBytes != 96 {
		t.Fatalf("MaxSubagentOutputBytes = %d, want 96", body.MaxSubagentOutputBytes)
	}
}

func assertSubagentOutputLimit(t *testing.T, output []byte, want int) {
	t.Helper()
	var body struct {
		RunManifest struct {
			MaxSubagentOutputBytes int `json:"max_subagent_output_bytes"`
		} `json:"run_manifest"`
		SubagentResults []struct {
			Summary         string `json:"summary"`
			OutputTruncated bool   `json:"output_truncated"`
		} `json:"subagent_results"`
	}
	if jsonErr := json.Unmarshal(output, &body); jsonErr != nil {
		t.Fatalf("output must be JSON: %v\n%s", jsonErr, string(output))
	}
	if body.RunManifest.MaxSubagentOutputBytes != want {
		t.Fatalf("MaxSubagentOutputBytes = %d, want %d", body.RunManifest.MaxSubagentOutputBytes, want)
	}
	if len(body.SubagentResults) == 0 || !body.SubagentResults[0].OutputTruncated {
		t.Fatalf("subagent results = %+v, want truncated first result", body.SubagentResults)
	}
	if strings.Contains(body.SubagentResults[0].Summary, "RAW_TAIL") {
		t.Fatalf("summary kept raw tail: %q", body.SubagentResults[0].Summary)
	}
}

func Test_HelperProcess_cli_subagent_output_model(t *testing.T) {
	if os.Getenv("GO_WANT_CLI_SUBAGENT_OUTPUT_MODEL") != "1" {
		return
	}
	longSummary := strings.Repeat("summary-detail-", 20) + "RAW_TAIL"
	longEvidence := strings.Repeat("evidence-detail-", 20) + "RAW_TAIL"
	longQuestion := strings.Repeat("question-detail-", 20) + "RAW_TAIL"
	os.Stdout.WriteString(`{"summary":"` + longSummary + `","evidence":["` + longEvidence + `"],"questions":["` + longQuestion + `"]}`)
	os.Exit(0)
}
