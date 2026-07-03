package config

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func Test_LoadWorkspace_reads_job_policy_when_config_exists(t *testing.T) {
	// Given
	root := t.TempDir()
	content := `{"job_timeout_ms":2500}`
	if err := os.WriteFile(filepath.Join(root, ".ceo-harness.json"), []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	// When
	cfg, err := LoadWorkspace(context.Background(), root)
	// Then
	if err != nil {
		t.Fatalf("LoadWorkspace returned error: %v", err)
	}
	if cfg.JobTimeoutMS != 2500 {
		t.Fatalf("JobTimeoutMS = %d, want 2500", cfg.JobTimeoutMS)
	}
}

func Test_LoadWorkspace_rejects_negative_job_policy(t *testing.T) {
	// Given
	root := t.TempDir()
	content := `{"job_timeout_ms":-1}`
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
