package config

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func Test_LoadWorkspace_resolves_auto_check_set_when_task_matches_keyword(t *testing.T) {
	// Given
	root := t.TempDir()
	content := `{"check_sets":{"quick":[["go","test","./..."]],"full":[["go","test","./..."],["go","vet","./..."]]},"auto_check_sets":[{"check_set":"full","keywords":["auth","security"]}]}`
	if err := os.WriteFile(filepath.Join(root, ".ceo-harness.json"), []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	// When
	cfg, err := LoadWorkspace(context.Background(), root)
	// Then
	if err != nil {
		t.Fatalf("LoadWorkspace returned error: %v", err)
	}
	commands, ok := cfg.AutoCheckCommandsForTask("Fix auth callback")
	if !ok {
		t.Fatal("expected auto check set match")
	}
	if len(commands) != 2 || commands[1][1] != "vet" {
		t.Fatalf("commands = %q, want full check set", commands)
	}
}

func Test_LoadWorkspace_rejects_auto_check_set_with_unknown_check_set(t *testing.T) {
	// Given
	root := t.TempDir()
	content := `{"check_sets":{"quick":[["go","test","./..."]]},"auto_check_sets":[{"check_set":"full","keywords":["auth"]}]}`
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

func Test_LoadWorkspace_rejects_auto_check_set_with_empty_keyword(t *testing.T) {
	// Given
	root := t.TempDir()
	content := `{"check_sets":{"quick":[["go","test","./..."]]},"auto_check_sets":[{"check_set":"quick","keywords":[""]}]}`
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
