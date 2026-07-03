package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"ceoharness/internal/config"
)

func Test_Run_limits_workspace_brief_files_when_flag_is_supplied(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	writeWorkspaceBriefFiles(t, root)

	// When
	err := Run(context.Background(), &out, []string{"--workspace", root, "--workspace-brief-max-files", "1", "Fix", "workspace", "bug"})
	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	var body struct {
		WorkspaceBrief *struct {
			FileCount int `json:"file_count"`
			Files     []struct {
				Path string `json:"path"`
			} `json:"files"`
			Truncated bool `json:"truncated"`
		} `json:"workspace_brief"`
	}
	if jsonErr := json.Unmarshal(out.Bytes(), &body); jsonErr != nil {
		t.Fatalf("output must be JSON: %v\n%s", jsonErr, out.String())
	}
	if body.WorkspaceBrief == nil {
		t.Fatal("workspace_brief is missing")
	}
	if body.WorkspaceBrief.FileCount != 3 || len(body.WorkspaceBrief.Files) != 1 || !body.WorkspaceBrief.Truncated {
		t.Fatalf("workspace_brief = %+v, want three files counted, one shown, truncated", body.WorkspaceBrief)
	}
}

func Test_Run_writes_workspace_brief_max_files_when_init_config_flag_is_supplied(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	args := []string{"--workspace", root, "--init-config", "--workspace-brief-max-files", "7"}

	// When
	err := Run(context.Background(), &out, args)
	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	cfg, loadErr := config.LoadWorkspace(context.Background(), root)
	if loadErr != nil {
		t.Fatalf("LoadWorkspace returned error: %v", loadErr)
	}
	if cfg.WorkspaceBriefMaxFiles != 7 {
		t.Fatalf("WorkspaceBriefMaxFiles = %d, want 7", cfg.WorkspaceBriefMaxFiles)
	}
	var body struct {
		WorkspaceBriefMaxFiles int `json:"workspace_brief_max_files"`
	}
	if jsonErr := json.Unmarshal(out.Bytes(), &body); jsonErr != nil {
		t.Fatalf("output must be JSON: %v\n%s", jsonErr, out.String())
	}
	if body.WorkspaceBriefMaxFiles != 7 {
		t.Fatalf("report WorkspaceBriefMaxFiles = %d, want 7", body.WorkspaceBriefMaxFiles)
	}
}

func Test_Run_prints_workspace_brief_max_files_config_check(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, ".ceo-harness.json"), []byte(`{"workspace_brief_max_files":7}`), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	// When
	err := Run(context.Background(), &out, []string{"--workspace", root, "--config-check"})
	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	var body struct {
		WorkspaceBriefMaxFiles int `json:"workspace_brief_max_files"`
	}
	if jsonErr := json.Unmarshal(out.Bytes(), &body); jsonErr != nil {
		t.Fatalf("output must be JSON: %v\n%s", jsonErr, out.String())
	}
	if body.WorkspaceBriefMaxFiles != 7 {
		t.Fatalf("WorkspaceBriefMaxFiles = %d, want 7", body.WorkspaceBriefMaxFiles)
	}
}

func writeWorkspaceBriefFiles(t *testing.T, root string) {
	t.Helper()
	for _, name := range []string{"app.go", "README.md", "notes.txt"} {
		if err := os.WriteFile(filepath.Join(root, name), []byte("content"), 0o644); err != nil {
			t.Fatalf("write %s fixture: %v", name, err)
		}
	}
}
