package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"ceoharness/internal/config"
)

func Test_Run_writes_model_command_timeout_when_init_config_flag_is_supplied(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	args := []string{
		"--workspace",
		root,
		"--init-config",
		"--model-command-timeout-ms",
		"2500",
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
	if cfg.ModelCommandTimeoutMS != 2500 {
		t.Fatalf("ModelCommandTimeoutMS = %d, want 2500", cfg.ModelCommandTimeoutMS)
	}
	var body struct {
		ModelCommandTimeoutMS int `json:"model_command_timeout_ms"`
	}
	if jsonErr := json.Unmarshal(out.Bytes(), &body); jsonErr != nil {
		t.Fatalf("output must be JSON: %v\n%s", jsonErr, out.String())
	}
	if body.ModelCommandTimeoutMS != 2500 {
		t.Fatalf("report ModelCommandTimeoutMS = %d, want 2500", body.ModelCommandTimeoutMS)
	}
}

func Test_Run_prints_model_command_timeout_config_check_when_config_check_flag_is_supplied(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	content := `{"model_command_timeout_ms":2500}`
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
		ModelCommandTimeoutMS int `json:"model_command_timeout_ms"`
	}
	if jsonErr := json.Unmarshal(out.Bytes(), &body); jsonErr != nil {
		t.Fatalf("output must be JSON: %v\n%s", jsonErr, out.String())
	}
	if body.ModelCommandTimeoutMS != 2500 {
		t.Fatalf("ModelCommandTimeoutMS = %d, want 2500", body.ModelCommandTimeoutMS)
	}
}

func Test_Run_writes_tool_command_timeout_when_init_config_flag_is_supplied(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	args := []string{
		"--workspace",
		root,
		"--init-config",
		"--tool-command-timeout-ms",
		"2500",
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
	if cfg.ToolCommandTimeoutMS != 2500 {
		t.Fatalf("ToolCommandTimeoutMS = %d, want 2500", cfg.ToolCommandTimeoutMS)
	}
	var body struct {
		ToolCommandTimeoutMS int `json:"tool_command_timeout_ms"`
	}
	if jsonErr := json.Unmarshal(out.Bytes(), &body); jsonErr != nil {
		t.Fatalf("output must be JSON: %v\n%s", jsonErr, out.String())
	}
	if body.ToolCommandTimeoutMS != 2500 {
		t.Fatalf("report ToolCommandTimeoutMS = %d, want 2500", body.ToolCommandTimeoutMS)
	}
}

func Test_Run_prints_tool_command_timeout_config_check_when_config_check_flag_is_supplied(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	content := `{"tool_command_timeout_ms":2500}`
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
		ToolCommandTimeoutMS int `json:"tool_command_timeout_ms"`
	}
	if jsonErr := json.Unmarshal(out.Bytes(), &body); jsonErr != nil {
		t.Fatalf("output must be JSON: %v\n%s", jsonErr, out.String())
	}
	if body.ToolCommandTimeoutMS != 2500 {
		t.Fatalf("ToolCommandTimeoutMS = %d, want 2500", body.ToolCommandTimeoutMS)
	}
}
