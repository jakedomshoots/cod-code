package config

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func Test_LoadWorkspace_reads_no_progress_stop_when_config_exists(t *testing.T) {
	// Given
	root := t.TempDir()
	content := `{"no_progress_stop":2}`
	if err := os.WriteFile(filepath.Join(root, ".ceo-harness.json"), []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	// When
	cfg, err := LoadWorkspace(context.Background(), root)

	// Then
	if err != nil {
		t.Fatalf("LoadWorkspace returned error: %v", err)
	}
	if cfg.NoProgressStop != 2 {
		t.Fatalf("NoProgressStop = %d, want 2", cfg.NoProgressStop)
	}
}

func Test_LoadWorkspace_rejects_negative_no_progress_stop(t *testing.T) {
	// Given
	root := t.TempDir()
	content := `{"no_progress_stop":-1}`
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
