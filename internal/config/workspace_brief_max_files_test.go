package config

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func Test_LoadWorkspace_reads_workspace_brief_max_files(t *testing.T) {
	// Given
	root := t.TempDir()
	content := `{"workspace_brief_max_files":7}`
	if err := os.WriteFile(filepath.Join(root, ".ceo-harness.json"), []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	// When
	cfg, err := LoadWorkspace(context.Background(), root)

	// Then
	if err != nil {
		t.Fatalf("LoadWorkspace returned error: %v", err)
	}
	if cfg.WorkspaceBriefMaxFiles != 7 {
		t.Fatalf("WorkspaceBriefMaxFiles = %d, want 7", cfg.WorkspaceBriefMaxFiles)
	}
}

func Test_LoadWorkspace_rejects_negative_workspace_brief_max_files(t *testing.T) {
	// Given
	root := t.TempDir()
	content := `{"workspace_brief_max_files":-1}`
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
