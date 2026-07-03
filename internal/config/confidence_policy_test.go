package config

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func Test_LoadWorkspace_reads_min_subagent_confidence_when_config_exists(t *testing.T) {
	// Given
	root := t.TempDir()
	content := `{"min_subagent_confidence":0.6}`
	if err := os.WriteFile(filepath.Join(root, ".ceo-harness.json"), []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	// When
	cfg, err := LoadWorkspace(context.Background(), root)
	// Then
	if err != nil {
		t.Fatalf("LoadWorkspace returned error: %v", err)
	}
	if cfg.MinSubagentConfidence != 0.6 {
		t.Fatalf("MinSubagentConfidence = %v, want 0.6", cfg.MinSubagentConfidence)
	}
}

func Test_LoadWorkspace_rejects_invalid_min_subagent_confidence(t *testing.T) {
	// Given
	root := t.TempDir()
	content := `{"min_subagent_confidence":1.2}`
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
