package config

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func Test_LoadWorkspace_reads_context_policy_when_config_exists(t *testing.T) {
	// Given
	root := t.TempDir()
	content := `{"max_context_bytes":1024}`
	if err := os.WriteFile(filepath.Join(root, ".ceo-harness.json"), []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	// When
	cfg, err := LoadWorkspace(context.Background(), root)
	// Then
	if err != nil {
		t.Fatalf("LoadWorkspace returned error: %v", err)
	}
	if cfg.MaxContextBytes != 1024 {
		t.Fatalf("MaxContextBytes = %d, want 1024", cfg.MaxContextBytes)
	}
}

func Test_LoadWorkspace_reads_workspace_brief_excludes(t *testing.T) {
	// Given
	root := t.TempDir()
	content := `{"workspace_brief_excludes":["generated","*.lock"]}`
	if err := os.WriteFile(filepath.Join(root, ".ceo-harness.json"), []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	// When
	cfg, err := LoadWorkspace(context.Background(), root)
	// Then
	if err != nil {
		t.Fatalf("LoadWorkspace returned error: %v", err)
	}
	if len(cfg.WorkspaceBriefExcludes) != 2 || cfg.WorkspaceBriefExcludes[0] != "generated" || cfg.WorkspaceBriefExcludes[1] != "*.lock" {
		t.Fatalf("WorkspaceBriefExcludes = %#v, want generated and *.lock", cfg.WorkspaceBriefExcludes)
	}
}

func Test_LoadWorkspace_rejects_escaping_workspace_brief_exclude(t *testing.T) {
	// Given
	root := t.TempDir()
	content := `{"workspace_brief_excludes":["../secrets"]}`
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

func Test_LoadWorkspace_rejects_negative_context_policy(t *testing.T) {
	// Given
	root := t.TempDir()
	content := `{"max_context_bytes":-1}`
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
