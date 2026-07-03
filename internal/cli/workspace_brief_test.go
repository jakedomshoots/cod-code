package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func Test_Run_prints_workspace_brief_when_workspace_is_set(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "app.go"), []byte("package main"), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	// When
	err := Run(context.Background(), &out, []string{"--workspace", root, "Fix", "workspace", "bug"})
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
		} `json:"workspace_brief"`
	}
	if err := json.Unmarshal(out.Bytes(), &body); err != nil {
		t.Fatalf("output must be JSON: %v\n%s", err, out.String())
	}
	if body.WorkspaceBrief == nil {
		t.Fatal("workspace_brief is missing")
	}
	if body.WorkspaceBrief.FileCount != 1 || body.WorkspaceBrief.Files[0].Path != "app.go" {
		t.Fatalf("workspace_brief = %+v, want app.go", body.WorkspaceBrief)
	}
}

func Test_Run_uses_workspace_brief_excludes_from_config_and_flag(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, ".ceo-harness.json"), []byte(`{"workspace_brief_excludes":["generated"]}`), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "app.go"), []byte("package main"), 0o644); err != nil {
		t.Fatalf("write app fixture: %v", err)
	}
	if err := os.Mkdir(filepath.Join(root, "generated"), 0o755); err != nil {
		t.Fatalf("mkdir generated: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "generated", "client.go"), []byte("package generated"), 0o644); err != nil {
		t.Fatalf("write generated fixture: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "package.lock"), []byte("lock"), 0o644); err != nil {
		t.Fatalf("write lock fixture: %v", err)
	}

	// When
	err := Run(context.Background(), &out, []string{"--workspace", root, "--workspace-brief-exclude", "*.lock", "Fix", "workspace", "bug"})
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
		} `json:"workspace_brief"`
	}
	if err := json.Unmarshal(out.Bytes(), &body); err != nil {
		t.Fatalf("output must be JSON: %v\n%s", err, out.String())
	}
	if body.WorkspaceBrief == nil {
		t.Fatal("workspace_brief is missing")
	}
	if body.WorkspaceBrief.FileCount != 1 || len(body.WorkspaceBrief.Files) != 1 || body.WorkspaceBrief.Files[0].Path != "app.go" {
		t.Fatalf("workspace_brief = %+v, want only app.go", body.WorkspaceBrief)
	}
}
