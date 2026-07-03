package config

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func Test_LoadWorkspace_reads_max_ceo_iterations_when_config_exists(t *testing.T) {
	// Given
	root := t.TempDir()
	content := `{"max_ceo_iterations":6}`
	if err := os.WriteFile(filepath.Join(root, ".ceo-harness.json"), []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	// When
	cfg, err := LoadWorkspace(context.Background(), root)

	// Then
	if err != nil {
		t.Fatalf("LoadWorkspace returned error: %v", err)
	}
	if cfg.MaxCEOIterations != 6 {
		t.Fatalf("MaxCEOIterations = %d, want 6", cfg.MaxCEOIterations)
	}
}

func Test_LoadWorkspace_rejects_negative_max_ceo_iterations(t *testing.T) {
	// Given
	root := t.TempDir()
	content := `{"max_ceo_iterations":-1}`
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
