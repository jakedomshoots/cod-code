package config

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func Test_LoadWorkspace_accepts_custom_subagent_provider_when_provider_exists(t *testing.T) {
	// Given
	root := t.TempDir()
	content := `{"providers":{"premium":{"model_command":["echo","premium"]}},"subagents":[{"name":"ux_reviewer","role":"review UX","provider":"premium"}]}`
	if err := os.WriteFile(filepath.Join(root, WorkspaceConfigName), []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	// When
	cfg, err := LoadWorkspace(context.Background(), root)

	// Then
	if err != nil {
		t.Fatalf("LoadWorkspace returned error: %v", err)
	}
	if cfg.Subagents[0].ProviderName != "premium" {
		t.Fatalf("ProviderName = %q, want premium", cfg.Subagents[0].ProviderName)
	}
}

func Test_LoadWorkspace_rejects_custom_subagent_provider_when_provider_is_missing(t *testing.T) {
	// Given
	root := t.TempDir()
	content := `{"subagents":[{"name":"ux_reviewer","role":"review UX","provider":"missing"}]}`
	if err := os.WriteFile(filepath.Join(root, WorkspaceConfigName), []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	// When
	_, err := LoadWorkspace(context.Background(), root)

	// Then
	if err == nil {
		t.Fatal("expected missing provider error")
	}
}
