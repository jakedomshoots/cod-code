package config

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func Test_LoadWorkspace_reads_subagent_output_policy_when_config_exists(t *testing.T) {
	// Given
	root := t.TempDir()
	content := `{"max_subagent_output_bytes":128}`
	if err := os.WriteFile(filepath.Join(root, ".ceo-harness.json"), []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	// When
	cfg, err := LoadWorkspace(context.Background(), root)
	// Then
	if err != nil {
		t.Fatalf("LoadWorkspace returned error: %v", err)
	}
	if cfg.MaxSubagentOutputBytes != 128 {
		t.Fatalf("MaxSubagentOutputBytes = %d, want 128", cfg.MaxSubagentOutputBytes)
	}
}

func Test_LoadWorkspace_rejects_negative_subagent_output_policy(t *testing.T) {
	// Given
	root := t.TempDir()
	content := `{"max_subagent_output_bytes":-1}`
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
