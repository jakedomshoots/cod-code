package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"ceoharness/internal/config"
)

func Test_Run_prints_config_check_when_config_check_flag_is_supplied(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	content := `{"model_command":["python3","-c","print(\"ok\")"]}`
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
		WorkspaceDir        string `json:"workspace_dir"`
		ConfigPath          string `json:"config_path"`
		ModelCommandSource  string `json:"model_command_source"`
		ModelCommandArgc    int    `json:"model_command_argc"`
		ModelCommandPresent bool   `json:"model_command_present"`
	}
	if jsonErr := json.Unmarshal(out.Bytes(), &body); jsonErr != nil {
		t.Fatalf("output must be JSON: %v\n%s", jsonErr, out.String())
	}
	if body.WorkspaceDir != root {
		t.Fatalf("WorkspaceDir = %q, want workspace root", body.WorkspaceDir)
	}
	if body.ConfigPath != filepath.Join(root, ".ceo-harness.json") {
		t.Fatalf("ConfigPath = %q, want workspace config path", body.ConfigPath)
	}
	if body.ModelCommandSource != "workspace" {
		t.Fatalf("ModelCommandSource = %q, want workspace", body.ModelCommandSource)
	}
	if body.ModelCommandArgc != 3 || !body.ModelCommandPresent {
		t.Fatalf("model command metadata = argc %d present %v, want 3 true", body.ModelCommandArgc, body.ModelCommandPresent)
	}
}

func Test_Run_prints_ceo_model_command_config_check_when_config_check_flag_is_supplied(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	content := `{"ceo_model_command":["python3","-c","print(\"review\")"]}`
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
		CEOModelCommandSource  string `json:"ceo_model_command_source"`
		CEOModelCommandArgc    int    `json:"ceo_model_command_argc"`
		CEOModelCommandPresent bool   `json:"ceo_model_command_present"`
	}
	if jsonErr := json.Unmarshal(out.Bytes(), &body); jsonErr != nil {
		t.Fatalf("output must be JSON: %v\n%s", jsonErr, out.String())
	}
	if body.CEOModelCommandSource != "workspace" {
		t.Fatalf("CEOModelCommandSource = %q, want workspace", body.CEOModelCommandSource)
	}
	if body.CEOModelCommandArgc != 3 || !body.CEOModelCommandPresent {
		t.Fatalf("CEO model command metadata = argc %d present %v, want 3 true", body.CEOModelCommandArgc, body.CEOModelCommandPresent)
	}
}

func Test_Run_prints_research_command_config_check_when_config_check_flag_is_supplied(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	content := `{"research_command":["python3","-c","print(\"research\")"]}`
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
		ResearchCommandSource  string `json:"research_command_source"`
		ResearchCommandArgc    int    `json:"research_command_argc"`
		ResearchCommandPresent bool   `json:"research_command_present"`
	}
	if jsonErr := json.Unmarshal(out.Bytes(), &body); jsonErr != nil {
		t.Fatalf("output must be JSON: %v\n%s", jsonErr, out.String())
	}
	if body.ResearchCommandSource != "workspace" {
		t.Fatalf("ResearchCommandSource = %q, want workspace", body.ResearchCommandSource)
	}
	if body.ResearchCommandArgc != 3 || !body.ResearchCommandPresent {
		t.Fatalf("research command metadata = argc %d present %v, want 3 true", body.ResearchCommandArgc, body.ResearchCommandPresent)
	}
}

func Test_Run_prints_ceo_revision_attempts_config_check_when_config_check_flag_is_supplied(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	content := `{"ceo_revision_attempts":2,"subagent_concurrency":3}`
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
		CEORevisionAttempts int `json:"ceo_revision_attempts"`
		SubagentConcurrency int `json:"subagent_concurrency"`
	}
	if jsonErr := json.Unmarshal(out.Bytes(), &body); jsonErr != nil {
		t.Fatalf("output must be JSON: %v\n%s", jsonErr, out.String())
	}
	if body.CEORevisionAttempts != 2 {
		t.Fatalf("CEORevisionAttempts = %d, want 2", body.CEORevisionAttempts)
	}
	if body.SubagentConcurrency != 3 {
		t.Fatalf("SubagentConcurrency = %d, want 3", body.SubagentConcurrency)
	}
}

func Test_Run_returns_config_error_when_config_check_finds_invalid_model_command(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	content := `{"model_command":["python3",""]}`
	if err := os.WriteFile(filepath.Join(root, ".ceo-harness.json"), []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	// When
	err := Run(context.Background(), &out, []string{"--workspace", root, "--config-check"})

	// Then
	if !errors.Is(err, config.ErrInvalidConfig) {
		t.Fatalf("error = %v, want ErrInvalidConfig", err)
	}
}
