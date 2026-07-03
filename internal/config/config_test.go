package config

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func Test_LoadWorkspace_returns_empty_config_when_file_is_missing(t *testing.T) {
	// Given
	root := t.TempDir()

	// When
	cfg, err := LoadWorkspace(context.Background(), root)

	// Then
	if err != nil {
		t.Fatalf("LoadWorkspace returned error: %v", err)
	}
	if len(cfg.ModelCommand) != 0 {
		t.Fatalf("ModelCommand length = %d, want 0", len(cfg.ModelCommand))
	}
}

func Test_LoadWorkspace_rejects_empty_model_command_arg(t *testing.T) {
	// Given
	root := t.TempDir()
	content := `{"model_command":["python3",""]}`
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

func Test_LoadWorkspace_rejects_empty_ceo_model_command_arg(t *testing.T) {
	// Given
	root := t.TempDir()
	content := `{"ceo_model_command":["python3",""]}`
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

func Test_LoadWorkspace_rejects_empty_research_command_arg(t *testing.T) {
	// Given
	root := t.TempDir()
	content := `{"research_command":["python3",""]}`
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

func Test_LoadWorkspace_rejects_empty_check_command_arg(t *testing.T) {
	// Given
	root := t.TempDir()
	content := `{"check_command":["go",""]}`
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

func Test_LoadWorkspace_rejects_empty_agent_model_command_arg(t *testing.T) {
	// Given
	root := t.TempDir()
	content := `{"agent_model_commands":{"scanner":["python3",""]}}`
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

func Test_LoadWorkspace_rejects_unknown_agent_provider(t *testing.T) {
	// Given
	root := t.TempDir()
	content := `{"agent_providers":{"scanner":"missing"}}`
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

func Test_CreateWorkspace_writes_config_when_file_is_missing(t *testing.T) {
	// Given
	root := t.TempDir()

	// When
	path, err := CreateWorkspace(context.Background(), root, Config{
		ModelCommand: []string{"python3", "-c", "print(\"ok\")"},
	})

	// Then
	if err != nil {
		t.Fatalf("CreateWorkspace returned error: %v", err)
	}
	if path != filepath.Join(root, ".ceo-harness.json") {
		t.Fatalf("path = %q, want workspace config path", path)
	}
	cfg, err := LoadWorkspace(context.Background(), root)
	if err != nil {
		t.Fatalf("LoadWorkspace returned error: %v", err)
	}
	if len(cfg.ModelCommand) != 3 {
		t.Fatalf("ModelCommand length = %d, want 3", len(cfg.ModelCommand))
	}
}

func Test_CreateWorkspace_refuses_to_overwrite_existing_config(t *testing.T) {
	// Given
	root := t.TempDir()
	path := filepath.Join(root, ".ceo-harness.json")
	if err := os.WriteFile(path, []byte(`{"model_command":["first"]}`), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	// When
	_, err := CreateWorkspace(context.Background(), root, Config{
		ModelCommand: []string{"second"},
	})

	// Then
	if !errors.Is(err, ErrConfigExists) {
		t.Fatalf("error = %v, want ErrConfigExists", err)
	}
	cfg, loadErr := LoadWorkspace(context.Background(), root)
	if loadErr != nil {
		t.Fatalf("LoadWorkspace returned error: %v", loadErr)
	}
	if cfg.ModelCommand[0] != "first" {
		t.Fatalf("ModelCommand[0] = %q, want original config", cfg.ModelCommand[0])
	}
}
