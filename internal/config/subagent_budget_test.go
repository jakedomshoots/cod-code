package config

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func Test_LoadWorkspace_reads_max_subagents_when_config_exists(t *testing.T) {
	// Given
	root := t.TempDir()
	content := `{"max_subagents":7}`
	if err := os.WriteFile(filepath.Join(root, ".ceo-harness.json"), []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	// When
	cfg, err := LoadWorkspace(context.Background(), root)

	// Then
	if err != nil {
		t.Fatalf("LoadWorkspace returned error: %v", err)
	}
	if cfg.MaxSubagents != 7 {
		t.Fatalf("MaxSubagents = %d, want 7", cfg.MaxSubagents)
	}
}

func Test_LoadWorkspace_rejects_max_subagents_above_hard_cap(t *testing.T) {
	// Given
	root := t.TempDir()
	content := `{"max_subagents":9}`
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
