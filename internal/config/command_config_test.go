package config

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func Test_LoadWorkspace_reads_model_command_when_config_exists(t *testing.T) {
	// Given
	root := t.TempDir()
	content := `{"model_command":["python3","-c","print(\"ok\")"]}`
	if err := os.WriteFile(filepath.Join(root, ".ceo-harness.json"), []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	// When
	cfg, err := LoadWorkspace(context.Background(), root)

	// Then
	if err != nil {
		t.Fatalf("LoadWorkspace returned error: %v", err)
	}
	if len(cfg.ModelCommand) != 3 {
		t.Fatalf("ModelCommand length = %d, want 3", len(cfg.ModelCommand))
	}
	if cfg.ModelCommand[0] != "python3" {
		t.Fatalf("ModelCommand[0] = %q, want python3", cfg.ModelCommand[0])
	}
}

func Test_LoadWorkspace_reads_ceo_model_command_when_config_exists(t *testing.T) {
	// Given
	root := t.TempDir()
	content := `{"ceo_model_command":["python3","-c","print(\"review\")"]}`
	if err := os.WriteFile(filepath.Join(root, ".ceo-harness.json"), []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	// When
	cfg, err := LoadWorkspace(context.Background(), root)

	// Then
	if err != nil {
		t.Fatalf("LoadWorkspace returned error: %v", err)
	}
	if len(cfg.CEOModelCommand) != 3 {
		t.Fatalf("CEOModelCommand length = %d, want 3", len(cfg.CEOModelCommand))
	}
	if cfg.CEOModelCommand[0] != "python3" {
		t.Fatalf("CEOModelCommand[0] = %q, want python3", cfg.CEOModelCommand[0])
	}
}

func Test_LoadWorkspace_reads_research_command_when_config_exists(t *testing.T) {
	// Given
	root := t.TempDir()
	content := `{"research_command":["python3","-c","print(\"research\")"]}`
	if err := os.WriteFile(filepath.Join(root, ".ceo-harness.json"), []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	// When
	cfg, err := LoadWorkspace(context.Background(), root)

	// Then
	if err != nil {
		t.Fatalf("LoadWorkspace returned error: %v", err)
	}
	if len(cfg.ResearchCommand) != 3 {
		t.Fatalf("ResearchCommand length = %d, want 3", len(cfg.ResearchCommand))
	}
	if cfg.ResearchCommand[0] != "python3" {
		t.Fatalf("ResearchCommand[0] = %q, want python3", cfg.ResearchCommand[0])
	}
}

func Test_LoadWorkspace_reads_agent_model_commands_when_config_exists(t *testing.T) {
	// Given
	root := t.TempDir()
	content := `{"agent_model_commands":{"scanner":["python3","-c","print(\"scan\")"]}}`
	if err := os.WriteFile(filepath.Join(root, ".ceo-harness.json"), []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	// When
	cfg, err := LoadWorkspace(context.Background(), root)

	// Then
	if err != nil {
		t.Fatalf("LoadWorkspace returned error: %v", err)
	}
	if len(cfg.AgentModelCommands["scanner"]) != 3 {
		t.Fatalf("scanner command length = %d, want 3", len(cfg.AgentModelCommands["scanner"]))
	}
	if cfg.AgentModelCommands["scanner"][0] != "python3" {
		t.Fatalf("scanner command = %q, want python3", cfg.AgentModelCommands["scanner"][0])
	}
}

func Test_LoadWorkspace_reads_check_command_when_config_exists(t *testing.T) {
	// Given
	root := t.TempDir()
	content := `{"check_command":["go","test","./..."]}`
	if err := os.WriteFile(filepath.Join(root, ".ceo-harness.json"), []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	// When
	cfg, err := LoadWorkspace(context.Background(), root)

	// Then
	if err != nil {
		t.Fatalf("LoadWorkspace returned error: %v", err)
	}
	if len(cfg.CheckCommand) != 3 {
		t.Fatalf("CheckCommand length = %d, want 3", len(cfg.CheckCommand))
	}
	if cfg.CheckCommand[0] != "go" {
		t.Fatalf("CheckCommand[0] = %q, want go", cfg.CheckCommand[0])
	}
}

func Test_LoadWorkspace_resolves_agent_provider_when_provider_exists(t *testing.T) {
	// Given
	root := t.TempDir()
	content := `{"providers":{"fast":{"model_command":["python3","-c","print(\"fast\")"]}},"agent_providers":{"scanner":"fast"}}`
	if err := os.WriteFile(filepath.Join(root, ".ceo-harness.json"), []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	// When
	cfg, err := LoadWorkspace(context.Background(), root)

	// Then
	if err != nil {
		t.Fatalf("LoadWorkspace returned error: %v", err)
	}
	commands := cfg.AgentCommands()
	if len(commands["scanner"]) != 3 {
		t.Fatalf("scanner command length = %d, want 3", len(commands["scanner"]))
	}
	if commands["scanner"][0] != "python3" {
		t.Fatalf("scanner command = %q, want python3", commands["scanner"][0])
	}
}
