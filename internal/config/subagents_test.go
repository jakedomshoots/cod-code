package config

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func Test_LoadWorkspace_reads_subagents_when_config_exists(t *testing.T) {
	// Given
	root := t.TempDir()
	content := `{"subagents":[{"name":"planner","role":"break down work"},{"name":"security","role":"review auth risks"}]}`
	if err := os.WriteFile(filepath.Join(root, ".ceo-harness.json"), []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	// When
	cfg, err := LoadWorkspace(context.Background(), root)
	// Then
	if err != nil {
		t.Fatalf("LoadWorkspace returned error: %v", err)
	}
	if len(cfg.Subagents) != 2 {
		t.Fatalf("Subagents length = %d, want 2", len(cfg.Subagents))
	}
	if cfg.Subagents[0].Name != "planner" || cfg.Subagents[1].Role != "review auth risks" {
		t.Fatalf("Subagents = %#v, want configured subagents", cfg.Subagents)
	}
}

func Test_LoadWorkspace_rejects_empty_subagent_role(t *testing.T) {
	// Given
	root := t.TempDir()
	content := `{"subagents":[{"name":"planner","role":""}]}`
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

func Test_LoadWorkspace_rejects_duplicate_subagent_names(t *testing.T) {
	// Given
	root := t.TempDir()
	content := `{"subagents":[{"name":"planner","role":"plan"},{"name":"planner","role":"review"}]}`
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
