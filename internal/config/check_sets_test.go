package config

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func Test_LoadWorkspace_resolves_default_check_set_when_set_exists(t *testing.T) {
	// Given
	root := t.TempDir()
	content := `{"default_check_set":"quick","check_sets":{"quick":[["go","test","./..."]],"full":[["go","test","./..."],["go","vet","./..."]]}}`
	if err := os.WriteFile(filepath.Join(root, ".ceo-harness.json"), []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	// When
	cfg, err := LoadWorkspace(context.Background(), root)

	// Then
	if err != nil {
		t.Fatalf("LoadWorkspace returned error: %v", err)
	}
	commands, ok := cfg.CheckCommandsForSet(cfg.DefaultCheckSet)
	if !ok {
		t.Fatalf("default check set %q was not found", cfg.DefaultCheckSet)
	}
	if len(commands) != 1 || len(commands[0]) != 3 || commands[0][1] != "test" {
		t.Fatalf("commands = %q, want quick go test command", commands)
	}
}

func Test_LoadWorkspace_rejects_unknown_default_check_set(t *testing.T) {
	// Given
	root := t.TempDir()
	content := `{"default_check_set":"missing","check_sets":{"quick":[["go","test","./..."]]}}`
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

func Test_LoadWorkspace_rejects_empty_check_set_command_arg(t *testing.T) {
	// Given
	root := t.TempDir()
	content := `{"check_sets":{"quick":[["go",""]]}}`
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
