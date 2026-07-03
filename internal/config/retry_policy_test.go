package config

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func Test_LoadWorkspace_reads_retry_policy_when_config_exists(t *testing.T) {
	// Given
	root := t.TempDir()
	content := `{"check_attempts":2,"check_backoff_ms":10,"subagent_attempts":3,"subagent_backoff_ms":20,"ceo_revision_attempts":1}`
	if err := os.WriteFile(filepath.Join(root, ".ceo-harness.json"), []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	// When
	cfg, err := LoadWorkspace(context.Background(), root)
	// Then
	if err != nil {
		t.Fatalf("LoadWorkspace returned error: %v", err)
	}
	if cfg.CheckAttempts != 2 || cfg.CheckBackoffMS != 10 {
		t.Fatalf("check retry policy = attempts %d backoff %d, want 2 and 10", cfg.CheckAttempts, cfg.CheckBackoffMS)
	}
	if cfg.SubagentAttempts != 3 || cfg.SubagentBackoffMS != 20 {
		t.Fatalf("subagent retry policy = attempts %d backoff %d, want 3 and 20", cfg.SubagentAttempts, cfg.SubagentBackoffMS)
	}
	if cfg.CEORevisionAttempts != 1 {
		t.Fatalf("CEORevisionAttempts = %d, want 1", cfg.CEORevisionAttempts)
	}
}

func Test_LoadWorkspace_rejects_negative_retry_policy_values(t *testing.T) {
	// Given
	root := t.TempDir()
	content := `{"check_backoff_ms":-1}`
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

func Test_LoadWorkspace_rejects_negative_ceo_revision_attempts(t *testing.T) {
	// Given
	root := t.TempDir()
	content := `{"ceo_revision_attempts":-1}`
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
