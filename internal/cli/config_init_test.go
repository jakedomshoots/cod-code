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

func Test_Run_writes_workspace_config_when_init_config_flag_is_supplied(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	args := []string{
		"--workspace",
		root,
		"--init-config",
		"--model-command",
		"python3",
		"-c",
		"print(\"ok\")",
		"--",
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
	if len(cfg.ModelCommand) != 3 {
		t.Fatalf("ModelCommand length = %d, want 3", len(cfg.ModelCommand))
	}
	var body struct {
		ConfigPath       string `json:"config_path"`
		Created          bool   `json:"created"`
		ModelCommandArgc int    `json:"model_command_argc"`
	}
	if jsonErr := json.Unmarshal(out.Bytes(), &body); jsonErr != nil {
		t.Fatalf("output must be JSON: %v\n%s", jsonErr, out.String())
	}
	if body.ConfigPath != filepath.Join(root, ".ceo-harness.json") {
		t.Fatalf("ConfigPath = %q, want workspace config path", body.ConfigPath)
	}
	if !body.Created || body.ModelCommandArgc != 3 {
		t.Fatalf("created = %v argc = %d, want true 3", body.Created, body.ModelCommandArgc)
	}
}

func Test_Run_refuses_to_overwrite_workspace_config_when_init_config_flag_is_supplied(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	path := filepath.Join(root, ".ceo-harness.json")
	if err := os.WriteFile(path, []byte(`{"model_command":["first"]}`), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	// When
	err := Run(context.Background(), &out, []string{"--workspace", root, "--init-config"})

	// Then
	if !errors.Is(err, config.ErrConfigExists) {
		t.Fatalf("error = %v, want ErrConfigExists", err)
	}
}

func Test_Run_writes_retry_policy_when_init_config_retry_flags_are_supplied(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	args := []string{
		"--workspace",
		root,
		"--init-config",
		"--check-attempts",
		"2",
		"--check-backoff-ms",
		"10",
		"--ceo-revision-attempts",
		"1",
		"--subagent-concurrency",
		"2",
		"--subagent-attempts",
		"3",
		"--subagent-backoff-ms",
		"20",
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
	if cfg.CheckAttempts != 2 || cfg.CheckBackoffMS != 10 || cfg.CEORevisionAttempts != 1 || cfg.SubagentConcurrency != 2 || cfg.SubagentAttempts != 3 || cfg.SubagentBackoffMS != 20 {
		t.Fatalf("retry policy = %#v, want init-config retry values", cfg)
	}
	var body struct {
		CheckAttempts       int `json:"check_attempts"`
		CheckBackoffMS      int `json:"check_backoff_ms"`
		CEORevisionAttempts int `json:"ceo_revision_attempts"`
		SubagentConcurrency int `json:"subagent_concurrency"`
		SubagentAttempts    int `json:"subagent_attempts"`
		SubagentBackoffMS   int `json:"subagent_backoff_ms"`
	}
	if jsonErr := json.Unmarshal(out.Bytes(), &body); jsonErr != nil {
		t.Fatalf("output must be JSON: %v\n%s", jsonErr, out.String())
	}
	if body.CheckAttempts != 2 || body.CheckBackoffMS != 10 || body.CEORevisionAttempts != 1 || body.SubagentConcurrency != 2 || body.SubagentAttempts != 3 || body.SubagentBackoffMS != 20 {
		t.Fatalf("init report retry policy = %#v, want retry values", body)
	}
}

func Test_Run_writes_check_command_when_init_config_check_flag_is_supplied(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	args := []string{
		"--workspace",
		root,
		"--init-config",
		"--check",
		"go",
		"test",
		"./...",
		"--",
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
	commands := cfg.CheckCommandList()
	if len(commands) != 1 || len(commands[0]) != 3 || commands[0][1] != "test" {
		t.Fatalf("check commands = %#v, want init check command", commands)
	}
	var body struct {
		CheckCommandArgc int `json:"check_command_argc"`
	}
	if jsonErr := json.Unmarshal(out.Bytes(), &body); jsonErr != nil {
		t.Fatalf("output must be JSON: %v\n%s", jsonErr, out.String())
	}
	if body.CheckCommandArgc != 3 {
		t.Fatalf("CheckCommandArgc = %d, want 3", body.CheckCommandArgc)
	}
}

func Test_Run_writes_research_command_when_init_config_research_flag_is_supplied(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	args := []string{
		"--workspace",
		root,
		"--init-config",
		"--research-command",
		"python3",
		"-c",
		"print(\"research\")",
		"--",
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
	if len(cfg.ResearchCommand) != 3 || cfg.ResearchCommand[0] != "python3" {
		t.Fatalf("ResearchCommand = %#v, want init research command", cfg.ResearchCommand)
	}
	var body struct {
		ResearchCommandArgc int `json:"research_command_argc"`
	}
	if jsonErr := json.Unmarshal(out.Bytes(), &body); jsonErr != nil {
		t.Fatalf("output must be JSON: %v\n%s", jsonErr, out.String())
	}
	if body.ResearchCommandArgc != 3 {
		t.Fatalf("ResearchCommandArgc = %d, want 3", body.ResearchCommandArgc)
	}
}
