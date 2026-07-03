package config

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func Test_LoadWorkspace_reads_check_commands_when_config_exists(t *testing.T) {
	// Given
	root := t.TempDir()
	content := `{"check_commands":[["go","test","./..."],["go","vet","./..."]]}`
	if err := os.WriteFile(filepath.Join(root, ".ceo-harness.json"), []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	// When
	cfg, err := LoadWorkspace(context.Background(), root)

	// Then
	if err != nil {
		t.Fatalf("LoadWorkspace returned error: %v", err)
	}
	commands := cfg.CheckCommandList()
	if len(commands) != 2 {
		t.Fatalf("CheckCommandList length = %d, want 2", len(commands))
	}
	if commands[1][1] != "vet" {
		t.Fatalf("second command = %q, want go vet", commands[1])
	}
}

func Test_LoadWorkspace_rejects_empty_check_commands_arg(t *testing.T) {
	// Given
	root := t.TempDir()
	content := `{"check_commands":[["go","test"],["go",""]]}`
	if err := os.WriteFile(filepath.Join(root, ".ceo-harness.json"), []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	// When
	_, err := LoadWorkspace(context.Background(), root)

	// Then
	if !errors.Is(err, ErrInvalidConfig) {
		t.Fatalf("error = %v, want ErrInvalidConfig", err)
	}
}
