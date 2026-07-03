package config

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func Test_LoadWorkspace_reads_ceo_provider_when_provider_exists(t *testing.T) {
	// Given
	root := t.TempDir()
	content := `{"providers":{"main":{"model_command":["echo","review"]}},"ceo_provider":"main"}`
	if err := os.WriteFile(filepath.Join(root, ".ceo-harness.json"), []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	// When
	cfg, err := LoadWorkspace(context.Background(), root)

	// Then
	if err != nil {
		t.Fatalf("LoadWorkspace returned error: %v", err)
	}
	if cfg.CEOProvider != "main" {
		t.Fatalf("CEOProvider = %q, want main", cfg.CEOProvider)
	}
}

func Test_LoadWorkspace_rejects_unknown_ceo_provider(t *testing.T) {
	// Given
	root := t.TempDir()
	content := `{"providers":{"main":{"model_command":["echo","review"]}},"ceo_provider":"missing"}`
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
